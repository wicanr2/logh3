"""驗HD原曲音訊、循環邊界及來源，保存可重現本機manifest。"""
import argparse
import array
import hashlib
import json
import math
import os
from pathlib import Path
import re
import shutil
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--root', type=Path, required=True)
args = parser.parse_args()
root, output = args.root, args.root/'music-hd'
assert output.stat().st_uid == os.getuid()
sources = json.loads((root/'music-source/SOURCE.json').read_text())
soundfont = json.loads((root/'music-preview/SOURCE.json').read_text())
rows = []
for source in sources:
    path = output/(source['track']+'.ogg')
    probe = json.loads(subprocess.check_output(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(path)],timeout=20))
    stream = probe['streams'][0]
    assert stream['codec_name']=='vorbis' and int(stream['sample_rate'])==44100 and stream['channels']==2
    duration = float(probe['format']['duration'])
    assert duration > 0 and source['notes'] > 0
    scan = subprocess.run(['ffmpeg','-nostdin','-v','info','-i',str(path),'-af','astats=metadata=0:reset=0','-f','null','-'],capture_output=True,text=True,check=True,timeout=60)
    (output/(source['track']+'.analysis.log')).write_text(scan.stderr)
    peaks = [float(v) for v in re.findall(r'Peak level dB: ([-+.\d]+)',scan.stderr)]
    rms = [float(v) for v in re.findall(r'RMS level dB: ([-+.\d]+)',scan.stderr)]
    assert peaks and max(peaks) < -.1 and rms and max(rms)>-50, (source['track'],peaks,rms)
    def edge(position):
        command = ['ffmpeg','-nostdin','-v','error','-ss',str(position),'-i',str(path),'-t','0.01','-ar','44100','-ac','2','-f','f32le','-']
        samples = array.array('f',subprocess.check_output(command,timeout=20))
        assert len(samples) >= 2
        return samples
    start, end = edge(0), edge(max(0,duration-.01))
    jump = max(abs(start[i]-end[-2+i]) for i in range(2))
    assert jump < .01, (source['track'],jump)
    rows.append(dict(source,file=path.name,bytes=path.stat().st_size,sha256=hashlib.sha256(path.read_bytes()).hexdigest(),
                     duration=duration,peak_db=max(peaks),rms_db=max(rms),loop_boundary_jump=jump))
manifest = {'schema':1,'classification':'原版樂曲的HD音色與混音，僅本機完整包；不公開散布',
            'listening':'待人耳確認；解碼及音量檢查不等於聽感驗收',
            'soundfont':soundfont,'exporter':'tools/scripts/hd_music_export.py',
            'renderer':'tools/scripts/hd_music_render.sh','sample_rate':44100,'channels':2,'files':rows}
shutil.copyfile(root/'music-preview/LICENSE.txt',output/'SOUNDFONT-LICENSE.txt')
(output/'MANIFEST.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'tracks':len(rows),'peak_db':max(r['peak_db'] for r in rows),'max_loop_jump':max(r['loop_boundary_jump'] for r in rows)},ensure_ascii=False))
