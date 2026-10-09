/*
 * GIN3PS.EXE 17A0:06D7  這個點被誰佔著：在場分隊清單裡，不屬於指定陣營、不是艦種 7、沒有 bit7 旗標，
 *   而且位置離 (x, y) 夠近的第一艘艦艇的池索引 ＋ 1；沒有就回 0xFFFF
 *
 * 原版：far 呼叫，三個堆疊字（序言 push bp／si／di、sub sp,0Ah，所以 x 在 [bp+14h]、y 在 [bp+16h]、faction 在 [bp+18h]），回 AX。
 * 逐條對應反組譯：
 *   i ← 0（word）、清單位移 ← 0x98BA（DS 段）
 *   迴圈：
 *     i >= word ds:9896（無號，jnb）→ 回 0xFFFF
 *     idx  ← word ds:[清單位移]
 *     艦艇 ← (word ds:CFB4 ＋ low16(idx × 0x4E)) : word ds:CFB6          ; 池基底每圈重讀
 *     byte 艦艇[+0] == faction 的低位元組 → 下一個（跳過）                 ; cmp al,[bp+18h] 是 byte 比較
 *     byte 艦艇[+4] == 7                                   → 下一個
 *     byte 艦艇[+0x1E] 的 bit7 有設                         → 下一個
 *     17A0:076B(x, y, word 艦艇[+0x34], word 艦艇[+0x36]) ≠ 0（x、y 距離都小於 16）→ 回 (重讀的 idx) ＋ 1（16 位元）
 *     i++；清單位移 += 2
 * 本函式不寫記憶體（076B 也沒有）。
 *
 * 呼叫者：2583:040C（戰場移動「撞到誰」的判定，docs/re/102 第 54 行「r ＝ 撞到誰(x, y, 陣營)」、docs/re/124 第 11 行
 *   「17A0:06D7(x, y, 陣營) 這個點有沒有被佔著」）。與 docs/re/124 的描述一致，補上：比較的是 byte（陣營高位元組被忽略）、
 *   艦種 7 與 +0x1E bit7 的艦艇不算、回傳是池索引 ＋ 1 而不是索引本身。
 *
 * 推論等級
 *   已證實（差異測試）：清單基底與筆數、池基底、艦艇大小、三個跳過條件的欄位與寬度、距離判定的參數次序、回傳值。
 *   強推論（docs/re/124、internals/38）：艦艇 +0 是陣營、+4 是艦種、+0x34／+0x36 是座標、+0x1E bit7 是「已離場／不佔格」類旗標
 *     （名稱為假說）、清單 ds:98BA 是在場分隊。
 * 輸入域：清單項目、筆數、池基底任意；筆數大到超過探針步數上限（200,000）的執行由探針丟棄。艦艇欄位的 word 存取在位移 0xFFFF 時
 *   高位元組取同一段的位移 0（rd16s）。
 */
#include "gin3_reg.h"

extern uint16_t f_17A0_076B(Mem *m, uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1);

/* word 讀取：位移各自 16 位元回繞（seg:FFFF 的高位元組在同一段的位移 0），與 dosgolem 一致 */
static uint16_t f17A0_06D7_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_17A0_06D7(Mem *m, uint16_t x, uint16_t y, uint16_t faction)
{
    uint16_t i = 0;
    uint16_t list = 0x98BAu;

    while (i < f17A0_06D7_rd16s(m, GIN3_DS, 0x9896u)) {
        uint16_t idx = f17A0_06D7_rd16s(m, GIN3_DS, list);
        uint16_t ship = (uint16_t)(f17A0_06D7_rd16s(m, GIN3_DS, 0xCFB4u) + (uint16_t)(idx * 0x4Eu));
        uint16_t seg = f17A0_06D7_rd16s(m, GIN3_DS, 0xCFB6u);

        if (rd8(m, LIN(seg, ship)) != (uint8_t)faction
            && rd8(m, LIN(seg, (uint16_t)(ship + 4u))) != 7
            && (rd8(m, LIN(seg, (uint16_t)(ship + 0x1Eu))) & 0x80u) == 0
            && f_17A0_076B(m, x, y, f17A0_06D7_rd16s(m, seg, (uint16_t)(ship + 0x34u)),
                           f17A0_06D7_rd16s(m, seg, (uint16_t)(ship + 0x36u))) != 0)
            return (uint16_t)(f17A0_06D7_rd16s(m, GIN3_DS, list) + 1u);
        i = (uint16_t)(i + 1u);
        list = (uint16_t)(list + 2u);
    }
    return 0xFFFFu;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_17A0_06D7(m, a[0], a[1], a[2]); }
GIN3_REGISTER("17A0_06D7", 3, run)
