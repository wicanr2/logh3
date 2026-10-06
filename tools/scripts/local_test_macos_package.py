"""#94：通用Mach-O及完整本機資料的macOS封包；在GUI Docker執行。"""
import argparse
import hashlib
import json
import math
import os
from pathlib import Path
import plistlib
import shutil
import struct
import zipfile


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda: f.read(1024*1024), b''):
            h.update(block)
    return h.hexdigest()


def record(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')


def code_signature(data, offset, size):
    blob = data[offset:offset+size]
    magic, length, count = struct.unpack_from('>III', blob)
    assert magic == 0xFADE0CC0 and length <= len(blob)
    directories = []
    for i in range(count):
        kind, start = struct.unpack_from('>II', blob, 12+i*8)
        if struct.unpack_from('>I', blob, start)[0] != 0xFADE0C02:
            continue
        fields = struct.unpack_from('>9I4B', blob, start)
        _, _, version, flags, hash_offset, ident, special, slots, limit, hash_size, hash_type, _, page_exponent = fields
        if version >= 0x20300:
            limit64 = struct.unpack_from('>Q', blob, start+56)[0]
            if limit64:
                limit = limit64
        page = 1 << page_exponent
        assert limit <= len(data) and slots == math.ceil(limit/page)
        algorithm = {1:'sha1', 2:'sha256', 3:'sha256', 4:'sha384'}[hash_type]
        for n in range(slots):
            got = hashlib.new(algorithm, data[n*page:min((n+1)*page, limit)]).digest()[:hash_size]
            expected = blob[start+hash_offset+n*hash_size:start+hash_offset+(n+1)*hash_size]
            assert got == expected, ('CodeDirectory page mismatch', n)
        directories.append({'ad_hoc': bool(flags & 2), 'flags': flags, 'algorithm': algorithm,
                            'code_pages': slots, 'code_limit': limit, 'all_page_hashes_match': True})
    assert directories and all(d['ad_hoc'] for d in directories)
    return directories


def macho_slice(data):
    magic, cpu, subtype, filetype, ncmds, cmdsize, flags, reserved = struct.unpack_from('<8I', data)
    assert magic == 0xFEEDFACF and filetype == 2
    arch = {0x1000007:'amd64', 0x100000C:'arm64'}[cpu]
    offset = 32
    deps, minimum, signature = [], [], None
    for _ in range(ncmds):
        command, size = struct.unpack_from('<II', data, offset)
        assert size >= 8 and offset+size <= 32+cmdsize
        if command in (0xC, 0x80000018, 0x8000001F):
            name_offset = struct.unpack_from('<I', data, offset+8)[0]
            name = data[offset+name_offset:offset+size].split(b'\0',1)[0].decode()
            assert name.startswith(('/usr/lib/', '/System/Library/')), name
            deps.append(name)
        elif command == 0x32:
            platform, value, sdk = struct.unpack_from('<III', data, offset+8)
            assert platform == 1
            minimum.append({'macos': f'{value>>16}.{(value>>8)&255}.{value&255}',
                            'sdk': f'{sdk>>16}.{(sdk>>8)&255}.{sdk&255}'})
        elif command == 0x24:
            value, sdk = struct.unpack_from('<II', data, offset+8)
            minimum.append({'macos': f'{value>>16}.{(value>>8)&255}.{value&255}',
                            'sdk': f'{sdk>>16}.{(sdk>>8)&255}.{sdk&255}'})
        elif command == 0x1D:
            signature = code_signature(data, *struct.unpack_from('<II', data, offset+8))
        offset += size
    assert minimum and all(tuple(map(int,m['macos'].split('.'))) <= (11,0,0) for m in minimum)
    # Apple Platform Security：arm64必須有有效簽章，Intel原生未簽章仍可執行。
    # 固定Go1.24.13的Link.NeedCodeSign同樣只要求Darwin/arm64。
    if arch == 'arm64':
        assert signature, 'Apple Silicon須具備可核對的ad-hoc程式頁簽章'
    return {'architecture': arch, 'dependencies': deps, 'deployment': minimum,
            'code_directory': signature or [], 'notarized': False, 'developer_id': False}


def universal(path):
    data = path.read_bytes()
    magic, count = struct.unpack_from('>II', data)
    assert magic == 0xCAFEBABE and count == 2
    result=[]
    for i in range(count):
        cpu, subtype, offset, size, align = struct.unpack_from('>5I', data, 8+20*i)
        item=macho_slice(data[offset:offset+size])
        item.update(offset=offset, bytes=size, sha256=hashlib.sha256(data[offset:offset+size]).hexdigest())
        result.append(item)
    assert {r['architecture'] for r in result} == {'amd64','arm64'}
    return result


def package(version, batch):
    assert Path.cwd()==Path('/src') and os.getuid()==1000
    root=Path('dist-all')/version
    build=json.loads((root/'BUILD.json').read_text())
    stage=Path('workplace')/batch
    binary=stage/'gin3-universal'
    slices=universal(binary)
    common=Path('workplace')/('package92-'+version)/'common'
    assert common.is_dir()
    folder=root/'full-local'/('LOGH3-'+version+'-macos-universal')
    assert not folder.exists()
    folder.mkdir()
    app=folder/'銀河英雄傳說III.app'
    contents=app/'Contents'
    resources=contents/'Resources'
    shutil.copytree(common, resources)
    (contents/'MacOS').mkdir()
    shutil.copyfile(binary, contents/'MacOS/gin3')
    (contents/'MacOS/gin3').chmod(0o755)
    launcher=contents/'MacOS/launch'
    launcher.write_text('#!/bin/sh\nset -eu\n'
        'LOGH3_APP_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)\n'
        'LOGH3_RES_DIR="$LOGH3_APP_DIR/../Resources"\n'
        'mkdir -p "$HOME/Library/Logs/logh3"\n'
        'exec "$LOGH3_APP_DIR/gin3" -orig "$LOGH3_RES_DIR/original" '
        '-data "$LOGH3_RES_DIR/data" -music "$LOGH3_RES_DIR/music" '
        '-ai -1 -scale 1.5 "$@" 2>>"$HOME/Library/Logs/logh3/game.log"\n')
    launcher.chmod(0o755)
    short=version.split('-')[0][2:]
    info={'CFBundleName':'銀河英雄傳說III', 'CFBundleDisplayName':'銀河英雄傳說III',
          'CFBundleIdentifier':'org.logh3.localtest', 'CFBundleExecutable':'launch',
          'CFBundlePackageType':'APPL', 'CFBundleShortVersionString':short,
          'CFBundleVersion':short, 'LSMinimumSystemVersion':'11.0',
          'NSHighResolutionCapable':True, 'LOGH3FullVersion':version,
          'CFBundleGetInfoString':version+' 本機測試版'}
    (contents/'Info.plist').write_bytes(plistlib.dumps(info))
    for mode, level in [('original',-1),('enhanced',1)]:
        script=folder/('start-'+mode+'.command')
        script.write_text('#!/bin/sh\nset -eu\n'
            'cd "$(CDPATH= cd "$(dirname "$0")" && pwd)"\n'
            'exec "./銀河英雄傳說III.app/Contents/MacOS/launch" -ai '+str(level)+' "$@"\n')
        script.chmod(0o755)
    (folder/'README.txt').write_bytes(b'\xef\xbb\xbf'+(
        '銀河英雄傳說III SP remake '+version+' macOS本機完整版測試包\n\n'
        'Intel與Apple Silicon共用同一個通用程式，最低macOS11。\n'
        '完整解壓後，打開銀河英雄傳說III.app使用原版AI。\n'
        '也可雙擊start-original.command或start-enhanced.command選兩種AI；強化預設難度1。\n'
        '程式與資料都在app內，.app可整個搬移，兩個.command須留在app旁。\n'
        '兩種戰略及戰術AI均已啟用，電腦能下令、移動與交戰。\n'
        '22首配樂、28段音效、音色庫與必要原版資料已包含，不用另下載。\n'
        '防拷題可用138進入，右鍵返回；主選單安靜是正常行為。\n'
        '存檔：~/Library/Application Support/logh3/save\n'
        '錯誤紀錄：~/Library/Logs/logh3/game.log\n\n'
        '此包已完成雙架構、系統依賴、Apple Silicon程式頁簽章與逐檔完整性核對。\n'
        'Intel切片未簽章，Apple Silicon切片為ad-hoc簽章。\n'
        '尚未在macOS真機執行，音訊、人耳、完整登陸分段演出仍待試玩。\n'
        '沒有Developer ID或Apple公證。若系統阻擋第一次開啟，依Apple說明在\n'
        '系統設定／隱私權與安全性確認本程式後選仍要打開；本包不變更系統安全設定。\n'
        'Apple說明：https://support.apple.com/102445\n\n'
        '請測新局／S3、兩種AI、艦隊命令、存檔退出再讀回、FM／GM及音效。\n'
        '回報版本、macOS版本、Intel或Apple Silicon、AI模式、存檔、步驟與畫面。\n'
        '完整版含原版及衍生素材，只限合法持有人本機測試，不公開散布。\n').encode())
    record(folder/'VERSION.json',{'version':version,'platform':'macos','architecture':'universal-amd64-arm64',
                                 'source_head':build['head'],'source_files':build['source_files']})
    files={str(p.relative_to(folder)):{'bytes':p.stat().st_size,'sha256':sha(p),
           'mode':oct(p.stat().st_mode&0o777)} for p in sorted(folder.rglob('*')) if p.is_file()}
    assert all(p.stat().st_uid==1000 and p.stat().st_gid==1000 for p in folder.rglob('*'))
    record(folder/'MANIFEST.json',{'version':version,'classification':build['classification'],'files':files})
    archive=folder.parent/(folder.name+'.zip')
    with zipfile.ZipFile(archive,'x',compression=zipfile.ZIP_DEFLATED,compresslevel=6) as z:
        for p in sorted(folder.rglob('*')):
            if p.is_file():z.write(p,str(Path(folder.name)/p.relative_to(folder)))
    with zipfile.ZipFile(archive) as z:
        assert z.testzip() is None
        for path, expected in files.items():
            name=folder.name+'/'+path
            content=z.read(name)
            assert len(content)==expected['bytes'] and hashlib.sha256(content).hexdigest()==expected['sha256']
            assert (z.getinfo(name).external_attr>>16)&0o777==int(expected['mode'],8)
        assert all(i.filename.isascii() or i.flag_bits&0x800 for i in z.infolist())
        assert z.read(folder.name+'/README.txt').startswith(b'\xef\xbb\xbf')
        assert plistlib.loads(z.read(folder.name+'/銀河英雄傳說III.app/Contents/Info.plist'))==info
    smoke=root/'smoke/macos'
    smoke.mkdir()
    for p in stage.glob('*.txt'):shutil.copyfile(p,smoke/p.name)
    result={'version':version,'slices':slices,'zip_members':len(files)+1,'zip_integrity':True,
            'executable_permissions_preserved':True,'runtime_test':'未在macOS真機執行',
            'signature_scope':'Apple Silicon CodeDirectory全頁雜湊；Intel原生未簽章；非Developer ID，app外層未公證',
            'archive':{'path':str(archive.relative_to(root)),'bytes':archive.stat().st_size,'sha256':sha(archive)}}
    record(smoke/'MACOS.json',result)
    build['macos']=result
    build['packages'].append(result['archive'])
    build['not_claimed']=[x for x in build['not_claimed'] if x!='macOS']+['macOS真機執行']
    record(root/'BUILD.json',build)
    print('PASS macOS通用包、ARM簽章全頁雜湊、雙架構依賴、權限與ZIP逐檔；',result['archive'])


if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--version',required=True)
    p.add_argument('--batch',required=True)
    args=p.parse_args()
    package(args.version,args.batch)
