"""Docker內凍結來源，建立不含素材的Windows執行檔替換包。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import zipfile

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--version', required=True)
p.add_argument('--with-posters', action='store_true', help='建立含本機海報的更新包，放full而非公開release')
a = p.parse_args()
assert re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}', a.version)
root = Path('/src')
assert Path.cwd() == root and os.getuid() == os.getgid() == 1000
out = root/'dist-all'/a.version
batch = root/'workplace'/('personnel-patch-'+a.version)
assert not out.exists() and not batch.exists()
assert (root/'dist-all').stat().st_uid == (root/'workplace').stat().st_uid == os.getuid()
base = root/'dist-all/v.1.5.0-20261019/release/LOGH3-v.1.5.0-20261019-windows-amd64'
assert base.is_dir() and (base/'gin3.exe').is_file()
sha = lambda f: hashlib.sha256(f.read_bytes()).hexdigest()
batch.mkdir()
source = batch/'source'
source.mkdir()
for name in ['cmd','internal']: shutil.copytree(root/name, source/name)
for name in ['go.mod','go.sum','LICENSE']: shutil.copyfile(root/name,source/name)
lock = source/'docs/internals/data/fm-import-lock.json'
lock.parent.mkdir(parents=True)
shutil.copyfile(root/'docs/internals/data/fm-import-lock.json',lock)
env = dict(os.environ,GOPROXY='off',GOSUMDB='off',GOWORK='off')
subprocess.run(['go','mod','vendor'],cwd=source,env=env,check=True)
files = {str(f.relative_to(source)):sha(f) for f in sorted(source.rglob('*')) if f.is_file()}
metadata = dict(version=a.version, base_version='v.1.5.0-20261019',platform='windows-amd64',
    head=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
    working_tree=subprocess.check_output(['git','status','--porcelain'],text=True),
    source_files=files,go_version=subprocess.check_output(['go','version'],text=True).strip())
package = out/('full' if a.with_posters else 'release')/('LOGH3-'+a.version+('-windows-amd64-poster-update' if a.with_posters else '-windows-amd64-patch'))
package.mkdir(parents=True)
command = ['go','build','-mod=vendor','-trimpath','-buildvcs=false',
    '-ldflags=-X main.buildVersion='+a.version,'-o',str(package/'gin3.exe'),'./cmd/gin3']
subprocess.run(command,cwd=source,env=dict(env,GOOS='windows',GOARCH='amd64',CGO_ENABLED='0'),check=True)
shutil.copyfile(root/'LICENSE',package/'LICENSE')
shutil.copytree(base/'licenses',package/'licenses')
readme = f'''銀河英雄傳說III SP {a.version} Windows修正包

適用v.1.5.0-20261019、v.1.5.1-20261024及v.1.5.2-20261024的既有遊戲資料夾。
關閉遊戲後，將本包gin3.exe複製到原1.5遊戲資料夾，取代同名檔案。
保留原有original、data、art、music與soundfont資料夾及啟動器。
存檔仍在%APPDATA%\\logh3\\save，原有存檔格式保持。

修正駐留前線後再次出擊仍從母星領船。編成、取消退船及新艦隊位置均使用提督所在地。
改善參謀與副官選人視窗反覆重建字模的延遲，保持字形、裁切與畫面順序。
新增「非戰艦旗艦」選項，預設關閉；開啟後手動編成及擊破轉移允許B／H／C／A／D。
修護艦E、運補艦T仍不能當旗艦。轉移仍須開啟「轉移旗艦」，成功率沿既有提督設定。
新選項會保存，重開及舊存檔沿用；不改寫已存在的旗艦或原版六表格式。
本包只含引擎與授權文字，不含原版或HD素材。未發布GitHub Release。
'''
if a.with_posters:
    poster_root = root/'assets/hd/poster'
    assert (poster_root/'empire.png').is_file() and (poster_root/'alliance.png').is_file()
    shutil.copytree(poster_root, package/'art/poster')
    readme = readme.replace('本包只含引擎與授權文字，不含原版或HD素材。未發布GitHub Release。', '''
新增設定頁右上的「海報」開關，預設關閉。開啟後左帝國萊因哈特、右同盟楊威利，中央完整4:3。
F7切換原版／HD時海報保持；海報側欄不接受遊戲操作。開關保存在原偏好目錄。
請將本包gin3.exe與art資料夾合併到既有遊戲資料夾，取代同名檔案；其餘素材及存檔保持。
本包包含由官方經典動畫參考重構的海報，僅本機使用，不公開散布。來源與提示見art/poster。
未發布GitHub Release。
''')
(package/'README.txt').write_bytes(b'\xef\xbb\xbf'+readme.replace('\n','\r\n').encode())
manifest = {str(f.relative_to(package)):dict(bytes=f.stat().st_size,sha256=sha(f)) for f in sorted(package.rglob('*')) if f.is_file()}
(package/'MANIFEST.json').write_text(json.dumps(dict(version=a.version,files=manifest),ensure_ascii=False,indent=2)+'\n')
archive = package.parent/(package.name+'.zip')
with zipfile.ZipFile(archive,'x',zipfile.ZIP_DEFLATED) as z:
    for f in sorted(package.rglob('*')):
        if f.is_file(): z.write(f,str(Path(package.name)/f.relative_to(package)))
with zipfile.ZipFile(archive) as z:
    assert z.testzip() is None
    for name,row in manifest.items():
        b=z.read(package.name+'/'+name)
        assert len(b)==row['bytes'] and hashlib.sha256(b).hexdigest()==row['sha256']
metadata.update(build_command=command,files=manifest,archive=dict(path=str(archive.relative_to(out)),bytes=archive.stat().st_size,sha256=sha(archive)))
(out/'BUILD.json').write_text(json.dumps(metadata,ensure_ascii=False,indent=2)+'\n')
(out/'SHA256SUMS.txt').write_text(sha(archive)+'  '+str(archive.relative_to(out))+'\n')
print(json.dumps(metadata['archive'],indent=2),flush=True)
