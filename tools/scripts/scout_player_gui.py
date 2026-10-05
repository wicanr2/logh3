"""第101輪正式60 Hz GUI，沿用已驗證的X11事件與完整擷取工具。"""
import hashlib
import json
from pathlib import Path
import subprocess
import sys

assert Path.cwd() == Path('/src')
parent = json.loads(Path('docs/internals/data/battle-finish-natural.json').read_text())
source = Path('tools/scripts/fleet_player_gui_driver.py')
assert hashlib.sha256(source.read_bytes()).hexdigest() == parent['sources'][str(source)]['sha256']
text = source.read_text()
anchor = "        before=wait(lambda s:match(s,action['before']),f'輸入{i}尚未就緒')"
assert text.count(anchor) == 1
text = text.replace(anchor, '''        if action['kind']=='wait_scout':
            st=wait(lambda s:s is not None and s.get('scout')==action['scout'],'正常索敵里程碑')
            snapshot(action['snapshot'])
            trace.append({'index':i,'kind':'wait_scout','state':st})
            continue
'''+anchor)
anchor = "        kind=action['kind']"
assert text.count(anchor) == 1
text = text.replace(anchor, anchor+'''
        if kind=='move':
            x('mousemove','--window',window,action['x']*a.scale,action['y']*a.scale)
            wait(lambda s:s is not None and s['cursor']==[action['x'],action['y']],'公開移動游標')
            after=observed_after(action['after'],update_offset,i)
            trace.append({'index':i,'kind':kind,'before':before,'after':after})
            continue
''')
assert text.count("    snapshot('final')") == 1
text = text.replace("    snapshot('final')", "    # 最後公開讀檔里程碑loaded已核全圖，不重複建立final快照。")
anchor = "    assert len(list(out.glob('*.DT0')))==6"
assert text.count(anchor) == 1
text = text.replace(anchor, anchor+'''
    public=list((planpath.parent/'player-save').glob('*.DT0'))
    assert len(public)==6
    for p in public:
        assert (out/'player-save'/p.name).read_bytes()==p.read_bytes(),'公開保存不同 '+p.name
''')
text = text.replace("'sound_device_tested':False", "'sound_device_tested':False,'test_tps':60,'normal_scout_expired':True,'public_save_load_equal_plan':True")
text = text.replace('正常主選單、S3、真實X11鍵鼠、命令與交戰；RND test-only，保存由診斷API', '正常主選單、S3、真實X11鍵鼠、四回合索敵失聯、冷卻、同池再派出及公開存讀；首次實行前固定seed')
revision = sys.argv[1]
assert revision.isascii() and revision.isalnum()
dest = Path(f'workplace/scout79-r101-driver-{revision}.py')
with dest.open('x') as stream:
    stream.write(text)
subprocess.run(['python3', str(dest), *sys.argv[2:]], check=True)
