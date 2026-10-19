"""143：正式v1.5提督編輯與實戰混剪；沿既有影片編碼、原曲連續混音與聲畫稽核。"""
import argparse
import json
import math
import os
from pathlib import Path
import shutil
import subprocess
from types import SimpleNamespace

import promo_v140_movie as movie

ROOT = Path('/src')
VERSION = 'v.1.5.0-20261019'
BATCH = ROOT/'workplace/admiral143'
OUT = BATCH/'movie'
PACKAGE = BATCH/'package'/f'LOGH3-{VERSION}-linux-amd64'
movie.VERSION = VERSION
movie.BATCH = BATCH
movie.args = SimpleNamespace(reuse_dir=None)
original_record = movie.record


def record(folder):
    if folder == BATCH/'editor-linux':
        value = movie.read(folder/'RESULT.json')
        assert value['passed'] and value['normal_entry']
        for video in value['videos']:
            assert movie.sha(folder/video['file']) == video['sha256']
        return dict(record_started_utc_ns=value['videos'][0]['record_started_utc_ns']), []
    return original_record(folder)


movie.record = record


def prepare():
    # 沿已驗S1鍵鼠路徑，依目前設定矩形更新選項操作。
    assert BATCH.is_dir()
    for faction in [0, 1]:
        plan = movie.read(ROOT/f'workplace/hero130/full-record-f{faction}.json')
        a = next(i for i, e in enumerate(plan) if e == dict(kind='click', x=544, y=125))
        b = next(i for i in range(a+1, len(plan)) if plan[i] == dict(kind='click', x=544, y=125))
        prefix = plan[:a]
        prefix += [dict(kind='wait_state', state='StateCommandPhase'),
                   dict(kind='capture', name='strategy-current')]
        prefix += [dict(kind='click', x=544, y=125), dict(kind='click', x=560, y=185),
                   dict(kind='wait_state', state='StateSettings'),
                   dict(kind='click', x=368, y=332), dict(kind='capture', name='ai-menu', wait=2),
                   dict(kind='click', x=368, y=284),
                   dict(kind='click', x=327, y=260), dict(kind='click', x=314, y=278)]
        prefix += [dict(kind='click', x=437, y=242) for _ in range(8)]
        prefix += [dict(kind='click', x=368, y=314),
                   dict(kind='capture', name='settings-recording'),
                   dict(kind='click', x=404, y=386),
                   dict(kind='wait_state', state='StateCommandPhase')]
        rest = plan[b:]
        # 既有正常實錄帝國艦隊戰從第4回合開始，同盟戰鬥機從第3回合開始。
        # 執行前固定終點；不換種子或重擲來追逐場次。
        ending = 'turn-04'
        end = next(i for i, e in enumerate(rest) if e.get('name') == ending)+1
        filename = f'game-f{faction}-final.json'
        movie.write(BATCH/filename, prefix+rest[:end])
    print('READY 兩陣營至第5回合正常輸入，無狀態／seed注入', flush=True)


def plan():
    scenes, inputs, anomalies = [], {}, []
    def scene(kind, duration, title, caption, **extra):
        scenes.append(dict(kind=kind, duration=math.ceil(duration*60)/60, title=title, caption=caption, **extra))
    def game(folder, start, duration, title, caption, **extra):
        scene('game', duration, title, caption, file=str(folder/'RAW.mp4'), start=start, **extra)
    old = ROOT/'workplace/hero130'
    scene('card', 4, '銀河英雄傳說 III SP｜v.1.5.0', '你的提督，你的艦隊，你的銀河', poster=str(old/'anime-sources/remaster-poster.png'))
    for start, title in [(238, '萊因哈特｜向銀河的頂點進軍'), (778, '楊威利｜魔術師的下一步')]:
        scene('anime', 3, title, '1988動畫來源，原生480p；本機推廣片', file=str(old/'ANIME-SOURCE-CLEAN.mp4'), start=start)
    folder = BATCH/'editor-linux'
    editor = movie.read(folder/'RESULT.json')
    assert editor['passed'] and editor['normal_entry']
    assert editor['binary_sha256'] == movie.sha(PACKAGE/'gin3')
    captures = {e['capture']: e for e in editor['events'] if 'capture' in e}
    for name, launch, duration, title, caption in [
        ('0-settings', 0, 3, '系統設定收進框內｜新增編輯提督', '沿用帝國與同盟樣式，設定中直接開啟提督能力編輯'),
        ('0-basic', 0, 3, '十三項提督能力，逐人調整', '統率、移動、光線砲、飛彈與戰鬥機；原版能力可分別編輯'),
        ('0-hidden', 0, 3, '內部能力，依已查用途顯示', '三項參數均可編輯；標籤為用途說明，原版欄名仍未知'),
        ('0-special', 0, 4, '全軍加成，依你的戰術設定', '萊因哈特與楊威利：移動預設+50%，全軍攻擊+30%'),
        ('0-kemp-fighter', 0, 4, '坎普｜戰鬥機攻擊預設+30%', '擔任提督或參謀均生效，只加強本人所在艦隊'),
        ('1-poplan-fighter', 1, 4, '波布蘭｜讓王牌展現實力', '本艦隊戰鬥機攻擊可編輯；全局保存，新遊戲與讀檔也套用')]:
        video = editor['videos'][launch]
        assert movie.sha(folder/video['file']) == video['sha256']
        # 顯示保留在拍攝時點後的正常操作，不把截圖偽裝成動態。
        start = max(0, (captures[name]['wall_ns']-video['record_started_utc_ns'])/1e9-.15)
        scene('game', duration, title, caption, file=str(folder/video['file']), start=start, editor=True)
    inputs[str(folder/'RESULT.json')] = movie.sha(folder/'RESULT.json')
    groups = []
    for faction in [0, 1]:
        folder = BATCH/f'gameplay-f{faction}-final'
        trace, rows = record(folder)
        assert trace['binary_sha256'] == movie.sha(PACKAGE/'gin3')
        origin = trace['record_started_utc_ns']
        captures = {e['name']: (e['wall_begin_ns']-origin)/1e9 for e in trace['events'] if e.get('kind') == 'capture'}
        game(folder, max(0, captures['entry']-.1), 3, '帝國的雄心' if faction == 0 else '同盟的信念', '新版完整版正常S1實錄；原版操作、陣營HD與強化AI')
        by_id = {}
        for row in rows:
            playback = row.get('playback') or {}
            if playback.get('kind') == 'fleet': by_id.setdefault(playback['id'], []).append(row)
        for identity, frames in by_id.items():
            if not frames[-1]['playback']['result']:
                anomalies.append(dict(scene=identity, source=str(folder), reason='來源未播完整，不選用'))
                continue
            applied, gaps = 0, []
            for frame in frames:
                pb = frame['playback']
                assert pb['applied'] == sum(p['tick'] <= pb['tick'] for p in pb['pictures'])
                if pb['applied'] > applied+1 and pb['interval'] > 0: gaps.append([applied, pb['applied']])
                applied = pb['applied']
            if gaps:
                anomalies.append(dict(scene=identity, source=str(folder), reason='來源漏格，不選用', gaps=gaps))
                continue
            start = math.floor((frames[0]['wall_ns']-origin)/1e9*60)/60
            end = math.ceil((frames[-1]['wall_ns']-origin)/1e9*60+2)/60
            groups.append((folder, start, end-start, identity, frames[-1]['playback']['weapon']))
        for name in ['TRACE.json', 'FRAMES.jsonl', 'FRAME-TRACE-CLOSE.json']:
            inputs[str(folder/name)] = movie.sha(folder/name)
    # 要塞砲規則與演出未變，沿前版已驗正常操作完整場次，明列來源。
    fortress = ROOT/'workplace/promo139/record-fortress'
    trace, rows = record(fortress)
    gun = [r for r in rows if r.get('fortress')]
    first = math.floor((gun[0]['wall_ns']-trace['record_started_utc_ns'])/1e9*60)/60
    last = math.ceil((gun[-1]['wall_ns']-trace['record_started_utc_ns'])/1e9*60+2)/60
    game(fortress, first, last-first, '伊謝爾倫要塞砲｜發射至戰損', '沿既有正式實錄，保留完整演出與音效', fortress_animation=True)
    for name in ['TRACE.json', 'FRAMES.jsonl', 'FRAME-TRACE-CLOSE.json']:
        inputs[str(fortress/name)] = movie.sha(fortress/name)
    # 優先挑各武器與兩陣營，再補足原速交火50秒。
    chosen, seen, seconds = [], set(), 0
    for group in sorted(groups, key=lambda g: (g[4], str(g[0]), g[1])):
        if group[4] not in seen:
            chosen.append(group); seen.add(group[4]); seconds += group[2]
    for group in groups:
        if seconds >= 50 and {g[0].name for g in chosen} == {'gameplay-f0-final', 'gameplay-f1-final'}: break
        if group not in chosen: chosen.append(group); seconds += group[2]
    assert seen == {0, 1, 2}, ('新來源不足三武器完整交火', seen)
    for folder, start, duration, identity, weapon in chosen:
        game(folder, start, duration, '光束交錯，飛彈破陣', '新版正式包原速交火；全片單首配樂連續播放', battle_animation=True, playback_id=identity, weapon=weapon)
    fresh_seconds = seconds
    previous = movie.read(ROOT/'dist-all/v.1.4.0-20261015/promo/PLAN.json')
    for old_scene in previous['scenes']:
        if seconds >= 50: break
        if not old_scene.get('battle_animation'): continue
        folder = Path(old_scene['file']).parent
        original_record(folder)
        game(folder, old_scene['start'], old_scene['duration'], '前版實戰回顧｜完整交火演出',
             '沿v.1.4.0已驗場次；新版能力編輯見前段，保留原速與完整結果',
             battle_animation=True, playback_id=old_scene['playback_id'], weapon=old_scene['weapon'], reused=True)
        seconds += old_scene['duration']
        for name in ['TRACE.json', 'FRAMES.jsonl', 'FRAME-TRACE-CLOSE.json']:
            inputs[str(folder/name)] = movie.sha(folder/name)
    assert seconds >= 50, '完整實戰篇幅不足'
    scene('card', 5, '你的戰術，決定銀河的明天', 'v.1.5.0｜Linux・Windows・macOS｜提督能力與特殊加成可編輯')
    result = dict(version=VERSION, background=str(ROOT/'workplace/hd127/promo-final/BACKGROUND.png'),
                  music=str(PACKAGE/'music/BGM_0F.ogg'), scenes=scenes, fresh_gameplay_inputs=inputs,
                  rights='含原版及衍生音畫、指定動畫，只留本機，不公開上傳',
                  unselected_source_anomalies=anomalies, playback_audit_scope='本片實際選用的完整艦隊與要塞場次；不宣稱整份來源無漏格',
                  fresh_fleet_combat_seconds=fresh_seconds, fleet_combat_seconds=seconds,
                  recording_scope='提督編輯及三武器艦隊交火含新版正式ZIP新錄；另明列沿用v1.4已驗完整交火及要塞砲；動畫及海報沿既有來源',
                  anime_source='https://www.bilibili.com/video/BV1G87szDEh7/')
    assert sum(s['duration'] for s in scenes) < 125, '單曲長度內保留完整片尾，不截短動畫'
    movie.write(BATCH/'PROMO-PLAN.json', result)
    print('READY', len(scenes), '段', sum(s['duration'] for s in scenes), '秒，新實戰', fresh_seconds, '秒，完整交火合計', seconds, flush=True)


def source_audit():
    data = movie.read(OUT/'PLAN.json')
    groups = {}
    for scene in data['scenes']:
        if scene.get('battle_animation'): groups.setdefault(Path(scene['file']).parent, []).append(scene['playback_id'])
    audits = []
    previous = movie.read(ROOT/'dist-all/v.1.4.0-20261015/promo/SOURCE-PLAYBACK.json')['recordings']
    for folder, ids in groups.items():
        if not folder.is_relative_to(BATCH):
            old = next(row for row in previous if row['source'] == str(folder))
            assert movie.sha(folder/'PLAYBACK-AUDIT.json') == old['sha256']
            value = old['audit']
            assert value['passed'] and not value['skipped_frames'] and set(ids) <= {s['id'] for s in value['scenes']}
            audits.append(dict(old, reused=True, selected_scene_ids=ids))
            continue
        cached = folder/'PLAYBACK-AUDIT.json'
        if cached.exists():
            value = movie.read(cached)
            closed = movie.read(folder/'FRAME-TRACE-CLOSE.json')
            if (value['passed'] and not value['skipped_frames'] and value['frame_trace_sha256'] == closed['frame_trace_sha256']
                    and value['media']['raw_sha256'] == movie.sha(folder/'RAW.mp4')
                    and set(ids) == set(value['scene_ids_requested'])):
                audits.append(dict(source=str(folder), sha256=movie.sha(cached), audit=value, cached_same_inputs=True))
                continue
        log_path = OUT/(folder.name+'-source-audit.log')
        attempt = 2
        while log_path.exists():
            log_path = OUT/(folder.name+f'-source-audit-r{attempt}.log')
            attempt += 1
        with log_path.open('x') as log:
            movie.run('python3', ROOT/'tools/scripts/battle_playback_audit.py', folder, '--kinds', 'fleet', '--scene-ids', *ids, stdout=log, stderr=subprocess.STDOUT)
        value = movie.read(folder/'PLAYBACK-AUDIT.json')
        assert value['passed'] and not value['skipped_frames']
        audits.append(dict(source=str(folder), sha256=movie.sha(folder/'PLAYBACK-AUDIT.json'), audit=value))
    fortress = ROOT/'workplace/promo139/record-fortress/PLAYBACK-AUDIT.json'
    value = movie.read(fortress)
    assert value['passed'] and not value['skipped_frames']
    audits.append(dict(source=str(fortress.parent), sha256=movie.sha(fortress), reused=True, audit=value))
    movie.write(OUT/'SOURCE-PLAYBACK.json', dict(passed=True, scope=data['playback_audit_scope'], recordings=audits,
                audit_generator_sha256=movie.sha(ROOT/'tools/scripts/battle_playback_audit.py'),
                unselected_source_anomalies=data['unselected_source_anomalies']))


def audit():
    movie.audit(OUT)
    movie.write(OUT/'FFPROBE.json', movie.read_probe(OUT/movie.MOVIE))
    data = movie.read(OUT/'CONTINUITY-AUDIT.json')
    data.update(all_gameplay_new=False, new_sources='新版正式ZIP提督編輯與三武器艦隊交火', reused_sources='既有正式完整交火、要塞砲、1988動畫及海報')
    movie.write(OUT/'CONTINUITY-AUDIT.json', data)
    (OUT/'RIGHTS.txt').write_text(movie.read(OUT/'PLAN.json')['rights']+'\n'+movie.read(OUT/'PLAN.json')['recording_scope']+'\n')


def publish(visual_reviewed):
    data = movie.read(OUT/'CONTINUITY-AUDIT.json')
    assert data['passed'] and visual_reviewed
    target = ROOT/'dist-all'/VERSION/'promo'
    assert target.is_dir() and not list(target.iterdir())
    data['visual_review'] = 'CONTACT及原尺寸編輯器、交火、要塞與片尾已目視'
    movie.write(OUT/'CONTINUITY-AUDIT.json', data)
    names = [movie.MOVIE, 'PLAN.json', 'AUDIT.json', 'FFPROBE.json', 'TEXT-BOUNDS.json', 'SOURCE-FRAMES.json', 'SOURCE-PLAYBACK.json',
             'CONTINUITY-AUDIT.json', 'SOUNDTRACK.json', 'FREEZE-REVIEW.json', 'AUDIO-AUDIT.log', 'VIDEO-AUDIT.log',
             'CONTACT.jpg', 'RIGHTS.txt', 'REVIEW-settings.png', 'REVIEW-fortress.png', 'REVIEW-ending.png']
    files = []
    for name in names:
        destination = target/(f'LOGH3-{VERSION}-promo.mp4' if name == movie.MOVIE else name)
        shutil.copyfile(OUT/name, destination)
        files.append(dict(path=str(destination.relative_to(ROOT)), bytes=destination.stat().st_size, sha256=movie.sha(destination)))
    movie.write(target/'SHA256SUMS.json', dict(version=VERSION, files=files))
    print('DELIVERED', target, flush=True)


if __name__ == '__main__':
    assert os.getuid() == 1000
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode', choices=['prepare', 'plan', 'render', 'source_audit', 'audit', 'publish'])
    parser.add_argument('--visual-reviewed', action='store_true')
    args = parser.parse_args()
    if args.mode == 'render': movie.render(OUT)
    elif args.mode == 'publish': publish(args.visual_reviewed)
    else: globals()[args.mode]()
