/*
 * GIN3PS.EXE 2AD2:0029  在艦艇池挑第一個空槽，依來源記錄與艦種建一筆新艦艇記錄（回槽索引）
 *
 * 原版：far 呼叫，三個參數（push bp／si／di、sub sp,10h，所以第一個參數在 [bp+1Ah]）：
 *   args[0..1] ＝ 來源記錄 S 的 far 指標（位移、段；推論為艦隊記錄）    args[2] ＝ 艦種 kind（word，用低位元組）
 *   args[3]    ＝ b（word，用低位元組：mov dl,[bp+20h]）
 * 回傳：AX ＝ 槽索引（2AD2:0151 的回傳；池滿時 0xFFFF，rec ＝ 池基底 + (0xFFFF × 0x4E 的低 16 位元) ＝ 池基底 − 0x4E）。
 *
 * 過程（已證實：反組譯逐指令；2AD2:0151、2703:1433、2703:14CF 由其各自的差異測試證實）。寫入與讀取的交錯順序如下，
 * S 與 rec 重疊時前面的寫入會影響後面的讀取（描述檔以別名 seeds 驗證，並有「不重讀」突變）：
 *   idx ← 2AD2:0151()；rec ← (ds:CFB6 段):(ds:CFB4 + 16 位元(idx × 0x4E))
 *   rec[+0] ← S[+1]；rec[+1] ← S[+3]；rec[+2] ← S[+6]；rec[+4] ← kind（byte）；rec[+3] ← b（byte）
 *   word rec[+0x38] ← 0；word rec[+0x3A] ← 0；word rec[+0x10]、[+0x0E]、[+0x14]、[+0x12] ← 0
 *     （最後四個 word 隨後被 2703:14CF 全部覆寫，所以它們的清零在最終記憶體看不出來）
 *   列 ← ds:(0x4E3A + S[+1] × 0xD0 + kind × 0x1A)（S[+1] 在上面的寫入之後重讀；16 位元乘法）
 *   word rec[+0x0A] ← word 列[+0]
 *   kind == 7（低位元組）：word rec[+8] ← word ds:(0x4E3A + rec[+0] × 0xD0 + 0xB6)（rec[+0] 重讀）；word rec[+0x0C] ← word rec[+0x0A]（重讀）
 *   否則：t ← 2703:1433(S[+6]（word）, S[+1]（word）, kind（word）)；word rec[+8] ← t；word rec[+0x0C] ← t
 *   2703:14CF(rec, word rec[+8])        ; 重讀 rec[+8]；寫 rec[+0x0E..0x15]
 *   回 idx
 * 語意（強推論，docs/re/344 §3、docs/re/345、docs/internals/64-ship-pool.md）：這是「建一個新編制組」，rec[+0] 陣營、rec[+1] 隊形槽？、
 *   rec[+4] 艦種、rec[+8] 艘數、rec[+0x0A] 滿編。名稱不是證據，位移與運算才是。
 * 副作用：寫 rec（池內一筆）的上述欄位；2703:1433 寫星域 +0x30 起的待命存量（kind != 7 時）；2703:14CF 寫 rec +0x0E..0x15。
 *
 * 輸入域與限制：池指標（ds:CFB4／CFB6）用描述檔 freeze 排除隨機突變，由 seeds 給定（基底、池滿、段內回繞、S 放進池內造別名）。
 *   池在 ROM 區（線性 0xF0000 以上）不涵蓋：原版對該區的寫入被安靜丟棄，而 2703:14CF 與 2703:1433 的 C 版沒有 ROM 寫入模型，
 *   本檔也不模型化（描述檔 seeds 與探索都把池放在 RAM）。word 存取在位移 0xFFFF 時高位元組取同一段位移 0（w0029_rd16s／wr16s）。
 *   rec 的 +0x0E 到 +0x15 位移落在 0xFFFF 的 word 不涵蓋（14CF 的 C 版用線性位址）。讀到 A0000 到 BFFFF 的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_2AD2_0151(Mem *m);
extern uint16_t f_2703_1433(Mem *m, uint16_t region, uint16_t side, uint16_t kind);
extern void f_2703_14CF(Mem *m, uint16_t rec_off, uint16_t rec_seg, uint16_t n);

static uint16_t w0029_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}


static void w0029_wr8s(Mem *m, uint16_t seg, uint16_t off, uint8_t v)
{
    wr8(m, LIN(seg, off), v);
}

static void w0029_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_2AD2_0029(Mem *m, uint16_t s_off, uint16_t s_seg, uint16_t kind_word, uint16_t b_word)
{
    uint16_t idx = f_2AD2_0151(m);
    uint16_t rseg = rd16(m, DS(0xCFB6));
    uint16_t roff = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(idx * 0x4Eu));
    uint8_t kind = (uint8_t)kind_word;
    uint16_t row, w;

    w0029_wr8s(m, rseg, roff, rd8(m, LIN(s_seg, (uint16_t)(s_off + 1u))));
    w0029_wr8s(m, rseg, (uint16_t)(roff + 1u), rd8(m, LIN(s_seg, (uint16_t)(s_off + 3u))));
    w0029_wr8s(m, rseg, (uint16_t)(roff + 2u), rd8(m, LIN(s_seg, (uint16_t)(s_off + 6u))));
    w0029_wr8s(m, rseg, (uint16_t)(roff + 4u), kind);
    w0029_wr8s(m, rseg, (uint16_t)(roff + 3u), (uint8_t)b_word);
    w0029_wr16s(m, rseg, (uint16_t)(roff + 0x38u), 0);
    w0029_wr16s(m, rseg, (uint16_t)(roff + 0x3Au), 0);
    w0029_wr16s(m, rseg, (uint16_t)(roff + 0x10u), 0);
    w0029_wr16s(m, rseg, (uint16_t)(roff + 0x0Eu), 0);
    w0029_wr16s(m, rseg, (uint16_t)(roff + 0x14u), 0);
    w0029_wr16s(m, rseg, (uint16_t)(roff + 0x12u), 0);

    row = (uint16_t)(0x4E3Au + (uint16_t)(rd8(m, LIN(s_seg, (uint16_t)(s_off + 1u))) * 0xD0u) + (uint16_t)(kind * 0x1Au));
    w = w0029_rd16s(m, GIN3_DS, row);
    w0029_wr16s(m, rseg, (uint16_t)(roff + 0x0Au), w);

    if (kind == 7) {
        uint16_t row7 = (uint16_t)(0x4E3Au + (uint16_t)(rd8(m, LIN(rseg, roff)) * 0xD0u) + 0xB6u);

        w0029_wr16s(m, rseg, (uint16_t)(roff + 8u), w0029_rd16s(m, GIN3_DS, row7));
        w0029_wr16s(m, rseg, (uint16_t)(roff + 0x0Cu), w0029_rd16s(m, rseg, (uint16_t)(roff + 0x0Au)));
    } else {
        uint16_t t = f_2703_1433(m, w0029_rd16s(m, s_seg, (uint16_t)(s_off + 6u)),
                                 w0029_rd16s(m, s_seg, (uint16_t)(s_off + 1u)), kind_word);

        w0029_wr16s(m, rseg, (uint16_t)(roff + 8u), t);
        w0029_wr16s(m, rseg, (uint16_t)(roff + 0x0Cu), t);
    }
    f_2703_14CF(m, roff, rseg, w0029_rd16s(m, rseg, (uint16_t)(roff + 8u)));
    return idx;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2AD2_0029(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("2AD2_0029", 4, run)
