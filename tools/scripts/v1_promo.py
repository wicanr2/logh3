"""v1推廣片：新封包真實畫面、固定音源及完整影音檢查。音畫Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess

VERSION = 'v.1.0.0-20261010'
PROJECT = Path('/src')
ROOT = Path('/work') / ('promo-' + VERSION)
FONT = Path('/fonts/NotoSansCJK-Regular.ttc')
BOLD = Path('/fonts/NotoSansCJK-Bold.ttc')
os.environ.setdefault('MAGICK_THREAD_LIMIT', '1')
BG, PANEL, LINE, CYAN, WHITE, DIM, GOLD = '#071820', '#102B35', '#365566', '#61D3F3', '#F3F3F3', '#B2C2CF', '#D3A261'


def run(*args, **kwargs):
    return subprocess.run([str(a) for a in args], check=True, **kwargs)


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def write(p, value):
    p.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def prepare():
    ROOT.mkdir()
    (ROOT / 'inputs').mkdir()
    full = PROJECT / 'dist-all' / VERSION
    gui = full / 'smoke/gui-0'
    captures = PROJECT / 'workplace' / ('package92-smoke-' + VERSION)
    sources = {'account': gui / '01-account-x11.png', 'budget': gui / '03-budget-changed-x11.png',
               'build': gui / '09-build-row1-x11.png', 'command': gui / '12-command-x11.png',
               'portrait': gui / '17-new-portrait-x11.png',
               'original': full / 'smoke/gui-original/17-new-portrait-x11.png',
               'alliance': full / 'smoke/gui-1/17-new-portrait-x11.png',
               'battle': captures / 'battle.png', 'fleet': captures / 'fleet.png',
               'ai-original': captures / 'ai-original.png', 'ai-enhanced': captures / 'ai-enhanced.png',
               'load': captures / 'load.png'}
    entries = []
    for name, path in sources.items():
        destination = ROOT / 'inputs' / (name + '.png')
        shutil.copyfile(path, destination)
        entries.append({'name': name, 'path': str(path), 'sha256': sha(path),
                        'kind': '正常X11玩家路徑' if path.is_relative_to(full / 'smoke') else '新封包正式renderer，公開CLI腳本抽驗'})
    music = full / 'full' / ('LOGH3-' + VERSION + '-linux-amd64/music/BGM_0F.ogg')
    shutil.copyfile(music, ROOT / 'inputs/BGM_0F.ogg')
    story = [
        {'id': 0, 'seconds': 6, 'layout': 'title', 'title': '銀河英雄傳說 III SP', 'subtitle': '回到指揮席，展開自己的銀河戰役'},
        {'id': 1, 'seconds': 7, 'layout': 'hero', 'source': 'command', 'title': '從一張星圖開始', 'subtitle': '選擇陣營、部署艦隊，推進戰略回合'},
        {'id': 2, 'seconds': 7, 'layout': 'briefing', 'source': 'budget', 'title': '每筆預算，都影響下一步', 'subtitle': '財政變更、預算箭頭與確認，從正常入口操作'},
        {'id': 3, 'seconds': 6, 'layout': 'briefing', 'source': 'build', 'title': '為下一場戰役造艦', 'subtitle': '調整艦種比例，完成次月份建造計畫'},
        {'id': 4, 'seconds': 8, 'layout': 'compare', 'title': '熟悉的面孔，更清楚的輪廓', 'subtitle': '82 張人物原畫風修復；左為重製原圖模式，右為 HD'},
        {'id': 5, 'seconds': 6, 'layout': 'dossier', 'source': 'alliance', 'title': '選擇你的提督與副官', 'subtitle': '人物、能力與艦隊編組，接到同一條遊戲路徑'},
        {'id': 6, 'seconds': 7, 'layout': 'hero', 'source': 'battle', 'title': '把部署帶進戰術戰場', 'subtitle': '下令、移動與交戰，原版與強化 AI 均已啟用'},
        {'id': 7, 'seconds': 7, 'layout': 'ai', 'title': '用兩種 AI，展開不同戰役', 'subtitle': '原版 AI ／強化 AI；兩種模式共用存檔'},
        {'id': 8, 'seconds': 6, 'layout': 'platform', 'title': '三種平台，同一個指揮席', 'subtitle': 'Windows・Linux・macOS Intel／Apple Silicon'},
        {'id': 9, 'seconds': 6, 'layout': 'ending', 'title': '下一場戰役，交給你', 'subtitle': 'v1.0.0 引擎包與本機完整版已備妥'},
    ]
    write(ROOT / 'PLAN.json', {'version': VERSION, 'sources': entries, 'storyboard': story,
          'duration': sum(s['seconds'] for s in story), 'audio': {'source': str(music), 'sha256': sha(music), 'origin': '已驗證原版MDS轉出的正式OGG，沿固定音源來源收據'},
          'fonts': [{'path': str(p), 'sha256': sha(p)} for p in (FONT, BOLD)],
          'rights': '含原版衍生圖像及配樂，只留本機；不隨引擎包散布',
          'comparison': '重製程式原圖／HD顯示模式展示，非新原版same-state證據',
          'differences_from_previous': ['新正式封包的正常操作畫面', '82人物HD前後及兩陣營資料展示', '以功能發行與素材範圍取代測試版世界觀敘事'],
          'intentional_static': '影片以實際畫面的靜態剪輯及兩段有界平移呈現，不宣稱連續實況錄影'})


def canvas(path):
    run('convert', '-size', '1920x1080', 'xc:' + BG, '-stroke', LINE, '-strokewidth', '2',
        '-fill', 'none', '-draw', 'line 80,95 1840,95 line 80,965 1840,965 circle 1510,520 1790,520 line 1510,175 1510,920', path)


def text(path, x, y, size, colour, content, bold=False, width=1740, height=160):
    font = BOLD if bold else FONT
    result = subprocess.check_output(['convert', '-font', str(font), '-pointsize', str(size), 'label:' + content, '-format', '%w %h', 'info:'], text=True)
    w, h = map(int, result.split())
    assert w <= width and h <= height and x + w <= 1840 and y + h <= 1080, (content, w, h)
    run('convert', path, '-font', font, '-pointsize', size, '-fill', colour, '-gravity', 'northwest', '-annotate', f'+{x}+{y}', content, path)
    TEXT_BOUNDS.append({'file': path.name, 'text': content, 'bounds': [x, y, w, h]})


def image(path, name, x, y, width, height):
    run('convert', path, '(', ROOT / 'inputs' / (name + '.png'), '-filter', 'point', '-resize', f'{width}x{height}', ')',
        '-gravity', 'northwest', '-geometry', f'+{x}+{y}', '-composite', path)


TEXT_BOUNDS = []


def cards():
    plan = json.loads((ROOT / 'PLAN.json').read_text())
    output = ROOT / 'cards'
    output.mkdir()
    for scene in plan['storyboard']:
        path = output / f'{scene["id"]:02d}.png'
        canvas(path)
        text(path, 90, 35, 25, DIM, '銀河英雄傳說 III SP  /  ' + VERSION, width=1650)
        layout = scene['layout']
        if layout in ('title', 'ending'):
            text(path, 110, 255, 88, WHITE, scene['title'], True)
            text(path, 115, 420, 40, CYAN, scene['subtitle'])
            text(path, 115, 610, 32, GOLD, '82 張人物修復   ／   兩種 AI   ／   三平台')
            text(path, 115, 720, 30, DIM, '引擎包：玩家自備原版資料')
            text(path, 115, 785, 30, DIM, '完整版及本片含原版衍生素材，只留本機')
            text(path, 90, 1000, 26, DIM, '完整登陸分段、真機與人耳驗收仍待完成；不宣稱全遊戲原版一致')
        elif layout in ('compare', 'ai'):
            text(path, 90, 125, 54, WHITE, scene['title'], True)
            left, right = ('original', 'portrait') if layout == 'compare' else ('ai-original', 'ai-enhanced')
            image(path, left, 90, 280, 820, 615)
            image(path, right, 1010, 280, 820, 615)
            text(path, 90, 225, 28, CYAN, '原圖模式' if layout == 'compare' else '原版 AI')
            text(path, 1010, 225, 28, CYAN, 'HD 模式' if layout == 'compare' else '強化 AI')
            text(path, 90, 1000, 32, WHITE, scene['subtitle'])
        elif layout == 'platform':
            text(path, 90, 160, 60, WHITE, scene['title'], True)
            for i, (label, info) in enumerate([('Windows', 'amd64'), ('Linux', 'amd64'), ('macOS', 'Intel / Apple Silicon')]):
                x = 100 + i * 600
                run('convert', path, '-fill', PANEL, '-stroke', LINE, '-draw', f'rectangle {x},350 {x+530},775', path)
                text(path, x+35, 440, 52, CYAN, label, True, width=450)
                text(path, x+35, 550, 28, WHITE, info, width=450)
                text(path, x+35, 640, 28, DIM, '引擎包 / 本機完整版', width=450)
            text(path, 90, 1000, 30, WHITE, scene['subtitle'])
        else:
            text(path, 90, 125, 54, WHITE, scene['title'], True)
            if layout == 'hero':
                image(path, scene['source'], 430, 225, 1066, 700)
            else:
                image(path, scene['source'], 90, 240, 920, 690)
                text(path, 1100, 335, 42, CYAN, '從財政與造艦開始下令', True, width=740)
                text(path, 1100, 455, 32, WHITE, '財政・造艦・人事・戰術', width=740)
                text(path, 1100, 565, 30, DIM, '原畫風人物修復與 4:3 畫布', width=740)
                text(path, 1100, 660, 30, DIM, '規則與存檔沿現有遊戲', width=740)
            text(path, 90, 1000, 32, WHITE, scene['subtitle'])
    write(ROOT / 'TEXT-BOUNDS.json', {'text_fields': TEXT_BOUNDS, 'clipped': 0})


def encode():
    plan = json.loads((ROOT / 'PLAN.json').read_text())
    output = ROOT / 'encode'
    output.mkdir()
    listing = []
    for scene in plan['storyboard']:
        frames = scene['seconds'] * 30
        motion = scene['id'] in (0, 9)
        filters = f"zoompan=z='1.015':x='6+on/30':y='4':d={frames}:s=1920x1080:fps=30," if motion else ''
        filters += 'format=yuv420p'
        target = output / f'{scene["id"]:02d}.mp4'
        args = ['ffmpeg', '-nostdin', '-loglevel', 'error', '-filter_threads', '1']
        if not motion:
            args += ['-loop', '1', '-framerate', '30']
        args += ['-i', str(ROOT / 'cards' / f'{scene["id"]:02d}.png'), '-vf', filters,
                 '-frames:v', str(frames), '-an', '-threads', '2', '-c:v', 'libx264', '-preset', 'veryfast', '-crf', '20', str(target)]
        run(*args)
        listing.append("file '" + str(target) + "'")
        print('影片片段', scene['id'], frames, flush=True)
    (output / 'concat.txt').write_text('\n'.join(listing) + '\n')
    mux()


def mux():
    plan = json.loads((ROOT / 'PLAN.json').read_text())
    plan['audio']['gain_db'] = 14
    plan['audio']['processing'] = '同一遊戲音源固定增益、循環裁齊及首尾淡化'
    write(ROOT / 'PLAN.json', plan)
    duration = plan['duration']
    target = ROOT / ('LOGH3-' + VERSION + '-promo.mp4')
    run('ffmpeg', '-nostdin', '-loglevel', 'error', '-f', 'concat', '-safe', '0', '-i', ROOT / 'encode/concat.txt',
        '-stream_loop', '-1', '-i', ROOT / 'inputs/BGM_0F.ogg', '-map', '0:v', '-map', '1:a', '-c:v', 'copy',
        '-af', f'atrim=0:{duration},volume=14dB,afade=t=in:st=0:d=1.5,afade=t=out:st={duration-3}:d=3',
        '-t', duration, '-c:a', 'aac', '-b:a', '192k', '-threads', '2', '-movflags', '+faststart', target)


def audit():
    plan = json.loads((ROOT / 'PLAN.json').read_text())
    video = ROOT / ('LOGH3-' + VERSION + '-promo.mp4')
    probe = json.loads(subprocess.check_output(['ffprobe', '-v', 'error', '-show_streams', '-show_format', '-of', 'json', str(video)]))
    v = next(s for s in probe['streams'] if s['codec_type'] == 'video')
    a = next(s for s in probe['streams'] if s['codec_type'] == 'audio')
    assert (v['width'], v['height'], v['r_frame_rate']) == (1920, 1080, '30/1')
    assert int(v['nb_frames']) == plan['duration'] * 30
    assert abs(float(v['duration']) - plan['duration']) < 0.05 and abs(float(a['duration']) - plan['duration']) < 0.1
    write(ROOT / 'FFPROBE.json', probe)
    with (ROOT / 'AUDIO-AUDIT.txt').open('w') as log:
        run('ffmpeg', '-nostdin', '-hide_banner', '-i', video, '-vn', '-af', 'volumedetect,silencedetect=n=-55dB:d=2', '-f', 'null', '-', stderr=log, stdout=subprocess.DEVNULL)
    audio = (ROOT / 'AUDIO-AUDIT.txt').read_text()
    assert 'mean_volume: -inf' not in audio and 'max_volume: -inf' not in audio
    assert 'silence_start:' not in audio
    with (ROOT / 'VIDEO-AUDIT.txt').open('w') as log:
        run('ffmpeg', '-nostdin', '-hide_banner', '-i', video, '-an', '-vf', 'blackdetect=d=0.25:pix_th=0.05,freezedetect=n=-60dB:d=2', '-threads', '2', '-f', 'null', '-', stderr=log, stdout=subprocess.DEVNULL)
    assert 'black_start:' not in (ROOT / 'VIDEO-AUDIT.txt').read_text()
    motion = []
    for index in (0, 9):
        segment = ROOT / 'encode' / f'{index:02d}.mp4'
        first = subprocess.check_output(['ffmpeg', '-nostdin', '-loglevel', 'error', '-i', str(segment), '-frames:v', '1', '-pix_fmt', 'rgb24', '-f', 'rawvideo', '-'])
        last = subprocess.check_output(['ffmpeg', '-nostdin', '-loglevel', 'error', '-sseof', '-0.05', '-i', str(segment), '-frames:v', '1', '-pix_fmt', 'rgb24', '-f', 'rawvideo', '-'])
        assert len(first) == len(last) == 1920 * 1080 * 3 and first != last
        motion.append({'scene': index, 'first_sha256': hashlib.sha256(first).hexdigest(), 'last_sha256': hashlib.sha256(last).hexdigest(), 'different': True})
    samples = []
    elapsed = 0
    for scene in plan['storyboard']:
        target = ROOT / f'sample-{scene["id"]:02d}.png'
        run('ffmpeg', '-nostdin', '-y', '-loglevel', 'error', '-ss', elapsed + scene['seconds'] / 2, '-i', video, '-frames:v', '1', '-threads', '2', target)
        thumbnail = ROOT / f'thumb-{scene["id"]:02d}.png'
        run('convert', '-limit', 'thread', '1', target, '-thumbnail', '480x270', thumbnail)
        samples.append(thumbnail)
        elapsed += scene['seconds']
    # 映像沒有預設字型；montage即使未寫標籤仍需明示可用字型。
    run('montage', '-font', FONT, '-limit', 'thread', '1', *samples, '-tile', '2x5', '-geometry', '+8+8', '-background', BG, ROOT / 'CONTACT.png')
    write(ROOT / 'AUDIT.json', {'version': VERSION, 'file': video.name, 'bytes': video.stat().st_size, 'sha256': sha(video),
          'seconds': plan['duration'], 'frames': int(v['nb_frames']), 'canvas': [1920, 1080], 'audio_present': True,
          'no_black_segment': True, 'subtitle_clipped': 0,
          'motion_checks': motion, 'no_silence_over_two_seconds': True,
          'freeze_policy': '第1至8段為分鏡明示的靜態遊戲畫面，凍結檢測保留原始結果；第0及9段有界平移，另驗首尾幀不同。',
          'rights': plan['rights'], 'comparison': plan['comparison']})
    print('PASS 影片影音軌、總幀數、非靜音、黑段及字幕幾何', flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode', choices=['prepare', 'cards', 'encode', 'mux', 'audit'])
    action = parser.parse_args().mode
    assert os.getuid() != 0 and FONT.is_file() and BOLD.is_file()
    globals()[action]()
