"""#92／93：保存第117輪既有證據與第118輪交付身份，不重跑原版。"""
import hashlib
import json
import os
from pathlib import Path


def entry(path):
    h=hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda:f.read(1024*1024),b''): h.update(block)
    return {'path':str(path),'bytes':path.stat().st_size,'sha256':h.hexdigest()}


def main():
    assert Path.cwd()==Path('/src') and os.getuid()==1000
    root=Path('dist-all/v.0.1.1-20261006')
    sums=json.loads((root/'SHA256SUMS.json').read_text())
    for record in sums['files']:
        found=entry(root/record['path'])
        assert (found['bytes'],found['sha256'])==(record['bytes'],record['sha256'])
    original=Path('workplace/oracle/normal79-r117-a')
    repeated=Path('workplace/oracle/normal79-r117-b')
    replay=[]
    assert sorted(p.name for p in original.iterdir())==sorted(p.name for p in repeated.iterdir())
    for p in sorted(original.iterdir()):
        if p.is_file():
            a,b=entry(p),entry(repeated/p.name)
            assert (a['bytes'],a['sha256'])==(b['bytes'],b['sha256'])
            replay.append({'file':p.name,**{k:a[k] for k in ('bytes','sha256')}})
    assert len(replay)==49
    observation=json.loads((original/'observations.json').read_text())
    rows=[{k:v for k,v in row.items() if k!='tables'} for row in observation['rows']]
    artifacts=[root/'BUILD.json',root/'FINALIZE.json',root/'SHA256SUMS.json',
               Path('workplace/round118-delivery-standard.log'),Path('workplace/round118-regression-complete.log'),
               Path('workplace/active-ai93-r118-b/AI.json'),Path('workplace/package92-smoke-v.0.1.1-20261006/SMOKE.json'),
               Path('workplace/music69-windows/package92-smoke-v.0.1.1-20261006/SMOKE.json'),
               Path('workplace/package92-r118-b/AUDIO.json'),Path('workplace/normal79-r117-probec/observations.json')]
    sources=['cmd/gin3/main.go','internal/game/game.go','internal/game/battle.go','internal/game/passivebattle.go',
             'internal/game/aimode.go','internal/ai/tactical.go','internal/ai/focused.go','internal/game/activebattle_test.go',
             'tools/scripts/local_test_package.py','tools/scripts/local_test_finalize.py','tools/scripts/local_test_promo.sh',
             'tools/scripts/local_test_smoke.py','tools/scripts/local_test_windows_smoke.py','tools/scripts/active_tactical_player.go',
             'tools/scripts/local_test_receipt.py','tools/battle_normal_turn_oracle.sh','tools/battle_normal_turn_native_run.sh',
             'tools/scripts/battle_normal_turn_native.go','tools/scripts/battle_normal_turn_prepare.py','tools/scripts/battle_normal_turn_probe.go']
    document={'schema':'logh3.local-test-package.v1','version':sums['version'],'date':'2026-10-06',
              'user_decisions':['停止新增雙玩家命令與原版戰場差異驗證；完成門檻改為機制與兩種AI實作',
                                '啟用原版及強化戰術AI，取代061正常被動設定','先交付本機完整版與推廣影片供試玩'],
              'delivery_files':sums['files'],'artifacts':[entry(p) for p in artifacts],
              'source_files':[entry(Path(p)) for p in sources],
              'ai':json.loads(Path('workplace/active-ai93-r118-b/AI.json').read_text()),
              'audio':json.loads(Path('workplace/package92-r118-b/AUDIO.json').read_text()),
              'original_round117':{'scope':'原版正常S3單人首回合；與當時061被動版remake無完整同態裁決，後續新增比較已依使用者停止',
                                  'metadata':{k:v for k,v in observation.items() if k!='rows'},'rows':rows,'replayed_files':replay},
              'limitations':['Windows僅Wine，未實機','Linux虛擬音效卡，未人耳','macOS未打包',
                             '完整登陸分段演出尚未接完','原版戰術參考點強推論保持','未宣稱完整原版戰場同態'],
              'promo':{'seconds':60,'width':1920,'height':1080,'fps':30,'kind':'實際遊戲畫面靜態剪輯',
                       'music':'BGM_0F.ogg','gain_db':16,'mean_db':-21.7,'max_db':-6.3,'visual_review':'六張抽樣幀已目視，字幕未裁切'}}
    output=Path('docs/internals/data/local-test-package.json')
    assert not output.exists()
    output.write_text(json.dumps(document,ensure_ascii=False,indent=2)+'\n')
    print('PASS 交付檔、來源與既有117原版49檔保存；來源索引',entry(output))


if __name__=='__main__':main()
