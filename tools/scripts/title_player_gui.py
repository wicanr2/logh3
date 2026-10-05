"""容器內正常cmd/gin3＋X11鍵鼠；不改TPS、不注入Game或時間。"""
import argparse, json, pathlib, shutil, subprocess, time

p=argparse.ArgumentParser()
p.add_argument("--case",choices=["left","replay"],required=True)
p.add_argument("--out",required=True)
p.add_argument("--plan",required=True)
a=p.parse_args()
out=pathlib.Path(a.out);out.mkdir()
plan=pathlib.Path(a.plan)
events=[]
begin=time.monotonic()
def run(*args):
    return subprocess.run(args,check=True,capture_output=True,text=True,timeout=10).stdout
def note(kind,**extra):
    events.append(dict(kind=kind,seconds=time.monotonic()-begin,**extra))
    print(json.dumps(events[-1],ensure_ascii=False),flush=True)
log=(out/"game.log").open("x")
ticks=5200 if a.case=="left" else 5480
cmd=["/src/workplace/title68-r97-gin3","-orig","/orig","-data","/src/docs/internals/data",
     "-save","/tmp/title-player-save","-shot",str(out/"final.png"),"-shot-bin",str(out/"final.bin"),
     "-state-out",str(out/"final.txt"),"-shot-after",str(ticks)]
proc=subprocess.Popen(cmd,stdout=log,stderr=log)
try:
    window=run("xdotool","search","--sync","--onlyvisible","--name","^銀河英雄傳說III SP$").splitlines()[0]
    run("xdotool","windowfocus",window)
    run("xdotool","mousemove","--window",window,"600","340")
    note("normal-start",command=cmd,window=window,tps=60,injection=False)
    time.sleep(78)
    def capture(name,expected):
        base=out/name
        shutil.copyfile(plan/(expected+".png"),str(base)+".png")
        note("capture",name=name,expected=expected)
        run("/src/workplace/title68-r97-capture","-window",window,"-base",str(base))
    capture("title","title")
    if a.case=="left":
        run("xdotool","mousedown","3");time.sleep(.15);run("xdotool","mouseup","3");time.sleep(.2)
        note("right",x=600,y=340);capture("right-ignored","title")
        run("xdotool","keydown","Return");time.sleep(.15);run("xdotool","keyup","Return");time.sleep(.2)
        note("key",key="Return");capture("key-ignored","title")
        run("xdotool","mousedown","1");time.sleep(.15);run("xdotool","mouseup","1");time.sleep(.2)
        note("left",x=600,y=340);capture("main","main")
    proc.wait(timeout=20)
    if proc.returncode: raise RuntimeError("正式遊戲退出失敗")
    state=dict(line.split("=",1) for line in (out/"final.txt").read_text().splitlines())
    expected="main" if a.case=="left" else "replay-60hz"
    wanted="StateMainMenu" if a.case=="left" else "StateIntro"
    assert state["state"]==wanted and int(state["updates"])==ticks,state
    assert (out/"final.bin").read_bytes()==(plan/(expected+".bin")).read_bytes()
    note("complete",state=state,full_index_pixels=307200,different=0,expected=expected)
    (out/"receipt.json").write_text(json.dumps(dict(schema="logh3.title-player-gui.v1",events=events),ensure_ascii=False,indent=2)+"\n")
finally:
    if proc.poll() is None:
        proc.terminate()
        try: proc.wait(timeout=5)
        except subprocess.TimeoutExpired: proc.kill();proc.wait(timeout=5)
    log.close()
