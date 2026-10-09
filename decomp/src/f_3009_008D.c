/*
 * GIN3PS.EXE 3009:008D  戰場電腦座位的防僵局規則：在場敵方分隊「一個都看不到」時，把敵方全部分隊標成可見
 * （far，無參數，無回傳值；ret = none。AX 在 retf 時是殘值，不是輸出）
 *
 * 反組譯逐條對應（IDA sub_3CC8D）：
 *   found ← 0
 *   第一趟（只讀）：對 i = 0 … 直到 i >= byte ds:9894 或 found（byte 比較，每圈重讀 ds:9894）：
 *     fl ← ds:C074 + (byte ds:9898[i]) × 0x7A                 ; 16 位元回繞；清單 ds:9898 每筆 1 byte 艦隊編號
 *     艦隊[+0x01] != ds:8F1D（byte）才處理（本方艦隊跳過）：
 *       對 j = 0 … 直到 j >= 艦隊[+0x0F] 或 found（byte，每圈重讀）：
 *         sidx ← word 艦隊[+0x10 + j × 4]                      ; 位置 16 位元回繞
 *         艦艇 ← (ds:CFB6 : ds:CFB4 + low16(sidx × 0x4E))      ; mul cx 取 AX；池基底每圈重讀
 *         艦艇[+0x17] != 0 → found ← 1
 *   found == 0 才做第二趟（寫入）：同樣的雙層迴圈、同樣的敵方篩選，但沒有 found 的提前結束，對每艘：
 *     艦艇[+0x16] ← 艦艇[+0x16] | 1；艦艇[+0x17] ← 3；艦艇[+0x18] ← 6
 *   找到任何一艘 +0x17 非 0 的敵方分隊就什麼都不寫。
 *   第一趟沒有副作用，found 的提前結束只影響讀取次數，輸出看不出來；本檔用同一個掃描函式 w3009_scan(write) 表達兩趟。
 *
 * 推論等級
 *   已證實（差異測試的位元組與寫入集合）：上列兩趟結構、欄位位移、常數 3 與 6、OR 1、byte 計數與 16 位元回繞。
 *   強推論（docs/internals/38 第 283、323 行）：ds:9898／ds:9894 是在場艦隊清單與筆數（byte）；艦隊[+0x01] 是陣營、
 *     [+0x0F] 是分隊數、[+0x10 + 4j] 是分隊的池索引；艦艇[+0x16] bit0 是「本輪被看到」、[+0x17] 是索敵等級、
 *     [+0x18] 是等級的倒數計時（給 6 讓它撐六個回合）。名稱不是證據，只有位移與常數是證據。
 *   未知：呼叫端 3009:0003 的條件（docs/internals/38 說電腦座位且 ds:8F1A >= 6，本函式不檢查這兩項）。
 *
 * 輸入域：ds:9894 取 byte 0..255（高位元組 ds:9895 不讀）；艦隊編號 0..255（>= 32 時讀到艦隊表之後的記憶體，
 *   位移 16 位元回繞，原版同樣如此）；艦隊[+0x0F] 與池索引取全域。每圈約 30 道指令，總步數超過 200,000 的案例被探針丟棄。
 *   線性 0xF0000 以上是 ROM，寫入在 dosgolem 被安靜丟棄（internal/machine/bios_rom.go），用 w3009_put8 表達。
 *   讀到 A0000 到 BFFFF 的案例由探針丟棄。
 */
#include "gin3_reg.h"

/* 段內 word 讀取：位移各自 16 位元回繞（seg:FFFF 的高位元組在 seg:0000）。 */
static uint16_t w3009_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效。 */
static void w3009_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

/* write == 0：第一趟，回傳是否有任何敵方艦艇 +0x17 非 0；write != 0：第二趟，寫入。 */
static int w3009_scan(Mem *m, int write)
{
    int found = 0;
    uint16_t i;
    uint16_t lp = 0x9898u;

    for (i = 0; i < rd8(m, DS(0x9894)) && !found; i++, lp = (uint16_t)(lp + 1u)) {
        uint16_t fl = (uint16_t)(0xC074u + (uint16_t)(rd8(m, DS(lp)) * 0x7Au));

        if (rd8(m, DS((uint16_t)(fl + 1u))) != rd8(m, DS(0x8F1D))) {
            uint16_t j;
            uint16_t up = (uint16_t)(fl + 0x10u);

            for (j = 0; j < rd8(m, DS((uint16_t)(fl + 0x0Fu))) && !found; j++, up = (uint16_t)(up + 4u)) {
                uint16_t pseg = rd16(m, DS(0xCFB6));
                uint16_t unit = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(w3009_rd16s(m, GIN3_DS, up) * 0x4Eu));

                if (!write) {
                    if (rd8(m, LIN(pseg, unit + 0x17u)) != 0)
                        found = 1;
                } else {
                    w3009_put8(m, LIN(pseg, unit + 0x16u), (uint8_t)(rd8(m, LIN(pseg, unit + 0x16u)) | 1u));
                    w3009_put8(m, LIN(pseg, unit + 0x17u), 3);
                    w3009_put8(m, LIN(pseg, unit + 0x18u), 6);
                }
            }
        }
    }
    return found;
}

void f_3009_008D(Mem *m)
{
    if (!w3009_scan(m, 0))
        (void)w3009_scan(m, 1);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_3009_008D(m);
    return 0;
}
GIN3_REGISTER("3009_008D", 0, run)
