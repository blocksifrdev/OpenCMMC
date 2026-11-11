#!/usr/bin/env python3
"""Fail closed on retired branding, unexpected distributed binaries or altered archive checksums.

Works without third-party packages or network. Source checks use only tracked files.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import re
import subprocess
import sys
import tarfile
import zipfile

ROOT = Path(__file__).resolve().parent.parent
RETIRED = "free" + "compliance"
PLATFORMS = (("windows", "arm64"), ("windows", "amd64"),
             ("linux", "arm64"), ("linux", "amd64"))


def scan_retired_text(path: str, content: bytes) -> list[str]:
    issues = []
    if RETIRED in path.casefold():
        issues.append(f"retired product name in filename: {path}")
    # Treat unknown binary content as opaque, except its tracked filename.
    if b"\0" in content:
        return issues
    text = content.decode("utf-8", errors="replace")
    for line_num, line in enumerate(text.splitlines(), start=1):
        if RETIRED in line.casefold():
            issues.append(f"retired product name: {path}:{line_num}")
    return issues


def verify_source(root: Path) -> list[str]:
    paths = subprocess.check_output(["git", "ls-files", "-z"], cwd=root).split(b"\0")
    issues = []
    for raw in paths:
        if not raw:
            continue
        name = os.fsdecode(raw)
        path = root / name
        if not path.is_file():
            issues.append(f"missing tracked file: {name}")
            continue
        issues.extend(scan_retired_text(name, path.read_bytes()))
    return issues


def verify_archive(archive: Path, version: str, system: str, arch: str) -> list[str]:
    prefix = f"opencmmc-{version}-{system}-{arch}"
    extension = ".zip" if system == "windows" else ".tar.gz"
    expected_name = prefix + extension
    if archive.name != expected_name:
        return [f"unexpected archive filename: {archive.name}"]

    if system == "windows":
        with zipfile.ZipFile(archive) as pack:
            members = [x for x in pack.namelist() if not x.endswith("/")]
            data = {x: pack.read(x) for x in members}
    else:
        with tarfile.open(archive, "r:gz") as pack:
            members = [x.name for x in pack.getmembers() if x.isfile()]
            data = {x: pack.extractfile(x).read() for x in members}

    issues = []
    if len(members) != len(set(members)):
        issues.append(f"duplicate archive paths in {archive.name}")
    for name in members:
        if PurePosixPath(name).is_absolute() or ".." in PurePosixPath(name).parts:
            issues.append(f"unsafe archive entry: {name}")
        if RETIRED in name.casefold():
            issues.append(f"retired product binary/file in archive: {name}")
    executable = "opencmmc.exe" if system == "windows" else "opencmmc"
    if executable not in data or not data[executable]:
        issues.append(f"missing OpenCMMC executable: {archive.name}")
    if system == "windows" and any(n.endswith(".exe") and n != executable for n in members):
        issues.append(f"unexpected second Windows executable in {archive.name}")
    if "PACKAGE-MANIFEST.json" not in data:
        issues.append(f"missing package manifest in {archive.name}")
    else:
        try:
            manifest = json.loads(data["PACKAGE-MANIFEST.json"])
            if manifest.get("schema") != "opencmmc.package.v1":
                issues.append(f"incorrect package schema in {archive.name}")
            if manifest.get("package") != prefix or manifest.get("version") != version:
                issues.append(f"incorrect package name/version in {archive.name}")
            if manifest.get("goos") != system or manifest.get("goarch") != arch:
                issues.append(f"incorrect package target in {archive.name}")
            if manifest.get("contents") != sorted(n for n in members if n != "PACKAGE-MANIFEST.json"):
                issues.append(f"incorrect archive contents manifest in {archive.name}")
        except (ValueError, TypeError) as exc:
            issues.append(f"invalid package manifest in {archive.name}: {exc}")
    return issues


def verify_artifacts(folder: Path, version: str) -> list[str]:
    issues = []
    manifest = folder / "SHA256SUMS"
    if not manifest.is_file():
        return ["SHA256SUMS is missing"]
    hashes = {}
    for line in manifest.read_text(encoding="utf-8").splitlines():
        parts = line.split()
        if len(parts) != 2 or not re.fullmatch(r"[0-9a-f]{64}", parts[0]):
            issues.append("invalid SHA256SUMS line")
            continue
        digest, name = parts
        if name in hashes:
            issues.append(f"duplicate SHA256SUMS entry: {name}")
        hashes[name] = digest
    expected = set()
    for system, arch in PLATFORMS:
        name = f"opencmmc-{version}-{system}-{arch}" + (
            ".zip" if system == "windows" else ".tar.gz")
        expected.add(name)
        archive = folder / name
        if not archive.is_file():
            issues.append(f"missing archive: {name}")
            continue
        digest = hashlib.sha256(archive.read_bytes()).hexdigest()
        if hashes.get(name) != digest:
            issues.append(f"SHA256 mismatch: {name}")
        try:
            issues.extend(verify_archive(archive, version, system, arch))
        except (OSError, ValueError, zipfile.BadZipFile, tarfile.TarError) as exc:
            issues.append(f"archive unreadable: {name}: {exc}")
    if not expected.issubset(hashes):
        issues.append("checksum manifest is missing one or more platforms")
    return issues


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", action="store_true", help="check current tracked repo")
    parser.add_argument("--artifacts", type=Path, help="directory of built archives")
    parser.add_argument("--version", help="release version for --artifacts")
    args = parser.parse_args()
    if not args.source and not args.artifacts:
        parser.error("choose --source and/or --artifacts")
    if args.artifacts and not args.version:
        parser.error("--artifacts requires --version")
    issues = []
    if args.source:
        issues.extend(verify_source(ROOT))
    if args.artifacts:
        issues.extend(verify_artifacts(args.artifacts, args.version))
    if issues:
        for issue in issues:
            print("FAIL:", issue, file=sys.stderr)
        return 1
    print("OpenCMMC source and distribution guard: PASS")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
