"""第139輪：建立兩陣營全新正常錄影輸入；保留既有流程，加入兩項設定及F7。"""
import json
import os
from pathlib import Path

root=Path('/src')
out=root/'workplace/promo139'
assert os.getuid()==1000 and not out.exists()
out.mkdir()
for faction in [0,1]:
    source=root/f'workplace/hero130/full-record-f{faction}.json'
    plan=json.loads(source.read_text())
    i=next(i for i,e in enumerate(plan) if e==dict(kind='click',x=424,y=310))
    extra=[dict(kind='capture',name='settings-on',wait=2),
           dict(kind='click',x=370,y=410),dict(kind='click',x=370,y=434),
           dict(kind='capture',name='settings-off',wait=2),
           dict(kind='click',x=370,y=410),dict(kind='click',x=370,y=434),
           dict(kind='key',key='F7',wait=2),dict(kind='capture',name='settings-original',wait=1),
           dict(kind='key',key='F7',wait=2),dict(kind='capture',name='settings-restored',wait=1)]
    plan[i:i]=extra
    # 前兩回合便會包含交火；保留到第5回合，讓成片選完整演出。
    (out/f'input-f{faction}.json').write_text(json.dumps(plan,ensure_ascii=False,indent=2)+'\n')
    if faction==1:
        end=next(i for i,e in enumerate(plan) if e.get('name')=='turn-02')+1
        (out/'input-f1-final.json').write_text(json.dumps(plan[:end],ensure_ascii=False,indent=2)+'\n')
print('READY 本輪兩陣營新錄影輸入，無-script／狀態或seed注入')
