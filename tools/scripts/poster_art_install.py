"""150：非破壞性保存四張image_gen群像與來源，v1原圖保持。Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil

ROOT = Path('/src')
OUT = ROOT/'assets/hd/poster'
FILES = {
    'v2-empire': 'exec-225d73de-43f4-4c38-a29e-c462e053447b.png',
    'v2-alliance': 'exec-a14e9af8-5e90-4805-877b-885d6db853ad.png',
    'v3-empire': 'exec-16acf567-da1c-4be2-9c5a-92584a8fe217.png',
    'v3-alliance': 'exec-58175410-4a48-4242-b49d-c092414c6e7a.png',
}


def main():
    parser = argparse.ArgumentParser(description=__doc__); parser.add_argument('--narrow', action='store_true'); args = parser.parse_args()
    files = FILES if not args.narrow else {
        'v2-empire-narrow': 'exec-af9ec732-3bf8-4c3c-b6a6-fb7c6f332081.png',
        'v2-alliance-narrow': 'exec-04bd7dfd-ff2b-46a2-b7f3-24b189ab7faa.png',
        'v3-empire-narrow': 'exec-953db526-796e-4bf8-bad4-1291251c23e8.png',
        'v3-alliance-narrow': 'exec-09dd2230-1ecc-4c41-9965-2dedcdc5d765.png',
    }
    assert os.getuid() == os.getgid() == OUT.stat().st_uid == 1000
    rows = []
    for name, source in files.items():
        target = OUT/(name+'.png'); original = Path('/generated')/source
        assert original.is_file() and not target.exists()
        shutil.copyfile(original, target)
        rows.append(dict(file=str(target.relative_to(ROOT)), generated_file=source,
                         sha256=hashlib.sha256(target.read_bytes()).hexdigest(), bytes=target.stat().st_size))
    value = dict(mode='built-in image_gen', files=rows,
                 prompts=json.loads((ROOT/'workplace/poster150'/('NARROW-PROMPTS.json' if args.narrow else 'PROMPTS.json')).read_text()),
                 references=json.loads((ROOT/'workplace/poster150/references/SOURCES.json').read_text()),
                 rights='官方經典動畫參考重構，本機使用，不加入公開Git或引擎包；v1原圖保持')
    (OUT/('NARROW.SOURCE.json' if args.narrow else 'SETS.SOURCE.json')).write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')
    print('PASS 四張群像原檔及完整提示、來源已保存；v1未改', flush=True)


if __name__ == '__main__': main()
