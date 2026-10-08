/*
 * GIN3PS.EXE 1930:15F4  戰略層：星域存量補給艦隊（飛彈與艦載機）
 *
 * 原版：far 呼叫，兩個 far 指標參數（位移在前、段在後）：艦隊記錄、星域記錄；無回傳值（描述檔 ret = none）。
 * 艦艇池基底在 ds:CFB4（位移）與 ds:CFB6（段）；艦種表在 ds:4E3A，列號 ＝ 陣營 × 208 ＋ 艦種 × 26。
 *
 * 逐條對應反組譯（IDA sub_27464）：
 *   for i ← 0；i < 艦隊[+0x0F](byte，每圈重讀)；i++：
 *     ship ← 池[艦隊[+0x10 + 4i](word)]                    ; 池基底每圈重讀
 *     列   ← 0x4E3A + low16(ship[+0x00] × 0xD0) + low16(ship[+0x04] × 0x1A)
 *     n    ← ship[+0x08](word)
 *     want1 ← 列[+0x10](byte) × n − ship[+0x0E..+0x11]     ; 32 位元，借位回繞（缺口為負時是接近 2^32 的大數）
 *     want2 ← 列[+0x14](byte) × n − ship[+0x12..+0x15]
 *     w     ← 星域[+0x12](byte) × 4
 *     stock1 ← 星域[+0x68 + w]（32 位元）；stock2 ← 星域[+0x70 + w]
 *     give1 ← stock1 < want1 ? stock1 : want1              ; 32 位元無號比較
 *     give2 ← stock2 < want2 ? stock2 : want2
 *     ship[+0x0E..] += give1；ship[+0x12..] += give2         ; 32 位元加（add 低字、adc 高字）
 *     w 重讀星域[+0x12] × 4；星域[+0x68 + w] −= give1         ; sub 低字、sbb 高字
 *     w 再重讀一次；  星域[+0x70 + w] −= give2
 *   星域[+0x12] 在寫入之前讀一次，兩次減法前各重讀一次（原版逐條如此），所以星域記錄與艦艇記錄重疊時三次讀到的值可以不同。
 *
 * 推論等級
 *   已證實（差異測試的位元組與寫入集合）：上述運算、欄位位移、三次讀星域 +0x12、無號比較與借位回繞（docs/re/349 §四「原版 bug」）。
 *   強推論（docs/re/349-1930-supply-family.md §二 的虛擬碼逐項相符；docs/internals/29-planets-and-economy.md、
 *     docs/internals/64-ship-pool.md）：星域 +0x12 是擁有者、+0x68／+0x70 是兩種補給存量、ship +0x0E／+0x12 是飛彈／艦載機存量、
 *     列 +0x10／+0x14 是每艘的飛彈數／艦載機數。名稱不是證據，只有位移與運算是證據。
 *   與文件的差異：docs/re/349 §二的虛擬碼沒有寫「星域 +0x12 在兩次減法前各重讀一次」，行為上等價於一次讀取，除非記錄別名。
 *
 * 呼叫的函式：3CC9:0009（32 位元乘）已 diff-tested，這裡用 gin3_lmul 表達同一個數學結果。
 *
 * 輸入域：艦隊指標與星域指標取自各自的表（描述檔 table）；艦隊 +0x0F 到 255、池索引任意、星域擁有者 byte 任意
 *   （w 最大 1020，讀寫星域記錄之後的記憶體，本版同樣照做）；描述檔另有一個星域指標與艦艇記錄別名的種子。
 *   池索引超出 352 筆時讀寫池之後的記憶體，本版同樣照做（平面視訊記憶體的讀取除外，見 docs/re/445 限制）。
 */
#include "gin3_reg.h"

static uint32_t rd32w(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint32_t)rd16(m, LIN(seg, off)) | ((uint32_t)rd16(m, LIN(seg, off + 2)) << 16);
}

static void add32w(Mem *m, uint16_t seg, uint16_t off, uint32_t v)
{
    uint32_t lo = (uint32_t)rd16(m, LIN(seg, off)) + (uint16_t)v;

    wr16(m, LIN(seg, off), (uint16_t)lo);
    wr16(m, LIN(seg, off + 2), (uint16_t)(rd16(m, LIN(seg, off + 2)) + (uint16_t)(v >> 16) + (lo >> 16)));
}

static void sub32w(Mem *m, uint16_t seg, uint16_t off, uint32_t v)
{
    uint16_t lo = rd16(m, LIN(seg, off));

    wr16(m, LIN(seg, off), (uint16_t)(lo - (uint16_t)v));
    wr16(m, LIN(seg, off + 2), (uint16_t)(rd16(m, LIN(seg, off + 2)) - (uint16_t)(v >> 16) - (lo < (uint16_t)v)));
}

void f_1930_15F4(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t reg_off, uint16_t reg_seg)
{
    uint8_t i = 0;
    uint16_t slot = (uint16_t)(fleet_off + 0x10);

    while (i < rd8(m, LIN(fleet_seg, fleet_off + 0x0F))) {
        uint16_t pseg = rd16(m, DS(0xCFB6));
        uint16_t ship = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(rd16(m, LIN(fleet_seg, slot)) * 0x4Eu));
        uint16_t row = (uint16_t)(0x4E3A + (uint16_t)(rd8(m, LIN(pseg, ship)) * 0xD0u) +
                                  (uint16_t)(rd8(m, LIN(pseg, ship + 0x04)) * 0x1Au));
        uint16_t n = rd16(m, LIN(pseg, ship + 0x08));
        uint32_t want1 = gin3_lmul(rd8(m, DS(row + 0x10)), n) - rd32w(m, pseg, (uint16_t)(ship + 0x0E));
        uint32_t want2 = gin3_lmul(rd8(m, DS(row + 0x14)), n) - rd32w(m, pseg, (uint16_t)(ship + 0x12));
        uint16_t w = (uint16_t)(rd8(m, LIN(reg_seg, reg_off + 0x12)) * 4u);
        uint32_t stock1 = rd32w(m, reg_seg, (uint16_t)(reg_off + 0x68 + w));
        uint32_t give1 = stock1 < want1 ? stock1 : want1;
        uint32_t stock2 = rd32w(m, reg_seg, (uint16_t)(reg_off + 0x70 + w));
        uint32_t give2 = stock2 < want2 ? stock2 : want2;

        add32w(m, pseg, (uint16_t)(ship + 0x0E), give1);
        add32w(m, pseg, (uint16_t)(ship + 0x12), give2);
        w = (uint16_t)(rd8(m, LIN(reg_seg, reg_off + 0x12)) * 4u);
        sub32w(m, reg_seg, (uint16_t)(reg_off + 0x68 + w), give1);
        w = (uint16_t)(rd8(m, LIN(reg_seg, reg_off + 0x12)) * 4u);
        sub32w(m, reg_seg, (uint16_t)(reg_off + 0x70 + w), give2);
        i++;
        slot = (uint16_t)(slot + 4);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1930_15F4(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("1930_15F4", 4, run)
