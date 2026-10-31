"""下一次HD發行的藝術stage：人物預縮，原始HD與既有封包保持。只在Docker執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess


def sha(path):
    with Path(path).open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def prepare(art, original, output):
    art, original, output = map(Path, (art, original, output))
    assert art.is_dir() and original.is_dir() and not output.exists()
    assert output.parent.stat().st_uid == os.getuid() == 1000
    assert not output.resolve().is_relative_to(art.resolve())
    before = {str(p.relative_to(art)): sha(p) for p in art.rglob('*') if p.is_file()}
    output.mkdir()
    for p in art.iterdir():
        if p.name == 'face':
            continue
        target = output / p.name
        if p.is_dir():
            shutil.copytree(p, target)
        else:
            shutil.copy2(p, target)
    subprocess.run(['go', 'run', './cmd/hd-pack', '-orig', str(original), '-art', str(art),
                    '-out', str(output/'face')], check=True, timeout=600)
    subprocess.run(['go', 'run', './cmd/hd-pack', '-orig', str(original), '-art', str(art),
                    '-out', str(output/'face'), '-check'], check=True, timeout=600)
    for p in (art/'face').iterdir():
        if p.is_file() and p.suffix != '.png':
            shutil.copy2(p, output/'face'/p.name)
    packed = json.loads((output/'face/RELEASE.json').read_text())
    by_file = {'face/'+r['file']: r for r in packed['faces']}
    manifest = output/'MANIFEST.json'
    if manifest.exists():
        m = json.loads(manifest.read_text())
        for row in m.get('files', []):
            if row['file'] in by_file:
                r = by_file[row['file']]
                assert row['sha256'] == r['source_sha256']
                row.update(hd_master_sha256=row['sha256'], sha256=r['sha256'],
                           width=r['size'][0], height=r['size'][1], bytes=(output/row['file']).stat().st_size,
                           release_method='runtime-compatible Scale3 bilinear')
        manifest.write_text(json.dumps(m, ensure_ascii=False, indent=2)+'\n')
    # 所有未縮圖片及來源文件保持，原始目錄逐檔SHA保持。
    assert before == {str(p.relative_to(art)): sha(p) for p in art.rglob('*') if p.is_file()}
    for name, digest in before.items():
        if name not in by_file and name != 'MANIFEST.json':
            assert sha(output/name) == digest, name
    summary = dict(format='logh3-hd-release-stage-v1', source=str(art), source_files=before,
                   output=str(output), changed_files=sorted(by_file),
                   before_bytes=sum((art/n).stat().st_size for n in by_file),
                   after_bytes=sum((output/n).stat().st_size for n in by_file),
                   preserved='masters, original-mode data, fonts, approved poster layers and source-bound atlases')
    (output/'PACKED-HD.json').write_text(json.dumps(summary, ensure_ascii=False, indent=2)+'\n')
    print('PASS HD stage', len(by_file), summary['before_bytes'], '->', summary['after_bytes'], flush=True)
    return summary


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--art', required=True)
    p.add_argument('--orig', required=True)
    p.add_argument('--out', required=True)
    a = p.parse_args()
    prepare(a.art, a.orig, a.out)
