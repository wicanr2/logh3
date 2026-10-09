/*
 * GIN3PS.EXE 1C03:0006  戰場移動階段開頭：在場艦隊清單裡每支艦隊，把 +0x52 的 dword 抄到 +0x56，再呼叫 1C61:0005
 *
 * 原版：far 呼叫，無參數，回傳 AX 是殘值（見下）。序言 push bp／si／di、sub sp,0Ch，區域變數：
 *   di ← 0（清單序號，word）、[bp+0] ← 0x9898（清單位移，word）、[bp+2] ← DS（清單段）
 *   迴圈：
 *     di >= word ds:9894（無號，jnb）→ 結束，AX ＝ di
 *     f   ← byte ds:[清單位移]                          ; 在場艦隊清單，一項 1 byte（艦隊編號）
 *     艦隊 ← ds:(0xC074 ＋ low16(f × 0x7A))              ; 艦隊表一筆 122 bytes
 *     ax ← word 艦隊[+0x54]；bx ← word 艦隊[+0x52]
 *     word 艦隊[+0x58] ← ax；word 艦隊[+0x56] ← bx        ; dword 艦隊[+0x52] 抄到 [+0x56]（來源與目的互不重疊）
 *     1C61:0005(艦隊 far 指標)
 *     di++；清單位移++
 * 本函式自己只寫艦隊 +0x56..+0x59 四個 bytes；其餘寫入都在 1C61:0005。
 *
 * 回傳：離開迴圈時 AX ＝ di（比較用的 `mov ax,di`），描述檔 ret = ax。呼叫者 181B:00FA 不用回傳值，
 *   AX ＝ di 是 codegen 的殘值，不屬於語意，但差異測試照比。
 * 呼叫者：181B:00FA（戰場移動階段的開頭，docs/spec/080 第 48 行「181B:00FA 先 1C03:0006 初始化，再依戰場清單順序逐艦隊」）。
 *
 * 推論等級
 *   已證實（差異測試）：清單基底 0x9898 與筆數 ds:9894（word 無號比較；注意 1588:0945 讀同一個筆數是 byte）、艦隊大小 0x7A 與基底
 *     0xC074、dword 搬移的來源 +0x52 與目的 +0x56、先搬移後呼叫 0005、`AX ＝ di`。
 *   強推論（docs/re/371、docs/spec/080）：艦隊 +0x52 是這回合的行進額度、+0x56 是剩餘額度；搬移是「重置」；
 *     ds:9894／9898 是在場艦隊清單與筆數。
 * 輸入域：清單項目 0..255、筆數任意 word；筆數大到超過探針步數上限（200,000）的執行由探針丟棄。艦隊 far 指標固定在 DS 段，
 *   位移 16 位元回繞（f × 0x7A ＋ 0xC074 可超過 0xFFFF），word 存取在位移 0xFFFF 時高位元組取位移 0。
 */
#include "gin3_reg.h"

extern void f_1C61_0005(Mem *m, uint16_t fleet_off, uint16_t fleet_seg);

/* DS 段內的 word 存取：位移 16 位元回繞，seg:FFFF 的高位元組在位移 0（與 dosgolem 一致） */
static uint16_t f1C03_0006_rd16(const Mem *m, uint16_t off)
{
    return (uint16_t)(rd8(m, DS(off)) | (rd8(m, DS((uint16_t)(off + 1u))) << 8));
}

static void f1C03_0006_wr16(Mem *m, uint16_t off, uint16_t v)
{
    wr8(m, DS(off), (uint8_t)v);
    wr8(m, DS((uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_1C03_0006(Mem *m)
{
    uint16_t di = 0;
    uint16_t list = 0x9898u;

    while (di < f1C03_0006_rd16(m, 0x9894u)) {
        uint16_t fleet = (uint16_t)(0xC074u + (uint16_t)(rd8(m, DS(list)) * 0x7Au));
        uint16_t hi = f1C03_0006_rd16(m, (uint16_t)(fleet + 0x54u));
        uint16_t lo = f1C03_0006_rd16(m, (uint16_t)(fleet + 0x52u));

        f1C03_0006_wr16(m, (uint16_t)(fleet + 0x58u), hi);
        f1C03_0006_wr16(m, (uint16_t)(fleet + 0x56u), lo);
        f_1C61_0005(m, fleet, GIN3_DS);
        di = (uint16_t)(di + 1u);
        list = (uint16_t)(list + 1u);
    }
    return di;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_1C03_0006(m);
}
GIN3_REGISTER("1C03_0006", 0, run)
