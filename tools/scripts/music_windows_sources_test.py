"""研究 363 固定來源邊界：容器內執行，不下載或覆寫真實快取。"""
import copy
import hashlib
import io
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import music_windows_sources as sources


class Response(io.BytesIO):
    def geturl(self):
        return 'https://zlib.net/fossils/zlib-1.2.13.tar.gz'


class SourceTests(unittest.TestCase):
    def setUp(self):
        self.lock = json.loads(sources.LOCK.read_text())

    def test_lock_rejects_missing_collision_and_unreviewed_url(self):
        sources.records_for(self.lock)
        mutations = []
        missing = copy.deepcopy(self.lock)
        missing['extra_sources'].pop()
        mutations.append(missing)
        collision = copy.deepcopy(self.lock)
        collision['extra_sources'][0]['filename'] = collision['sources'][0]['description']['filename']
        mutations.append(collision)
        wrong_url = copy.deepcopy(self.lock)
        wrong_url['extra_sources'][0]['url'] += '?head=latest'
        mutations.append(wrong_url)
        for lock in mutations:
            with self.subTest(lock=lock['extra_sources']), self.assertRaises(ValueError):
                sources.records_for(lock)

    def test_corrupt_cache_stops_before_network_and_preserves_bytes(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            lock = root/'LOCK.json'
            lock.write_text(json.dumps(self.lock))
            cache = root/'source-cache'
            cache.mkdir()
            bad = cache/self.lock['sources'][0]['description']['filename']
            bad.write_bytes(b'corrupted')
            with patch.multiple(sources, ROOT=root, LOCK=lock, LINUX=root/'missing'), \
                    patch.object(sources.urllib.request, 'urlopen') as network:
                with self.assertRaises(ValueError):
                    sources.prepare(download=True)
                network.assert_not_called()
            self.assertEqual(bad.read_bytes(), b'corrupted')

    def test_symlink_and_offline_missing_rejected(self):
        row = self.lock['extra_sources'][0]
        with tempfile.TemporaryDirectory() as tmp:
            cache = Path(tmp)
            with self.assertRaises(ValueError):
                sources.obtain_extra(cache, row, False)
            target = cache/row['filename']
            target.symlink_to(cache/'does-not-exist')
            with self.assertRaises((ValueError, FileNotFoundError)):
                sources.obtain_extra(cache, row, True)
            self.assertTrue(target.is_symlink())

    def test_download_size_and_hash_fail_atomically(self):
        row = copy.deepcopy(next(r for r in self.lock['extra_sources'] if r['source'] == 'zlib-upstream'))
        row.update(bytes=3, sha256=hashlib.sha256(b'abc').hexdigest())
        with tempfile.TemporaryDirectory() as tmp:
            cache = Path(tmp)
            for data in [b'abcd', b'abx']:
                with self.subTest(data=data), patch.object(sources.urllib.request, 'urlopen', return_value=Response(data)):
                    with self.assertRaises(ValueError):
                        sources.obtain_extra(cache, row, True)
                    self.assertEqual(list(cache.iterdir()), [])

    def test_fixed_download_and_offline_recheck(self):
        row = copy.deepcopy(next(r for r in self.lock['extra_sources'] if r['source'] == 'zlib-upstream'))
        row.update(bytes=3, sha256=hashlib.sha256(b'abc').hexdigest())
        with tempfile.TemporaryDirectory() as tmp:
            cache = Path(tmp)
            with patch.object(sources.urllib.request, 'urlopen', return_value=Response(b'abc')):
                sources.obtain_extra(cache, row, True)
            with patch.object(sources.urllib.request, 'urlopen') as network:
                sources.obtain_extra(cache, row, False)
                network.assert_not_called()
            self.assertEqual((cache/row['filename']).read_bytes(), b'abc')


if __name__ == '__main__':
    unittest.main()
