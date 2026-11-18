"""Tests use throwaway synthetic Ed25519 identities, never real signoffs."""
from contextlib import redirect_stderr, redirect_stdout
from datetime import datetime, timedelta, timezone
import hashlib
import io
import json
from pathlib import Path
import subprocess
import tempfile
import unittest

import verify_ga_signoffs as gate

SHA = "a" * 40
VERSION = "v1.0.0"


class GAAttestationTests(unittest.TestCase):
    def setUp(self):
        temp = tempfile.TemporaryDirectory()
        self.addCleanup(temp.cleanup)
        self.root = Path(temp.name)
        self.dist, self.records, self.keys = (self.root / n for n in ("dist", "records", "trusted"))
        for directory in (self.dist, self.records, self.keys):
            directory.mkdir()
        lines = []
        for platform, arch in (("linux", "amd64"), ("linux", "arm64"),
                               ("windows", "amd64"), ("windows", "arm64")):
            suffix = ".zip" if platform == "windows" else ".tar.gz"
            name = f"opencmmc-{VERSION}-{platform}-{arch}{suffix}"
            data = f"SYNTHETIC-ONLY:{name}".encode()
            (self.dist / name).write_bytes(data)
            lines.append(f"{hashlib.sha256(data).hexdigest()}  {name}\n")
        (self.dist / "SHA256SUMS").write_text("".join(lines), encoding="ascii")
        self.fingerprint = gate.archive_fingerprint(self.dist, VERSION)
        self.now = datetime.now(timezone.utc).replace(microsecond=0)
        for role in gate.ROLES:
            private = self.root / f"{role}.private"
            subprocess.run(["openssl", "genpkey", "-algorithm", "Ed25519", "-out", str(private)],
                           check=True, capture_output=True)
            with (self.keys / f"{role}.pub.pem").open("wb") as output:
                subprocess.run(["openssl", "pkey", "-in", str(private), "-pubout"],
                               stdout=output, stderr=subprocess.PIPE, check=True)
            self.sign(role, self.statement(role))

    def statement(self, role):
        return {
            "schema": gate.SCHEMA, "role": role, "release_version": VERSION,
            "source_sha": SHA, "archives_sha256": self.fingerprint,
            "report_sha256": hashlib.sha256(("SYNTHETIC REPORT:"+role).encode()).hexdigest(),
            "reviewer": "Synthetic fixture reviewer for "+role,
            "reviewed_at": self.now.isoformat(), "decision": "APPROVE",
        }

    def sign(self, role, statement):
        path = self.records / f"{role}.json"
        path.write_text(json.dumps(statement, sort_keys=True) + "\n")
        subprocess.run(["openssl", "pkeyutl", "-sign", "-inkey", str(self.root / f"{role}.private"),
                        "-rawin", "-in", str(path), "-out", str(self.records / f"{role}.sig")],
                       capture_output=True, check=True)

    def verify(self, source=SHA):
        with redirect_stdout(io.StringIO()), redirect_stderr(io.StringIO()):
            return gate.run(["--version", VERSION, "--source-sha", source,
                             "--dist", str(self.dist), "--records-dir", str(self.records),
                             "--trust-dir", str(self.keys)])

    def test_six_signatures_accept_exact_source_and_archives(self):
        self.assertEqual(0, self.verify())

    def test_missing_signature_fails(self):
        (self.records / "independent-security.sig").unlink()
        self.assertEqual(1, self.verify())

    def test_wrong_commit_fails(self):
        self.assertEqual(1, self.verify(source="b" * 40))

    def test_archive_tampering_fails_even_with_old_sha_manifest(self):
        path = self.dist / f"opencmmc-{VERSION}-windows-amd64.zip"
        path.write_bytes(path.read_bytes() + b"MODIFIED")
        self.assertEqual(1, self.verify())

    def test_changed_manifest_and_archive_fails_against_signed_fingerprint(self):
        path = self.dist / f"opencmmc-{VERSION}-linux-amd64.tar.gz"
        path.write_bytes(b"modified")
        lines = (self.dist / "SHA256SUMS").read_text().splitlines()
        lines[0] = f"{hashlib.sha256(b'modified').hexdigest()}  {path.name}"
        (self.dist / "SHA256SUMS").write_text("\n".join(lines) + "\n")
        self.assertEqual(1, self.verify())

    def test_tampered_signoff_fails(self):
        path = self.records / "independent-security.json"
        path.write_bytes(path.read_bytes().replace(b"APPROVE", b"DENY"))
        self.assertEqual(1, self.verify())

    def test_wrong_trusted_public_key_fails(self):
        (self.keys / "independent-security.pub.pem").write_bytes(
            (self.keys / "release-owner.pub.pem").read_bytes())
        self.assertEqual(1, self.verify())

    def test_stale_review_fails_even_when_validly_signed(self):
        role = "cmmc-methodology"
        data = self.statement(role)
        data["reviewed_at"] = (self.now - timedelta(days=100)).isoformat()
        self.sign(role, data)
        self.assertEqual(1, self.verify())

    def test_duplicate_json_keys_fail_even_when_signed(self):
        role = "release-owner"
        path = self.records / f"{role}.json"
        raw = path.read_text().rstrip()
        path.write_text(raw[:-1] + ',"decision":"APPROVE"}')
        subprocess.run(["openssl", "pkeyutl", "-sign", "-inkey", str(self.root / f"{role}.private"),
                        "-rawin", "-in", str(path), "-out", str(self.records / f"{role}.sig")],
                       capture_output=True, check=True)
        self.assertEqual(1, self.verify())

    def test_missing_trust_key_fails(self):
        (self.keys / "accessibility-review.pub.pem").unlink()
        self.assertEqual(1, self.verify())

    def test_sbom_append_does_not_change_archive_fingerprint(self):
        sbom = b"synthetic"
        (self.dist / "SBOM.spdx.json").write_bytes(sbom)
        with (self.dist / "SHA256SUMS").open("a") as out:
            out.write(f"{hashlib.sha256(sbom).hexdigest()}  SBOM.spdx.json\n")
        self.assertEqual(self.fingerprint, gate.archive_fingerprint(self.dist, VERSION))


if __name__ == "__main__":
    unittest.main()
