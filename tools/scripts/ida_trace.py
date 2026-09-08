"""在 `seg_to_com.py` 切出來的段上追一個位址：函式起點、反組譯、誰呼叫它。

    IDA_TRACE_OFFSETS=5CA,0B03 tools/ida.sh script seg113A.com ida_trace.py <輸出路徑>

**只吃一個命令列參數**：headless 這條路上多給的參數不一定進得了
`sys.argv`，位移改走環境變數。

`.COM` 前面補了 0x100 bytes，IDA 又從 `0100h` 起載，所以**段內位移 X 的
位址是 `min_ea + 0x100 + X`**（這顆 image 上 `min_ea` ＝ 0x10100）。
直接用 `0x100 + X` 會落在載入範圍外，而且**不會報錯**，只會什麼都查不到。
沿用 `ida_probe.py` 驗過的形狀：**用 `sys.argv`、用 `ida_pro.qexit`、
只信輸出檔**（headless 的 print 不進 stdout）。

⚠ **容器裡的預設編碼是 ASCII**：`open(out, 'w')` 之後寫中文會丟
`UnicodeEncodeError`，而檔案已經被建立並清空了——症狀是「有檔案、
0 bytes、沒有任何錯誤訊息」。一律 `encoding='utf-8'`。
"""
import sys

import ida_auto
import ida_bytes
import ida_funcs
import ida_pro
import idautils
import idc

import os
import traceback

out = sys.argv[1] if len(sys.argv) > 1 else '/work/ida_trace.txt'
fh = open(out, 'w', encoding='utf-8')
fh.write('起動 argv=%r\n' % (sys.argv,))
fh.flush()

try:
    ida_auto.auto_wait()
    offsets = [int(v, 16) for v in os.environ.get('IDA_TRACE_OFFSETS', '').split(',') if v]
    BASE = idc.get_inf_attr(idc.INF_MIN_EA) + 0x100
    lines = ['位移：%s  base=%X' % ([hex(o) for o in offsets], BASE)]

    for off in offsets:
        ea = BASE + off
        # IDA 對這種「從記憶體切出來的段」只認得到少數入口，其餘留成資料。
        # 先把那一帶整段取消定義，再從指定位址開始逐條建指令，
        # 讓自動分析順著流程走下去。
        ida_bytes.del_items(ea, ida_bytes.DELIT_EXPAND, 0x400)
        cur = ea
        for _ in range(400):
            if not idc.create_insn(cur):
                break
            nxt = idc.next_head(cur)
            if nxt == idc.BADADDR or nxt > ea + 0x400:
                break
            cur = nxt
        ida_auto.auto_wait()
        f = ida_funcs.get_func(ea)
        if not f:
            ida_funcs.add_func(ea)
            ida_auto.auto_wait()
            f = ida_funcs.get_func(ea)
        head = f.start_ea if f else ea
        end = f.end_ea if f else ea + 0x100
        lines.append('')
        lines.append('== 段內 %04X（位址 %04X）函式 %04X–%04X  首位元組 %s'
                     % (off, ea, head - BASE, end - BASE,
                        ' '.join('%02X' % ida_bytes.get_byte(ea + k) for k in range(6))))
        cur = head
        for _ in range(300):
            lines.append('  %04X  %s' % (cur - BASE, idc.GetDisasm(cur)))
            nxt = idc.next_head(cur)
            if nxt == idc.BADADDR or nxt >= end:
                break
            cur = nxt
        lines.append('  -- 呼叫這個函式的地方 --')
        n = 0
        for x in idautils.XrefsTo(head):
            lines.append('    %04X  type=%d' % (x.frm - BASE, x.type))
            n += 1
        if n == 0:
            lines.append('    （沒有）')

    fh.write('\n'.join(lines))
    fh.flush()
    ida_pro.qexit(0)
except Exception:
    fh.write(traceback.format_exc())
    fh.flush()
    ida_pro.qexit(1)
