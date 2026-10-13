"""從正式v.1.2.0正常玩家實錄重生README截圖，保留完整視窗與來源身份。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import struct
import subprocess

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('mode',choices=['prepare','install','check'])
a=p.parse_args()
ROOT=Path('/src')
VERSION='v.1.2.0-20261012'
batch=ROOT/'workplace/hero130'
stage=Path('/work/hero130/readme-shots') if Path('/work').is_dir() else batch/'readme-shots'
images=ROOT/'docs/images'


def sha(path):return hashlib.sha256(path.read_bytes()).hexdigest()
def read(path):return json.loads(path.read_text())
def write(path,value):
    assert path.parent.stat().st_uid==os.getuid()
    path.write_text(json.dumps(value,ensure_ascii=False,indent=2)+'\n')
def dimensions(path):
    data=path.read_bytes()
    assert data[:8]==b'\x89PNG\r\n\x1a\n'
    return struct.unpack('>II',data[16:24])


if a.mode=='prepare':
    assert os.getuid()==1000 and stage.parent.stat().st_uid==1000
    stage.mkdir(exist_ok=True)
    binary=ROOT/'dist-all'/VERSION/'full'/('LOGH3-'+VERSION+'-linux-amd64')/'gin3'
    binary_hash=sha(binary)
    choices=[
        ('empire-map','full-record-f0','strategy.png','帝國藍金介面與戰略星圖'),
        ('alliance-map','full-record-f1-r2','strategy.png','同盟赤焰戰線介面與戰略星圖'),
        ('scenarios','full-record-f1-r2','load.png','正常取進度清單的S1、S2、S3'),
        ('tactical-ai','full-record-f1-r2','turn-03.png','同盟視角的第4回合戰術戰場，顯示敵軍已開啟'),
        ('beam','full-record-f1-r2','review-00.png','正常速度中子光線砲交火'),
        ('fighters','full-record-f1-r2','review-07.png','正常速度戰鬥機交火'),
    ]
    receipts=[]
    for name,run,capture,caption in choices:
        folder=batch/run;trace=read(folder/'TRACE.json')
        assert trace['normal_x11_input'] and not any(trace[k] for k in ['state_injection','seed_injection','script_flag'])
        assert trace['binary_sha256']==binary_hash
        source=folder/capture;destination=stage/('readme-v120-'+name+'.png')
        assert dimensions(source)==(960,720)
        shutil.copyfile(source,destination)
        entry=dict(file=destination.name,caption=caption,source=str(source.relative_to(ROOT)),
            source_sha256=sha(source),sha256=sha(destination),width=960,height=720,
            trace=str((folder/'TRACE.json').relative_to(ROOT)),trace_sha256=sha(folder/'TRACE.json'),graphics='hd')
        if capture.startswith('review-'):
            audit=read(folder/'PLAYBACK-AUDIT.json');assert audit['passed']
            sample=next(s for s in audit['visual_samples'] if s['file']==capture)
            entry.update(raw=str((folder/'RAW.mp4').relative_to(ROOT)),raw_sha256=sha(folder/'RAW.mp4'),
                         video_seconds=sample['video_seconds'],playback_id=sample['id'])
        receipts.append(entry)
    folder=batch/'record-f0-ai1';trace=read(folder/'TRACE.json')
    assert trace['binary_sha256']==binary_hash and trace['normal_x11_input'] and not trace['state_injection']
    assert sha(folder/'RAW.mp4')==trace['raw_sha256']
    rows=[json.loads(line) for line in (folder/'FRAMES.jsonl').open()]
    for mode,seconds in [('original',33.3),('hd',35.4)]:
        wall=trace['record_started_utc_ns']+int(seconds*1e9)
        row=min(rows,key=lambda r:abs(r['wall_ns']-wall))
        assert row['state']=='StateBattle' and row['graphics']==mode and row['battle_ready']
        destination=stage/('readme-v120-switch-'+mode+'.png')
        subprocess.run(['ffmpeg','-nostdin','-v','error','-y','-ss',str(seconds),'-i',str(folder/'RAW.mp4'),
                        '-frames:v','1',str(destination)],check=True,timeout=30)
        assert dimensions(destination)==(960,720)
        receipts.append(dict(file=destination.name,caption='同次正常戰場F7切換：'+mode,
            source=str((folder/'RAW.mp4').relative_to(ROOT)),source_sha256=sha(folder/'RAW.mp4'),
            video_seconds=seconds,sha256=sha(destination),width=960,height=720,graphics=mode,
            trace=str((folder/'TRACE.json').relative_to(ROOT)),trace_sha256=sha(folder/'TRACE.json'),
            tactical=row['tactical']))
    assert receipts[-2]['tactical']==receipts[-1]['tactical']
    write(stage/'readme-v120.json',dict(version=VERSION,binary_sha256=binary_hash,normal_player_path=True,
        state_injection=False,cropping=False,rescaling=False,images=receipts,
        scope='正式Linux封包由正常選單載入S1的實錄。F7兩張同戰術狀態；不宣稱DOS原版同態。',
        rights='僅私有儲存庫的README畫面；完整遊戲、原版檔與動畫影片維持本機'))
    print('PASS 八張完整視窗截圖，正式程式身份及F7同戰術狀態相同')
elif a.mode=='install':
    assert images.is_dir() and images.stat().st_uid==os.getuid()
    receipt=read(stage/'readme-v120.json')
    for item in receipt['images']:
        source=stage/item['file'];assert sha(source)==item['sha256']
        target=images/item['file']
        assert not target.exists() or target.stat().st_uid==os.getuid()
        shutil.copyfile(source,target)
    shutil.copyfile(stage/'readme-v120.json',images/'readme-v120.json')
    print('PASS README截圖已安裝至docs/images，原研究截圖保留')
else:
    receipt=read(images/'readme-v120.json');assert len(receipt['images'])==8
    for item in receipt['images']:
        target=images/item['file'];assert target.stat().st_uid==os.getuid()
        assert dimensions(target)==(960,720) and sha(target)==item['sha256']
        assert sha(ROOT/item['source'])==item['source_sha256']
        assert sha(ROOT/item['trace'])==item['trace_sha256']
    print('PASS 八張README截圖與實錄來源身份')
