/*
 * GIN3PS.EXE 307D:0465  收集「距離參數艦隊 500.0 以內的敵對陣營艦隊」：在場艦隊清單裡陣營等於 對方陣營 的艦隊，與參數艦隊的 16.16 距離 <= 500.0 者，
 *   把艦隊的遠指標（位移、段）依序寫進 ds:D61C 起的陣列，計數放 ds:D61B，回傳是否至少找到一支
 *
 * 原版（188 bytes，far 呼叫，一個遠指標堆疊參數；序言 push bp／si／di、sub sp,10h、mov bp,sp，所以參數 F 的位移在 [bp+1Ah]、段在 [bp+1Ch]）。
 *   區域變數：[bp+0] ＝ i（byte）、[bp+2]／[bp+4] ＝ 清單指標（位移從 ds:9898 起、段 DS）、[bp+6] ＝ 對方陣營（byte）、
 *   [bp+8]／[bp+0Ah] ＝ 目前艦隊指標（位移、段 DS）、[bp+0Ch]／[bp+0Eh] ＝ 輸出指標（位移從 ds:D61C 起、段 DS）。DI ＝ 回傳旗標。
 *   回傳 AX ＝ DI（0 或 1）。
 * 過程（已證實：反組譯逐指令；3982:0250、3310:0006、3CC9:00B9 各由其差異測試證實）：
 *   enemy ← 3982:0250(byte F[+1])                 ; 陣營 0 → 1，其餘 → 0；原版推入 word F[+1]，3982:0250 只看低位元組
 *   ds:D61B ← 0；i ← 0；DI ← 0
 *   當 i < byte ds:9894（無號 jnb，每圈重讀）：
 *     f ← 0xC074 + byte ds:(9898 + i) × 0x7A       ; mul cx 之後 16 位元相加
 *     f[+1] == enemy（byte）：
 *       d ← 3310:0006(F[+5C], F[+5E], f[+5C], f[+5E])   ; 16.16 距離，AX:BX
 *       3CC9:00B9(d, 0x01F40000)；jg（ZF ＝ 0 且 SF ＝ OF；有號 32 位元 d > 500.0）→ 略過
 *       否則：word ds:(D61C + 4k)[+2] ← DS、[+0] ← f（k ＝ 目前計數）；byte ds:D61B 加一（讀改寫）；DI ← 1
 *     i 加一（byte）
 *   回 DI
 *   F 的欄位（F[+5C]、F[+5E]）每圈重讀（`les bx,[bp+1Ah]` 在迴圈內），輸出寫入若蓋到這些欄位，後面各圈會看到。
 * 語意（強推論）：F[+1] ＝ 陣營、F[+5C]／F[+5E] ＝ 座標（docs/re/288-fleet-record-remaining-fields.md、docs/internals/33-fleet-table.md、
 *   0FF4:0E5A／297B:0C74 的差異測試）；ds:9894／9898 ＝ 在場艦隊清單的筆數與編號（1588:0945、1C03:0006 的差異測試）；
 *   「500.0」是 16.16 的 0x01F40000。輸出陣列與計數的用途未查（ds:D61B／D61C）。
 * 副作用：ds:D61B ← 0 再逐筆加一；ds:D61C 起每筆 4 bytes（位移、段）。3310:0006 內部的 ds:8EDC..8EED 暫存由被呼叫者的 C 版寫入。
 *
 * 輸入域：基底 fcmd-base.state（被呼叫的 39B5 C 版寫死進入點 SP = 1F8A 的堆疊位址）；清單項與艦隊座標由突變與種子涵蓋。筆數 ds:9894 最多 255。
 *   word 存取在位移 0xFFFF 時 dosgolem 的高位元組取同一段的位移 0（rd16s／wr16s）。輸出陣列最大 D61C + 4 × 255，不回繞。
 *   參數 F 的段不是 DS 時也成立（F 以 farptr 讀取）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_3982_0250(uint8_t faction);
extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

/* word 讀寫在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 的 read16／write16） */
static uint16_t f307D_0465_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void f307D_0465_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

#define F0465_ZF 0x0040u
#define F0465_SF 0x0080u
#define F0465_OF 0x0800u

uint16_t f_307D_0465(Mem *m, uint16_t f_off, uint16_t f_seg)
{
    uint8_t enemy;
    uint8_t i = 0;
    uint16_t list = 0x9898u;
    uint16_t out = 0xD61Cu;
    uint16_t found = 0;

    /* 原版推入 word F[+1]；3982:0250 只取低位元組 */
    enemy = (uint8_t)f_3982_0250(rd8(m, LIN(f_seg, (uint16_t)(f_off + 1u))));
    wr8(m, DS(0xD61Bu), 0);
    while (i < rd8(m, DS(0x9894u))) {
        uint16_t f = (uint16_t)(0xC074u + (uint16_t)(rd8(m, DS(list)) * 0x7Au));

        if (rd8(m, DS((uint16_t)(f + 1u))) == enemy) {
            uint16_t fx = f307D_0465_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu));
            uint16_t fy = f307D_0465_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu));
            uint16_t gx = f307D_0465_rd16s(m, GIN3_DS, (uint16_t)(f + 0x5Cu));
            uint16_t gy = f307D_0465_rd16s(m, GIN3_DS, (uint16_t)(f + 0x5Eu));
            uint32_t d = f_3310_0006(m, fx, fy, gx, gy);
            uint16_t fl = f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, 0x01F4u, 0x0000u);
            int jg = !(fl & F0465_ZF) && (((fl & F0465_SF) != 0) == ((fl & F0465_OF) != 0));

            if (!jg) {
                f307D_0465_wr16s(m, GIN3_DS, (uint16_t)(out + 2u), GIN3_DS);
                f307D_0465_wr16s(m, GIN3_DS, out, f);
                out = (uint16_t)(out + 4u);
                wr8(m, DS(0xD61Bu), (uint8_t)(rd8(m, DS(0xD61Bu)) + 1u));
                found = 1;
            }
        }
        i = (uint8_t)(i + 1u);
        list = (uint16_t)(list + 1u);
    }
    return found;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_307D_0465(m, a[0], a[1]); }
GIN3_REGISTER("307D_0465", 2, run)
