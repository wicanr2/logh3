/*
 * GIN3PS.EXE 2E54:1876  一串艦隊的戰力總和（far，四個參數，回 AX:BX = 32 位元總和，高字在 AX、低字在 BX，不寫記憶體）
 *
 * 參數（堆疊字，由低位址到高位址）：
 *   region   一個字：星域編號。函式沒有讀它（asm 全程沒有碰 [bp+1Ch]）
 *   list     far 指標（位移、段）：艦隊索引清單，每筆 1 個 byte；讀完一筆位移加 1（16 位元回繞）
 *   count    一個字，只用低位元組：清單筆數（cmp al,[bp+22h] 是 byte 比較，無號 jb）
 *   faction  一個字：陣營。函式沒有讀它（asm 沒有碰 [bp+24h]）
 *
 * 算式：
 *   對清單的每一筆索引 idx（i = 0 … count-1）：
 *     fleet = ds:C074 + idx × 122（16 位元位移回繞）
 *     fleet[+0x00] bit0 有設 → 整支跳過
 *     n = fleet[+0x0F]；對 j = 0 … n-1（byte 計數，無號比較）：
 *       sidx = word @ fleet[+0x10 + j × 4]
 *       ship = (word ds:CFB6):(word ds:CFB4 + sidx × 78)   ； 艦艇池基底每次重讀，sidx × 78 取 16 位元
 *       w = byte ds:6032[ ship[+0x04] ]                    ； 權重表，索引是 byte
 *       總和 += ((w × ship[+0x08]) × 1000) ÷ ship[+0x0A]    ； 乘用 3CC9:0009（取低 32 位元），除用 3CC9:0020（無號，除數 0 → 0xFFFFFFFF）
 *   總和是 32 位元，加法 modulo 2^32。
 *
 * 推論等級
 *   已證實（差異測試，含乘積取低 32 位元與除以 0 的情形）：上列算式、位移、常數。
 *   強推論（docs/re/79、163、internals/33、54）：艦艇[+0x04] = 艦種（0..7）、[+0x08] = 現有艘數、[+0x0A] = 滿編艘數；
 *     權重表 ds:6032 = 4 3 3 3 2 1 3 0；fleet[+0x00] bit0 = 無法戰鬥；fleet[+0x0F] = 分隊數。
 * 輸入域：list 取一塊暫存緩衝區（內容 0/0xFF/任意 byte；fleet 索引 0..255 全域，>= 32 時讀到艦隊表之後的記憶體，原版同樣如此），
 *   count 低位元組 0..40（艦隊表 32 筆，一個陣營最多 16 支；更大的 count 會讓執行步數超過探針上限而被丟棄），
 *   高位元組任意；region、faction 任意。fleet[+0x0F] 與艦艇池內容由突變決定，0..255 全域。
 */
#include "gin3_reg.h"

uint32_t f_2E54_1876(Mem *m, uint16_t region, uint16_t list_off, uint16_t list_seg, uint16_t count_w, uint16_t faction)
{
    uint32_t sum = 0;
    uint16_t lp = list_off;
    uint8_t count = (uint8_t)count_w;
    uint8_t i;

    (void)region;
    (void)faction;
    for (i = 0; i < count; i++) {
        uint8_t idx = rd8(m, LIN(list_seg, lp));
        uint16_t fleet = (uint16_t)(0xC074u + idx * 122u);
        uint8_t j;

        lp = (uint16_t)(lp + 1u);
        if (rd8(m, DS(fleet)) & 1u)
            continue;
        for (j = 0; j < rd8(m, DS(fleet + 0x0Fu)); j++) {
            uint16_t sidx = rd16(m, DS((uint16_t)(fleet + 0x10u + j * 4u)));
            uint16_t pool_off = rd16(m, DS(0xCFB4));
            uint16_t pool_seg = rd16(m, DS(0xCFB6));
            uint16_t ship = (uint16_t)(pool_off + (uint16_t)(sidx * 78u));
            uint8_t kind = rd8(m, LIN(pool_seg, ship + 4u));
            uint32_t weight = rd8(m, DS((uint16_t)(0x6032u + kind)));
            uint32_t cur = rd16(m, LIN(pool_seg, ship + 8u));
            uint32_t full = rd16(m, LIN(pool_seg, ship + 0x0Au));
            uint32_t prod = gin3_lmul(gin3_lmul(weight, cur), 1000u);

            sum += gin3_uldiv(prod, full);
        }
    }
    return sum;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2E54_1876(m, a[0], a[1], a[2], a[3], a[4]); }
GIN3_REGISTER("2E54_1876", 5, run)
