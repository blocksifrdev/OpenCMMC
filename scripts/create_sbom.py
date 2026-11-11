#!/usr/bin/env python3
"""SPDX 2.3 Go module inventory. Not a vulnerability scan or complete OS SBOM."""
import argparse
import datetime
import json
import os
from pathlib import Path
import re
import subprocess

def modules_from_json(raw):
    decoder = json.JSONDecoder()
    pos, modules = 0, []
    while pos < len(raw):
        while pos < len(raw) and raw[pos].isspace():
            pos += 1
        if pos >= len(raw):
            break
        entry, end = decoder.raw_decode(raw, pos)
        if not isinstance(entry, dict) or not entry.get('Path'):
            raise ValueError('invalid Go module metadata')
        modules.append(entry)
        pos = end
    if not modules:
        raise ValueError('empty Go module inventory')
    return modules

def make_document(modules, version, source_sha, toolchain, created):
    if not re.fullmatch(r'v[0-9A-Za-z.-]{1,49}', version):
        raise ValueError('invalid version')
    if not re.fullmatch(r'[0-9a-f]{40}', source_sha):
        raise ValueError('source commit SHA required')
    identifiers, packages = {}, []
    for index, module in enumerate(modules):
        spdx_id = f'SPDXRef-GoModule-{index}'
        path = module['Path']
        package = {
            'SPDXID': spdx_id, 'name': path,
            'versionInfo': version if module.get('Main') else module.get('Version', 'UNKNOWN'),
            'downloadLocation': 'NOASSERTION', 'filesAnalyzed': False,
            'licenseConcluded': 'NOASSERTION',
            'licenseDeclared': 'Apache-2.0' if module.get('Main') else 'NOASSERTION',
            'copyrightText': 'NOASSERTION', 'supplier': 'NOASSERTION',
        }
        if module.get('Replace'):
            package['comment'] = 'Module replacement present; review metadata before high assurance deployment'
        identifiers[path] = spdx_id
        packages.append(package)
    packages.append({
        'SPDXID': 'SPDXRef-GoToolchain', 'name': 'Go toolchain and standard library',
        'versionInfo': toolchain, 'downloadLocation': 'NOASSERTION',
        'filesAnalyzed': False, 'licenseConcluded': 'NOASSERTION',
        'licenseDeclared': 'NOASSERTION', 'copyrightText': 'NOASSERTION',
    })
    main = next((mod for mod in modules if mod.get('Main')), None)
    if main is None:
        raise ValueError('Go main module missing')
    root = identifiers[main['Path']]
    relations = [
        {'spdxElementId': 'SPDXRef-DOCUMENT', 'relationshipType': 'DESCRIBES', 'relatedSpdxElement': root},
        {'spdxElementId': root, 'relationshipType': 'DEPENDS_ON', 'relatedSpdxElement': 'SPDXRef-GoToolchain'},
    ]
    for mod in modules:
        if not mod.get('Main'):
            relations.append({'spdxElementId': root, 'relationshipType': 'DEPENDS_ON',
                              'relatedSpdxElement': identifiers[mod['Path']]})
    return {
        'spdxVersion': 'SPDX-2.3', 'dataLicense': 'CC0-1.0', 'SPDXID': 'SPDXRef-DOCUMENT',
        'name': f'OpenCMMC-{version}-Go-module-inventory',
        'documentNamespace': f'https://github.com/blocksifrdev/OpenCMMC/spdx/{version}/{source_sha}',
        'creationInfo': {'creators': ['Tool: opencmmc/scripts/create_sbom.py'], 'created': created},
        'documentComment': 'Go module/toolchain inventory for pinned source checkout. Not independently validated complete software or OS inventory, vulnerability scan, or certification.',
        'packages': packages, 'relationships': relations,
    }

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--version', required=True)
    parser.add_argument('--source-sha', default=os.environ.get('GITHUB_SHA', ''))
    parser.add_argument('--out', required=True, type=Path)
    args = parser.parse_args()
    modules = modules_from_json(subprocess.check_output(['go','list','-m','-json','all'], text=True))
    toolchain = subprocess.check_output(['go','version'], text=True).strip()
    created = datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
    doc = make_document(modules, args.version, args.source_sha, toolchain, created)
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(json.dumps(doc, indent=2, sort_keys=True) + '\n', encoding='utf-8')
    print(f'SPDX module inventory: {len(doc["packages"])} packages')

if __name__ == '__main__':
    main()
