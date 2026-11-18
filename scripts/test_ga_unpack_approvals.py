"""Negative fixtures: no real review materials and no private keys."""
import base64
from contextlib import redirect_stderr, redirect_stdout
import io
import os
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import zipfile

from ga_unpack_approvals import RECORD_FILES, KEY_FILES, expand, run


def archive(names, content=b"SYNTHETIC"):
    blob = io.BytesIO()
    with zipfile.ZipFile(blob, "w") as zipout:
        for name in names:
            zipout.writestr(name, content)
    return base64.b64encode(blob.getvalue()).decode("ascii")


class ProtectedApprovalBundleTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)

    def invoke(self, records=None, keys=None):
        env = {
            "GA_SIGNOFF_RECORDS_B64": records or "",
            "GA_TRUSTED_PUBLIC_KEYS_B64": keys or "",
        }
        with patch.dict(os.environ, env), redirect_stdout(io.StringIO()), redirect_stderr(io.StringIO()):
            return run(["--destination", str(self.root / "approval")])

    def test_full_exact_bundle_unpacks_with_private_files(self):
        self.assertEqual(0, self.invoke(archive(sorted(RECORD_FILES)), archive(sorted(KEY_FILES))))
        for folder, names in (("records", RECORD_FILES), ("trusted", KEY_FILES)):
            for filename in names:
                path = self.root / "approval" / folder / filename
                self.assertTrue(path.is_file())
                self.assertEqual("SYNTHETIC", path.read_text())
                if os.name != "nt":
                    self.assertEqual(0o600, path.stat().st_mode & 0o777)

    def test_missing_signed_statement_fails(self):
        self.assertEqual(1, self.invoke(archive(sorted(RECORD_FILES)[:-1]), archive(sorted(KEY_FILES))))

    def test_missing_trusted_public_key_fails(self):
        self.assertEqual(1, self.invoke(archive(sorted(RECORD_FILES)), archive(sorted(KEY_FILES)[:-1])))

    def test_traversal_member_fails_even_if_other_items_present(self):
        malicious = archive(sorted(RECORD_FILES | {"../outside"}))
        self.assertEqual(1, self.invoke(malicious, archive(sorted(KEY_FILES))))
        self.assertFalse((self.root / "outside").exists())

    def test_duplicate_member_fails(self):
        self.assertEqual(1, self.invoke(archive(sorted(RECORD_FILES) + [sorted(RECORD_FILES)[0]]),
                                         archive(sorted(KEY_FILES))))

    def test_missing_both_secrets_fails(self):
        self.assertEqual(1, self.invoke())

    def test_invalid_base64_fails(self):
        self.assertEqual(1, self.invoke("not-base64%%", archive(sorted(KEY_FILES))))

    def test_destination_collision_fails(self):
        destination = self.root / "approval"
        destination.mkdir()
        self.assertEqual(1, self.invoke(archive(sorted(RECORD_FILES)), archive(sorted(KEY_FILES))))


if __name__ == "__main__":
    unittest.main()
