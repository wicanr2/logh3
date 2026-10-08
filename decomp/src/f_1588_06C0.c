/*
 * GIN3PS.EXE 1588:06C0  戰場地圖 RLE 解壓縮（128 欄 × 102 列 ＝ 13,056 格）
 *
 * 原版：far 呼叫，兩個 far 指標參數：目的地（位移在前、段在後；[bp+8]）、壓縮資料來源（[bp+0xC]）。無回傳值。
 * 唯一呼叫者 1588:06A3（載入 MAP<nn>.CMP 的那一支，docs/re/314 §四；目的地 ＝ ds:3C1A 指的地圖緩衝區）。
 *
 * 語意（已證實：與反組譯逐條對過；檔案格式的文字版說明在 docs/re 尚無專文，格式推論等級同「強推論」）：
 *   共輸出 0x66 × 0x80 個 byte，一個 byte 一個 byte 輸出：
 *     若 cnt == 0（需要讀下一個封包）：
 *       若 *src 的 bit7 ＝ 1：val ← *src & 0x7F，src++，cnt ← 1             （單格封包，1 byte）
 *       否則：val ← *src，src++，cnt ← *src，src++                           （重複封包：值、次數，2 bytes）
 *     輸出 val 到 *dst，dst++，cnt-- （cnt 是 byte，0 減 1 變 0xFF：次數 0 的封包實際重複 256 次）
 *   cnt 跨列保留（不在列首重置），初值 0。src、dst 的位移各自 16 位元回繞，段不變。
 *   原版對 src 的 bit7 測試與取值是各讀一次，本版也各讀一次（目的地與來源重疊時次序有差別）。
 *
 * 輸入域：目的地 ＝ 地圖緩衝區（13,056 bytes，基底狀態裡 ds:3C1A 所指），來源是任意 bytes（含全 0、開頭 700 個 0xFF，見下面的步數限制）。
 *   全由 0x80 以上的 byte 組成的來源會消耗 13,056 個來源 byte；差異測試的來源緩衝區 8192 bytes，後面接未填隨機的基底內容。
 *   探針每案上限 200,000 步（decomp/harness/diffprobe.go 的 stepBudget）。實測：輸出 13,056 個 byte、封包都很長時約
 *   158,400 步；每個單格封包比一般輸出多約 10 步。因此整張地圖裡單格封包（或次數 1 的重複封包）超過約 4,000 個的來源
 *   會被探針丟棄、不在驗證範圍。差異測試的單格封包種子放 700 個（約 165,400 步）。
 */
#include "gin3_reg.h"

void f_1588_06C0(Mem *m, uint16_t dst_off, uint16_t dst_seg, uint16_t src_off, uint16_t src_seg)
{
    uint8_t val = 0, cnt = 0;
    uint8_t row, col;

    for (row = 0; row < 0x66u; row++) {
        for (col = 0; col < 0x80u; col++) {
            if (cnt == 0) {
                if (rd8(m, LIN(src_seg, src_off)) & 0x80u) {
                    val = (uint8_t)(rd8(m, LIN(src_seg, src_off)) & 0x7Fu);
                    src_off = (uint16_t)(src_off + 1u);
                    cnt = 1;
                } else {
                    val = rd8(m, LIN(src_seg, src_off));
                    src_off = (uint16_t)(src_off + 1u);
                    cnt = rd8(m, LIN(src_seg, src_off));
                    src_off = (uint16_t)(src_off + 1u);
                }
            }
            wr8(m, LIN(dst_seg, dst_off), val);
            dst_off = (uint16_t)(dst_off + 1u);
            cnt = (uint8_t)(cnt - 1u);
        }
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1588_06C0(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("1588_06C0", 4, run)
