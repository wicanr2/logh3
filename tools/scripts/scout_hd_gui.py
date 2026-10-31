"""170：沿169已保存的正常入口驅動器，追加P選點與載入量測。僅以真實X11輸入。"""
import hashlib
from pathlib import Path
import subprocess
import sys

base = Path('tools/scripts/battle_crash_gui.py')
assert hashlib.sha256(base.read_bytes()).hexdigest() == '6247c283a1767268704f2ba420c76514e67fd1cd263a6db37e60cec4e804008d'
text = base.read_text()
anchor = "game=subprocess.Popen(command"
assert text.count(anchor) == 1
text = text.replace(anchor, "result['process_start_utc_ns']=time.time_ns()\n    "+anchor)
anchor = "        if kind=='stop':break"
assert text.count(anchor) == 1
text = text.replace(anchor, '''        if kind=='scout_check':
            before=observe()['tactical'];view=before['view']
            layout=json.loads((package/'data/screen-battle.json').read_text());field=layout['field'];mini=layout['minimap']
            menus=json.loads((package/'data/battle-command-menus.json').read_text())['layout']
            visible=[]
            for u in before['units']:
                sx=field['x']+u['X']-view[0]*field['tile_w']+8;sy=field['y']+u['Y']-view[1]*field['tile_h']+8
                if u['Faction']==event.get('faction',1) and u['Squadron']==0 and not u['Out'] and field['x']+8<=sx<field['x']+field['cols']*field['tile_w']-8 and field['y']+8<=sy<field['y']+field['rows']*field['tile_h']-8:
                    visible.append((u,sx,sy))
            assert visible,'沒有可點的己方旗艦'
            u,sx,sy=visible[0];click(sx,sy)
            wait(lambda r:r.get('tactical',{}).get('command_open'))
            assert observe()['tactical']['command_kind']=='fleet'
            point=observe()['tactical']['command_point'];click(point[0],point[1]+2*menus['row_pitch'])
            wait(lambda r:r.get('tactical',{}).get('scout_target'))
            capture('scout-range')
            click(mini['x']+mini['w']//2,mini['y']+mini['h']//2);assert observe()['tactical']['scout_target']
            capture('scout-other-view')
            click(field['x']+field['cols']*field['tile_w']//2,field['y']+field['rows']*field['tile_h']//2)
            row=observe()['tactical'];assert not row['scout_target'] and row['view']==view
            assert not row['message_open'],'選到不可通行處'
            capture('scout-restored');events.append(dict(kind='scout_check',before_view=view,after_view=row['view']))
            continue
        if kind=='font_check':
            for style in ['sans','serif','classic']:
                move(400,3);wait(lambda r:r.get('toolbar_visible'))
                bounds=observe()['toolbar_controls']['font'];click((bounds[0]+bounds[2])//2,(bounds[1]+bounds[3])//2)
                wait(lambda r:r.get('toolbar_menu')=='font')
                bounds=observe()['toolbar_controls']['font-'+style];click((bounds[0]+bounds[2])//2,(bounds[1]+bounds[3])//2)
                wait(lambda r:r.get('font_style')==style);capture('font-'+style)
            continue
        if kind=='stop':break''')
anchor = "result['observed_until_utc_ns']=time.time_ns()"
assert text.count(anchor) == 1
text = text.replace(anchor, anchor+'''
    result['memory_peak_bytes']=int(Path('/sys/fs/cgroup/memory.peak').read_text())
''')
output = Path(sys.argv[1]); assert not output.exists() and output.parent.stat().st_uid == 1000
output.write_text(text)
subprocess.run([sys.executable, str(output), *sys.argv[2:]], check=True)
