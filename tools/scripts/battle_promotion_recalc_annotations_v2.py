"""#79：審查後的重算語意，以原始定位與bytes防護，不改名稱。"""
from pathlib import Path
import hashlib
import json

path = Path('docs/internals/data/battle-promotion-recalc-v2.json')
raw = path.read_bytes()
assert hashlib.sha256(raw).hexdigest() == 'cda14b5b122e4e875a825948c61194c96b534a07c7a563c770b26570b35291e6'
data = json.loads(raw)
notes = {
 '297B:1131':'全局維持費helper也呼叫12FB重算移動費；第一版漏追此callee。',
 '297B:1158':'移動費高word寫WORK+6E+陣營*4+2。',
 '297B:115C':'移動費低word寫WORK+6E+陣營*4。',
 '297B:1305':'駐留狀態2不計入移動費加總。',
 '297B:1309':'停駐狀態5不計入移動費加總。',
 '297B:1315':'移動費FLEET+0C以signed16延伸供陣營加總。',
 '297B:0209':'依提督等級回傳應有組數；九門檻與66案符合。',
 '297B:05B2':'升級後重算應有組數與11項艦隊能力；53次重算全表符合。',
 '297B:0E8B':'依現有組數及SHIP陣營／艦種重算速度、最小速度、分隊士氣與艦隊完整度。',
 '297B:0F38':'回轉基礎速度先截成byte寫SHIP+1F，折扣使用此byte。',
 '297B:0F64':'巡航基礎速度先截成byte寫SHIP+21。',
 '297B:0F8D':'戰鬥基礎速度先截成byte寫SHIP+22。',
 '297B:0FB7':'機動力折扣乘積取低word；後續CWD／IDIV按signed16除100。',
 '297B:0FBC':'以低word符號延伸；不能使用未截斷的整數乘積。',
 '297B:0FBD':'signed16折扣乘積除100，向零取整。',
 '297B:0FD0':'回轉byte乘折扣後清DX，後續只以低word除100。',
 '297B:0FD4':'回轉byte減商的低byte；保留byte回繞。',
 '297B:0FE2':'巡航速度折扣乘積取低word。',
 '297B:0FF7':'戰鬥速度折扣乘積取低word。',
 '297B:1051':'先把FLEET+3C設100，之後1261:032B重算完整度；零艘分支保留100。',
 '297B:1059':'FLEET+4D保存各組SHIP+1F最小值。',
 '297B:1060':'FLEET+4E保存各組SHIP+21最小值。',
 '297B:1067':'FLEET+4F保存各組SHIP+22最小值。',
 '297B:106E':'FLEET+50保存各組SHIP+23最小值；不因本輪尚無值變化而省略。',
 '297B:10B0':'獨立索敵隊回轉速度直接取自身陣營的艦種7，不吃能力折扣。',
 '297B:10BC':'獨立索敵隊巡航速度直接取艦種7。',
 '297B:10C8':'獨立索敵隊戰鬥速度直接取艦種7。',
 '297B:10D0':'呼叫1261:032B，按新提督+0B重算分隊SHIP+06、FLEET+3C及bit0／2／3。',
 '297B:116C':'依各組SHIP陣營／艦種／艘數重算兩項費用；全陣營現役艦隊均更新。',
 '297B:1253':'維持費的unsigned32加總先除1000。',
 '297B:1282':'營運折扣為signed16的200-FLEET+47，後續乘積截unsigned32。',
 '297B:12AC':'營運折扣乘積以unsigned32除100；高能力255診斷保留原版回繞。',
 '297B:12B7':'維持費商的低word寫FLEET+0A。',
 '297B:12E2':'其他費用商的低word寫FLEET+0C。',
 '297B:12EF':'維持費低word以signed16延伸為回傳值，供陣營費用加總。',
 '090A:0E6D':'先呼叫2703:02E5重算經濟力、預估稅與未安定度；本輪完整STELLAR僅原版觀測。',
 '090A:0E75':'全局重算第一份陣營預估，參數0。',
 '090A:0E80':'全局重算第二份陣營預估，參數1。',
 '090A:0EA5':'維持費加總低word寫WORK+5E+陣營*4；高word由前一指令寫入。',
 '090A:0EC5':'造艦可用款再減WORK+6E的既有移動費；297B:10DE已先重算移動費，此處再扣該快取。',
 '090A:0F04':'可用款達21474836時先除100再乘造艦百分比，避開signed32溢位。',
 '090A:0F7A':'小額分支先乘造艦百分比再除100，寫WORK+46+陣營*4。',
}
by_location = {i['original_location']:i for i in data['instructions']+data['global_cost_instructions']}
rows = []
for location, semantic in notes.items():
    ins = by_location[location]
    rows.append({'original_location':location,'original_bytes_hex':ins['original_bytes_hex'],
                 'semantic':semantic,'grade':'proven','source':'docs/re/419-battle-promotion-recalc.md；原始指令與66案／53次三表獨立模型',
                 'source_sha256':hashlib.sha256(raw).hexdigest()})
result = {'schema':'logh3.battle-promotion-recalc-annotations.v2',
          'source_sha256':hashlib.sha256(raw).hexdigest(),'rows':rows}
b = (json.dumps(result,ensure_ascii=False,indent=2)+'\n').encode()
for target in ['docs/internals/data/battle-promotion-recalc-annotations-v2.json','workplace/ida/promotion79-r85-recalc-annotations-v2.json']:
    with Path(target).open('xb') as f:
        f.write(b)
print('重算語意',len(rows),'SHA-256',hashlib.sha256(b).hexdigest())
