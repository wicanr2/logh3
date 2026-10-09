/*
 * GIN3PS.EXE 2BE1:050F  把艦隊每個編制組所屬的「分組」標成已用：ds:D072 + 組別 × 10 + 4 ← 1
 *
 * 原版：far 呼叫，一個 far 指標參數＝艦隊記錄（push bp、push si、sub sp,6，所以參數在 [bp+0Eh] 與 [bp+10h]），
 *   無回傳值（AX 是殘留值，呼叫者 2BE1:0328 不用；描述檔 ret=none）。
 * 過程（i 是 byte 計數器，從 0 開始，組指標從艦隊 +0x10 起每組 4 bytes）：
 *   i >= 艦隊[+0x0F]（無號 byte 比較；每圈用堆疊上的指標重讀，所以寫入若蓋到艦隊記錄自己，下一圈看到新值）→ 結束
 *   id ← 組[+3]（byte）
 *   ds:(0xD072 + id × 10 + 4) ← 1（byte；id × 10 是 16 位元 mul，最大 2550；位移 16 位元運算）
 *   i++，指標 += 4（16 位元回繞）
 * 寫入位址在 DGROUP（DS = 54B3）內，最大 ds:D072 + 2550 + 4 = ds:DA6C，不會到 0xF0000 以上。
 *
 * 欄位語意（強推論）：艦隊 +0x0F ＝ 編制組數、組[+3] ＝ 隊形槽（docs/re/229 第 32 到 38 行）；ds:D072 是 10 bytes 一筆的分組表，
 *   +4 的 byte 是「已有人」旗標（docs/re/130 第 20、63 行：「已經有人的格子標成 1」）。位移已證實，名稱是強推論。
 * 輸入域：艦隊指標任意（艦隊表任一筆、暫存緩衝區或與 ds:D072 表重疊的別名）；組別 id 全部 byte 值。
 * 證據：asm 逐指令。推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

void f_2BE1_050F(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint8_t i = 0;
    uint16_t p = (uint16_t)(fleet_off + 0x10);

    while (i < rd8(m, LIN(fleet_seg, fleet_off + 0x0F))) {
        uint8_t id = rd8(m, LIN(fleet_seg, p + 3));
        uint16_t si = (uint16_t)(0xD072u + (uint16_t)(id * 10u));

        wr8(m, DS(si + 4), 1);
        i++;
        p = (uint16_t)(p + 4);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2BE1_050F(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2BE1_050F", 2, run)
