/*
 * GIN3PS.EXE 321F:01CA  艦隊的每一組編制 -> 對應艦艇池記錄的 far 指標，依序寫進輸出陣列，回組數
 *
 * 原版：far 呼叫，序言 push bp／si／di、sub sp,2、mov bp,sp。堆疊參數（由低位址到高位址）：
 *   out（far 指標，位移 [bp+0Ch]／段 [bp+0Eh]）、fleet（far 指標，位移 [bp+10h]／段 [bp+12h]）。
 *   區域變數 [bp+0] 是已輸出的組數。回傳 AX ＝ 組數。呼叫者 321F:000A 以 out ＝ ds:D78E 呼叫兩次（先處理第一支艦隊，
 *   再把輸出指標推進 4 × 組數處處理第二支），把 ds:D7EC 設成總數。
 *
 * 資料（docs/internals/33 艦隊表、docs/re/287 §307D:03B5 也用同一個讀法）：
 *   艦隊記錄 +0x0F byte ＝ 編制組數；+0x10 起每組 4 bytes，第一個 word ＝ 艦艇池索引（另兩個 byte 本函式不讀）。
 *   艦艇池基底 ds:CFB4（位移）／ds:CFB6（段），一筆 78（0x4E）bytes。
 *
 * 語意（已證實，本檔與反組譯逐條對過）：
 *   for (i ＝ 0; i < byte fleet[+0x0F]; i++) {                         ; 16 位元無號比較，每次迭代重讀 +0x0F
 *       idx ＝ word fleet[+0x10 ＋ 4i]
 *       p ＝ (ds:CFB6) : ((ds:CFB4) ＋ (idx × 0x4E 的低 16 位元))      ; 位移 16 位元回繞
 *       word out[+2] ← 段；word out[+0] ← 位移；out ＝ out ＋ 4
 *   }
 *   return 組數
 * 原版的細節（本版照做）：
 *   - 不讀艦艇池記錄本身，只算指標；idx × 0x4E 取低 16 位元（idx >= 0x35F 時回繞）。
 *   - 組數 +0x0F 與池基底 ds:CFB4／CFB6 在每次迭代都從記憶體重讀；輸出陣列與它們別名時（例如 out 指到艦隊 +0x0C 使 +0x0F 被覆寫，
 *     或 out 指到 ds:CFB4）後面的迭代看得到前面的寫入。先寫段（out[+2]），再寫位移（out[+0]）。
 *   - word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（dosgolem cpu.go read16／write16）；位移加法在 16 位元下回繞。
 *   - 寫入落在線性 0xF0000 以上（ROM 區）時 dosgolem 安靜丟棄；本版自己模型化。
 *
 * 輸入域：fleet 與 out 指向 A0000 到 BFFFF 的案例被探針丟棄；組數最多 255，輸出最多 1020 bytes。
 * 呼叫者（台帳）：321F:000A、321F:008A。
 */
#include "gin3_reg.h"

/* 段內 word 讀取：位移各自在段內回繞（dosgolem cpu.go read16），不是線性位址加 1。 */
static uint16_t f321F01CA_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go，真機同）。 */
static void f321F01CA_wr8_skip_rom(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void f321F01CA_wr16s_rom(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    f321F01CA_wr8_skip_rom(m, LIN(seg, off), (uint8_t)v);
    f321F01CA_wr8_skip_rom(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_321F_01CA(Mem *m, uint16_t out_off, uint16_t out_seg, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint16_t di = 0;
    uint16_t count = 0;

    for (;;) {
        uint16_t cnt = rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x0Fu)));
        uint16_t bx, idx, cx, si, obx;

        if (di >= cnt)
            break;
        bx = (uint16_t)(fleet_off + 0x10u + (uint16_t)(di << 2));
        idx = f321F01CA_rd16s(m, fleet_seg, bx);
        cx = rd16(m, DS(0xCFB6));
        si = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(idx * 0x4Eu));
        obx = out_off;
        out_off = (uint16_t)(out_off + 4u);
        f321F01CA_wr16s_rom(m, out_seg, (uint16_t)(obx + 2u), cx);
        f321F01CA_wr16s_rom(m, out_seg, obx, si);
        count = (uint16_t)(count + 1u);
        di = (uint16_t)(di + 1u);
    }
    return count;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_321F_01CA(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("321F_01CA", 4, run)
