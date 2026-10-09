/*
 * GIN3PS.EXE 303F:0061  數在場分隊清單裡有幾支不是目前陣營的，存進 ds:D610
 * （far，無參數，無回傳值；ret = none。AX 在 retf 時是迴圈殘值，不是輸出）
 *
 * 反組譯逐條對應（IDA sub_3CFC1）：
 *   ds:D610 ← 0（word）
 *   對 i = 0 … 直到 i >= word ds:9896（無號 word 比較，每圈重讀 ds:9896）：
 *     idx   ← word ds:98BA[i]                           ; 位置 98BA + 2 × i，16 位元回繞
 *     艦艇  ← (ds:CFB6 : ds:CFB4 + low16(idx × 0x4E))   ; mul cx 取 AX，位移 16 位元回繞；池基底每圈重讀
 *     艦艇[+0x00] != ds:8F1D（byte 比較，jz 反向） → word ds:D610 ← ds:D610 + 1（inc word ptr，16 位元）
 *
 * 推論等級
 *   已證實（差異測試的位元組與寫入集合）：上列算式、欄位位移、比較方向與常數、先清零再累加。
 *   強推論（docs/re/145 第 77 行、docs/internals/50 第 61 行）：D610 是「敵方總數」（要塞砲開火門檻的分母），
 *     ds:98BA／ds:9896 是在場分隊清單與筆數（docs/internals/38 第 291 行），艦艇[+0x00] 是陣營。
 *   與舊文件的差異：docs/re/145 第 77 行寫「艦艇池裡 +0x00 不等於 ds:8F1D 的筆數」、docs/internals/50 第 61 行寫
 *     「數整個艦艇池」。原版走的是在場分隊清單 ds:98BA（筆數 ds:9896），不是整個艦艇池；清單裡的池索引可重複，
 *     重複的索引會重複計數。
 *
 * 輸入域：ds:9896 取 0..約 5000（每圈約 22 道指令，步數上限 200,000 使更大的筆數被探針丟棄；描述檔凍結 ds:9896，
 *   筆數由 seeds 的 pokes 指定）；清單內容取 word 全域，池索引超出 352 筆時讀到池之後的記憶體，本版同樣照讀
 *   （平面視訊記憶體的讀取除外，見 docs/re/445 限制）。在場分隊清單在所有可用基底狀態裡是空的（ds:9896 = 0），
 *   內容靠突變與描述檔 seeds 的 pokes 造出。
 */
#include "gin3_reg.h"

void f_303F_0061(Mem *m)
{
    uint16_t i;

    wr16(m, DS(0xD610), 0);
    for (i = 0; i < rd16(m, DS(0x9896)); i++) {
        uint16_t idx = rd16(m, DS((uint16_t)(0x98BAu + 2u * i)));
        uint16_t pool_seg = rd16(m, DS(0xCFB6));
        uint16_t unit = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(idx * 0x4Eu));

        if (rd8(m, LIN(pool_seg, unit)) != rd8(m, DS(0x8F1D)))
            wr16(m, DS(0xD610), (uint16_t)(rd16(m, DS(0xD610)) + 1u));
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_303F_0061(m);
    return 0;
}
GIN3_REGISTER("303F_0061", 0, run)
