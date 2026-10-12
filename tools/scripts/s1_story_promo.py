"""雙旗、歷史抉擇與S1正常遊玩實錄。於影片Docker執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import v1_promo as layout

VERSION = 'v.1.0.1-20261011'
ROOT = Path('/work/hd123-s1-20261011')
OUT = ROOT/'promo'
FONT = '/fonts/NotoSansCJK-Regular.ttc'
SCENES = None
os.environ['MAGICK_THREAD_LIMIT'] = '1'

def run(*args, **kwargs):
    return subprocess.run([str(x) for x in args], check=True, **kwargs)

def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()

def write(p, obj):
    p.write_text(json.dumps(obj, ensure_ascii=False, indent=2)+'\n')

def prepare():
    assert os.getuid() == 1000 and not OUT.exists()
    OUT.mkdir()
    scenes = [
        dict(id=0, seconds=6, title='如果，這一戰由你指揮？', subtitle='兩面旗幟，一片銀河。', source=None),
        dict(id=1, seconds=11, title='三個歷史轉折。從這一戰開始。', subtitle='S1 亞姆力札會戰　／　S2 諸神的黃昏　／　S3 巴米利恩會戰', source='record-f', start=3.3),
        dict(id=2, seconds=6, title='宇宙曆796年。亞姆力札。', subtitle='艦隊奉命出征。命運，正向這片星域集結。', source='record-f', start=30),
        dict(id=3, seconds=8, title='這一次，你站上指揮席。', subtitle='選擇帝國，或為同盟爭取轉機。', source='record-f', start=60),
        dict(id=4, seconds=9, title='集結、進逼。下一道命令，由你下達。', subtitle='正常遊玩實錄：選取艦隊 → 戰鬥命令 → 目的地與朝向 → 實行', source='record-f', start=68),
        dict(id=5, seconds=10, title='艦隊正在接近。勝負，尚未決定。', subtitle='從星圖部署，走進真正的戰場。', source='record-d', start=64),
        dict(id=6, seconds=12, title='當炮火亮起，歷史不再只是回憶。', subtitle='S1戰鬥實錄：艦隊開火、命中與爆炸', source='record-d', start=98),
        dict(id=7, seconds=22, title='延續攻勢，或逆轉戰局？', subtitle='亞姆力札會戰實錄。把下一個可能，交給你的指揮。', source='record-d', start=262),
        dict(id=8, seconds=6, title='下一頁歷史，由你寫下。', subtitle='銀河英雄傳說 III SP', source=None),
    ]
    if SCENES:
        scenes = json.loads(Path(SCENES).read_text())
        assert [s['id'] for s in scenes] == list(range(9))
    sources = json.loads((ROOT/'FLAG-SOURCES.json').read_text())
    for name, crop in [('empire','460x314+20+90'), ('alliance','414x318+40+88')]:
        run('convert', ROOT/(name+'-flag.jpg'), '-crop', crop, '+repage', OUT/(name+'-flag.png'))
        run('convert', OUT/(name+'-flag.png'), '-resize', '960x1080^', '-gravity', 'center', '-extent', '960x1080', OUT/(name+'-wall.png'))
    run('convert', OUT/'empire-wall.png', OUT/'alliance-wall.png', '+append',
        '-fill', '#030811a8', '-draw', 'rectangle 0,0 1919,1079', OUT/'BACKGROUND.png')
    timestamp = 0
    for scene in scenes:
        path = OUT/f'card-{scene["id"]:02}.png'
        shutil.copyfile(OUT/'BACKGROUND.png', path)
        if scene['source']:
            # 兩側小旗保留完整徽章，中央1280x960實錄不裁切遊戲內容。
            for faction, x in [('empire',16),('alliance',1616)]:
                run('convert', path, '(', OUT/(faction+'-flag.png'), '-resize', '288x220', ')',
                    '-gravity','northwest','-geometry',f'+{x}+400','-composite',path)
            layout.text(path, 46, 340, 28, '#dfc58d', '銀河帝國', width=270)
            layout.text(path, 1630, 340, 28, '#83dcf4', '自由行星同盟', width=270)
            layout.text(path, 330, 10, 31, '#f3f1e8', scene['title'], True, width=1280, height=50)
            layout.text(path, 330, 1023, 27, '#f3f1e8', scene['subtitle'], width=1510, height=45)
            trace = json.loads((ROOT/scene['source']/'TRACE.json').read_text())
            scene['raw_sha256'] = trace['raw_sha256']
            scene['normal_x11_input'] = trace['normal_x11_input']
            scene['binary_sha256'] = trace['binary_sha256']
        else:
            layout.text(path, 120, 85, 30, '#dfc58d', '銀河帝國', width=760)
            layout.text(path, 1250, 85, 30, '#83dcf4', '自由行星同盟', width=590)
            layout.text(path, 170, 230, 76, '#f3f1e8', scene['title'], True, width=1670)
            layout.text(path, 175, 355, 41, '#dfc58d', scene['subtitle'], width=1665)
            layout.text(path, 175, 860, 38, '#f3f1e8', '銀河英雄傳說 III SP', True, width=1650)
            layout.text(path, 175, 935, 25, '#bcc7d5', VERSION+'　Windows / Linux / macOS', width=1650)
        scene['timeline_start'] = timestamp
        scene['frames'] = scene['seconds']*30
        timestamp += scene['seconds']
    assert timestamp == 90
    write(OUT/'PLAN.json', dict(version=VERSION, seconds=timestamp, fps=30, storyboard=scenes,
        visual='左帝國、右同盟旗幟；中央保留完整4:3遊戲畫面。',
        editing='正常X11實錄剪接；交火取自同一S1會戰較後回合，不宣稱90秒內連續完成。',
        audio='畫面實錄，遊戲配樂後配；不宣稱現場音效同步。',
        flags=sources, official_reference='https://ginei.club/infoDetail.php?iKey=1224',
        story_reference='https://ginei.club/infoDetail.php?iKey=384',
        intentional_static=[0,8], reading_pause='第1段保留實際讀檔清單停留，供閱讀三劇本。',
        changed_from_previous=['雙旗左右背景','歷史抉擇故事主軸','三劇本與S1公開讀檔實錄','艦隊下令、移動與交火動畫實錄'],
        rights='旗幟、遊戲音畫及衍生素材僅供本機檢視，沒有取得公開散布權。',
        hd_scope=('原配置A介面、HD字型、地圖及操作符號；45組AI艦型基底、原幀表面與光效。片頭沿原鏡頭重建表面；不宣稱動畫每幀AI新繪。' if SCENES else
                  '影片使用現行HD呈現；除人物外仍為邊緣細化，不聲稱動畫逐幀HD重製完成。')))
    write(OUT/'TEXT-BOUNDS.json', layout.TEXT_BOUNDS)
    (OUT/'RIGHTS.txt').write_text('本片含原版遊戲及衍生音畫、銀河英雄傳說旗幟商品照片。僅留本機，未取得公開散布權。旗幟照片只裁去背景及邊界，中央徽章保留。來源、雜湊與配樂見PLAN.json。\n')
    print('雙旗故事分鏡完成，90秒／7段實錄', flush=True)

def encode():
    plan=json.loads((OUT/'PLAN.json').read_text())
    for scene in plan['storyboard']:
        card=OUT/f'card-{scene["id"]:02}.png'
        target=OUT/f'scene-{scene["id"]:02}.mp4'
        if scene.get('pieces'):
            assert sum(p['seconds'] for p in scene['pieces']) == scene['seconds']
            pieces=[]
            for n,piece in enumerate(scene['pieces']):
                part=OUT/f'piece-{scene["id"]:02}-{n:02}.mp4'
                run('ffmpeg','-nostdin','-y','-v','error','-loop','1','-framerate','30','-i',card,
                    '-ss',str(piece['start']),'-i',ROOT/scene['source']/'RAW.mp4',
                    '-filter_complex','[1:v]scale=1280:960:flags=lanczos,fps=30,setpts=PTS-STARTPTS[game];[0:v][game]overlay=320:60:shortest=1,format=yuv420p[v]',
                    '-map','[v]','-an','-frames:v',str(piece['seconds']*30),'-c:v','libx264','-crf','19','-preset','veryfast','-threads','2',part)
                pieces.append(part)
            listing=OUT/f'pieces-{scene["id"]:02}.txt'
            listing.write_text(''.join(f"file '{p.name}'\n" for p in pieces))
            run('ffmpeg','-nostdin','-y','-v','error','-f','concat','-safe','0','-i',listing,'-c','copy',target)
            print('完成故事段落',scene['id'],flush=True)
            continue
        args=['ffmpeg','-nostdin','-y','-v','error','-loop','1','-framerate','30','-i',card]
        if scene['source']:
            args+=['-ss',str(scene['start']),'-i',ROOT/scene['source']/'RAW.mp4',
                   '-filter_complex','[1:v]scale=1280:960:flags=lanczos,fps=30,setpts=PTS-STARTPTS[game];[0:v][game]overlay=320:60:shortest=1,format=yuv420p[v]', '-map','[v]']
        else:
            args+=['-vf','format=yuv420p']
        args+=['-an','-frames:v',str(scene['frames']),'-c:v','libx264','-crf','19','-preset','veryfast','-threads','2',target]
        run(*args)
        print('完成故事段落',scene['id'],flush=True)
    listing=OUT/'concat.txt'
    listing.write_text(''.join(f"file 'scene-{s['id']:02}.mp4'\n" for s in plan['storyboard']))
    run('ffmpeg','-nostdin','-y','-v','error','-f','concat','-safe','0','-i',listing,'-c','copy',OUT/'VIDEO.mp4')
    music=Path('/src/dist-all')/VERSION/'full'/f'LOGH3-{VERSION}-linux-amd64/music/BGM_0F.ogg'
    plan['music']={'source':str(music),'sha256':sha(music),'gain_db':14}
    write(OUT/'PLAN.json',plan)
    run('ffmpeg','-nostdin','-y','-v','error','-i',OUT/'VIDEO.mp4','-stream_loop','-1','-i',music,
        '-map','0:v','-map','1:a','-c:v','copy','-af','volume=14dB,afade=t=in:d=1,afade=t=out:st=88:d=2',
        '-c:a','aac','-b:a','192k','-t','90','-movflags','+faststart',OUT/f'LOGH3-{VERSION}-promo.mp4')

def audit():
    movie=OUT/f'LOGH3-{VERSION}-promo.mp4'
    plan=json.loads((OUT/'PLAN.json').read_text())
    probe=json.loads(subprocess.check_output(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(movie)]))
    write(OUT/'FFPROBE.json',probe)
    video=next(s for s in probe['streams'] if s['codec_type']=='video')
    audio=next(s for s in probe['streams'] if s['codec_type']=='audio')
    assert (video['width'],video['height'],video['r_frame_rate'],int(video['nb_frames']))==(1920,1080,'30/1',2700)
    assert abs(float(probe['format']['duration'])-90)<.1 and abs(float(audio['duration'])-90)<.1
    with (OUT/'AUDIO-AUDIT.txt').open('w') as log:
        run('ffmpeg','-nostdin','-i',movie,'-af','volumedetect,silencedetect=noise=-50dB:d=2','-vn','-f','null','-',stdout=log,stderr=subprocess.STDOUT)
    with (OUT/'VIDEO-AUDIT.txt').open('w') as log:
        run('ffmpeg','-nostdin','-i',movie,'-vf','blackdetect=d=0.25:pix_th=0.1,freezedetect=n=-50dB:d=2','-an','-f','null','-',stdout=log,stderr=subprocess.STDOUT)
    volume=(OUT/'AUDIO-AUDIT.txt').read_text()
    assert 'mean_volume: -inf' not in volume and 'max_volume: -inf' not in volume and 'silence_start:' not in volume
    assert 'black_start:' not in (OUT/'VIDEO-AUDIT.txt').read_text()
    samples=[]
    for s in plan['storyboard']:
        target=OUT/f'sample-{s["id"]:02}.png'
        run('ffmpeg','-nostdin','-y','-v','error','-ss',str(s['timeline_start']+s['seconds']/2),'-i',movie,'-frames:v','1',target)
        samples.append(target)
    run('montage','-font',FONT,*samples,'-geometry','640x360+8+8','-tile','3x3','-background','#07111e',OUT/'CONTACT.png')
    # 實際遊戲區逐秒取樣，旗幟及字幕不計入動態證據。
    hashes=subprocess.check_output(['ffmpeg','-nostdin','-v','error','-ss','50','-i',movie,'-t','34',
          '-vf','crop=1280:960:320:60,fps=1,scale=160:120','-an','-f','framemd5','-'],text=True)
    rows=[line.rsplit(',',1)[-1].strip() for line in hashes.splitlines() if line and not line.startswith('#')]
    assert len(rows)>=30 and len(set(rows))>=20, '交火實錄必須有實際動態'
    (OUT/'COMBAT-FRAMES.md5').write_text(hashes)
    write(OUT/'AUDIT.json',dict(version=VERSION,seconds=90,frames=2700,resolution=[1920,1080],fps=30,
         video_sha256=sha(movie),movie_bytes=movie.stat().st_size,non_silent=True,black_segments=0,
         sampled_combat_frames=len(rows),distinct_combat_frames=len(set(rows)),
         frozen_intervals='依VIDEO-AUDIT逐段比對；首尾故事卡與真實讀檔閱讀停留為刻意靜態。',
         text_bounds=len(json.loads((OUT/'TEXT-BOUNDS.json').read_text())),visual_review='待人工目視抽樣幀',
         sources='PLAN.json與正常X11 TRACE.json',rights='本機限定'))
    print('90秒影片技術檢查通過；待目視CONTACT與交火抽樣',flush=True)

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('mode',choices=['prepare','encode','audit'])
    p.add_argument('--version', default=VERSION)
    p.add_argument('--root', default=str(ROOT))
    p.add_argument('--scenes', help='已目視新實錄後的九段分鏡JSON；舊分鏡預設保持')
    p.add_argument('--font-dir', default='/fonts')
    args=p.parse_args()
    VERSION=args.version
    ROOT=Path(args.root)
    OUT=ROOT/'promo'
    SCENES=args.scenes
    FONT=str(Path(args.font_dir)/'NotoSansCJK-Regular.ttc')
    layout.FONT=Path(FONT)
    layout.BOLD=Path(args.font_dir)/'NotoSansCJK-Bold.ttc'
    {'prepare':prepare,'encode':encode,'audit':audit}[args.mode]()
