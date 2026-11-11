#!/usr/bin/env python3
"""Reproducible OpenCMMC and legacy-compatible Go archive packaging, using only the Python stdlib.

Cross-builds the offline CLI. Does not assert that it has a graphical interface or a signed publisher identity.
"""
import argparse
import hashlib
import io
import json
import os
from pathlib import Path
import subprocess
import tarfile
import zipfile

PLATFORMS = [('windows', 'amd64'), ('windows', 'arm64'), ('linux', 'amd64'), ('linux', 'arm64')]
STAMP = (2026, 11, 2, 0, 0, 0)
ROOT = Path(__file__).resolve().parent.parent


def archive_file(name, blob, mode=0o644):
    return name, blob, mode


def manifest_bytes(items, name, version, goos, goarch):
    return (json.dumps({'schema': 'freecompliance.package.v1',
                        'package': name, 'version': version, 'goos': goos,
                        'goarch': goarch, 'contents': sorted(items)},
                       indent=2, sort_keys=True) + '\n').encode('utf-8')


def write_zip(path, files):
    with zipfile.ZipFile(path, 'w', zipfile.ZIP_DEFLATED, compresslevel=9) as out:
        for name, payload, mode in sorted(files):
            info = zipfile.ZipInfo(name, STAMP)
            info.compress_type = zipfile.ZIP_DEFLATED
            info.external_attr = (0o100000 | mode) << 16
            out.writestr(info, payload, compress_type=zipfile.ZIP_DEFLATED, compresslevel=9)


def write_tar(path, files):
    import gzip
    with path.open('wb') as fp:
        with gzip.GzipFile(filename='', fileobj=fp, mode='wb', mtime=0, compresslevel=9) as gz:
            with tarfile.open(fileobj=gz, mode='w', format=tarfile.PAX_FORMAT) as out:
                for name, payload, mode in sorted(files):
                    info = tarfile.TarInfo(name)
                    info.size = len(payload)
                    info.mode = mode
                    info.uid = info.gid = 0
                    info.uname = info.gname = ''
                    info.mtime = 0
                    out.addfile(info, io.BytesIO(payload))


def build(version, out_dir):
    if not version.startswith('v') or len(version) > 50 or any(ch not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-' for ch in version):
        raise SystemExit('invalid release version')
    out_dir.mkdir(parents=True, exist_ok=True)
    outputs = []
    for goos, goarch in PLATFORMS:
        # Historic v0 package identifiers and filenames remain stable.
        prefix = 'freecompliance' if version.startswith('v0.') else 'opencmmc'
        name = f'{prefix}-{version}-{goos}-{goarch}'
        exe_name = f'{prefix}.exe' if goos == 'windows' else prefix
        stage = out_dir / f'.{name}.binary'
        env = dict(os.environ, CGO_ENABLED='0', GOOS=goos, GOARCH=goarch, GOTOOLCHAIN='local')
        cmd = ['go', 'build', '-trimpath', '-buildvcs=false',
               '-ldflags', f'-s -w -buildid= -X main.releaseVersion={version}',
               '-o', str(stage), './cmd/freecompliance']
        subprocess.run(cmd, cwd=ROOT, env=env, check=True)
        binary = stage.read_bytes()
        files = [archive_file(exe_name, binary, 0o755),
                 archive_file('LICENSE', (ROOT / 'LICENSE').read_bytes()),
                 archive_file('README.md', (ROOT / 'README.md').read_bytes()),
                 archive_file('CMMC_AIRGAP.md', (ROOT / 'docs/CMMC_AIRGAP.md').read_bytes()),
                 archive_file('CMMC_OBJECTIVE_WORKBENCH.md', (ROOT / 'docs/CMMC_OBJECTIVE_WORKBENCH.md').read_bytes()),
                 archive_file('CMMC_ASSESSMENT_NEXT.md', (ROOT / 'docs/CMMC_ASSESSMENT_NEXT.md').read_bytes()),
                 archive_file('LOCAL_UI.md', (ROOT / 'docs/LOCAL_UI.md').read_bytes()),
                 archive_file('INSTALL-OFFLINE.txt', (ROOT / 'docs/DOWNLOADS.md').read_bytes())]
        # A byte-identical legacy CLI alias supports existing local scripts.
        if prefix == 'opencmmc':
            alias = 'freecompliance.exe' if goos == 'windows' else 'freecompliance'
            files.append(archive_file(alias, binary, 0o755))
        if goos == 'windows':
            files.append(archive_file('START-CMMC-DEMO.cmd', (ROOT / 'scripts/START-CMMC-DEMO.cmd').read_bytes()))
            files.append(archive_file('START-CMMC-UI.cmd', (ROOT / 'scripts/START-CMMC-UI.cmd').read_bytes()))
        else:
            files.append(archive_file('start-cmmc-demo.sh', (ROOT / 'scripts/start-cmmc-demo.sh').read_bytes(), 0o755))
            files.append(archive_file('start-cmmc-ui.sh', (ROOT / 'scripts/start-cmmc-ui.sh').read_bytes(), 0o755))
        files.append(archive_file('PACKAGE-MANIFEST.json', manifest_bytes([name for name, _, _ in files], name, version, goos, goarch)))
        output = out_dir / f'{name}'
        if goos == 'windows':
            output = Path(str(output) + '.zip')
            write_zip(output, files)
        else:
            output = Path(str(output) + '.tar.gz')
            write_tar(output, files)
        outputs.append(output)
        stage.unlink()
    # SHA256SUMS is an integrity record only, not an authentication signature.
    checksums = ''.join(f'{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.name}\n' for p in sorted(outputs))
    (out_dir / 'SHA256SUMS').write_text(checksums, encoding='utf-8')
    (out_dir / 'SHA256SUMS.sha256').write_text(hashlib.sha256(checksums.encode()).hexdigest() + '  SHA256SUMS\n', encoding='utf-8')
    print(checksums, end='')


if __name__ == '__main__':
    ap = argparse.ArgumentParser()
    ap.add_argument('--version', required=True)
    ap.add_argument('--out', type=Path, default=Path('dist'))
    args = ap.parse_args()
    build(args.version, args.out)
