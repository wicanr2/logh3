/*
 * GIN3PS.EXE 0FF4:0E84  這個艦種在目前星域的存量，夠編幾組（存量 ÷ 每組艘數，取低 16 位元）
 *
 * 原版：far 呼叫，一個參數 kind（序言 push bp／si／di、sub sp,4，所以在 [bp+0Eh]；只讀低位元組）。回 AX。
 *
 * 反組譯逐條對應：
 *   k   ← kind 的低位元組（零延伸）、s ← byte ds:8F1D（陣營）
 *   rec ← far 指標 ds:9872（位移）／ds:9874（段）
 *   存量 ← dword  rec[+0x30 + s × 0x1C + k × 4]            （位移 16 位元回繞）
 *   每組 ← word   ds:[4E3A + s × 0xD0 + k × 0x1A]           （位移 16 位元回繞；高位字當 0）
 *   回 AX ← (存量 ÷ 每組) 的低 16 位元                       （無號 32 位元除法，3CC9:0020；除數 0 時商 0xFFFFFFFF）
 * 函式沒有寫入任何記憶體。
 *
 * 推論等級
 *   已證實（差異測試）：上列運算、常數、byte 參數、除數 0 回 0xFFFF（不引發 #DE，3CC9:0020 內部處理）。
 *   強推論（docs/internals/56-fleet-composition-panel.md 第 157 行、docs/re/216、229）：ds:9872 指的是艦隊所在
 *     星域記錄，星域 +0x30 起每陣營 28 bytes（7 個艦種各 4 bytes dword 存量）；ds:4E3A 是艦種表
 *     （陣營 × 208 + 艦種 × 26 的列，第 0 欄是每組艘數，docs/internals/64 的艦種表 `ds:4E3A + 陣營 × 208 + 艦種 × 26`）。
 *   未知：艦種超過 6 時兩張表的讀法沒有語意（讀到鄰近資料）。本版照原版運算。
 *
 * 輸入域：kind 任意 word（只用低位元組）；ds:8F1D、ds:9872／9874 與兩張表任意。
 *   存量 dword 的兩個 word 讀取是 16 位元位移：位移 0xFFFF 的 word，高位元組取同一段的位移 0（見 0FF4:0EF1 的差異測試）。
 */
#include "gin3_reg.h"

/* word 讀取：高位元組的位移是 off + 1，16 位元回繞（off ＝ 0xFFFF 時取同一段的位移 0） */
static uint16_t w0E84_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, off + 1)) << 8));
}

uint16_t f_0FF4_0E84(Mem *m, uint16_t kind_word)
{
    uint16_t kind = (uint8_t)kind_word;
    uint16_t side = rd8(m, DS(0x8F1D));
    uint16_t rec_off = rd16(m, DS(0x9872));
    uint16_t rec_seg = rd16(m, DS(0x9874));
    uint16_t at = (uint16_t)(rec_off + 0x30u);
    uint16_t per_off;
    uint32_t stock, per, q;

    at = (uint16_t)(at + (uint16_t)(side * 0x1Cu));
    at = (uint16_t)(at + (uint16_t)(kind << 2));
    stock = ((uint32_t)w0E84_rd16(m, rec_seg, (uint16_t)(at + 2)) << 16) | w0E84_rd16(m, rec_seg, at);
    per_off = (uint16_t)(0x4E3Au + (uint16_t)(side * 0xD0u) + (uint16_t)(kind * 0x1Au));
    per = rd16(m, DS(per_off));
    q = gin3_uldiv(stock, per);
    return (uint16_t)q;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_0FF4_0E84(m, a[0]); }
GIN3_REGISTER("0FF4_0E84", 1, run)
