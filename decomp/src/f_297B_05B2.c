/*
 * GIN3PS.EXE 297B:05B2  提督 A 與副官 B（可為空）的能力值合成進艦隊記錄 D
 *
 * 原版：far 呼叫，三個 far 指標參數（push bp／si／di、sub sp,2，所以第一個參數在 [bp+0Ch]），回 AX（見下）：
 *   args[0..1] ＝ D（輸出的艦隊記錄）  args[2..3] ＝ A（提督記錄，ds:B77C 起每筆 28 bytes）  args[4..5] ＝ B（副官記錄；0:0 ＝ 沒有副官）
 * 過程（已證實：反組譯逐指令；297B:0209 由其自己的差異測試證實）：
 *   D[+0x0E] ← 297B:0209(word A[+6]) 的 AL    ; 以 A[+6]（等級）查應有組數；先寫，後面的讀取會看到
 *   nb ← (B 的段 | B 的位移) == 0              ; AX = 1 表示沒有 B（`mov ax,[bp+16h]; or ax,[bp+14h]`）
 *   逐欄位（共 11 筆，依下表順序；每一筆都是：讀 A[f]、讀 B[f]（nb 為真時不讀）、算值、寫 D[g]，寫完才處理下一筆）：
 *     A 群（平均）：f ＝ 0x0F、0x10、0x11、0x13、0x12 → g ＝ 0x3D、0x3E、0x3F、0x40、0x41
 *       nb 為假且 A[f] < B[f]（無號 byte，jb）→ D[g] ← (A[f] + B[f]) >> 1（16 位元相加，不會溢位）；否則 D[g] ← A[f]
 *     B 群（取大）：f ＝ 0x0C、0x0D、0x14、0x16、0x17、0x0E → g ＝ 0x42、0x43、0x44、0x45、0x46、0x47
 *       nb 為假且 A[f] < B[f] → D[g] ← B[f]；否則 D[g] ← A[f]
 *   回 AX ＝ 最後一筆（f ＝ 0x0E）寫進 D[+0x47] 的值（零延伸成 16 位元：`xor ah,ah`）
 * 欄位語意（強推論：docs/re/70 以收據對出寫入順序，docs/re/287 §+22 與 docs/internals/33 第 181 行說這支是
 *   「提督與副官逐項取大」；A 是提督、B 是副官，D 是艦隊記錄，A[+6] ＝ 提督等級，傳給 297B:0209 得應有組數）：
 *   A[+0x0F..0x13] 光線砲、飛彈、戰鬥機、防禦(0x12)、占領(0x13)，A[+0x0C] 移動、A[+0x0D] 回轉、A[+0x14] 與 A[+0x16]（畫面上沒印）、
 *   A[+0x17] 運用，A[+0x0E] 另一項。名稱來自 docs/re/70，不是本檔的證據。
 *   與 docs/re/91 相符（該文以存檔對出同一條規則）；docs/re/66 第 27 行、docs/re/70 第 96 與 115 行、docs/spec/013 第 36 行、
 *   docs/re/287 第 102 行、docs/re/288 第 89 行仍寫 11 項「逐項取大」，反組譯只有後 6 項（D[+0x42..0x47]）是取大；
 *   前 5 項（D[+0x3D..0x41]）是「A < B 時取 (A + B) >> 1（捨去），否則取 A」。
 *   呼叫者 297B:0546（hw-dependent）與 07F9:0D72（unknown）。
 * 副作用：寫 D[+0x0E] 與 D[+0x3D..0x47] 共 12 個 byte。線性 0xF0000 以上（ROM）的寫入在 dosgolem 被安靜丟棄，本檔按 byte 模型化。
 *   D 與 A、B 重疊時，前面的寫入會影響後面的讀取（每筆欄位都重新讀 A、B；描述檔以別名 seeds 驗證）。
 * 輸入域：A、B、D 任意 far 指標（取暫存區或 seeds）；讀到 A0000 到 BFFFF 的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_297B_0209(Mem *m, uint16_t ax_in, uint16_t level_word);

/* 0xF0000 以上是 ROM：寫入無效 */
static void w05B2_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static uint16_t w05B2_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static const struct {
    uint8_t src;
    uint8_t dst;
    uint8_t avg;
} w05B2_rows[11] = {
    {0x0F, 0x3D, 1}, {0x10, 0x3E, 1}, {0x11, 0x3F, 1}, {0x13, 0x40, 1}, {0x12, 0x41, 1},
    {0x0C, 0x42, 0}, {0x0D, 0x43, 0}, {0x14, 0x44, 0}, {0x16, 0x45, 0}, {0x17, 0x46, 0}, {0x0E, 0x47, 0},
};

uint16_t f_297B_05B2(Mem *m, uint16_t d_off, uint16_t d_seg, uint16_t a_off, uint16_t a_seg, uint16_t b_off,
                     uint16_t b_seg)
{
    int nb = (b_seg | b_off) == 0;
    uint8_t v = 0;
    int i;

    w05B2_put8(m, LIN(d_seg, (uint16_t)(d_off + 0x0Eu)),
               (uint8_t)f_297B_0209(m, 0, w05B2_rd16s(m, a_seg, (uint16_t)(a_off + 6u))));
    for (i = 0; i < 11; i++) {
        uint8_t a = rd8(m, LIN(a_seg, (uint16_t)(a_off + w05B2_rows[i].src)));

        v = a;
        if (!nb) {
            uint8_t b = rd8(m, LIN(b_seg, (uint16_t)(b_off + w05B2_rows[i].src)));

            if (a < b)
                v = w05B2_rows[i].avg ? (uint8_t)(((uint16_t)b + a) >> 1) : b;
        }
        w05B2_put8(m, LIN(d_seg, (uint16_t)(d_off + w05B2_rows[i].dst)), v);
    }
    return v;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_297B_05B2(m, a[0], a[1], a[2], a[3], a[4], a[5]); }
GIN3_REGISTER("297B_05B2", 6, run)
