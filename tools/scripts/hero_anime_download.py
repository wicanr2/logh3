"""規格140：下載使用者指定的公開清晰串流，來源與輸出僅留本機。"""
import gzip
import hashlib
import json
import os
from pathlib import Path
import subprocess
import urllib.parse
import urllib.request
from concurrent.futures import ThreadPoolExecutor

root = Path('/work/hero130')
assert root.stat().st_uid == os.getuid() == 1000
page = 'https://www.bilibili.com/video/BV1G87szDEh7/'
headers = {'User-Agent': 'Mozilla/5.0', 'Referer': page}
meta = json.loads((root/'BILIBILI-STATE.json').read_text())['videoData']
play = json.loads((root/'BILIBILI-PLAY.json').read_text())
assert play['code'] == 0 and meta['bvid'] == 'BV1G87szDEh7'
dash = play['data']['dash']
video = max((v for v in dash['video'] if v['codecs'].startswith('avc1')), key=lambda v: v['height'])
audio = max(dash['audio'], key=lambda v: v['bandwidth'])
receipt = {'source': page, 'title': meta['title'], 'cid': meta['cid'], 'duration_ms': play['data']['timelength'],
           'classification': '使用者指定；本機推廣片限定，未取得公開散布權',
           'native_width': video['width'], 'native_height': video['height'],
           'no_login_or_drm': True, 'files': []}
for kind, stream in [('video', video), ('audio', audio)]:
    url = stream.get('baseUrl') or stream['base_url']
    assert urllib.parse.urlparse(url).scheme == 'https'
    target = root/('anime-'+kind+'.m4s')
    if not target.exists():
        partial = target.with_suffix('.partial')
        offset = partial.stat().st_size if partial.exists() else 0
        request_headers = dict(headers)
        if offset:
            request_headers['Range'] = f'bytes={offset}-'
        with urllib.request.urlopen(urllib.request.Request(url, headers=request_headers), timeout=60) as response, partial.open('ab' if offset else 'xb') as output:
            if offset:
                assert response.status == 206 and response.headers['Content-Range'].startswith(f'bytes {offset}-'), '伺服器未接受續傳'
            size = offset
            while block := response.read(1024*1024):
                size += len(block)
                assert size < 200*1024*1024, '串流超出本次容量上限'
                output.write(block)
        partial.rename(target)
    receipt['files'].append({'kind': kind, 'file': target.name, 'bytes': target.stat().st_size,
                             'sha256': hashlib.sha256(target.read_bytes()).hexdigest(), 'codec': stream['codecs']})
# CDN大串流曾在中途傳回不連續內容。逐段驗Content-Range與長度後重組，
# 原失敗檔留存；不把容器可開啟當作有效影格證據。
clean_video = root/'anime-video-checked.m4s'
if not clean_video.exists():
    url = video.get('baseUrl') or video['base_url']
    total = (root/'anime-video.m4s').stat().st_size
    chunk = 2*1024*1024
    def part(index):
        start, end = index*chunk, min(total,(index+1)*chunk)-1
        target = root/f'anime-part-{index:03}.bin'
        if target.exists() and target.stat().st_size == end-start+1:
            return target
        for attempt in range(3):
            try:
                h = dict(headers,Range=f'bytes={start}-{end}')
                with urllib.request.urlopen(urllib.request.Request(url,headers=h),timeout=60) as response:
                    assert response.status==206 and response.headers['Content-Range']==f'bytes {start}-{end}/{total}'
                    data=response.read()
                assert len(data)==end-start+1
                target.write_bytes(data)
                return target
            except Exception:
                if attempt==2:raise
    with ThreadPoolExecutor(max_workers=3) as pool:
        parts=list(pool.map(part,range((total+chunk-1)//chunk)))
    with clean_video.open('xb') as stream:
        for item in parts:stream.write(item.read_bytes())
    assert clean_video.stat().st_size==total
movie = root/'ANIME-SOURCE-CLEAN.mp4'
if not movie.exists():
    subprocess.run(['ffmpeg', '-nostdin', '-v', 'warning', '-i', str(clean_video),
                    '-i', str(root/'anime-audio.m4s'), '-map', '0:v:0', '-map', '1:a:0', '-c', 'copy',
                    '-movflags', '+faststart', str(movie)], check=True, timeout=120)
probe = json.loads(subprocess.check_output(['ffprobe', '-v', 'error', '-show_streams', '-show_format', '-of', 'json', str(movie)]))
assert {s['codec_type'] for s in probe['streams']} == {'video', 'audio'}
assert all(abs(float(s['duration'])-receipt['duration_ms']/1000)<1 for s in probe['streams'])
with (root/'ANIME-CLEAN-DECODE.log').open('w') as log:
    subprocess.run(['ffmpeg','-nostdin','-v','error','-xerror','-i',str(movie),'-f','null','-'],
                   check=True,stdout=log,stderr=subprocess.STDOUT,timeout=150)
assert (root/'ANIME-CLEAN-DECODE.log').stat().st_size==0
receipt['checked_video']={'file':clean_video.name,'sha256':hashlib.sha256(clean_video.read_bytes()).hexdigest()}
receipt['movie'] = {'file': movie.name, 'sha256': hashlib.sha256(movie.read_bytes()).hexdigest(), 'probe': probe}
(root/'ANIME-SOURCE-CLEAN.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n')
print('PASS', meta['title'], video['width'], video['height'], movie.stat().st_size, flush=True)
