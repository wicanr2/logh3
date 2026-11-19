"""177：v1.7.0推廣片。沿v1.6.4已驗故事片，重繪三張字卡的版號，於結尾字卡前插入Android實錄。Docker內執行。

舊片的片段來源已在11-04清理時刪除，不能整片重剪；這裡直接沿用已驗的成片：
- 字卡用與舊片相同的算法（poster_story_promo.render 的 story 分支）重繪；analyze 先以舊版號重繪、
  與舊片逐格比對，證明算法與剪接格位置相同，才換新版號。
- 舊片其餘影格原樣解碼再編碼，不增刪、不變速。
- 配樂：舊片的BGM_0F只剩2秒，不循環；Android段改用該畫面在遊戲中實際的平時曲BGM_3F
  （S3存檔WORK+1F音源=1 FM、+20平時曲=3），音量沿舊片混音的同一倍率，交界淡出淡入0.8秒。
"""
import argparse
import hashlib
import json
import math
import random
import shutil
import subprocess
import wave
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw, ImageFont

ROOT = Path('/src'); VERSION = 'v.1.7.0-20261119'; OLD_VERSION = 'v.1.6.4-20261031'
OLD = ROOT/'dist-all/promo/LOGH3-v.1.6.4-20261031-promo.mp4'
OLD_SHA = '432fb2cef509043fd95734abfc0eeeaf917057ed91a3a8c9a396c069b0b6f50a'
FLAGS = ROOT/'workplace/hd127/promo-final/BACKGROUND.png'
MUSIC_OLD = ROOT/'workplace/round132/music/BGM_0F.ogg'
MUSIC_NEW = ROOT/'workplace'/('package92-'+VERSION+'-release177')/'common/music/BGM_3F.ogg'
REC = ROOT/'workplace/promo177/emu-rec2'
APK = ROOT/'dist-all'/VERSION/'full'/('LOGH3-'+VERSION+'-android.apk')
WORK = ROOT/'workplace/promo177/movie'; OUT = ROOT/'dist-all'/VERSION/'promo'; NAME = 'LOGH3-'+VERSION+'-promo.mp4'
FONT = Path('/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc')
BOLD = Path('/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc')
RATE = 44100; FPS = 60

# 舊片字卡：(第一格, 格數, 標題, 主句, 副句)。格位由 analyze 量測並核對。
CARDS = [
    (0, 120, '銀河英雄傳說 III SP', '帝國與同盟，對峙百餘年。', '戰火橫越星海，兩個國家仍在追尋各自的未來。'),
    (427, 120, '亞姆力札，遠征軍陷入危局', '深入帝國領土的艦隊，面臨補給困境。', '如果這一次，前線的下一道命令由你下達？'),
]
FINAL_OLD = ('下一頁歷史，由你下令', '銀河英雄傳說 III SP｜v1.6.4', 'P 索敵・整隊行軍・前線整軍・三組海報｜Windows / Linux / macOS')
FINAL_NEW = ('下一頁歷史，由你下令', '銀河英雄傳說 III SP｜v1.7.0', 'Android版・需要時才載入・參謀更換｜Windows / Linux / macOS / Android')
OLD_FRAMES = 7613; FINAL_START = 7493

# Android段：(錄影, 起秒, 秒數, 標題, 說明)。錄影時間＝events-ms的輕觸時間減約0.9秒（戰略）／1.2秒（戰場）。
ANDROID = [
    ('strat', 1.5, 5.0, 'Android版：戰略地圖放大到滿高', '點星域選取，右側卡片隨即顯示星域摘要。'),
    ('strat', 6.8, 4.7, '再點一次，開啟星系情報', '情報視窗沿用原版畫面與操作。'),
    ('strat', 14.9, 4.3, '側欄按鈕，打開原版命令選單', '戰略命令、一覽表、系統指令都在左側欄；左下「退回」等同右鍵。'),
    ('battle', 4.0, 6.5, '戰場用方向鈕捲動視野', '右側方向鈕一下捲三格，沿原版捲動規則。'),
    ('battle', 15.2, 4.8, '點分隊，下達作戰命令', '移動、戰鬥、索敵沿用原版選單；含HD人物、星圖與介面。'),
]


def sha(p):
    with Path(p).open('rb') as f: return hashlib.file_digest(f, 'sha256').hexdigest()
def write(p, v): Path(p).write_text(json.dumps(v, ensure_ascii=False, indent=2)+'\n')
def read(p): return json.loads(Path(p).read_text())
def run(*a): subprocess.run([str(x) for x in a], check=True, timeout=1800)
def capture(*a): return subprocess.run([str(x) for x in a], check=True, timeout=1800, capture_output=True).stdout


def text(image, xy, value, size, bounds, log, bold=False):
    d = ImageDraw.Draw(image); font = ImageFont.truetype(str(BOLD if bold else FONT), size, index=3)
    box = d.textbbox(xy, value, font=font, anchor='mm')
    assert box[0] >= bounds[0] and box[1] >= bounds[1] and box[2] <= bounds[2] and box[3] <= bounds[3], (value, box, bounds)
    d.text(xy, value, font=font, anchor='mm', fill='#f2e4c6', stroke_width=1, stroke_fill='#08101a')
    log.append(dict(text=value, ink_box=list(box), safe_box=list(bounds)))


def card_overlay(flags, title, caption, subtitle, version, log):
    """與 poster_story_promo.render 的 story 分支相同。"""
    overlay = Image.new('RGBA', flags.size); d = ImageDraw.Draw(overlay)
    for y in range(785, 1080): d.line((0, y, 1919, y), fill=(2, 6, 14, round((y-785)/295*195)))
    text(overlay, (960, 70), title, 46, (60, 20, 1860, 115), log, True)
    text(overlay, (480, 160), '銀河帝國', 48, (60, 115, 900, 210), log, True)
    text(overlay, (1440, 160), '自由行星同盟', 48, (1020, 115, 1860, 210), log, True)
    text(overlay, (960, 875), caption, 58, (60, 810, 1860, 930), log, True)
    text(overlay, (960, 970), subtitle, 34, (60, 935, 1860, 1010), log)
    text(overlay, (960, 1040), version, 26, (60, 1015, 1860, 1075), log)
    return overlay


def card_frames(flags, overlay, count):
    rng = random.Random(126); stars = [(rng.randrange(1920), rng.randrange(1080), rng.randrange(1, 3), rng.randrange(25, 70)) for _ in range(100)]
    for n in range(count):
        im = flags.copy(); d = ImageDraw.Draw(im); t = n/60
        for x, y, r, speed in stars:
            xx = int((x+t*speed/8) % 1920); yy = int((y+t*speed) % 1080); d.ellipse((xx, yy, xx+r, yy+r), fill=(155, 165, 185))
        im.paste(overlay, (0, 0), overlay); yield im


def old_frames(start, count, size=None):
    vf = f'select=between(n\\,{start}\\,{start+count-1})' + (f',scale={size[0]}:{size[1]}' if size else '')
    w, h = size or (1920, 1080)
    raw = capture('ffmpeg', '-nostdin', '-v', 'error', '-threads', '2', '-i', OLD, '-vf', vf, '-vsync', '0', '-f', 'rawvideo', '-pix_fmt', 'rgb24', '-')
    arr = np.frombuffer(raw, np.uint8).reshape(-1, h, w, 3); assert len(arr) == count, (len(arr), count); return arr


def psnr(a, b):
    m = np.mean((a.astype(np.float32)-b.astype(np.float32))**2); return 99.0 if m == 0 else 10*math.log10(255**2/m)


def decode_audio(path):
    return np.frombuffer(capture('ffmpeg', '-nostdin', '-v', 'error', '-threads', '2', '-i', path, '-vn', '-ar', RATE, '-ac', '2', '-f', 'f32le', '-'), dtype='<f4').reshape(-1, 2).copy()


def analyze():
    assert sha(OLD) == OLD_SHA
    WORK.mkdir(parents=True, exist_ok=True)
    flags = Image.open(FLAGS).convert('RGB'); report = dict(cards=[], cuts={})
    # 剪接格：字卡前後一格的差異應該突然變大。
    for name, (start, count) in dict(card0_end=(115, 10), card1_begin=(422, 10), card1_end=(542, 10), final_begin=(FINAL_START-5, 10)).items():
        f = old_frames(start, count, (320, 180)).astype(np.float32)
        diffs = [float(np.mean(np.abs(f[i+1]-f[i]))) for i in range(count-1)]
        jump = int(np.argmax(diffs)); report['cuts'][name] = dict(first_frame_after_cut=start+jump+1, diffs=[round(x, 2) for x in diffs])
    assert report['cuts']['card0_end']['first_frame_after_cut'] == 120
    assert report['cuts']['card1_begin']['first_frame_after_cut'] == 427
    assert report['cuts']['card1_end']['first_frame_after_cut'] == 547
    assert report['cuts']['final_begin']['first_frame_after_cut'] == FINAL_START
    # 字卡算法：用舊版號重繪，逐格（抽樣）比對舊片。
    for start, count, title, caption, subtitle in CARDS+[(FINAL_START, 120, *FINAL_OLD)]:
        overlay = card_overlay(flags, title, caption, subtitle, OLD_VERSION, [])
        mine = [np.asarray(im) for im in card_frames(flags, overlay, count)]
        old = old_frames(start, count)
        values = [psnr(mine[i], old[i]) for i in range(0, count, 7)]
        report['cards'].append(dict(start=start, frames=count, title=title, psnr_min=round(min(values), 2), psnr_mean=round(sum(values)/len(values), 2)))
        assert min(values) > 33, (title, values)
    # 配樂倍率：舊片字卡（只有配樂）那段對BGM_0F做最小平方，順便量出AAC延遲。
    old_a = decode_audio(OLD); music = decode_audio(MUSIC_OLD)
    a0, a1 = round(7.3*RATE), round(9.0*RATE); seg = old_a[a0:a1, 0]
    best = None
    for lag in range(-4096, 4097, 1):
        ref = music[a0+lag:a1+lag, 0]
        g = float(np.dot(seg, ref)/np.dot(ref, ref)); err = float(np.mean((seg-g*ref)**2))
        if best is None or err < best[2]: best = (lag, g, err)
    lag, gain, err = best; snr = 10*math.log10(float(np.mean(seg**2))/err)
    report['music'] = dict(lag_samples=lag, gain=round(gain, 6), fit_snr_db=round(snr, 2), old_audio_samples=len(old_a))
    assert snr > 25, report['music']
    write(WORK/'ANALYZE.json', report); print(json.dumps(report, ensure_ascii=False)[:1500])



ENC = ['-c:v', 'libx264', '-preset', 'veryfast', '-crf', '18', '-threads', '2', '-pix_fmt', 'yuv420p', '-video_track_timescale', '60000']


def pipe_frames(target, frames):
    cmd = ['ffmpeg', '-nostdin', '-v', 'error', '-f', 'rawvideo', '-pixel_format', 'rgb24', '-video_size', '1920x1080', '-framerate', '60', '-i', 'pipe:0', '-an', *ENC, str(target)]
    p = subprocess.Popen(cmd, stdin=subprocess.PIPE)
    n = 0
    for im in frames: p.stdin.write(im.tobytes()); n += 1
    p.stdin.close(); assert p.wait(timeout=600) == 0; return n


def android_layers(flags, title, caption, log):
    """手機畫面 1280×576 放大 1.25 倍成 1600×720，置中；外加深色邊框與上下字幕列（同舊片 wide 版式）。"""
    under = flags.copy(); d = ImageDraw.Draw(under)
    d.rounded_rectangle((136, 156, 1783, 923), radius=36, fill=(10, 14, 22), outline=(70, 80, 98), width=4)
    over = Image.new('RGBA', flags.size); d = ImageDraw.Draw(over); left, right = 160, 1759
    d.rectangle((left, 0, right, 59), fill=(4, 8, 16, 255)); d.rectangle((left, 1023, right, 1079), fill=(4, 8, 16, 255))
    text(over, (960, 29), title, 32, (left+12, 0, right-12, 60), log, True)
    text(over, (960, 1050), caption, 26, (left+12, 1025, right-12, 1079), log)
    return under, over


def render():
    analysis = read(WORK/'ANALYZE.json'); assert sha(OLD) == OLD_SHA
    clips = WORK/'clips'; assert not clips.exists(); clips.mkdir()
    flags = Image.open(FLAGS).convert('RGB'); log = []; plan = []
    def card(name, count, title, caption, subtitle):
        target = clips/(name+'.mp4'); ov = card_overlay(flags, title, caption, subtitle, VERSION, log)
        Image.alpha_composite(flags.convert('RGBA'), ov).convert('RGB').save(clips/(name+'.png'))
        n = pipe_frames(target, card_frames(flags, ov, count)); plan.append(dict(kind='card', file=target.name, frames=n, title=title, caption=caption, subtitle=subtitle))
    def old(name, start, stop):
        target = clips/(name+'.mp4')
        run('ffmpeg', '-nostdin', '-v', 'error', '-threads', '2', '-i', OLD, '-vf', f'trim=start_frame={start}:end_frame={stop},setpts=PTS-STARTPTS', '-an', '-r', '60', *ENC, target)
        plan.append(dict(kind='old', file=target.name, frames=stop-start, old_first_frame=start))
    card('00-card', *CARDS[0][1:])
    old('01-old', 120, 427)
    card('02-card', *CARDS[1][1:])
    old('03-old', 547, FINAL_START)
    for i, (rec, start, dur, title, caption) in enumerate(ANDROID):
        under, over = android_layers(flags, title, caption, log)
        u, o = clips/f'a{i}-under.png', clips/f'a{i}-over.png'; under.save(u); over.save(o)
        target = clips/f'{4+i:02}-android.mp4'; count = round(dur*FPS); source = REC/(rec+'.mp4')
        run('ffmpeg', '-nostdin', '-v', 'error', '-loop', '1', '-framerate', '60', '-i', u, '-ss', start, '-i', source, '-loop', '1', '-i', o,
            '-filter_complex_threads', '1', '-filter_complex',
            '[1:v]fps=60,scale=1600:720:flags=lanczos,setsar=1,setpts=PTS-STARTPTS[g];[0:v][g]overlay=160:180:shortest=1[b];[b][2:v]overlay=0:0:shortest=1[out]',
            '-map', '[out]', '-an', '-frames:v', count, '-r', '60', *ENC, target)
        plan.append(dict(kind='android', file=target.name, frames=count, source=str(source.relative_to(ROOT)), source_sha256=sha(source), start=start, title=title, caption=caption))
    card(f'{4+len(ANDROID):02}-final', 120, *FINAL_NEW)
    total = sum(x['frames'] for x in plan)
    (WORK/'concat.txt').write_text(''.join(f"file '{clips/x['file']}'\n" for x in plan))
    # 配樂：舊片音軌到結尾字卡前；之後接BGM_3F，同一倍率，交界0.8秒淡出淡入，片尾淡出1秒。
    old_a = decode_audio(OLD); new = decode_audio(MUSIC_NEW); n = round(total/FPS*RATE)
    cut = round(FINAL_START/FPS*RATE); ov = round(.8*RATE); gain = analysis['music']['gain']
    mix = np.zeros((n, 2), np.float32); first = old_a[:cut].copy(); first[-ov:] *= np.linspace(1, 0, ov, dtype=np.float32)[:, None]; mix[:cut] = first
    need = n-(cut-ov); assert len(new) >= need, (len(new), need)
    second = new[:need]*gain; second[:ov] *= np.linspace(0, 1, ov, dtype=np.float32)[:, None]
    peak_second = float(np.max(np.abs(second))); limit = 10**(-1.5/20)
    scale = min(1.0, limit/peak_second) if peak_second > 0 else 1.0; second *= scale; mix[cut-ov:] += second
    fade = RATE; mix[-fade:] *= np.linspace(1, 0, fade, dtype=np.float32)[:, None]
    with wave.open(str(WORK/'MIX.wav'), 'wb') as f:
        f.setnchannels(2); f.setsampwidth(2); f.setframerate(RATE); f.writeframes(np.round(np.clip(mix, -1, 1)*32767).astype('<i2').tobytes())
    movie = WORK/NAME; assert not movie.exists()
    run('ffmpeg', '-nostdin', '-v', 'error', '-f', 'concat', '-safe', '0', '-i', WORK/'concat.txt', '-i', WORK/'MIX.wav', '-map', '0:v:0', '-map', '1:a:0',
        '-frames:v', total, '-vf', 'fps=60', '-r', '60', *ENC, '-c:a', 'aac', '-b:a', '256k', '-ar', RATE, '-ac', '2', '-movflags', '+faststart', movie)
    write(WORK/'PLAN.json', dict(version=VERSION, frames=total, seconds=total/FPS, segments=plan, text_bounds=log,
        inputs={str(p.relative_to(ROOT)): sha(p) for p in [OLD, FLAGS, MUSIC_OLD, MUSIC_NEW, REC/'strat.mp4', REC/'battle.mp4', REC/'events-ms.txt', APK, Path(__file__).resolve()]},
        music=dict(old_part='v1.6.4成片音軌原樣（BGM_0F＋遊戲音效）至第%d格' % FINAL_START, new_part='BGM_3F從零開始，倍率%.4f×%.4f' % (gain, scale),
                   crossfade_seconds=.8, final_fade_seconds=1, loops=0),
        scope='舊片影格原樣重新編碼；三張字卡以相同算法重繪新版號；Android段為release APK於模擬器的正常玩家路徑實錄（S3存檔讀入），原速，未加速。',
        rights='含原版及衍生音畫，只留本機，不公開發布。'))
    print('RENDER', total, 'frames', round(total/FPS, 3), 's', flush=True)


def audit():
    plan = read(WORK/'PLAN.json'); movie = WORK/NAME
    info = json.loads(capture('ffprobe', '-v', 'error', '-show_entries', 'stream=codec_name,width,height,r_frame_rate,nb_frames,sample_rate,channels', '-show_entries', 'format=duration', '-of', 'json', movie))
    v = next(s for s in info['streams'] if s['codec_name'] == 'h264'); a = next(s for s in info['streams'] if s['codec_name'] == 'aac')
    assert (v['width'], v['height'], v['r_frame_rate']) == (1920, 1080, '60/1') and int(v['nb_frames']) == plan['frames'], v
    assert (a['sample_rate'], a['channels']) == ('44100', 2), a
    # 舊片影格保持：新片中舊段的位置抽樣比對舊片同一格。
    checks = []; pos = 0
    for seg in plan['segments']:
        if seg['kind'] == 'old':
            for k in range(0, seg['frames'], max(1, seg['frames']//12)):
                a_ = new_frame(OLD, seg['old_first_frame']+k); b_ = new_frame(movie, pos+k)
                checks.append(round(psnr(a_, b_), 2))
        pos += seg['frames']
    # 跳轉取格本身要先對：舊片第8格用兩種方法取，必須相同。
    assert np.array_equal(new_frame(OLD, 8), old_frames(8, 1)[0])
    assert min(checks) > 33, checks
    black = subprocess.run(['ffmpeg', '-nostdin', '-i', str(movie), '-vf', 'blackdetect=d=0.1:pix_th=0.05', '-an', '-f', 'null', '-'], capture_output=True, text=True, timeout=1800).stderr
    blacks = [l for l in black.splitlines() if 'black_start' in l]
    audio = decode_audio(movie); old_a = decode_audio(OLD); cut = round(FINAL_START/FPS*RATE)-round(.8*RATE)
    n = min(cut, len(audio)); err = float(np.mean((audio[:n]-old_a[:n])**2)); snr = 10*math.log10(float(np.mean(old_a[:n]**2))/err)
    rms, peak = float(np.sqrt(np.mean(audio**2))), float(np.max(np.abs(audio)))
    tail = audio[cut:]; tail_rms = float(np.sqrt(np.mean(tail**2)))
    assert snr > 30 and peak < 1 and rms > .01 and tail_rms > .01 and not blacks, (snr, peak, rms, tail_rms, blacks)
    # 目視用：每段開頭、中間、結尾各一格。
    thumbs = []; pos = 0
    for seg in plan['segments']:
        for k in sorted({0, seg['frames']//2, seg['frames']-1}):
            if seg['kind'] == 'old' and k not in (0, seg['frames']-1): continue
            im = Image.fromarray(new_frame(movie, pos+k)).resize((480, 270)); d = ImageDraw.Draw(im)
            d.rectangle((0, 0, 300, 16), fill='black'); d.text((2, 2), f'{seg["file"]} +{k} ({(pos+k)/FPS:.2f}s)', fill='yellow'); thumbs.append(im)
        pos += seg['frames']
    cols = 4; g = Image.new('RGB', (480*cols, 270*math.ceil(len(thumbs)/cols)))
    for i, im in enumerate(thumbs): g.paste(im, ((i % cols)*480, (i//cols)*270))
    g.save(WORK/'AUDIT-GRID.png')
    write(WORK/'AUDIT.json', dict(passed=True, file=NAME, sha256=sha(movie), bytes=movie.stat().st_size, frames=plan['frames'], seconds=plan['seconds'],
        old_frames_psnr_min=min(checks), old_frames_samples=len(checks), black_segments=len(blacks), audio=dict(rms=round(rms, 5), peak=round(peak, 5),
        old_part_snr_db=round(snr, 2), new_part_rms=round(tail_rms, 5)), grid=sha(WORK/'AUDIT-GRID.png'), visual_review='待目視'))
    print('AUDIT PASS', plan['frames'], 'frames', 'old psnr min', min(checks), 'snr', round(snr, 2), 'peak', round(peak, 4), flush=True)


def new_frame(movie, n):
    """第n格（定格率60）：-ss 在輸入前做精確跳轉，取 pts ≥ n/60 的第一格。"""
    raw = capture('ffmpeg', '-nostdin', '-v', 'error', '-threads', '2', '-ss', f'{n/FPS:.6f}', '-i', movie, '-frames:v', '1', '-f', 'rawvideo', '-pix_fmt', 'rgb24', '-')
    return np.frombuffer(raw, np.uint8).reshape(1080, 1920, 3)


def publish():
    audit_ = read(WORK/'AUDIT.json'); assert audit_['passed'] and audit_['visual_review'].startswith('已目視'), audit_['visual_review']
    movie = WORK/NAME; assert sha(movie) == audit_['sha256']
    OUT.mkdir(parents=True, exist_ok=False); shutil.copyfile(movie, OUT/NAME)
    for n in ['PLAN.json', 'AUDIT.json', 'AUDIT-GRID.png', 'ANALYZE.json']: shutil.copyfile(WORK/n, OUT/n)
    assert sha(OUT/NAME) == audit_['sha256']; print('PUBLISH', OUT/NAME, audit_['sha256'], flush=True)

if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__); p.add_argument('mode', choices=['analyze', 'render', 'audit', 'publish'])
    globals()[p.parse_args().mode]()
