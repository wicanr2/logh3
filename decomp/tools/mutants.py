"""負向對照：套用 specs/<id>.mutants.json 的每個突變到 C 參考實作，重編並跑向量檔，
每個突變都必須被抓到（檢查程式退出碼非 0）。有任何突變存活就失敗。

用法（Docker 內，python3 -I）：
    mutants.py <id> <向量檔所在目錄> <vectors檔名>
"""
import json
import os
import subprocess
import sys

d = "/d"
fid, vdir, vname = sys.argv[1:4]
muts = json.load(open(f"{d}/specs/{fid}.mutants.json"))
src_name = f"f_{fid}.c"
orig = open(f"{d}/src/{src_name}").read()
done_ids = sorted(n[:-4] for n in os.listdir(f"{d}/ledger/status.d") if n.endswith(".tsv")) if os.path.isdir(f"{d}/ledger/status.d") else []
others = [f"{d}/src/f_{i}.c" for i in done_ids if f"f_{i}.c" != src_name and os.path.exists(f"{d}/src/f_{i}.c")]
survivors = []
timeouts = []
os.makedirs("/build/m", exist_ok=True)
for i, m in enumerate(muts):
    if orig.count(m["find"]) != 1:
        print(f"突變 {m['name']}：原始碼中找到 {orig.count(m['find'])} 處（必須剛好 1 處）")
        sys.exit(2)
    mut_src = f"/build/m/{src_name}"
    open(mut_src, "w").write(orig.replace(m["find"], m["replace"]))
    exe = f"/build/m/check_{i}"
    cc = subprocess.run(["gcc", "-std=gnu99", "-O1", "-Wall", "-Wextra", f"-I{d}/include", mut_src, *others,
                         f"{d}/test/check_diff.c", "-o", exe], capture_output=True, text=True)
    if cc.returncode != 0:
        print(f"突變 {m['name']}：編譯失敗\n{cc.stderr}")
        sys.exit(2)
    try:
        run = subprocess.run([exe, f"{vdir}/{vname}"], capture_output=True, text=True, timeout=180)
        caught = run.returncode != 0
        last = run.stdout.strip().splitlines()[-1] if run.stdout.strip() else ""
    except subprocess.TimeoutExpired:
        caught = True      # 突變造成無窮迴圈或極慢，視為被抓到；但只是弱證據，另外計數
        timeouts.append(m["name"])
        last = "（逾時 180 秒，視為被抓到，弱證據）"
    print(f"  {'抓到' if caught else '存活'}  {m['name']}  {last}")
    if not caught:
        survivors.append(m["name"])
print(f"突變 {len(muts)} 個，存活 {len(survivors)} 個" + (f"：{', '.join(survivors)}" if survivors else ""))
if timeouts:
    print(f"其中 {len(timeouts)} 個是逾時才算抓到（沒有實際比出差異）：{', '.join(timeouts)}")
sys.exit(1 if survivors else 0)
