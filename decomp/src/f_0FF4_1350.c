/*
 * GIN3PS.EXE 0FF4:1350  依替代順序表挑第一個編得出至少一組的艦種（0xFE 是「依艦隊現有編制決定 5、6 的先後」的占位）（回 AL）
 *
 * 原版：far 呼叫，三個堆疊字（序言 push bp／si、sub sp,8，所以 fleet 在 [bp+10h]／[bp+12h]，a 在 [bp+14h]），回 AL。
 * 逐條對應反組譯：
 *   p ← ds:(0x295C ＋ low16(a(byte) × 6))              ; 替代順序表一列 6 bytes，以 0xFF 結尾（docs/internals/data/ship-kind-fallback.json）
 *   迴圈：
 *     k ← byte ds:p
 *     k == 0xFF → 回 0xFF
 *     k != 0xFE：
 *         0FF4:0E84(k) ≠ 0（艦種 k 在目前星域的存量夠編 >= 1 組，見 f_0FF4_0E84.c）→ 回「重讀的 byte ds:p」（＝ k）
 *     k == 0xFE：
 *         0FF4:1404(fleet, &A, &B)                    ; A ＝ 艦隊編制裡艦種 5 的組數、B ＝ 艦種 6 的組數（見 f_0FF4_1404.c）
 *         A >= B（無號）：(c1, c2) ＝ (6, 5)；否則 (c1, c2) ＝ (5, 6)
 *         0E84(c1) ≠ 0 → 回 c1；0E84(c2) ≠ 0 → 回 c2        ; 都不行就接著走下一格
 *     p++（16 位元）
 * 也就是 0xFE 表示「5 與 6 兩種艦種，現有組數較少的那一種優先」（A 是 5 的組數，B 是 6 的組數；A >= B 先試 6）。
 *
 * 堆疊局部變數：A、B 在原版放在堆疊（ss:bp+6、ss:bp+7，相鄰的兩個 byte），1404 透過 far 指標改它們。差異測試排除堆疊區，
 *   所以本版借用線性 0x9D000（探針的暫存緩衝區，本函式不用）：先存原來的 2 bytes、呼叫 1404 讓它寫那兩格、讀回 A、B、再還原，
 *   對記憶體的淨影響為零。
 *
 * 回傳：只比較 AL（描述檔 ret = pack:al）。AH 是 codegen 殘值（回 k 的路徑是 0E84 回傳值的高位元組，回 0xFF 的路徑
 *   是 `a × 6` 的高位元組或 0）。呼叫者 0FF4:11F4 存 AL（mov [bp+1],al），0FF4:131E 直接回傳，AH 不屬語意。
 * 呼叫者：0FF4:11F4、0FF4:131E（docs/re/212、224、231：艦種不足時的替代順序）。
 *
 * 推論等級
 *   已證實（差異測試）：表基底 0x295C、列寬 6、0xFF 結尾、0xFE 的特殊處理與 A／B 的比較方向（無號 jb）、各次呼叫的參數與順序。
 *   強推論（docs/re/231「艦種 0 的替代順序 2 → 4 → FE → 3 → 1」、docs/internals/data/ship-kind-fallback.json）：
 *     表內容是各艦種的替代順序；5、6 是補給／修護艦類（名稱見 docs/internals/64，此函式不依賴）。
 * 輸入域：fleet 取艦隊表任一筆；a 任意 word（取低位元組）；表內容與 0E84 讀的星域、艦種表任意。列沒有 0xFF 又沒有任何艦種編得出時，
 *   原版會繞過整個 64 KB 段，超過探針步數上限的案例由探針丟棄。
 */
#include "gin3_reg.h"

extern uint16_t f_0FF4_0E84(Mem *m, uint16_t kind_word);
extern void f_0FF4_1404(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t a_off, uint16_t a_seg, uint16_t b_off,
                        uint16_t b_seg);

#define F1350_TMP_SEG 0x9D00u /* 線性 0x9D000：探針的暫存緩衝區，本函式不用；借來放 A、B 兩個堆疊局部變數 */

uint8_t f_0FF4_1350(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t a_word)
{
    uint16_t p = (uint16_t)(0x295Cu + (uint16_t)((a_word & 0xFFu) * 6u));

    for (;;) {
        uint8_t k = rd8(m, DS(p));

        if (k == 0xFFu)
            return 0xFFu;
        if (k != 0xFEu) {
            if (f_0FF4_0E84(m, k) != 0)
                return rd8(m, DS(p));
        } else {
            uint8_t saved0 = rd8(m, LIN(F1350_TMP_SEG, 0));
            uint8_t saved1 = rd8(m, LIN(F1350_TMP_SEG, 1));
            uint8_t cnt_a, cnt_b, c1, c2;

            f_0FF4_1404(m, fleet_off, fleet_seg, 0, F1350_TMP_SEG, 1, F1350_TMP_SEG);
            cnt_a = rd8(m, LIN(F1350_TMP_SEG, 0));
            cnt_b = rd8(m, LIN(F1350_TMP_SEG, 1));
            wr8(m, LIN(F1350_TMP_SEG, 0), saved0);
            wr8(m, LIN(F1350_TMP_SEG, 1), saved1);
            if (cnt_a < cnt_b) {
                c1 = 5;
                c2 = 6;
            } else {
                c1 = 6;
                c2 = 5;
            }
            if (f_0FF4_0E84(m, c1) != 0)
                return c1;
            if (f_0FF4_0E84(m, c2) != 0)
                return c2;
        }
        p = (uint16_t)(p + 1u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_0FF4_1350(m, a[0], a[1], a[2]); }
GIN3_REGISTER("0FF4_1350", 3, run)
