"""規格 076：同名 DLL、未審依賴與工具路徑的合併拒絕測試。"""
import copy
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import music_windows_bundle as bundle


class BundleTests(unittest.TestCase):
    def setUp(self):
        self.rows = {n: {'sha256': 'a'*64, 'bytes': 10, 'imports': ['KERNEL32.dll']}
                     for n in ['fluidsynth.exe', 'ffmpeg.exe', 'opl2wav.exe', 'dosgolem-probe.exe']}

    def test_identical_overlap_keeps_both_origins(self):
        result, overlaps = bundle.merge_records([('gm', self.rows), ('fm', {'ffmpeg.exe': self.rows['ffmpeg.exe']})])
        self.assertEqual(overlaps, ['ffmpeg.exe'])
        self.assertEqual(result['ffmpeg.exe']['groups'], ['gm', 'fm'])

    def test_case_collision_rejects_different_content_or_imports(self):
        for field, value in [('sha256', 'b'*64), ('bytes', 11), ('imports', ['USER32.dll'])]:
            row = dict(self.rows['ffmpeg.exe'], **{field: value})
            with self.subTest(field=field), self.assertRaises(ValueError):
                bundle.merge_records([('gm', self.rows), ('fm', {'FFMPEG.exe': row})])

    def test_missing_tool_unknown_import_and_path_rejected(self):
        missing = copy.deepcopy(self.rows)
        del missing['dosgolem-probe.exe']
        unknown = copy.deepcopy(self.rows)
        unknown['opl2wav.exe']['imports'].append('some-system-fallback.dll')
        traversal = dict(self.rows, **{'../ffmpeg.exe': self.rows['ffmpeg.exe']})
        for rows in [missing, unknown, traversal]:
            with self.subTest(rows=rows), self.assertRaises(ValueError):
                bundle.merge_records([('probe', rows)])

    def test_nonregular_or_link_input_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            normal = root/'normal'
            normal.write_bytes(b'valid')
            link = root/'link'
            link.symlink_to(normal)
            self.assertEqual(len(bundle.sha(normal)), 64)
            for p in [link, root, root/'missing']:
                with self.subTest(path=p), self.assertRaises(ValueError):
                    bundle.sha(p)

    def test_output_parent_link_rejected_before_loading_inputs(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            other = root/'other'
            other.mkdir()
            (root/'combined-stage').symlink_to(other, target_is_directory=True)
            with patch.object(bundle, 'ROOT', root), patch.object(bundle, 'sha') as digest:
                with self.assertRaises(ValueError):
                    bundle.build()
                digest.assert_not_called()
            self.assertEqual(list(other.iterdir()), [])

    def test_existing_failed_package_preserved(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            package = root/'combined-stage/封裝 測試'
            package.mkdir(parents=True)
            marker = package/'BUNDLE.json'
            marker.write_bytes(b'failed receipt')
            with patch.object(bundle, 'ROOT', root), self.assertRaises(ValueError):
                bundle.build()
            self.assertEqual(marker.read_bytes(), b'failed receipt')


if __name__ == '__main__':
    unittest.main()
