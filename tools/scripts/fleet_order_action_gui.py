"""第93輪正常首戰的旗艦命令；真實X11輸入，唯讀觀測，不控制鏡頭。"""
import json
import os
from pathlib import Path
import subprocess
import time

root = Path("/src")
out = root/"workplace/order79-r93-command-gui"
assert not out.exists() and out.parent.stat().st_uid == os.getuid()
out.mkdir()
plan = json.loads((root/"workplace/order79-r93-plan-1/plan.json").read_text())
statuspath = out/"status.json"
trace = []
snapshots = []
started = time.monotonic()
log = (out/"app.log").open("w")
env = dict(os.environ, LOGH3_FLEET_GUI_OUT=str(out))
proc = subprocess.Popen([str(root/"workplace/order79-r93-command-gin3"), "-orig", "/orig",
    "-data", str(root/"docs/internals/data"), "-save", str(out/"player-save"), "-scale", "1"], env=env, stdout=log, stderr=log)
inputproc = None

def x(*args):
    return subprocess.run(["xdotool", *map(str,args)], capture_output=True, text=True, check=True, timeout=30).stdout.strip()

def status():
    assert proc.poll() is None, "app提前退出"
    try:
        return json.loads(statuspath.read_text())
    except FileNotFoundError:
        return None

def wait(test, label, timeout=90):
    deadline = time.monotonic()+timeout
    while time.monotonic()<deadline:
        s = status()
        if s is not None and test(s):
            return s
        time.sleep(.002)
    raise RuntimeError(label+str(status()))

def send(data):
    inputproc.stdin.write(json.dumps(data)+"\n")
    inputproc.stdin.flush()
    assert json.loads(inputproc.stdout.readline())["queued"]

def press(kind, px=0, py=0, key=0):
    before = wait(lambda s:True, "觀測器")
    if kind == "key":
        slot, down, up = 2, {"Type":2,"Key":key}, {"Type":3,"Key":key}
    else:
        x("mousemove","--window",window,px,py)
        wait(lambda s:s["cursor"]==[px,py],"游標")
        slot = 0 if kind=="click" else 1
        button = 1 if slot==0 else 3
        down, up = {"Type":4,"Button":button}, {"Type":5,"Button":button}
    edge = before["edges"][slot]+1
    send(down)
    try:
        wait(lambda s:s["edges"][slot]>=edge,"按下",5)
    finally:
        send(up)
    after = wait(lambda s:not s["buttons"][slot],"放開",5)
    trace.append({"kind":kind,"x":px,"y":py,"key":key,"at_seconds":time.monotonic()-started,"before":before,"after":after})

def request(data):
    st = status()
    path = out/"request.json"
    assert not path.exists()
    tmp = out/"request.tmp"
    tmp.write_text(json.dumps(data))
    tmp.rename(path)
    return wait(lambda s:s["requests"]==st["requests"]+1,"請求",10)

def snapshot(name):
    st = request({"Name":name})
    wait(lambda s:s["app_ticks"]>=st["app_ticks"]+3,"Draw",10)
    subprocess.run([str(root/"workplace/round79-capture"),"-window",window,"-base",str(out/name),"-scale","1"],check=True)
    comparison = json.loads((out/(name+"-x11-comparison.json")).read_text())
    assert comparison["different"]==0,comparison
    snapshots.append({"name":name,"comparison":comparison})

try:
    inputproc = subprocess.Popen([str(root/"workplace/order79-r93-input")],stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=log,text=True)
    window = x("search","--sync","--onlyvisible","--all","--pid",proc.pid,"--name","^銀河英雄傳說III SP$").splitlines()[0]
    x("windowfocus",window)
    x("windowmove",window,0,0)
    for action in plan["actions"]:
        if action["kind"]=="seed":
            break
        wait(lambda s:all(s.get(k)==v for k,v in action["before"].items()),"正常入口")
        press(action["kind"],action["x"],action["y"],action["key"])
        wait(lambda s:all(s.get(k)==v for k,v in action["after"].items()),"入口結果")
    st = wait(lambda s:s["state"]=="StateBattle" and s["idle"],"首戰")
    left, top, width, height = st["field"]
    flags = [u for u in st["flagships"] if u["faction"]==1 and not u["out"]
        and left<=u["x"] and top<=u["y"] and u["x"]+16<=left+width and u["y"]+16<=top+height]
    assert flags, "初始鏡頭沒有可見己方旗艦"
    flag = flags[0]
    assert flag["action"]==flag["raw_action"]==127
    snapshot("before-order")
    for row in (0,1):
        press("click",flag["x"]+8,flag["y"]+8)
        st = wait(lambda s:s["battle_cmd"],"艦隊選單",5)
        mx,my = st["command_xy"]
        dx,dy,pitch = st["command_layout"]
        press("click",mx+dx+8,my+dy+row*pitch+8)
        wait(lambda s:s["order"]==1,"目的地",5)
        press("click",flag["x"]+8,flag["y"]+8)
        wait(lambda s:s["order"]==2,"朝向",5)
        press("click",flag["x"]+8,flag["y"]+8)
        st = wait(lambda s:s["order"]==0 and not s["battle_cmd"],"完成命令",5)
        current = next(u for u in st["flagships"] if u["pool"]==flag["pool"])
        assert current["action"]==current["raw_action"]==127,current
        snapshot("move-order" if row==0 else "battle-order")
    request({"Save":True})
    st = status()
    current = next(u for u in st["flagships"] if u["pool"]==flag["pool"])
    assert current["action"]==current["raw_action"]==127
    assert len(list(out.glob("*.DT0")))==6
    receipt = {"classification":"正常首戰與公開移動／戰鬥命令；X11真實輸入；保存由診斷API",
        "no_seed_or_state_injection":True,"no_camera_control":True,"flagship":flag,
        "after":current,"trace":trace,"snapshots":snapshots,"six_diagnostic_save_files":True,
        "limitations":["未執行下一回合或交戰","不是原版同狀態對拍","不是自然勝負GUI"]}
    (out/"receipt.json").write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+"\n")
    print("正常旗艦GUI命令通過，Action7F保留，三張原生X11畫面零差，六檔診斷保存")
except BaseException:
    (out/"failed-trace.json").write_text(json.dumps(trace,ensure_ascii=False,indent=2)+"\n")
    raise
finally:
    for child in (inputproc,proc):
        if child is None:
            continue
        child.terminate()
        try:
            child.wait(timeout=5)
        except subprocess.TimeoutExpired:
            child.kill()
            child.wait(timeout=5)
    log.close()
