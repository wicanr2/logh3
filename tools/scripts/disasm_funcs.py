"""把記憶體倒出來的程式碼段裡的幾支函式反組譯成文字。

    idat -A -S"disasm_funcs.py <輸出.txt> <段內位移(hex)> ..." code.com

`GIN3PS.EXE` 是 GLUE 1.70 包過的，靜態反組譯只看得到解包器，所以程式碼
一律從執行期記憶體倒。倒出來的段前面補 0x100 bytes 之後當 `.COM` 給 IDA
載，段內位移 X 就落在「第一個段的起點 ＋ 0x100 ＋ X」。

headless 的 print 不進 stdout，所以連診斷（ARGV、段清單、例外）都寫進
輸出檔——沒有輸出檔就是沒跑到。
"""
import traceback

import idc

lines = []
try:
    import ida_auto
    import ida_bytes
    import ida_funcs
    import ida_lines
    import ida_segment
    import ida_ua

    argv = list(idc.ARGV)
    lines.append("ARGV = %r" % (argv,))
    out = argv[1] if len(argv) > 1 else "/work/disasm.txt"
    # 參數寫成 `位移` 或 `位移:長度`（都是十六進位，長度預設 0x100）。
    offsets = []
    for a in argv[2:]:
        if ":" in a:
            o, n = a.split(":", 1)
            offsets.append((int(o, 16), int(n, 16)))
        else:
            offsets.append((int(a, 16), 0x100))

    ida_auto.auto_wait()
    segs = []
    n = ida_segment.get_segm_qty()
    for i in range(n):
        s = ida_segment.getnseg(i)
        segs.append("%d: %X-%X" % (i, s.start_ea, s.end_ea))
    lines.append("段 %d 個：%s" % (n, "  ".join(segs)))
    base = ida_segment.getnseg(0).start_ea

    # 不靠 IDA 的函式辨識——它把這段當資料，`add_func` 只會生出一條指令的
    # 假函式。改成從指定位址**線性逐條解**，要解多長由參數決定。
    for off, length in offsets:
        ea = base + 0x100 + off
        lines.append("")
        lines.append("==== 段內 +0x%04X（%d bytes）====" % (off, length))
        cur, end = ea, ea + length
        while cur < end:
            ida_bytes.del_items(cur, ida_bytes.DELIT_SIMPLE, 1)
            size = ida_ua.create_insn(cur)
            if size == 0:
                size = 1
                txt = "db %02X" % (ida_bytes.get_byte(cur),)
            else:
                txt = ida_lines.generate_disasm_line(cur, ida_lines.GENDSM_REMOVE_TAGS)
            raw = ida_bytes.get_bytes(cur, size) or b""
            lines.append("  +%04X  %-18s %s" % (cur - base - 0x100, raw.hex(" "), txt or ""))
            cur += size
except Exception:
    out = "/work/disasm-error.txt"
    lines.append(traceback.format_exc())

with open(out, "w", encoding="utf-8") as fh:
    fh.write("\n".join(lines) + "\n")
idc.qexit(0)
