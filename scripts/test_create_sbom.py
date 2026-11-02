#!/usr/bin/env python3
import unittest
from create_sbom import make_document, modules_from_json

class TestSBOM(unittest.TestCase):
    def test_modules_stream(self):
        data='{"Path":"github.com/blocksifrdev/freecompliance","Main":true}\n{"Path":"example.com/lib","Version":"v1.0.0"}'
        self.assertEqual(len(modules_from_json(data)),2)
    def test_inventory_and_relationships(self):
        modules=[{'Path':'github.com/blocksifrdev/freecompliance','Main':True},
                 {'Path':'example.com/lib','Version':'v1.0.0'}]
        doc=make_document(modules,'v0.5.0-preview.4','a'*40,'go1.23','2026-11-02T00:00:00Z')
        self.assertEqual(doc['spdxVersion'],'SPDX-2.3')
        self.assertEqual(len(doc['packages']),3)
        self.assertEqual(len({p['SPDXID'] for p in doc['packages']}),3)
        self.assertEqual(doc['relationships'][0]['relationshipType'],'DESCRIBES')
    def test_missing_main_or_sha_fails(self):
        with self.assertRaises(ValueError):
            modules_from_json('')
        with self.assertRaises(ValueError):
            make_document([{'Path':'example.com/lib'}],'v0.5.0-preview.4','a'*40,'go','2026-11-02T00:00:00Z')
        with self.assertRaises(ValueError):
            make_document([{'Path':'x','Main':True}],'v0.5.0-preview.4','invalid','go','2026-11-02T00:00:00Z')

if __name__ == '__main__':
    unittest.main()
