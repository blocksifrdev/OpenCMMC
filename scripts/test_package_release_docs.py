"""Packaged offline documentation must match the *installed binary*, not a prior RC."""
from pathlib import Path
import tempfile
import unittest

from package_release import ROOT, render_package_doc


class OfflinePackageDocumentationTests(unittest.TestCase):
    def test_both_help_files_render_exact_ga_version(self):
        for path in ("PACKAGE_README.md", "PACKAGE_INSTALL.md"):
            payload = render_package_doc(ROOT / "docs" / path, "v1.0.0")
            self.assertIn(b"OpenCMMC v1.0.0", payload)
            self.assertNotIn(b"@OPENCMMC_VERSION@", payload)
            self.assertNotIn(b"v1.0.0-rc.11", payload)
            self.assertNotIn(b"OpenCMMC v0.4.0-dev", payload)

    def test_other_versions_render_from_same_source_without_rc_leaks(self):
        for version in ("v1.0.1", "v1.1.0-rc.1"):
            for path in ("PACKAGE_README.md", "PACKAGE_INSTALL.md"):
                content = render_package_doc(ROOT / "docs" / path, version)
                self.assertIn(("OpenCMMC " + version).encode(), content)
                self.assertNotIn(b"v1.0.0-rc.11", content)

    def test_template_without_version_refused(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "broken.md"
            path.write_text("Old release instructions with no templated version.")
            with self.assertRaises(ValueError):
                render_package_doc(path, "v1.0.0")

    def test_offline_help_discloses_encryption_and_compliance_boundaries(self):
        helpdata = render_package_doc(ROOT / "docs/PACKAGE_INSTALL.md", "v1.0.0").lower()
        for phrase in (b"not encrypted", b"cui", b"not a certified cmmc", b"private",
                       b"sha256", b"restore", b"origin", b"csrf"):
            self.assertIn(phrase, helpdata)


if __name__ == "__main__":
    unittest.main()
