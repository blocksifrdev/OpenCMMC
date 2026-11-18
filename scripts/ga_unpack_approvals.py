#!/usr/bin/env python3
"""Safely expand two independently supplied, private GA verification bundles.

The protected GitHub environment must inject:
  GA_SIGNOFF_RECORDS_B64  - ZIP of six ROLE.json and ROLE.sig files.
  GA_TRUSTED_PUBLIC_KEYS_B64 - ZIP of six independently provisioned ROLE.pub.pem files.
Do not put real review reports, private keys, CUI or secrets into source control.
"""
import argparse
import base64
import binascii
import io
import os
from pathlib import Path
import stat
import sys
import zipfile

from verify_ga_signoffs import ROLES

RECORD_FILES = {f"{role}.{suffix}" for role in ROLES for suffix in ("json", "sig")}
KEY_FILES = {f"{role}.pub.pem" for role in ROLES}


def expand(secret, directory: Path, expected: set[str]) -> None:
    if not secret:
        raise ValueError("missing mandatory protected-environment secret")
    try:
        archive = base64.b64decode(secret, validate=True)
    except (binascii.Error, ValueError) as e:
        raise ValueError("invalid protected-environment bundle encoding") from e
    if len(archive) > 128 * 1024:
        raise ValueError("approval bundle exceeds compressed limit")
    with zipfile.ZipFile(io.BytesIO(archive)) as zipped:
        entries = zipped.infolist()
        names = [entry.filename for entry in entries]
        if len(names) != len(expected) or set(names) != expected or len(names) != len(set(names)):
            raise ValueError("approval bundle must contain exactly the required role files")
        if sum(x.file_size for x in entries) > 160 * 1024:
            raise ValueError("approval files exceed uncompressed limit")
        directory.mkdir(mode=0o700, parents=False, exist_ok=False)
        for member in entries:
            mode = (member.external_attr >> 16) & 0xFFFF
            if (mode & 0o170000) == stat.S_IFLNK:
                raise ValueError("links are not accepted in approval bundles")
            if member.is_dir() or member.file_size > 16 * 1024:
                raise ValueError("invalid approval bundle file")
            # The exact whitelist prevents traversal, nested paths, and aliases.
            path = directory / member.filename
            payload = zipped.read(member)
            if len(payload) != member.file_size:
                raise ValueError("approval bundle size mismatch")
            descriptor = os.open(path, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
            with os.fdopen(descriptor, "wb") as out:
                out.write(payload)


def run(argv=None):
    parser = argparse.ArgumentParser()
    parser.add_argument("--destination", type=Path, required=True)
    args = parser.parse_args(argv)
    try:
        root = args.destination
        if root.exists() or root.is_symlink():
            raise ValueError("approval destination must not exist")
        os.umask(0o077)
        root.mkdir(mode=0o700)
        expand(os.environ.get("GA_SIGNOFF_RECORDS_B64", ""), root / "records", RECORD_FILES)
        expand(os.environ.get("GA_TRUSTED_PUBLIC_KEYS_B64", ""), root / "trusted", KEY_FILES)
        print("PASS: protected GA statement/key files unpacked independently (content withheld)")
        return 0
    except (OSError, ValueError, zipfile.BadZipFile, RuntimeError) as exc:
        print(f"GA BLOCKED: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(run())
