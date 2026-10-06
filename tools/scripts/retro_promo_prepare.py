"""#95：收集既有真實音畫、來源與分鏡，不新增原版戰場對拍。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import struct


def sha(p):
    h=hashlib.sha256()
    with p.open('rb') as f:
        for b in iter(lambda:f.read(1024*1024),b''):h.update(b)
    return h.hexdigest()


def prepare(version):
    root=Path('workplace')/('promo95-'+version)
    assert not root.exists()
    root.mkdir()
    (root/'inputs').mkdir()
    old=Path('workplace/package92-smoke-v.0.1.1-20261006')
    source_files={'command':old/'input.png','battle':old/'battle.png','load':old/'load.png',
                  'ship':old/'ship.png','fleet':old/'fleet.png',
                  'admiral':Path('docs/images/09-admiral-detail.png'),
                  'original':Path('workplace/oracle/normal79-r117-a/01-turn-entry.png'),
                  'ai-original':Path('workplace/active-ai93-r118-b/original/turn3/frame.png'),
                  'ai-enhanced':Path('workplace/active-ai93-r118-b/enhanced/turn3/frame.png')}
    entries=[]
    for name,p in source_files.items():
        data=p.read_bytes()
        assert data[:8]==b'\x89PNG\r\n\x1a\n'
        assert struct.unpack_from('>II',data,16)==(640,480)
        q=root/'inputs'/(name+'.png')
        shutil.copyfile(p,q)
        assert sha(p)==sha(q)
        entries.append({'name':name,'source':str(p),'bytes':p.stat().st_size,'sha256':sha(p),
                        'canvas':[640,480], 'type':'dosgolem原版' if name=='original' else '正式remake renderer'})
    music=Path('dist-all')/version/'full-local'/('LOGH3-'+version+'-linux-amd64/music/BGM_0F.ogg')
    assert music.is_file()
    shutil.copyfile(music,root/'inputs/BGM_0F.ogg')
    plan=[
        {'id':'00','layout':'title','seconds':7,'motion':True,'title':'銀河英雄傳說 III SP','subtitle':'銀河的命運，回到你的指揮席'},
        {'id':'01','layout':'world','seconds':8,'motion':True,'title':'兩個國家，一場漫長的戰爭','subtitle':'銀河帝國與自由惑星同盟，交戰一百五十年'},
        {'id':'02','layout':'briefing','seconds':6,'title':'先看星圖，再下命令','subtitle':'財政、情報、造艦與星域部署'},
        {'id':'03','layout':'split','seconds':8,'title':'從 DOS 畫面，走到新的平台','subtitle':'原版／重製：不同回合的版面展示'},
        {'id':'04','layout':'dossier','seconds':6,'title':'你的提督，你的艦隊','subtitle':'任命將領、配置副官、編組作戰部隊'},
        {'id':'05','layout':'wide','seconds':6,'title':'把部署帶進戰術戰場','subtitle':'下令、移動、交戰，推進下一個回合'},
        {'id':'06','layout':'ai','seconds':8,'title':'選擇你的對手','subtitle':'原版與強化：戰略、戰術均已啟用'},
        {'id':'07','layout':'save','seconds':5,'title':'每一次決定，都能留下','subtitle':'保存進度，繼續自己的戰役'},
        {'id':'08','layout':'music','seconds':6,'title':'熟悉的聲音，隨艦隊啟航','subtitle':'22 首配樂與 28 段音效已包含'},
        {'id':'09','layout':'platform','seconds':5,'title':'三種平台，同一套遊戲機制','subtitle':'Windows · Linux · macOS 通用完整版'},
        {'id':'10','layout':'ending','seconds':7,'title':'下一場戰役，交給你','subtitle':'完整版測試包，現在開始試玩'},
    ]
    assert sum(r['seconds'] for r in plan)==72
    value={'version':version,'source_frames':entries,'music':{'source':str(music),'bytes':music.stat().st_size,'sha256':sha(music),
           'provenance':'原版MDS與已驗固定音源工具轉出的正式遊戲BGM_0F；非新造配樂'},
           'storyboard':plan,'layouts':sorted({r['layout'] for r in plan}),
           'comparisons':[{'segment':'03','level':'nearby-state','reason':'dosgolem原版首回合與remake入口不同；完整640×480版面展示，不宣稱戰場同態'},
                          {'segment':'06','level':'模式展示','reason':'既有正常S3兩模式第3回合收據；不新增原版差異驗證'}],
           'original_metadata':{'dosgolem_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c','receipt':'docs/internals/data/local-test-package.json'},
           'differences_from_previous':['功能巡覽改為世界觀及原版／重製對照','11種版面與星圖母題','片段5至8秒、72秒總長','兩段安全單幀動態'],
           'story_sources':['https://www.production-ig.co.jp/works/ginei/story.html','https://youngjump.jp/manga/gineiden/'],
           'limits':['macOS畫面未實機擷取，平台卡只展示封包支援','完整登陸分段、人耳與macOS真機仍待驗收'],
           'rights':'含原版衍生音畫，只留本機，不加入Git或公開上傳'}
    (root/'PLAN.json').write_text(json.dumps(value,ensure_ascii=False,indent=2)+'\n')
    print('PASS 收集9張真實畫面、正式BGM及72秒分鏡',root)


if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--version',required=True)
    prepare(p.parse_args().version)
