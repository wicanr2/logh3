/*
 * GIN3PS.EXE 2B5F:02B2  每回合的星域補給回充（飛彈與艦載機存量）
 *
 * 原版：far 呼叫，沒有參數。星域表 ds:A79C 起 32 筆 × 126 bytes（0x7E）。對每一筆 k ＝ 0..31：
 *   星域[+0x12] == 2（無主）                → 跳過
 *   星域[+0x13] 的 bit0 為 1（路徑成本過高）  → 跳過
 *   上限 S ＝ 星域[+0x1A..+0x1D]（32 位元，經濟力）＋ 星域[+0x20]（word，防衛力零延伸），32 位元相加，進位溢出丟棄
 *   補量 Q ＝ S ÷ 10（無號 32 位元除法，3CC9:0020）
 *   星域[+0x68 + 4 × 星域[+0x12]]（32 位元，飛彈）← min(原值 + Q, S)
 *   星域[+0x70 + 4 × 星域[+0x12]]（32 位元，艦載機）← min(原值 + Q, S)
 *   min 是無號 32 位元比較（先比高字，高字相同再比低字）；原值 + Q 的進位溢出丟棄。
 * 兩個存量格都是「擁有者（+0x12）」索引的 dword；第二段會重讀 +0x12（原版的行為，第一段寫入若改到 +0x12
 *   會影響第二段，正常資料不會發生）。+0x12 不是 0／1／2 以外的值時遊戲資料不會出現，本版與原版一樣照寫。
 * 語意見 docs/spec/039、docs/re/273、docs/internals/27（已證實：反組譯逐指令）。
 * 回傳：沒有回傳值（離開時 AX 是最後一輪的殘值，唯一的呼叫者 0654:0007 之後馬上改寫 AX，不使用），不比較。
 * 呼叫：3CC9:0020（無號除法）。副作用：寫星域的 +0x68／+0x70 兩個 dword。輸入域：全部可到達（記憶體任意內容）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

static uint32_t rd32w(const Mem *m, uint16_t off)
{
    return (uint32_t)rd16(m, DS(off)) | ((uint32_t)rd16(m, DS((uint16_t)(off + 2))) << 16);
}

static void wr32w(Mem *m, uint16_t off, uint32_t v)
{
    wr16(m, DS((uint16_t)(off + 2)), (uint16_t)(v >> 16));
    wr16(m, DS(off), (uint16_t)v);
}

void f_2B5F_02B2(Mem *m)
{
    uint16_t region = 0xA79C;
    uint8_t i;

    for (i = 0; i < 0x20; i++) {
        if (rd8(m, DS((uint16_t)(region + 0x12))) != 2 && (rd8(m, DS((uint16_t)(region + 0x13))) & 1) == 0) {
            uint32_t limit = rd32w(m, (uint16_t)(region + 0x1A)) + rd16(m, DS((uint16_t)(region + 0x20)));
            uint32_t refill = gin3_uldiv(limit, 10);
            uint16_t slot;
            uint32_t value;

            slot = (uint16_t)(region + 0x68 + 4u * rd8(m, DS((uint16_t)(region + 0x12))));
            value = rd32w(m, slot) + refill;
            if (value > limit)
                value = limit;
            wr32w(m, slot, value);

            slot = (uint16_t)(region + 0x70 + 4u * rd8(m, DS((uint16_t)(region + 0x12))));
            value = rd32w(m, slot) + refill;
            if (value > limit)
                value = limit;
            wr32w(m, slot, value);
        }
        region = (uint16_t)(region + 0x7E);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2B5F_02B2(m);
    return 0;
}
GIN3_REGISTER("2B5F_02B2", 0, run)
