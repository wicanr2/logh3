"""171：實際封包的P索敵與整隊行軍補錄；沿固定正常入口工具，只有X11輸入。"""
import hashlib
import json
from pathlib import Path
import subprocess
import sys

source=Path('tools/scripts/scout_hd_gui.py')
receipt=json.loads(Path('docs/internals/data/scout-formation-hd170.json').read_text())
assert hashlib.sha256(source.read_bytes()).hexdigest()==receipt['tools'][str(source)]
prefix=source.read_text().split('output = Path(sys.argv[1])')[0]
namespace={};exec(compile(prefix,str(source),'exec'),namespace)
text=namespace['text']

def replace(old,new):
    global text
    assert text.count(old)==1,old
    text=text.replace(old,new)

replace('a=p.parse_args();',"p.add_argument('--appimage',action='store_true');p.add_argument('--video',action='store_true')\na=p.parse_args();")
replace('game=None;window=None','game=None;window=None;recorder=None')
replace('    result.update(command=command', '''    if a.appimage:
        env['XDG_CONFIG_HOME']=str(out/'config')
        command=[a.binary,'-intro=false','-frame-trace',str(trace),'-graphics',a.graphics,'-scale','1.5',
            '-poster='+str(a.poster!='off').lower(),'-poster-set',a.poster if a.poster!='off'else'v1','-locale',a.locale,'-ai',str(a.ai)]
        if a.audio=='off':command+=['-music','']
    result.update(appimage=a.appimage,automatic_resources=a.appimage)
    result.update(command=command''')
replace('    until=time.monotonic()+900', '''    if a.video:
        result['record_started_utc_ns']=time.time_ns()
        recorder=subprocess.Popen(['ffmpeg','-nostdin','-v','warning','-f','x11grab','-draw_mouse','0','-framerate','60',
            '-video_size','1280x720','-i',os.environ['DISPLAY']+'.0+0,0','-t','840','-an','-c:v','libx264','-preset','ultrafast',
            '-crf','18','-threads','2','-pix_fmt','yuv420p',str(out/'RAW.mp4')],env=env,stdout=log('record.log'),stderr=subprocess.STDOUT)
        time.sleep(.4)
    until=time.monotonic()+900''')
replace("    print('CAPTURE',name,observe()['state'],flush=True)", "    print('CAPTURE',name,observe()['state'],flush=True)\n    if a.video:time.sleep(.8)")
replace("        if kind=='stop':break", '''        if kind=='march_order':
            row=observe()['tactical'];view=row['view'];units=row['units']
            field=json.loads((package/'data/screen-battle.json').read_text())['field']
            avoid=next((u['Fleet']for u in units if u['Squadron']<0 and u['Faction']==event.get('faction',1)),None)
            chosen=None
            for u in units:
                sx=field['x']+u['X']-view[0]*field['tile_w']+8;sy=field['y']+u['Y']-view[1]*field['tile_h']+8
                if u['Faction']==event.get('faction',1)and u['Squadron']==0 and not u['Out']and u['Fleet']!=avoid and field['x']+8<=sx<field['x']+field['cols']*field['tile_w']-80 and field['y']+8<=sy<field['y']+field['rows']*field['tile_h']-8:
                    chosen=(u,sx,sy);break
            assert chosen,'沒有可點的整隊行軍旗艦'
            u,sx,sy=chosen;capture('march-before');click(sx,sy)
            wait(lambda r:r.get('tactical',{}).get('command_open'))
            assert observe()['tactical']['command_kind']=='fleet'
            point=observe()['tactical']['command_point'];click(*point)
            wait(lambda r:r.get('tactical',{}).get('fleet_order')==1)
            click(sx+56,sy-8);wait(lambda r:r.get('tactical',{}).get('fleet_order')==2)
            click(sx+88,sy-8);assert observe()['tactical']['fleet_order']==0
            events.append(dict(kind='march_order',fleet=u['Fleet'],from_xy=[u['X'],u['Y']],to_xy=[u['X']+64,u['Y']],wall_ns=time.time_ns()))
            capture('march-ordered');continue
        if kind=='stop':break''')
replace("    result['observed_until_utc_ns']=time.time_ns()",'''    result['observed_until_utc_ns']=time.time_ns()
    if recorder is not None:
        recorder.send_signal(signal.SIGINT)
        recorder.wait(timeout=20)
        assert recorder.returncode in (0,255),recorder.returncode
        result['video_sha256']=sha(out/'RAW.mp4')''')
output=Path(sys.argv[1]);assert not output.exists()and output.parent.stat().st_uid==1000
output.write_text(text)
subprocess.run([sys.executable,str(output),*sys.argv[2:]],check=True)
