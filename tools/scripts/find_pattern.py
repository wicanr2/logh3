"""在記憶體 dump 裡找「和某個分組模式一致」的陣列。

知道 32 個星域各屬於誰（從畫面數出來的）之後，記憶體裡一定有一份
32 筆的歸屬陣列。但不知道它用什麼值表示帝國與同盟——所以不能直接
搜位元組，要搜**模式**：同一群的位置值要相同、不同群的值要不同。

群編號寫 `-1` 表示這個位置不管（判不準的就別讓它否決整個候選）。

用法：tools/py.sh tools/scripts/find_pattern.py <dump> <群編號,逗號分隔> [每筆位元組數]
"""
import pathlib, sys

mem = pathlib.Path(sys.argv[1]).read_bytes()
groups = [int(v) for v in sys.argv[2].split(',')]
step = int(sys.argv[3]) if len(sys.argv) > 3 else 1
n = len(groups)
span = n * step

hits = []
for base in range(len(mem) - span):
    seen, ok = {}, True
    used = set()
    for i, g in enumerate(groups):
        if g < 0:            # -1 ＝ 這個位置的值不管
            continue
        v = mem[base + i * step] if step == 1 else int.from_bytes(
            mem[base + i * step:base + i * step + step], 'little')
        if g in seen:
            if seen[g] != v:
                ok = False
                break
        else:
            if v in used:
                ok = False
                break
            seen[g] = v
            used.add(v)
    if ok:
        hits.append((base, dict(sorted(seen.items()))))

print('%d 處符合（每筆 %d bytes）' % (len(hits), step))
for base, m in hits[:20]:
    print('  %05X  群→值 %s' % (base, m))
