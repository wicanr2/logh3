"""來源封裝邊界：阻止 archive 路徑、身份或類型錯誤破壞既有輸出。"""
import io
from pathlib import Path
import tarfile
import tempfile
import unittest
from unittest.mock import patch

import music_source_kit as kit


class SourceBoundaryTests(unittest.TestCase):
    def archive(self, root, name='go.mod', kind=tarfile.REGTYPE, head='a'*40, duplicate=False):
        p = root/'input.tar'
        with tarfile.open(p, 'w', format=tarfile.PAX_FORMAT, pax_headers={'comment': head}) as archive:
            for _ in range(2 if duplicate else 1):
                entry = tarfile.TarInfo(name)
                entry.type = kind
                if kind == tarfile.REGTYPE:
                    entry.size = 4
                    archive.addfile(entry, io.BytesIO(b'test'))
                else:
                    entry.linkname = '/outside'
                    archive.addfile(entry)
        return p

    def test_git_head_and_forbidden_paths_or_types_rejected(self):
        variants = [('go.mod', tarfile.REGTYPE, 'b'*40),
                    ('../go.mod', tarfile.REGTYPE, 'a'*40),
                    ('org_game/GIN3.COM', tarfile.REGTYPE, 'a'*40),
                    ('go.mod', tarfile.SYMTYPE, 'a'*40),
                    ('go.mod', tarfile.LNKTYPE, 'a'*40)]
        for name, kind, head in variants:
            with self.subTest(name=name, kind=kind, head=head), tempfile.TemporaryDirectory() as tmp:
                p = self.archive(Path(tmp), name, kind, head)
                with self.assertRaises(ValueError):
                    kit.read_git(p, 'a'*40, 'logh3')

    def test_duplicate_member_and_missing_module_file_rejected(self):
        for name, duplicate in [('go.mod', True), ('cmd/probe/main.go', False)]:
            with self.subTest(name=name), tempfile.TemporaryDirectory() as tmp:
                p = self.archive(Path(tmp), name, duplicate=duplicate)
                with self.assertRaises(ValueError):
                    kit.read_git(p, 'a'*40, 'logh3')

    def test_existing_output_is_preserved_before_loading_inputs(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            destination = root/'source-kit'
            destination.mkdir()
            marker = destination/'BUNDLE.json'
            marker.write_bytes(b'existing receipt')
            with patch.object(kit, 'ROOT', root), patch.object(kit, 'read_git') as read:
                with self.assertRaises(ValueError):
                    kit.build(destination)
                read.assert_not_called()
            self.assertEqual(marker.read_bytes(), b'existing receipt')

    def test_license_cannot_change_template_clauses(self):
        template = 'clause @YEAR@ repo @REPO@ again @REPO@ end'
        valid = 'clause 2026 repo https://github.com/wicanr2/logh3 again https://github.com/wicanr2/logh3 end'
        self.assertEqual(kit.license_mapping(template, valid)['@YEAR@'], '2026')
        for wrong in [valid.replace('clause', 'new clause'), valid.replace('2026', '2025'), valid+' added']:
            with self.subTest(wrong=wrong), self.assertRaises(ValueError):
                kit.license_mapping(template, wrong)


if __name__ == '__main__':
    unittest.main(verbosity=2)
