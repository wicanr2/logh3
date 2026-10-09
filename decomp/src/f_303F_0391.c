/*
 * GIN3PS.EXE 303F:0391  數候選清單前 n 筆裡有幾筆不是目前陣營的艦艇（far，一個參數，回 AX = 計數）
 *
 * 參數（堆疊字；序言 push bp／si／di、sub sp,8，所以第一個參數在 [bp+12h]）：
 *   n    一個字（mov di,[bp+12h]；cmp ax,di；jnb 是整個字的無號比較）：要檢查的筆數
 *
 * 反組譯逐條對應（IDA sub_3D2F1）：
 *   count ← 0；對 i = 0 … n−1（word 計數，無號）：
 *     idx   ← word ds:A49A[i]                           ; 位置 A49A + 2 × i，16 位元回繞
 *     艦艇  ← (ds:CFB6 : ds:CFB4 + low16(idx × 0x4E))   ; mul cx 取 AX，位移 16 位元回繞；池基底每圈重讀
 *     艦艇[+0x00] != ds:8F1D（byte 比較，jz 反向） → count ← count + 1
 *   回傳 count（AX，16 位元計數）。不寫記憶體。
 *
 * 推論等級
 *   已證實（差異測試的位元組與回傳值）：上列算式、欄位位移、比較方向與常數。
 *   強推論（docs/re/145 第 184 至 194 行、docs/re/90「名單：ds:A49A」）：ds:A49A 是候選艦艇池索引清單（由
 *     17A0:00AC 的空間查詢填入，word 陣列）、艦艇[+0x00] 是陣營、ds:8F1D 是目前陣營；
 *     回傳值是「這個方位掃到幾個敵方」（docs/re/145 第 69 行的 303F:0391(n)）。名稱不是證據。
 *
 * 輸入域：n 取 0..3000（每圈約 22 道指令，步數上限 200,000 使更大的 n 被探針丟棄）；清單內容（A49A 起）取 word 全域，
 *   池索引超出 352 筆時讀到池之後的記憶體，本版同樣照讀（平面視訊記憶體的讀取除外，見 docs/re/445 限制）。
 *   A49A 起的清單在所有可用基底狀態裡是全 0，內容靠突變與描述檔 seeds 的 pokes 造出。
 */
#include "gin3_reg.h"

uint16_t f_303F_0391(Mem *m, uint16_t n)
{
    uint16_t count = 0;
    uint16_t i;

    for (i = 0; i < n; i++) {
        uint16_t idx = rd16(m, DS((uint16_t)(0xA49Au + 2u * i)));
        uint16_t pool_seg = rd16(m, DS(0xCFB6));
        uint16_t unit = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(idx * 0x4Eu));

        if (rd8(m, LIN(pool_seg, unit)) != rd8(m, DS(0x8F1D)))
            count++;
    }
    return count;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_303F_0391(m, a[0]); }
GIN3_REGISTER("303F_0391", 1, run)
