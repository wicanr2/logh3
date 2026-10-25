"""只抓AppImage runtime的官方授權全文；保存本機快照及來源雜湊，後續離線打包。"""
import hashlib
import json
import os
from pathlib import Path
import urllib.request

out=Path('/src/workplace/appimage153/runtime-licenses')
assert not out.exists() and out.parent.stat().st_uid==os.getuid()==1000
out.mkdir()
urls={
    'AppImage-runtime.txt':'https://raw.githubusercontent.com/AppImage/type2-runtime/main/LICENSE',
    'musl.txt':'https://git.musl-libc.org/cgit/musl/plain/COPYRIGHT',
    'libfuse-LGPL2.txt':'https://raw.githubusercontent.com/libfuse/libfuse/master/LGPL2.txt',
    'squashfuse.txt':'https://raw.githubusercontent.com/vasi/squashfuse/master/LICENSE',
    'zstd.txt':'https://raw.githubusercontent.com/facebook/zstd/dev/LICENSE',
    'zlib.txt':'https://raw.githubusercontent.com/madler/zlib/develop/LICENSE',
}
rows={}
for name,url in urls.items():
    with urllib.request.urlopen(url,timeout=45) as response: content=response.read()
    assert len(content)>100 and b'<html' not in content[:100].lower()
    (out/name).write_bytes(content)
    rows[name]=dict(url=url,bytes=len(content),sha256=hashlib.sha256(content).hexdigest())
    print('SAVED',name,flush=True)
(out/'SOURCES.json').write_text(json.dumps(rows,indent=2)+'\n')
