"""152：凍結來源，建立供v1.6.0沿用素材的Windows引擎更新包。Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import zipfile

root = Path('/src')
p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--version', required=True)
a = p.parse_args()
version = a.version
assert re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}',version)
assert Path.cwd() == root and os.getuid() == os.getgid() == 1000
out = root/'dist-all'/version
batch = root/'workplace'/('pointer-patch-'+version)
assert not out.exists() and not batch.exists()
for p in [root/'dist-all', root/'workplace']:
    assert p.is_dir() and (p.stat().st_uid, p.stat().st_gid) == (1000,1000)
licenses = root/'dist-all/v.1.5.0-20261019/release/LOGH3-v.1.5.0-20261019-windows-amd64/licenses'
assert licenses.is_dir()
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
batch.mkdir(); source = batch/'source'; source.mkdir()
for name in ['cmd', 'internal']: shutil.copytree(root/name, source/name)
for name in ['go.mod', 'go.sum', 'LICENSE']: shutil.copyfile(root/name, source/name)
lock = source/'docs/internals/data/fm-import-lock.json'
lock.parent.mkdir(parents=True)
shutil.copyfile(root/'docs/internals/data/fm-import-lock.json', lock)
env = dict(os.environ, GOPROXY='off', GOSUMDB='off', GOWORK='off')
subprocess.run(['go','mod','vendor'], cwd=source, env=env, check=True)
files = {str(f.relative_to(source)):sha(f) for f in sorted(source.rglob('*')) if f.is_file()}
package = out/'release'/('LOGH3-'+version+'-windows-amd64-patch')
package.mkdir(parents=True)
commands = []
for platform, target in [('windows', package/'gin3.exe'), ('linux', batch/'gin3')]:
    cmd = ['go','build','-mod=vendor','-trimpath','-buildvcs=false',
           '-ldflags=-X main.buildVersion='+version,'-o',str(target),'./cmd/gin3']
    subprocess.run(cmd, cwd=source, env=dict(env, GOOS=platform, GOARCH='amd64', CGO_ENABLED='0' if platform=='windows' else '1'), check=True)
    commands.append(dict(platform=platform, command=cmd, sha256=sha(target)))
shutil.copyfile(root/'LICENSE',package/'LICENSE')
shutil.copytree(licenses,package/'licenses')
readme = f'''銀河英雄傳說III SP {version} Windows更新包

適用現有v.1.6.0-20261024完整版。
關閉遊戲，把本包gin3.exe複製到原遊戲資料夾，取代同名檔案。
原有啟動器、original、data、art、music、soundfont及存檔保留。
存檔仍在%APPDATA%\\logh3\\save，原版六表格式保持。

滑鼠移動改為獨立繪製游標，減少整張HD畫面重建造成的延遲。
主指令、下拉選單及按鈕新增滑鼠停留變色，原版與HD顯示均可使用。
滑鼠移至中央遊戲畫面頂端，工具列可直接選介面字型與電腦戰略AI。
字型使用既有原版點陣、原版強化、思源宋體及思源黑體。
思源黑體使用現有Noto／思源字集；設定頁採既有專用黑體。
原版顯示固定原字形，選擇的字型在HD畫面生效。
AI可選原版或強化0至3，影響後續決策。本次執行有效。
字型與AI與系統設定共享當前選擇；F1說明、F7原版／HD保持。
選單可按Esc或右鍵取消，開啟與取消不會點穿底下遊戲。
五語、三組海報及其他v1.6.0功能沿用原素材。

本包僅含引擎與授權文字，不含原版或海報素材。僅準備本機更新，未發布GitHub Release。
Windows程式以Wine正常玩家路徑驗證，尚未作Windows真機驗收。
'''
(package/'README.txt').write_bytes(b'\xef\xbb\xbf'+readme.replace('\n','\r\n').encode('utf-8'))
(package/'VERSION.txt').write_bytes((version+'\r\n').encode())
manifest = {str(f.relative_to(package)):dict(bytes=f.stat().st_size,sha256=sha(f)) for f in sorted(package.rglob('*')) if f.is_file()}
(package/'MANIFEST.json').write_text(json.dumps(dict(version=version, files=manifest),ensure_ascii=False,indent=2)+'\n')
archive = package.parent/(package.name+'.zip')
with zipfile.ZipFile(archive,'x',zipfile.ZIP_DEFLATED) as z:
    for f in sorted(package.rglob('*')):
        if f.is_file(): z.write(f,str(Path(package.name)/f.relative_to(package)))
with zipfile.ZipFile(archive) as z:
    assert z.testzip() is None
    for info in z.infolist():
        if not info.filename.isascii(): assert info.flag_bits & 0x800
    for name,row in manifest.items():
        b=z.read(package.name+'/'+name)
        assert len(b)==row['bytes'] and hashlib.sha256(b).hexdigest()==row['sha256']
    b=z.read(package.name+'/README.txt')
    assert b.startswith(b'\xef\xbb\xbf') and b'\n' not in b.replace(b'\r\n',b'')
    # 正式收據必須執行從實際ZIP解出的程式。
    installed = root/'workplace/music69-windows/pointer152'/version/'installed'
    assert not installed.exists() and installed.parent.parent.is_dir()
    installed.mkdir(parents=True); z.extractall(installed)
metadata = dict(version=version, base_version='v.1.6.0-20261024', platform='windows-amd64',
    head=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
    working_tree=subprocess.check_output(['git','status','--porcelain'],text=True),
    source_files=files, go_version=subprocess.check_output(['go','version'],text=True).strip(),
    commands=commands, files=manifest, source=str(source.relative_to(root)),
    archive=dict(path=str(archive.relative_to(out)),bytes=archive.stat().st_size,sha256=sha(archive)))
(out/'BUILD.json').write_text(json.dumps(metadata,ensure_ascii=False,indent=2)+'\n')
(out/'SHA256SUMS.txt').write_text(sha(archive)+'  '+str(archive.relative_to(out))+'\n')
print(json.dumps(metadata['archive'],indent=2),flush=True)
