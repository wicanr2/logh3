"""規格140：從已完成的正常錄影選取完整艦隊演出，保留原速與來源時碼。"""
import json
from pathlib import Path

batch=Path('/work/hero130')
version='v.1.2.0-20261012'
scenes=[]


def scene(kind,duration,title,caption,**kw):
    scenes.append(dict(kind=kind,duration=duration,title=title,caption=caption,**kw))


def anime(start,duration,title,caption):
    scene('anime',duration,title,caption+'｜1988動畫，來源原生480p',
          file=str(batch/'ANIME-SOURCE-CLEAN.mp4'),start=start)


def game(run,start,duration,title,caption,**kw):
    scene('game',duration,title,caption,file=str(batch/run/'RAW.mp4'),start=start,**kw)


scene('card',7,'銀河英雄傳說 III SP｜兩面旗幟，兩種命運',
      '帝國的雄心，同盟的信念。下一場會戰，由你改寫。｜1988／1993修復海報',
      poster=str(batch/'anime-sources/remaster-poster.png'))
anime(238,3,'萊因哈特｜向銀河的頂點進軍','以集中突擊撕開戰線')
anime(778,3,'楊威利｜魔術師的下一步','當優勢不在你這邊，戰術就是轉機')
anime(718,4,'艦隊集結，歷史再次分岔','你會追擊，還是看穿對手的誘餌？')
game('full-record-f1-r2',7.6,6,'從S1亞姆力扎會戰開始','正常讀檔實錄｜三個額外劇本完整保留')
game('full-record-f1-r2',25.8,3,'看清對手，觀察戰術','設定選單可開啟顯示敵軍｜遊戲中隨時切換')
game('record-f0-ai-1',33.6,5,'同一場會戰｜既有原版規則模式','同初始艦隊與等價操作，先看既有AI的部署')
game('full-record-f0',34.5,7,'面對魔術師楊｜強化模式','前衛牽制、雙翼包抄。別讓側後方成為破口。')
anime(898,3,'每一道航跡，都可能是伏筆','看穿戰線，才能扭轉局勢')
game('record-f1-ai-1',29.2,5,'同一場會戰｜既有原版規則模式','換到同盟視角，對照帝國艦隊的行動')
game('full-record-f1-r2',35.8,5,'面對萊因哈特｜強化模式','主力集中突擊，側翼封鎖退路')
game('full-record-f1-r2',63.9,7,'名將在場，全軍協同','存活帶隊名將所屬陣營，新移動額度1.5倍｜僅強化模式')
game('record-f0-ai1',31.3,5.8,'按F7，即時切換原版／HD','實際遊戲切換｜操作位置與原版部隊圖示保留')


def ranges(run):
    folder=batch/run
    audit=json.loads((folder/'PLAYBACK-AUDIT.json').read_text())
    assert audit['passed'] and not audit['skipped_frames']
    trace=json.loads((folder/'TRACE.json').read_text())
    origin=trace['record_started_utc_ns'];groups={}
    for line in (folder/'FRAMES.jsonl').open():
        row=json.loads(line);p=row.get('playback')
        if p and p['kind']=='fleet':groups.setdefault(p['id'],[]).append(row)
    result=[]
    for identity,rows in groups.items():
        assert rows[-1]['playback']['result']
        first=(rows[0]['wall_ns']-origin)/1e9
        last=(rows[-1]['wall_ns']-origin)/1e9
        # Use frame-aligned duration, enclosing the whole real displayed scene.
        start=int(first*60)/60
        end=(int(last*60)+2)/60
        result.append((start,end-start,identity,rows[-1]['playback']['weapon']))
    return result


yang=ranges('full-record-f0')
rhine=ranges('full-record-f1-r2')
selected=[('full-record-f0',r) for r in yang]
total=sum(r[1] for _,r in selected)
# Diverse intact volleys rather than accelerating an entire battle.
for r in rhine[7:]:
    selected.append(('full-record-f1-r2',r));total+=r[1]
    if total>=50:break
assert total>=50 and {r[3] for _,r in selected}=={0,1,2}
for i,(run,(start,duration,identity,weapon)) in enumerate(selected):
    title,caption=(('光束交錯，飛彈破陣','正常速度實戰錄影｜HD艦影、攻擊與受損演出'),
                   ('包抄正在收緊','在敵軍完成合圍前，找出能反擊的缺口'),
                   ('一個命令，改變整條戰線','保留遊戲配樂與交火音效｜沒有加速戰鬥動畫'))[min(2,i*3//len(selected))]
    game(run,start,round(duration,6),title,caption,battle_animation=True,playback_id=identity,weapon=weapon)
anime(958,3,'若你是楊威利，下一步會怎麼下令？','把敗局變成轉機，讓歷史出現另一種可能')
scene('card',6,'你的戰術，決定銀河的明天','三平台完整版｜原版／HD即時切換｜名將協同AI')
plan=dict(version=version,background='/src/workplace/hd127/promo-final/BACKGROUND.png',
          music='/src/dist-all/'+version+'/full/LOGH3-'+version+'-linux-amd64/music/BGM_0F.ogg',
          scenes=scenes,rights='全部動畫、原版及衍生素材僅本機使用，未公開上傳',
          anime_source='https://www.bilibili.com/video/BV1G87szDEh7/',
          comparison='remake既有原版規則模式對強化模式，同初始資料及等價玩家輸入')
(batch/'PROMO-PLAN.json').write_text(json.dumps(plan,ensure_ascii=False,indent=2)+'\n')
print('READY',round(sum(s['duration'] for s in scenes),3),'秒；實戰動畫',round(total,3),'秒；',len(scenes),'段')
