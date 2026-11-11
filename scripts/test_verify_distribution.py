import io
import json
import tempfile
import unittest
from pathlib import Path
from tarfile import TarFile, TarInfo
from zipfile import ZipFile

from verify_distribution import (
    scan_retired_text, verify_archive, verify_artifacts,
)


class BrandGuardTests(unittest.TestCase):
    def test_no_false_positive_for_opencmmc(self):
        self.assertEqual([], scan_retired_text("README.md", b"OpenCMMC\n"))

    def test_retired_brand_is_rejected_even_when_case_differs(self):
        deprecated = "Free" + "Compliance"
        result = scan_retired_text("docs/intro.md", deprecated.encode())
        self.assertEqual(1, len(result))
        self.assertIn("docs/intro.md:1", result[0])

    def test_retired_path_is_rejected(self):
        deprecated_path = "cmd/" + "free" + "compliance" + "/main.go"
        result = scan_retired_text(deprecated_path, b"package main\n")
        self.assertEqual(1, len(result))

    def test_package_manifest_and_checksum(self):
        version = "v1.0.0-rc.9"
        with tempfile.TemporaryDirectory() as temp:
            folder = Path(temp)
            import hashlib
            manifest_lines = []
            for system, arch in (
                ("windows", "arm64"), ("windows", "amd64"),
                ("linux", "arm64"), ("linux", "amd64"),
            ):
                prefix = f"opencmmc-{version}-{system}-{arch}"
                binary = "opencmmc.exe" if system == "windows" else "opencmmc"
                data = {
                    binary: b"synthetic executable",
                    "LICENSE": b"synthetic license",
                }
                data["PACKAGE-MANIFEST.json"] = json.dumps({
                    "schema": "opencmmc.package.v1",
                    "package": prefix,
                    "version": version,
                    "goos": system,
                    "goarch": arch,
                    "contents": sorted(data),
                }).encode()
                if system == "windows":
                    archive = folder / (prefix + ".zip")
                    with ZipFile(archive, "w") as pack:
                        for name, content in data.items():
                            pack.writestr(name, content)
                else:
                    archive = folder / (prefix + ".tar.gz")
                    with TarFile.open(archive, "w:gz") as pack:
                        for name, content in data.items():
                            member = TarInfo(name)
                            member.size = len(content)
                            pack.addfile(member, io.BytesIO(content))
                manifest_lines.append(
                    f"{hashlib.sha256(archive.read_bytes()).hexdigest()}  {archive.name}\n"
                )
            (folder / "SHA256SUMS").write_text("".join(manifest_lines))
            self.assertEqual([], verify_artifacts(folder, version))
            target = folder / f"opencmmc-{version}-windows-arm64.zip"
            target.write_bytes(target.read_bytes() + b"tampered")
            self.assertTrue(any("SHA256 mismatch" in x for x in verify_artifacts(folder, version)))

    def test_unexpected_executable_is_rejected(self):
        version = "v1.0.0-rc.9"
        with tempfile.TemporaryDirectory() as temp:
            archive = Path(temp) / f"opencmmc-{version}-windows-arm64.zip"
            with ZipFile(archive, "w") as pack:
                pack.writestr("opencmmc.exe", b"synthetic")
                pack.writestr("other.exe", b"unexpected")
                pack.writestr("PACKAGE-MANIFEST.json", json.dumps({
                    "schema": "opencmmc.package.v1",
                    "package": f"opencmmc-{version}-windows-arm64",
                    "version": version,
                    "goos": "windows",
                    "goarch": "arm64",
                    "contents": ["opencmmc.exe", "other.exe"],
                }))
            self.assertTrue(any("unexpected second Windows executable" in x for x in
                                verify_archive(archive, version, "windows", "arm64")))


if __name__ == "__main__":
    unittest.main()
