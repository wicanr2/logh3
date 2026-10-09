/*
 * GIN3PS.EXE 297B:131C  列出在場艦隊的槽號，再依條件原地過濾（所在星域、本方、敵方），回個數
 *
 * 原版：far 呼叫，四個堆疊字（push bp／si、sub sp,0Ah，所以第一個參數在 [bp+12h]），回 AL（AH 殘值，描述檔 ret = pack:al）：
 *   args[0..1] ＝ 輸出緩衝區 far 指標（位移、段）：艦隊槽號（byte）的清單
 *   args[2]    ＝ v（word，只用低位元組：mov al,[bp+16h]; cmp al,0FFh）：所在星域條件，0xFF ＝ 不過濾
 *   args[3]    ＝ flags（word，只用低位元組：test byte ptr [bp+18h],1／2）：bit0 ＝ 只留本方、bit1 ＝ 只留敵方
 * 過程（已證實：反組譯逐指令；297B:13C3、142A、1491 由其各自的差異測試證實）：
 *   n ← 0；對 i ＝ 0..31（byte）：艦隊 ＝ ds:(0xC074 + i × 0x7A)；艦隊[+4] != 0xFF（現役）→ 緩衝區[n] ← i（byte 寫入，
 *     寫入位移 16 位元回繞）、n++。每圈都重讀艦隊[+4]，所以緩衝區與艦隊表重疊時，前面的寫入會影響後面的判斷。
 *   v 的低位元組 != 0xFF：n ← 297B:13C3(緩衝區, n, v)     ; 留下艦隊[+6]（所在星域）== v 者
 *   flags & 1：n ← 297B:142A(緩衝區, n, word ds:8F1D)       ; 留下艦隊[+1]（陣營）== ds:8F1D 者
 *   flags & 2：n ← 297B:1491(緩衝區, n, word ds:8F1D)       ; 留下艦隊[+1] != ds:8F1D 者
 *   回 n（AL）。
 * 被呼叫者的 n 參數是 word：低位元組是 n，高位元組是迴圈結束時的 i ＝ 0x20（[bp+0] 與 [bp+1] 相鄰），被呼叫者只取低位元組。
 * 語意（強推論，docs/re/342、docs/re/330 第 302 到 303 行）：這是「列出本星域（或本方、敵方）現役艦隊」的小工具；
 *   欄位名稱不是證據，位移與比較才是。
 * 副作用：緩衝區前 n 個 byte（初次列舉寫的是全部現役艦隊，過濾只會把它們往前壓）。
 * 輸入域：緩衝區在 DS 之外的普通記憶體（暫存區、段內回繞）；緩衝區在 ROM 區（線性 0xF0000 以上）不涵蓋，因為被呼叫者
 *   的 C 版（13C3、142A、1491）沒有 ROM 寫入模型。艦隊表 ds:C074 起 32 筆 × 122 bytes。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_297B_13C3(Mem *m, uint16_t buf_off, uint16_t buf_seg, uint16_t n_word, uint16_t value_word);
extern uint16_t f_297B_142A(Mem *m, uint16_t buf_off, uint16_t buf_seg, uint16_t n_word, uint16_t value_word);
extern uint16_t f_297B_1491(Mem *m, uint16_t buf_off, uint16_t buf_seg, uint16_t n_word, uint16_t value_word);

uint8_t f_297B_131C(Mem *m, uint16_t buf_off, uint16_t buf_seg, uint16_t v_word, uint16_t flags_word)
{
    uint16_t wp = buf_off;
    uint8_t n = 0;
    uint8_t i;

    for (i = 0; i < 0x20u; i++) {
        uint16_t fleet = (uint16_t)(0xC074u + (uint16_t)(i * 0x7Au));

        if (rd8(m, DS((uint16_t)(fleet + 4u))) != 0xFF) {
            wr8(m, LIN(buf_seg, wp), i);
            wp = (uint16_t)(wp + 1u);
            n++;
        }
    }
    if ((uint8_t)v_word != 0xFF)
        n = (uint8_t)f_297B_13C3(m, buf_off, buf_seg, (uint16_t)(0x2000u | n), v_word);
    if (flags_word & 1u)
        n = (uint8_t)f_297B_142A(m, buf_off, buf_seg, (uint16_t)(0x2000u | n), rd16(m, DS(0x8F1Du)));
    if (flags_word & 2u)
        n = (uint8_t)f_297B_1491(m, buf_off, buf_seg, (uint16_t)(0x2000u | n), rd16(m, DS(0x8F1Du)));
    return n;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_297B_131C(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("297B_131C", 4, run)
