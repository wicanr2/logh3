"""155：從正常入口擷取固定構圖草稿；不是正式功能驗收。"""
import sys
from pathlib import Path

# 沿154已驗X11前置；完整執行來源及SHA-256寫入收據。
base=Path('/src/tools/scripts/poster_decoration_gui.py')
prefix=base.read_text().split('# 154：正常鍵鼠、裝飾留白、循環與存讀；')[0]
exec(compile(prefix,str(base),'exec'))

try:
    command=start('preview')
    click(300,entry_y);state('StateCopyProtection')
    for k in ['1','3','8','Return']:key(k)
    click(300,240,3);click(300,240,3);state('StateCommandPhase')
    settings();poster('v1');close();snap('elegant')
    w,h=geometry();log=(out/'ffmpeg.log').open('w');logs.append(log)
    recorder=subprocess.Popen(['ffmpeg','-v','warning','-nostdin','-y','-threads','1','-filter_threads','1',
        '-f','x11grab','-framerate','15','-video_size',str(w)+'x'+str(h),'-i',env['DISPLAY']+'+0,0',
        '-t','18','-an','-c:v','libx264','-threads','1','-preset','veryfast','-crf','18','-pix_fmt','yuv420p',str(out/'elegant.mp4')],env=env,stdout=log,stderr=subprocess.STDOUT)
    assert recorder.wait(timeout=90)==0;recorder=None
    snap('elegant-after')
    (out/'RESULT.json').write_text(json.dumps(dict(passed=True,prototype=True,normal_entry=True,command=command,
        binary_sha256=sha(a.binary),driver_sha256=sha(__file__),prefix_sha256=sha(base),events=events,
        video_sha256=sha(out/'elegant.mp4')),ensure_ascii=False,indent=2)+'\n')
finally:
    stop()
    for log in logs:log.close()
