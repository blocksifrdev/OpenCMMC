#!/usr/bin/env python3
"""Fail-closed v1 GA approval verifier; not a substitute for independent reviews.

Uses only stdlib and openssl. Trust anchors MUST be independently provisioned,
never copied from the source checkout or alongside reviewer-supplied statements.
"""
import argparse
from datetime import datetime, timezone, timedelta
import hashlib
import json
from pathlib import Path
import re
import stat
import subprocess
import sys

SCHEMA = "opencmmc.ga.signoff.v1"
ROLES = ("independent-security", "cmmc-methodology", "chromebook-acceptance",
         "airgap-acceptance", "accessibility-review", "release-owner")
FIELDS = {"schema", "role", "release_version", "source_sha",
          "archives_sha256", "report_sha256", "reviewer", "reviewed_at",
          "decision"}
HEX40 = re.compile(r"[0-9a-f]{40}\Z")
HEX64 = re.compile(r"[0-9a-f]{64}\Z")


class GateFailure(ValueError):
    pass


def safe_file(path, maxbytes=16384):
    try:
        s = path.lstat()
        if not stat.S_ISREG(s.st_mode) or s.st_size > maxbytes:
            raise GateFailure(f"unsafe or oversized gate file: {path.name}")
        data = path.read_bytes()
        if len(data) > maxbytes:
            raise GateFailure(f"oversized gate file: {path.name}")
        return data
    except OSError as e:
        raise GateFailure(f"unavailable gate file: {path.name}") from e


def unique_json(pairs):
    result = {}
    for key, val in pairs:
        if key in result:
            raise GateFailure("duplicate signed JSON field")
        result[key] = val
    return result


def archive_fingerprint(folder, version):
    """Digest only the *four actual archive hashes*, not volatile signing data."""
    if version != "v1.0.0":
        raise GateFailure("GA verifier only supports exact v1.0.0")
    expected = set()
    for system in ("linux", "windows"):
        for arch in ("amd64", "arm64"):
            ext = ".zip" if system == "windows" else ".tar.gz"
            expected.add(f"opencmmc-{version}-{system}-{arch}{ext}")
    lines = safe_file(folder / "SHA256SUMS", 4096).decode("ascii").splitlines()
    sums = {}
    for line in lines:
        match = re.fullmatch(r"([0-9a-f]{64})  ([A-Za-z0-9._-]+)", line)
        if not match:
            raise GateFailure("invalid SHA256SUMS format")
        digest, name = match.groups()
        if name in sums:
            raise GateFailure("duplicate digest in SHA256SUMS")
        sums[name] = digest
    if not expected.issubset(sums):
        raise GateFailure("one or more exact GA archive digests are missing")
    normalized = []
    for name in sorted(expected):
        p = folder / name
        s = p.lstat()
        if not stat.S_ISREG(s.st_mode):
            raise GateFailure(f"unsafe archive: {name}")
        sha = hashlib.sha256()
        with p.open("rb") as stream:
            for chunk in iter(lambda: stream.read(1048576), b""):
                sha.update(chunk)
        if sha.hexdigest() != sums[name]:
            raise GateFailure(f"archive digest mismatch: {name}")
        normalized.append(f"{sums[name]}  {name}\n")
    return hashlib.sha256("".join(normalized).encode("ascii")).hexdigest()


def verify_role(role, records, trusted, version, source_sha, fingerprint, now):
    payload = records / f"{role}.json"
    signature = records / f"{role}.sig"
    pubkey = trusted / f"{role}.pub.pem"
    raw = safe_file(payload)
    safe_file(signature, 512)
    safe_file(pubkey, 4096)
    try:
        statement = json.loads(raw.decode("utf-8"), object_pairs_hook=unique_json)
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise GateFailure(f"malformed {role} signed statement") from exc
    if not isinstance(statement, dict) or set(statement) != FIELDS:
        raise GateFailure(f"missing or extra approval fields: {role}")
    must = dict(schema=SCHEMA, role=role, release_version=version,
                source_sha=source_sha, archives_sha256=fingerprint,
                decision="APPROVE")
    for key, val in must.items():
        if statement[key] != val:
            raise GateFailure(f"{role} does not approve exact {key}")
    if not isinstance(statement["report_sha256"], str) or not HEX64.fullmatch(statement["report_sha256"]):
        raise GateFailure(f"report digest absent: {role}")
    who = statement["reviewer"]
    if not isinstance(who, str) or not 5 <= len(who.strip()) <= 160:
        raise GateFailure(f"reviewer identity absent: {role}")
    try:
        if not isinstance(statement["reviewed_at"], str):
            raise ValueError("invalid datetime")
        reviewed = datetime.fromisoformat(statement["reviewed_at"].replace("Z", "+00:00"))
        if reviewed.tzinfo is None:
            raise ValueError("missing timezone")
        reviewed = reviewed.astimezone(timezone.utc)
    except ValueError as exc:
        raise GateFailure(f"review timestamp invalid: {role}") from exc
    if reviewed > now + timedelta(minutes=5) or now - reviewed > timedelta(days=90):
        raise GateFailure(f"review timestamp stale or future: {role}")
    try:
        result = subprocess.run(
            ["openssl", "pkeyutl", "-verify", "-pubin", "-inkey", str(pubkey),
             "-rawin", "-in", str(payload), "-sigfile", str(signature)],
            stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=10, check=False)
    except (OSError, subprocess.TimeoutExpired) as exc:
        raise GateFailure("Ed25519 verifier unavailable") from exc
    if result.returncode != 0:
        raise GateFailure(f"invalid trusted Ed25519 signature: {role}")


def run(argv=None):
    parser = argparse.ArgumentParser(description="Independently signed, exact-artifact OpenCMMC GA gate")
    parser.add_argument("--version", required=True)
    parser.add_argument("--source-sha", required=True)
    parser.add_argument("--dist", type=Path, required=True)
    parser.add_argument("--records-dir", type=Path)
    parser.add_argument("--trust-dir", type=Path)
    parser.add_argument("--print-archive-fingerprint", action="store_true")
    opts = parser.parse_args(argv)
    try:
        if not HEX40.fullmatch(opts.source_sha):
            raise GateFailure("exact lowercase 40-character release commit SHA required")
        fingerprint = archive_fingerprint(opts.dist, opts.version)
        if opts.print_archive_fingerprint:
            print("GA_ARCHIVES_SHA256=" + fingerprint)
            return 0
        if opts.records_dir is None or opts.trust_dir is None:
            raise GateFailure("signed reviews and independently provisioned public keys are required")
        for role in ROLES:
            verify_role(role, opts.records_dir, opts.trust_dir, opts.version,
                        opts.source_sha, fingerprint, datetime.now(timezone.utc))
        print("PASS: all six signed reviews match the exact GA source and four archive digests")
        return 0
    except (GateFailure, OSError, UnicodeError) as exc:
        print(f"GA BLOCKED: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(run())
