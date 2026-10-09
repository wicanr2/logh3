/*
 * GIN3PS.EXE 181B:0083  戰場回合第 ① 步：目前星域切片內每一顆行星的防衛力回復一步
 *
 * 原版：far 呼叫，無參數，無回傳值（描述檔 ret = none）。切片由全域變數指定：
 *   ds:CFF8（位移）、ds:CFFA（段）＝ 行星切片的第一筆；ds:CFFC（byte）＝ 筆數。一筆 82（0x52）bytes。
 *
 * 語意（已證實：docs/re/339 §2.1 "181B:0083() 行星防衛力回復"；docs/internals/29 行星 +0x14 防衛力、
 *   +0x16 防衛力上限、+0x18 每次回復的百分比；本檔與反組譯逐條對過）：
 *   對切片每一顆行星 p：
 *     若 p[+0x14] < p[+0x16]（無號）：
 *         inc ＝ ((p[+0x16] × p[+0x18]) 的低 16 位元) ÷ 100     ; 16 位元商
 *         p[+0x14] ← (p[+0x14] + inc) 的低 16 位元               ; 會繞回，不先判溢位
 *         若 p[+0x14] >= p[+0x16]：p[+0x14] ← p[+0x16]            ; 夾上限（cmp ax,cx ＋ jb，無號）
 *
 * 與既有文件的差異（以本檔，也就是差異測試通過者為準）：
 *   - docs/re/339 §2.1 寫它與戰略層 20B4:15E1「同一條算式」。算式不同：15E1 用 32 位元乘（3CC9:0009）與
 *     32 位元除（3CC9:0020），這裡是 `mul word ptr` 之後 `xor dx,dx` 再 `div cx`，乘積的高 16 位元被丟掉，
 *     只拿低 16 位元去除 100。cap × rate ≥ 65536 時兩者不同（例如 cap 30000、rate 5：這裡 189，15E1 是 1500）。
 *   - 15E1 沒有「cur < cap 才進來」的守衛（先加再夾）；這裡有。cur >= cap 的行星完全不動，
 *     所以 cur > cap 時不會被拉回 cap。
 *   - docs/re/339 寫夾值條件 "> max"；原版是 jb（新值 < cap 就跳過），即 >=。相等時寫回同一個值，行為上無法區分。
 *
 * 原版的細節（本版照做）：
 *   - 筆數 ds:CFFC 每圈重讀，計數器是 byte 且無號比較；切片指標只在開頭讀一次。
 *   - 加法是對記憶體做的 16 位元 add（add es:[bx],ax），完成後再讀回來與上限比較；上限也是再讀一次。
 *   - 位移運算 16 位元回繞（add bx,14h／add si,16h／add [bp+2],52h）。
 *   - word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（rd16s／wr16s_rom），不是下一段的第一個 byte。
 *   - 切片段被改到 0xF000 以上時，寫入落在 BIOS ROM 區（線性 0xF0000 起），機器忽略該寫入
 *     （dosgolem internal/machine/bios_rom.go）；讀回來的仍是舊值。本版用 wr16s_rom 表達。
 *
 * 輸入域：切片指標與筆數取基底狀態（星域 0，4 顆）並由突變導向讀取集與種子改動；筆數可到 255，
 *   超出 81 顆時讀寫行星表之後的記憶體，本版同樣照做。平面視訊記憶體的寫入由探針丟棄。
 *   戰場局面本身不在輸入域：基底是戰略層狀態，戰場的行星切片由種子與突變構造。
 */
#include "gin3_reg.h"

/*
 * 段內存取的輔助。gin3_mem.h 的 rd16／wr16 吃線性位址，高位元組取「線性位址＋1」；真正的 CPU 位移各自回繞，
 * seg:FFFF 的 word 高位元組在 seg:0000（dosgolem internal/cpu/cpu.go read16／write16，8086 語意）。
 * 位移可能是 0xFFFF 時用 rd16s／wr16s_rom。寫入另外表達 ROM：0xF0000 以上寫入無效
 * （dosgolem internal/machine/bios_rom.go），用本檔的 wr8_skip_rom。這兩點 gin3_mem.h 都沒有（回報給主工作階段）。
 */
static uint16_t rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go，真機同）。gin3_mem.h 的 wr8 沒有這個語意（回報給主工作階段）。 */
static void wr8_skip_rom(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void wr16s_rom(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8_skip_rom(m, LIN(seg, off), (uint8_t)v);
    wr8_skip_rom(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_181B_0083(Mem *m)
{
    uint16_t off = rd16(m, DS(0xCFF8));
    uint16_t seg = rd16(m, DS(0xCFFA));
    uint8_t i = 0;

    while (i < rd8(m, DS(0xCFFC))) {
        uint16_t cur_off = (uint16_t)(off + 0x14u);
        uint16_t cap_off = (uint16_t)(off + 0x16u);

        if (rd16s(m, seg, cur_off) < rd16s(m, seg, cap_off)) {
            uint16_t cap = rd16s(m, seg, cap_off);
            uint16_t rate = rd16s(m, seg, (uint16_t)(off + 0x18u));
            uint16_t inc = (uint16_t)((uint16_t)(cap * rate) / 100u);

            wr16s_rom(m, seg, cur_off, (uint16_t)(rd16s(m, seg, cur_off) + inc));
            if (rd16s(m, seg, cur_off) >= rd16s(m, seg, cap_off))
                wr16s_rom(m, seg, cur_off, rd16s(m, seg, cap_off));
        }
        i++;
        off = (uint16_t)(off + 0x52u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_181B_0083(m);
    return 0;
}
GIN3_REGISTER("181B_0083", 0, run)
