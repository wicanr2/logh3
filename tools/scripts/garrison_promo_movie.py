"""146：剪輯本輪正常視窗實錄，保留完整遊戲畫布及連續配樂。Docker內執行。"""
import argparse
import hashlib
import json
import math
import os
from pathlib import Path
import random
import re
import subprocess

import numpy as np
from PIL import Image, ImageDraw, ImageFont

ROOT = Path('/src')
VERSION = 'v.1.5.2-20261024'
SOURCE = ROOT/'workplace/garrison146-recording-r2'
WORK = ROOT/'workplace/garrison146-movie'
OUT = ROOT/'dist-all'/VERSION/'promo'
MUSIC = ROOT/'workplace/round132/music/BGM_0F.ogg'
MOVIE = OUT/f'LOGH3-{VERSION}-garrison-promo.mp4'
FONT = Path('/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc')
BOLD = Path('/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc')


def sha(path):
    h = hashlib.sha256()
    with Path(path).open('rb') as f:
        for chunk in iter(lambda: f.read(1024*1024), b''): h.update(chunk)
    return h.hexdigest()


def write(path, value): Path(path).write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')
def read(path): return json.loads(Path(path).read_text())
def run(*args): return subprocess.run(list(map(str, args)), check=True, timeout=600)
def capture(*args): return subprocess.check_output(list(map(str, args)), timeout=180)
def probe(path): return json.loads(capture('ffprobe', '-v', 'error', '-show_streams', '-show_format', '-of', 'json', path))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--resume-mix', action='store_true', help='沿已編碼鏡頭，只重做配樂與驗收；逐張比對重生字幕及字卡')
    parser.add_argument('--audit-only', action='store_true', help='保留已合成影片，重生相同字卡並補做音畫驗收')
    args = parser.parse_args()
    assert not (args.resume_mix and args.audit_only)
    reuse = args.resume_mix or args.audit_only
    assert SOURCE.is_dir()
    assert WORK.parent.stat().st_uid == OUT.parent.stat().st_uid == os.getuid()
    if args.audit_only:
        assert WORK.is_dir() and OUT.is_dir() and MOVIE.is_file()
    elif args.resume_mix:
        assert WORK.is_dir() and OUT.is_dir() and MOVIE.is_file() and not (WORK/'QUIET-CANDIDATE.mp4').exists()
        MOVIE.rename(WORK/'QUIET-CANDIDATE.mp4')
    else:
        assert not WORK.exists() and not OUT.exists()
        WORK.mkdir(); OUT.mkdir()
    recorded = read(SOURCE/'RESULT.json')
    assert recorded['passed'] and recorded['normal_entry']
    assert not any(recorded[k] for k in ['state_injection', 'seed_injection', 'script_flag'])
    assert sha(SOURCE/'RAW.mp4') == recorded['video_sha256']
    assert sha(MUSIC) == '48c22ed0b5d9fec5d0b9bed70e7a06ae1ab3a07faef99da80ddfd21b4644739c'
    raw_duration = float(probe(SOURCE/'RAW.mp4')['format']['duration'])
    origin = recorded['record_started_utc_ns']
    marks = {e['name']: (e['wall_begin_ns']-origin)/1e9 for e in recorded['events'] if e['kind'] == 'capture'}
    for e in recorded['events']:
        if e['kind'] == 'capture': assert sha(SOURCE/e['file']) == e['sha256']
    # 140.5至197.5秒連續呈現，只有章節文字改變；不加速鍵鼠或替換畫面。
    scenes = [
        dict(kind='card', duration=4, title='前線整軍，再次出發', caption='銀河英雄傳說 III SP｜同盟軍實際操作', lines=['駐地領船', '改選提督', '重新出擊']),
        dict(kind='game', start=60.5, duration=5, title='將艦隊駐留前線', caption='艦隊駐留後，船艦與提督留在當地。', lines=['01', '前線駐留', '下達命令']),
        dict(kind='game', start=127.5, duration=13, title='非戰艦旗艦，由你開啟', caption='新選項預設關閉；在系統設定中啟用。', lines=['02', '系統設定', '預設關閉']),
        dict(kind='game', start=140.5, duration=18, title='從駐地改選提督', caption='先查看一位提督，再改選派特，編成第3艦隊。', lines=['03', '改選提督', '第3艦隊']),
        dict(kind='game', start=158.5, duration=25, title='領取當地船艦，重新編成', caption='巡航艦可以擔任旗艦；修護艦E、運補艦T仍會被拒絕。', lines=['04', '駐地船艦', 'E／T除外']),
        dict(kind='game', start=183.5, duration=14, title='第3艦隊，從前線出發', caption='確認編成，選擇目的地，再執行本回合命令。', lines=['05', '下達移動', '前線出發']),
        dict(kind='game', start=265.8, duration=4, title='下一回合，艦隊抵達', caption='此處剪去回合等待，接續同一局的實際抵達結果。', lines=['06', '抵達目的地', '回合等待剪輯']),
        dict(kind='card', duration=4, title='讓你的艦隊接續遠征', caption='非戰艦旗艦預設關閉｜E／T始終不能擔任旗艦', lines=['前線領船', '就地整軍', '再次出發']),
    ]
    if 'SCENE_FACTORY' in globals(): scenes = SCENE_FACTORY()
    wide = globals().get('WIDE', False)
    elapsed = 0
    for i, scene in enumerate(scenes):
        scene.update(id=i, start_at=elapsed); elapsed += scene['duration']
        if scene['kind'] == 'game': assert scene['start']+scene['duration'] <= scene.get('raw_duration', raw_duration)
    assert elapsed < float(probe(MUSIC)['format']['duration'])
    inputs = {str(p.relative_to(ROOT)): sha(p) for p in [SOURCE/'RESULT.json', SOURCE/'RAW.mp4', SOURCE/'FRAMES.jsonl', MUSIC, Path(__file__).resolve(), FONT, BOLD] if p.is_relative_to(ROOT)}
    inputs.update(globals().get('EXTRA_INPUTS', {}))
    plan = dict(version=VERSION, seconds=elapsed, scenes=scenes, input_sha256=inputs,
                raw_seconds=raw_duration, capture_marks=marks,
                presentation=globals().get('PRESENTATION', '本輪Linux正常視窗錄影，原速剪輯；完整960×720畫布縮放至1280×960，無裁切或替換UI。'),
                cuts=globals().get('CUTS', '駐留命令後剪至到達；編成、換提督與出發127.5至197.5秒連續；執行回合後剪至抵達。'),
                audio='遊戲BGM_0F單首連續配樂，離線加入，無遊戲音效或配音；不是即時遊戲混音。',
                music_volume=6.0,
                rights='含原版及衍生音畫，只留本機，不公開上傳。',
                font=dict(file=str(FONT), sha256=sha(FONT), bold_sha256=sha(BOLD)), text_bounds=[])
    background = Image.new('RGBA', (1920, 1080), (7, 13, 23, 255))
    d = ImageDraw.Draw(background)
    rng = random.Random(146)  # 僅用於片頭星點美術，與遊戲亂數無關。
    for _ in range(210):
        x, y = rng.randrange(1920), rng.randrange(1080)
        d.ellipse((x, y, x+1, y+1), fill=(73, 89, 105, 255))
    d.rectangle((316, 58, 1603, 1021), outline=(151, 122, 68, 255), width=2)
    background.save(WORK/'BACKGROUND.png')

    def text(draw, xy, value, size, color, bounds, center=False, bold=False):
        font = ImageFont.truetype(str(BOLD if bold else FONT), size, index=3)
        box = draw.textbbox(xy, value, font=font, anchor='mm' if center else 'lt')
        assert box[0] >= bounds[0] and box[1] >= bounds[1] and box[2] <= bounds[2] and box[3] <= bounds[3], (value, box, bounds)
        draw.text(xy, value, font=font, fill=color, anchor='mm' if center else 'lt')
        plan['text_bounds'].append(dict(text=value, ink_box=list(box), safe_box=list(bounds)))

    for scene in scenes:
        i = scene['id']
        overlay = Image.new('RGBA', (1920, 1080))
        draw = ImageDraw.Draw(overlay)
        draw.rectangle((240 if wide else 0, 0, 1679 if wide else 1919, 59), fill=(7, 13, 23, 255))
        draw.rectangle((240 if wide else 0, 1023, 1679 if wide else 1919, 1079), fill=(7, 13, 23, 255))
        text(draw, (960, 29), scene['title'], 32, '#eddcb5', (252 if wide else 48, 0, 1668 if wide else 1872, 60), center=True, bold=True)
        text(draw, (960, 1050), scene['caption'], 28, '#e4eaf0', (252 if wide else 48, 1025, 1668 if wide else 1872, 1079), center=True)
        if scene['kind'] == 'game' and not wide:
            text(draw, (36, 110), '同盟軍', 28, '#a4bcd3', (24, 90, 288, 200))
            for j, line in enumerate(scene['lines']):
                text(draw, (36, 210+j*67), line, 32 if j else 48, '#e5cf9d' if not j else '#e4eaf0', (24, 170, 296, 460), bold=not j)
            for j, line in enumerate(['實際遊戲畫面', '原速操作剪輯', '完整畫布保留']):
                text(draw, (1640, 700+j*46), line, 24, '#9db0c0', (1625, 650, 1904, 870))
            text(draw, (1640, 935), 'v.1.5.2', 25, '#e5cf9d', (1625, 900, 1904, 1000))
        overlay_path = WORK/f'overlay-{i:02}.png'
        old_overlay = sha(overlay_path) if reuse else None
        overlay.save(overlay_path)
        if reuse: assert sha(overlay_path) == old_overlay
        clip = WORK/f'scene-{i:02}.mp4'
        if scene['kind'] == 'card':
            card = background.copy(); draw = ImageDraw.Draw(card)
            draw.rectangle((245, 170, 1675, 900), fill=(8, 18, 31, 248), outline='#bfa16b', width=3)
            text(draw, (960, 305), '銀河英雄傳說 III SP', 52, '#b7cbd9', (300, 240, 1620, 400), center=True)
            text(draw, (960, 445), scene['title'], 74, '#f0dab0', (300, 380, 1620, 540), center=True, bold=True)
            for j, line in enumerate(scene['lines']):
                text(draw, (560+j*400, 635), line, 42, '#d8e6ee', (300+j*400, 590, 820+j*400, 690), center=True)
            text(draw, (960, 800), VERSION, 30, '#a8bac9', (400, 750, 1520, 860), center=True)
            card = Image.alpha_composite(card, overlay).convert('RGB')
            card_path = WORK/f'card-{i:02}.png'; old_card = sha(card_path) if reuse else None
            card.save(card_path)
            if reuse: assert sha(card_path) == old_card
            inputs = ['-loop', '1', '-framerate', '60', '-i', WORK/f'card-{i:02}.png']
            filters = ['-vf', 'format=yuv420p']
        else:
            inputs = ['-loop', '1', '-framerate', '60', '-i', WORK/'BACKGROUND.png', '-ss', scene['start'], '-i', scene.get('source', SOURCE/'RAW.mp4'), '-loop', '1', '-i', WORK/f'overlay-{i:02}.png']
            scale, position = ('1920:1080', '0:0') if wide else ('1280:960', '320:60')
            filters = ['-filter_complex', f'[1:v]scale={scale}:flags=neighbor,setpts=PTS-STARTPTS[game];[0:v][game]overlay={position}:shortest=1[base];[base][2:v]overlay=0:0:shortest=1,format=yuv420p[out]', '-map', '[out]']
        if not reuse:
            run('ffmpeg', '-nostdin', '-v', 'error', '-filter_complex_threads', '1', *inputs, *filters, '-t', scene['duration'], '-an', '-r', '60', '-c:v', 'libx264', '-preset', 'veryfast', '-crf', '19', '-threads', '2', '-pix_fmt', 'yuv420p', '-video_track_timescale', '60000', clip)
        else: assert clip.is_file() and abs(float(probe(clip)['format']['duration'])-scene['duration']) < .02
        scene['clip_sha256'] = sha(clip)
        print('RENDER', i, scene['title'], flush=True)
    (WORK/'concat.txt').write_text(''.join("file '"+str(WORK/f'scene-{s["id"]:02}.mp4')+"'\n" for s in scenes))
    # 首次0.72候選為RMS 0.00953、peak 0.0552；提升至6仍保留約6dB峰值餘裕。
    fade = f'volume={plan["music_volume"]},afade=t=in:st=0:d=1,afade=t=out:st={elapsed-1}:d=1'
    # 不等長鏡頭的容器時長可能四捨五入至毫秒，concat後需重整為固定60fps。
    video = ['-vf', 'fps=60', '-r', '60', '-c:v', 'libx264', '-preset', 'veryfast', '-crf', '18', '-video_track_timescale', '60000'] if wide else ['-c:v', 'copy']
    if not args.audit_only:
        run('ffmpeg', '-nostdin', '-v', 'error', '-f', 'concat', '-safe', '0', '-i', WORK/'concat.txt', '-i', MUSIC, '-map', '0:v:0', '-map', '1:a:0', '-t', elapsed, *video, '-af', fade, '-ar', '44100', '-ac', '2', '-c:a', 'aac', '-b:a', '192k', '-threads', '2', '-movflags', '+faststart', MOVIE)
    reference = WORK/'REFERENCE-FINAL.wav'
    run('ffmpeg', '-nostdin', '-y', '-v', 'error', '-i', MUSIC, '-t', elapsed, '-af', fade, '-ar', '44100', '-ac', '2', '-c:a', 'pcm_s16le', reference)
    media = probe(MOVIE); v = next(s for s in media['streams'] if s['codec_type'] == 'video'); a = next(s for s in media['streams'] if s['codec_type'] == 'audio')
    assert (v['width'], v['height'], v['avg_frame_rate']) == (1920, 1080, '60/1')
    assert v['codec_name'] == 'h264' and a['codec_name'] == 'aac' and a['channels'] == 2
    assert abs(float(v['duration'])-elapsed) < .02 and abs(float(a['duration'])-elapsed) < .05
    def decode(path): return np.frombuffer(capture('ffmpeg', '-nostdin', '-v', 'error', '-threads', '2', '-i', path, '-vn', '-ar', '44100', '-ac', '2', '-f', 'f32le', '-'), dtype='<f4').reshape(-1, 2)
    got = decode(MOVIE); ref = decode(reference); n = min(len(got), len(ref)); got = got[:n]; ref = ref[:n]
    rms = float(np.sqrt(np.mean(got**2))); peak = float(np.max(np.abs(got)))
    snr = float(10*np.log10(np.mean(ref**2)/np.mean((got-ref)**2)))
    assert rms > .01 and peak < .999 and snr > 25, (rms, peak, snr)
    cut_checks = []
    for s in scenes[1:]:
        t = int(s['start_at']*44100); part = got[t-4410:t+4410]; expected = ref[t-4410:t+4410]
        cut_snr = float(10*np.log10(np.mean(expected**2)/np.mean((part-expected)**2)))
        assert cut_snr > 22, (s['id'], cut_snr)
        cut_checks.append(dict(seconds=s['start_at'], snr_db=cut_snr))
    detect = subprocess.run(['ffmpeg', '-nostdin', '-hide_banner', '-threads', '2', '-i', str(MOVIE), '-an', '-vf', 'blackdetect=d=0.25:pix_th=0.10:pic_th=0.98,freezedetect=n=0.002:d=2', '-f', 'null', '-'], capture_output=True, text=True, check=True, timeout=240)
    (WORK/'DETECT.log').write_text(detect.stderr)
    black = re.findall(r'black_start:[^\n]+', detect.stderr); assert not black, black
    freeze = re.findall(r'freeze_(?:start|end|duration): [\d.]+', detect.stderr)
    sheet = Image.new('RGB', (1280, math.ceil(len(scenes)/2)*390), '#101924'); draw = ImageDraw.Draw(sheet)
    for i, s in enumerate(scenes):
        at = s['start_at']+(s['duration']-1 if s['kind'] == 'game' else 2)
        frame = WORK/f'frame-{i:02}.png'
        run('ffmpeg', '-nostdin', '-y', '-v', 'error', '-threads', '2', '-ss', at, '-i', MOVIE, '-frames:v', '1', '-threads', '1', frame)
        image = Image.open(frame); image.thumbnail((640, 360)); x = (i%2)*640; y = (i//2)*390
        sheet.paste(image, (x, y)); draw.text((x+12, y+363), f'{i+1:02}  {s["title"]}', font=ImageFont.truetype(str(FONT), 20, index=3), fill='#e4eaf0')
    sheet.save(OUT/'CONTACT.png')
    write(OUT/'PLAN.json', plan)
    audit = dict(passed=True, version=VERSION, file=MOVIE.name, sha256=sha(MOVIE), bytes=MOVIE.stat().st_size,
                 seconds=elapsed, video=dict(codec=v['codec_name'], width=v['width'], height=v['height'], fps=v['avg_frame_rate'], frames=v['nb_frames']),
                 audio=dict(codec=a['codec_name'], sample_rate=a['sample_rate'], channels=a['channels'], rms=rms, peak=peak, snr_db=snr, cut_checks=cut_checks, listening='未作人耳驗收'),
                 black_segments=black, freeze_detections=freeze, freeze_classification='片頭／片尾為4秒靜態字卡。其餘為實際介面停留；沒有以截圖替換實錄或重複影格延長片段。',
                 text_safe_bounds=len(plan['text_bounds']), subtitle_clipping=False, full_canvas=True, source_audio_sha256=sha(MUSIC),
                 plan_sha256=sha(OUT/'PLAN.json'), contact_sha256=sha(OUT/'CONTACT.png'),
                 commands='tools/scripts/garrison_promo_movie.py；logh3-hd-preview:20261011-r1，FFmpeg及Pillow內建於既有映像',
                 ffmpeg_version=capture('ffmpeg', '-version').decode().splitlines()[0],
                 visual_review='待檢查CONTACT.png及完整抽樣畫面；數值通過不取代目視')
    write(OUT/'AUDIT.json', audit)
    (OUT/'SHA256SUMS.txt').write_text(''.join(sha(p)+'  '+p.name+'\n' for p in [MOVIE, OUT/'PLAN.json', OUT/'AUDIT.json', OUT/'CONTACT.png']))
    print(json.dumps(dict(passed=True, seconds=elapsed, sha256=audit['sha256'], audio_snr_db=snr, freeze=freeze), ensure_ascii=False), flush=True)


if __name__ == '__main__': main()
