"""規格 073 邊界驗證；不執行遊戲、來源 archive 或網路。"""
import copy
import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock
import music_linux_sources as src


class SourceClosureTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.lock = json.loads(src.LOCK.read_text())
        cls.row = cls.lock['sources'][0]
        cls.data = (src.ROOT/'source-index'/cls.row['description']['filename']).read_bytes()

    def parse(self, data):
        return src.parse_dsc(data, self.row['source'], self.row['version'])

    def test_all_real_descriptions_cover_exact_source_parts(self):
        for row in self.lock['sources']:
            data = (src.ROOT/'source-index'/row['description']['filename']).read_bytes()
            parts = src.parse_dsc(data, row['source'], row['version'])
            self.assertEqual(parts, {n: {k: r[k] for k in ('sha256', 'bytes')} for n, r in row['files'].items()})
        self.assertEqual(len(src.files_for(self.lock)), 513)

    def test_wrong_identity_and_duplicate_fields_refused(self):
        for old, new in [(f'Source: {self.row["source"]}', 'Source: wrong'),
                         (f'Version: {self.row["version"]}', 'Version: 0'),
                         ('Source:', 'source:')]:
            data = self.data.replace(old.encode(), new.encode(), 1)
            if new == 'source:':
                self.assertEqual(self.parse(data), self.parse(self.data))
            else:
                with self.assertRaises(ValueError):
                    self.parse(data)
        marker = f'Source: {self.row["source"]}'.encode()
        with self.assertRaises(ValueError):
            self.parse(self.data.replace(marker, marker+b'\nSOURCE: wrong', 1))

    def test_unsafe_filename_and_missing_hash_refused(self):
        name = next(iter(self.row['files']))
        with self.assertRaises(ValueError):
            self.parse(self.data.replace((' '+name+'\n').encode(), (' ../'+name+'\n').encode()))
        with self.assertRaises(ValueError):
            self.parse(self.data.replace(b'Checksums-Sha256:', b'Checksums-Other:'))
        for url in ['http://deb.debian.org/debian/pool/main/a/a/a.dsc',
                    'https://example.org/a.dsc', 'https://deb.debian.org/etc/passwd']:
            with self.assertRaises(ValueError):
                src.safe_url(url)

    def test_source_mapping_missing_or_name_collision_refused(self):
        lock = copy.deepcopy(self.lock)
        lock['sources'].pop()
        with self.assertRaises(ValueError):
            src.files_for(lock)
        lock = copy.deepcopy(self.lock)
        first, second = lock['sources'][:2]
        n, r = next(iter(first['files'].items()))
        second['files'][n] = dict(r, sha256='0'*64)
        with self.assertRaises(ValueError):
            src.files_for(lock)

    def test_missing_and_corrupt_cache_never_access_network(self):
        row = self.row['description']
        with tempfile.TemporaryDirectory() as d, mock.patch.object(src.urllib.request, 'urlopen', side_effect=AssertionError('不得連網')) as net:
            cache = Path(d)
            with self.assertRaisesRegex(ValueError, '離線快取缺檔'):
                src.obtain(cache, row, False)
            p = cache/row['filename']
            p.write_bytes(self.data)
            src.obtain(cache, row, False)
            p.write_bytes(bytes([self.data[0]^1])+self.data[1:])
            with self.assertRaisesRegex(ValueError, 'SHA-256'):
                src.obtain(cache, row, True)
            p.unlink()
            p.symlink_to(src.ROOT/'source-index'/row['filename'])
            with self.assertRaisesRegex(ValueError, '連結'):
                src.obtain(cache, row, False)
            net.assert_not_called()


if __name__ == '__main__':
    unittest.main(verbosity=2)
