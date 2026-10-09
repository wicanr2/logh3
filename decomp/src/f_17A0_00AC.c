/*
 * GIN3PS.EXE 17A0:00AC  在場分隊名冊裡，陣營相符、沒出局、距離 (x, y) 不超過門檻的艦艇：池索引依序寫進輸出表，補 0xFFFF 結尾，回命中數
 *
 * 原版（192 bytes，far 呼叫，六個 word 堆疊參數；序言 push bp／si／di、sub sp,10h、mov bp,sp，所以第一個參數在 [bp+1Ah]）：
 *   faction ＝ word [bp+1Ah]（只比低位元組）、x ＝ [bp+1Ch]、y ＝ [bp+1Eh]、thr ＝ [bp+20h]、out ＝ far 指標（位移 [bp+22h]、段 [bp+24h]）。
 *   回傳 AX ＝ 命中數（描述檔 ret ＝ ax）。區域變數：[bp+0] ＝ i、[bp+2] ＝ 命中數、[bp+4]／[bp+6] ＝ 名冊指標（位移／段，段 ＝ DS）、
 *   [bp+8]／[bp+0Ah] ＝ thr << 16（低字 0、高字 thr）、[bp+0Ch]／[bp+0Eh] ＝ 艦艇記錄指標（只存不讀）。
 * 過程（已證實：反組譯逐指令；3310:0006、3CC9:00B9 由各自的差異測試證實）：
 *   i ← 0；名冊指標 ← DS:0x98BA；命中 ← 0
 *   重複：i >= word ds:9896（無號 jnb，筆數每圈重讀）→ 結束
 *     idx ← word 名冊指標[0]；rec ← (word ds:CFB4) ＋ low16(idx × 0x4E)，段 ← word ds:CFB6（段、位移每圈重讀）
 *     byte rec[0] ≠ byte faction → 跳過
 *     byte rec[+0x1E] 的 bit7 有設 → 跳過
 *     d ← 3310:0006(x, y, rec[+0x34], rec[+0x36])             ; 16.16 距離（推入順序：先 +0x36、+0x34，再 y、x；只有讀取，順序無可觀察差異）
 *     3CC9:00B9(d, thr << 16) 之後 jg（ZF ＝ 0 且 SF ＝ OF，有號 32 位元 d > thr << 16）→ 跳過
 *     out 的段與目前位移都是 0（or ax,[bp+22h]；jz）→ 不寫；否則 word out[0] ← 名冊指標[0]（重新讀一次），out 的位移 ＋= 2（16 位元，段不變）
 *     命中 ＋= 1                                              ; 不論有沒有寫
 *     i ＋= 1；名冊指標位移 ＋= 2
 *   結束時 word out[0] ← 0xFFFF（out 位移是目前值；out 為 0:0 時寫線性位址 0），回命中數
 * 語意（強推論）：docs/re/90-battle-engagement-pipeline.md 第 39 到 58 行（名冊 ds:98BA、筆數 ds:9896、三個條件、補 0xFFFF 結尾，回命中數）、
 *   docs/re/403-weapon-candidates.md 第 30 行、docs/re/290-battle-spotting-solved.md 第 117 行、docs/re/145-fortress-gun-aim-and-fire.md 第 192、193 行，
 *   與本檔逐項相符。補充（docs 沒寫）：陣營只比低位元組；門檻是 thr << 16 的有號 32 位元（thr >= 0x8000 時為負，距離 >= 0 者全被擋掉）；
 *   與 3310:0006 回傳的 16.16 距離同一個比例；比較是有號，距離的高字 >= 0x8000（座標差 >= 0x8000）時距離為負，反而被收下；
 *   輸出指標為 0:0 時不寫索引、仍計數，結尾的 0xFFFF 照寫到線性位址 0。
 *   +0x1E bit7 是 docs/re/132-scout-order.md 第 96 行所說「跳過 bit7 設起來的（偵察隊）」。
 * 副作用：輸出表（索引與結尾 0xFFFF）；3310:0006 內 39B5:005C 的 ds:8EDC..8EED 暫存。
 *
 * 機器模型：C 版沒有堆疊。word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 的 read16／write16），用本檔的 w00AC_rd16s／w00AC_wr16s。
 *   寫入落在線性 0xF0000 以上（ROM）時原版安靜丟棄，用 w00AC_wr8（逐 byte）表達。
 * 輸入域：名冊筆數 ds:9896 取 seeds 與突變（超過探針步數上限的案例由探針丟棄）；輸出表指向暫存區、DGROUP 或 seeds 給的位址，不指向堆疊帶與 A0000 到 BFFFF，
 *   輸出蓋到呼叫堆疊的情形不涵蓋。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

#define F00AC_ZF 0x0040u
#define F00AC_SF 0x0080u
#define F00AC_OF 0x0800u

static uint16_t w00AC_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w00AC_wr8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w00AC_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w00AC_wr8(m, LIN(seg, off), (uint8_t)v);
    w00AC_wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_17A0_00AC(Mem *m, uint16_t faction, uint16_t x, uint16_t y, uint16_t thr, uint16_t out_off, uint16_t out_seg)
{
    uint16_t i = 0;
    uint16_t count = 0;
    uint16_t lp = 0x98BAu;

    while (i < w00AC_rd16s(m, GIN3_DS, 0x9896u)) {
        uint16_t idx = w00AC_rd16s(m, GIN3_DS, lp);
        uint16_t pseg = w00AC_rd16s(m, GIN3_DS, 0xCFB6u);
        uint16_t rec = (uint16_t)(w00AC_rd16s(m, GIN3_DS, 0xCFB4u) + (uint16_t)(idx * 0x4Eu));

        if (rd8(m, LIN(pseg, rec)) == (uint8_t)faction && !(rd8(m, LIN(pseg, (uint16_t)(rec + 0x1Eu))) & 0x80u)) {
            uint16_t x2 = w00AC_rd16s(m, pseg, (uint16_t)(rec + 0x34u));
            uint16_t y2 = w00AC_rd16s(m, pseg, (uint16_t)(rec + 0x36u));
            uint32_t d = f_3310_0006(m, x, y, x2, y2);
            uint16_t f = f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, thr, 0);
            int jg = !(f & F00AC_ZF) && (((f & F00AC_SF) != 0) == ((f & F00AC_OF) != 0));

            if (!jg) {
                if ((uint16_t)(out_seg | out_off) != 0) {
                    w00AC_wr16s(m, out_seg, out_off, w00AC_rd16s(m, GIN3_DS, lp));
                    out_off = (uint16_t)(out_off + 2u);
                }
                count = (uint16_t)(count + 1u);
            }
        }
        i = (uint16_t)(i + 1u);
        lp = (uint16_t)(lp + 2u);
    }
    w00AC_wr16s(m, out_seg, out_off, 0xFFFFu);
    return count;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_17A0_00AC(m, a[0], a[1], a[2], a[3], a[4], a[5]);
}
GIN3_REGISTER("17A0_00AC", 6, run)
