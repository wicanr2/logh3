/*
 * GIN3PS.EXE 203A:04D1  行星附近的守軍量：行星查詢半徑內、陣營相符的艦艇池項，只加艦種 0／1 的 word（+0x08），飽和在 65535，再除以 100
 *
 * 原版（194 bytes，far 呼叫，一個 far 指標參數 P（行星，位移 [bp+2D8h]、段 [bp+2DAh]）；序言 push bp／si／di、sub sp,2CEh、mov bp,sp），
 *   回傳 AX ＝ 加總 ÷ 100（無號 16 位元除法，被除數 <= 65535，不會 #DE；描述檔 ret ＝ ax）。
 *   區域變數：[bp+0] ＝ i、[bp+2] ＝ 加總（word）、[bp+4] 起 0x2C0 bytes ＝ 17A0:00AC 的輸出表（word 池索引，結尾 0xFFFF）、
 *   [bp+2C4h] ＝ 2AF5:01F7 的結果、[bp+2C6h]／[bp+2C8h] ＝ 艦艇記錄指標（位移／段，只存不讀）、[bp+2CCh] ＝ 該項的 word（只存不讀）。
 * 過程（已證實：反組譯逐指令；2AF5:01F7、17A0:00AC 由其各自的差異測試證實）：
 *   ds:A764 ← P 的位移；ds:A766 ← P 的段                         ; 兩個全域字（用途未知，只寫）
 *   thr ← 2AF5:01F7(P)                                              ; 行星的查詢半徑（60..240）
 *   n ← 17A0:00AC(faction ＝ word P[+0x11], x ＝ P[+0x46], y ＝ P[+0x48], thr, far SS:bp+4)  ; 推入順序：P[+0x48]、P[+0x46]、P[+0x11] 在最後
 *   sum ← 0；對 i ＝ 0 .. n−1（無號 jnb）：
 *     idx ← 輸出表[i]；rec ← DS:(CFB4) ＋ low16(idx × 0x4E)，段 ← word DS:CFB6
 *     byte rec[+4] ＝ 0 或 1 → v ← word rec[+8]；sum ← sum ＋ v，若進位則 sum ← 0xFFFF（add／adc／cmp ax,1，ax ＝ 進位）
 *   回 sum ÷ 100
 * 語意（強推論）：docs/spec/110-battle-finish.md 第 72 行（守軍量：僅艦種 0／1 艘數，加總飽和 65535 後除 100）、docs/spec/019-planet-assault.md 第 107 到 113、127 到 135 行、
 *   docs/internals/53-strategic-planet-assault.md 第 110 行、docs/re/424-battle-finish-attack.md 第 26 行（17A0:00AC 返回符合條件的池筆數；203A:04D1 才是守軍艘數消費端），與本檔一致。
 *   rec[+4] 為艦種、rec[+8] 為該分隊艘數（強推論，同 docs）。
 * 副作用：ds:A764、ds:A766；17A0:00AC 的輸出表（堆疊，比對排除）；3310:0006 內 39B5:005C 的暫存（ds:8EDC..8EED）。
 *
 * 機器模型：C 版沒有堆疊。輸出表放在原版的實際位置（基底狀態 fcmd-base.state：SS 固定 0x63EB，進入點 SP ＝ 1F8A，探針推 2 個參數字 4 bytes、返回位址 4、
 *   push bp／si／di 6、sub sp,2CEh，所以 bp ＝ 1CAE，輸出表從 SS:1CB2 起），交給 f_17A0_00AC 寫、本檔讀。表尾到 SS:1F72，在堆疊帶（SS:1B8A..1F89）內，雙方排除；
 *   39B5 的 C 版暫存在 SS:1F82..1F89，不與輸出表重疊。
 * 輸入域：P 取行星表內指標；名冊筆數 ds:9896 列入 freeze 由 seeds 給定（命中數超過 351 時輸出表會蓋到原版的區域變數，C 版不涵蓋）；池基底 ds:CFB4..CFB7 列入 freeze。
 *   名冊其餘內容、艦艇記錄與行星欄位取記憶體與突變。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_2AF5_01F7(Mem *m, uint16_t planet_off, uint16_t planet_seg);
extern uint16_t f_17A0_00AC(Mem *m, uint16_t faction, uint16_t x, uint16_t y, uint16_t thr, uint16_t out_off, uint16_t out_seg);

#define F04D1_SS 0x63EBu /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define F04D1_LIST 0x1CB2u /* 進入點 SP 1F8A − 4 − 4 − 6 − 2CEh ＝ bp(1CAE)，輸出表在 bp+4 */

static uint16_t w04D1_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_203A_04D1(Mem *m, uint16_t p_off, uint16_t p_seg)
{
    uint16_t thr, n, i, sum = 0;
    uint16_t faction, px, py;

    wr16(m, DS(0xA766u), p_seg);
    wr16(m, DS(0xA764u), p_off);
    thr = f_2AF5_01F7(m, p_off, p_seg);
    faction = w04D1_rd16s(m, p_seg, (uint16_t)(p_off + 0x11u));
    px = w04D1_rd16s(m, p_seg, (uint16_t)(p_off + 0x46u));
    py = w04D1_rd16s(m, p_seg, (uint16_t)(p_off + 0x48u));
    n = f_17A0_00AC(m, faction, px, py, thr, F04D1_LIST, F04D1_SS);
    for (i = 0; i < n; i++) {
        uint16_t idx = w04D1_rd16s(m, F04D1_SS, (uint16_t)(F04D1_LIST + 2u * i));
        uint16_t seg = w04D1_rd16s(m, GIN3_DS, 0xCFB6u);
        uint16_t rec = (uint16_t)(w04D1_rd16s(m, GIN3_DS, 0xCFB4u) + (uint16_t)(idx * 0x4Eu));
        uint8_t type = rd8(m, LIN(seg, (uint16_t)(rec + 4u)));

        if (type == 0 || type == 1) {
            uint16_t v = w04D1_rd16s(m, seg, (uint16_t)(rec + 8u));
            uint32_t s = (uint32_t)sum + v;

            sum = (s > 0xFFFFu) ? 0xFFFFu : (uint16_t)s;
        }
    }
    return (uint16_t)(sum / 100u);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_203A_04D1(m, a[0], a[1]); }
GIN3_REGISTER("203A_04D1", 2, run)
