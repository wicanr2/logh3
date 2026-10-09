/*
 * GIN3PS.EXE 3870:08FF  Cohen–Sutherland 線段裁切：把線段兩端點推進裁切矩形，回 1（可見）或 0（整段在矩形外）
 *
 * 原版：far 呼叫，四個 far 指標參數（序言 push bp、sub sp,2、mov bp,sp，沒有 push si／di，所以第一個參數在 [bp+8]）：
 *   args[0..1] ＝ XA（端點 A 的 x，word 的位址）  args[2..3] ＝ YA（A 的 y）  args[4..5] ＝ XB（B 的 x）  args[6..7] ＝ YB（B 的 y）
 *   回傳 AX ＝ 1 或 0（`mov ax,1`／`xor ax,ax`）。區域變數兩個 byte：[bp+0] ＝ A 的 outcode c0、[bp+1] ＝ B 的 outcode c1。
 * 過程（已證實：反組譯逐指令）：
 *   c0 ← AL of 3870:09F1(*XA, *YA)；c1 ← AL of 3870:09F1(*XB, *YB)
 *   c0 & c1 ≠ 0 → 回 0                                       ; 兩端在矩形同一側之外，整段不可見
 *   迴圈：
 *     c0 ＝ 0 且 c1 ＝ 0 → 回 1
 *     c0 ≠ 0：c0 ← AL of 3870:0A30(code ＝ word [bp+0], A 的兩個指標 XA／YA, c ＝ *XB, d ＝ *YB)   ; 把 A 沿線段推到邊界
 *     c0 ＝ 0 且 c1 ＝ 0 → 回 1
 *     c0 & c1 ≠ 0 → 回 0
 *     c1 ≠ 0：c1 ← AL of 3870:0A30(code ＝ word [bp+1], B 的兩個指標 XB／YB, c ＝ *XA, d ＝ *YA)   ; 把 B 推到邊界
 *     c0 & c1 ≠ 0 → 回 0；否則回到迴圈開頭
 *   0A30 的 c、d 參數在每次呼叫前才從指標重讀（前一步可能已改過 *XA、*YA）。0A30 的 code 參數是 word（c0 | 下一個 byte << 8），
 *   0A30 只測低位元組的 bit 3／2／0／1，高位元組無影響；本版仍照原版組出 word。
 * 副作用：只經由 3870:0A30 寫兩端點的 x、y（四個指標指到的 word）；本函式自己不寫記憶體。指標彼此別名或指到矩形 ds:8D6A..8D70 時，
 *   後面的讀取看得到前面的寫入（逐步照做）。
 * 語意（強推論）：矩形 ds:8D6A..8D70 是 3870:089D 設定的裁切矩形（docs/internals/52 第 83 行「畫線之前先裁切端點」、docs/re/155 第 68 行、
 *   docs/re/135 第 111 行）；呼叫者 3870:07C9、3870:0833。
 * 機器模型：*P 是 word 存取，位移 0xFFFF 時高位元組取同段位移 0（dosgolem 的 read16），用本檔的 rd16s。
 * 輸入域：指標指到暫存區、DGROUP 或 seeds 給的位址；不指向堆疊帶、A0000 到 BFFFF 與 ROM 區（寫入在 ROM 時原版可能不終止）。
 *   迴圈在原版若不終止（整數溢位使端點推不進矩形）超過步數上限，探針丟棄該案，C 版不涵蓋；C 版迴圈以 4096 圈為上限（超過回 0xFFFF），
 *   原版每圈至少數百步，在 200000 步上限內到不了，所以上限只影響突變版本，不影響收錄的案例。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_3870_09F1(Mem *m, uint16_t x, uint16_t y);
extern uint16_t f_3870_0A30(Mem *m, uint16_t code, uint16_t p1_off, uint16_t p1_seg, uint16_t p2_off, uint16_t p2_seg,
                            uint16_t c, uint16_t d);

static uint16_t f3870_08FF_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_3870_08FF(Mem *m, uint16_t xa_off, uint16_t xa_seg, uint16_t ya_off, uint16_t ya_seg, uint16_t xb_off,
                     uint16_t xb_seg, uint16_t yb_off, uint16_t yb_seg)
{
    uint8_t c0, c1;
    unsigned guard;

    c0 = (uint8_t)f_3870_09F1(m, f3870_08FF_rd16s(m, xa_seg, xa_off), f3870_08FF_rd16s(m, ya_seg, ya_off));
    c1 = (uint8_t)f_3870_09F1(m, f3870_08FF_rd16s(m, xb_seg, xb_off), f3870_08FF_rd16s(m, yb_seg, yb_off));
    if (c0 & c1)
        return 0;
    for (guard = 0; guard < 4096u; guard++) {
        if (c0 == 0 && c1 == 0)
            return 1;
        if (c0 != 0) {
            uint16_t c = f3870_08FF_rd16s(m, xb_seg, xb_off);
            uint16_t d = f3870_08FF_rd16s(m, yb_seg, yb_off);

            c0 = (uint8_t)f_3870_0A30(m, (uint16_t)(c0 | ((uint16_t)c1 << 8)), xa_off, xa_seg, ya_off, ya_seg, c, d);
        }
        if (c0 == 0 && c1 == 0)
            return 1;
        if (c0 & c1)
            return 0;
        if (c1 != 0) {
            uint16_t c = f3870_08FF_rd16s(m, xa_seg, xa_off);
            uint16_t d = f3870_08FF_rd16s(m, ya_seg, ya_off);

            c1 = (uint8_t)f_3870_0A30(m, c1, xb_off, xb_seg, yb_off, yb_seg, c, d);
        }
        if (c0 & c1)
            return 0;
    }
    return 0xFFFFu; /* 迭代上限 4096：原版在 200000 步內不可能走到這裡（每圈至少數百步），只為了讓突變版本不會無窮迴圈 */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3870_08FF(m, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7]);
}
GIN3_REGISTER("3870_08FF", 8, run)
