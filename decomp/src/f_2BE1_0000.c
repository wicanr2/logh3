/*
 * GIN3PS.EXE 2BE1:0000  艦隊的每個艦艇組依隊形槽算出戰場目的地：對 F[+0x10 ＋ 4i]（word 池索引）逐一呼叫 2BE1:0059(F, 艦艇)
 *
 * 原版（89 bytes，far 呼叫，一個 far 指標參數 F＝艦隊記錄；序言 push bp／si、sub sp,6，所以 F 的位移在 [bp+0Eh]、段在 [bp+10h]），無回傳值：
 *   區域變數：[bp+0] ＝ i（byte）、[bp+2]／[bp+4] ＝ 清單指標（位移／段）
 *   清單指標 ← (F 的位移 ＋ 0x10 ＜16 位元＞, F 的段)                      ; 一次算好
 *   迴圈：
 *     i >= byte F[+0x0F]（無號 jnb；F 每圈重讀 `les bx,[bp+0Eh]`）→ 結束
 *     idx ← word 清單指標[0]（段內位移回繞）
 *     艦艇 ← (word ds:CFB4 ＋ low16(idx × 0x4E)) : word ds:CFB6             ; 池基底每圈重讀
 *     2BE1:0059(F, 艦艇)                                                   ; 推入順序：艦艇段、艦艇位移、F 段、F 位移
 *     i++（byte）；清單指標位移 ＋= 4（16 位元）
 * 過程（已證實：反組譯逐指令；2BE1:0059 由其差異測試證實）。
 * 語意（強推論）：docs/internals/38-battle-units.md 第 237 行「297B:091D → 2BE1:0000 → 2BE1:0059，逐支讀艦艇記錄 +0x05」、
 *   docs/re/300-opening-deployment-uses-formation-slots.md 第 20 到 25 行（戰場初始化 297B:091D 呼叫 2BE1:0000(艦隊)）、
 *   docs/spec/025-battle-formation-order.md 第 17 行，與本檔一致。F[+0x10 ＋ 4i] 的 word 是艦艇池索引（4 bytes 一組，另 2 bytes 未讀）。
 * 副作用：全部來自 0059（艦艇 S 的 +0x28／+0x2A 兩個 word，經 3A92:0001 旋轉的 x、y 在堆疊）。S 與 F 或與 ds:CFB4 別名時，
 *   前面一圈的寫入影響後面的讀取（F[+0x0F]、清單、池基底每圈重讀），本檔逐步照做。
 * 機器模型：被呼叫的 f_2BE1_0059 把旋轉用的 x、y 放在 SS:1F52／1F54（它自己的進入點框架，基底 fcmd-s3-base.state 的 SS:1F68 起算）。
 *   從本函式呼叫時原版的實際框架更深（0059 的 x、y 在 SS:1F40 起），位置不同，但兩處都在探針的堆疊帶（SS:1B68..1F67）內，寫入集合比對時排除，
 *   而 C 版自寫自讀，不影響結果。
 * 輸入域：F 指到艦隊表、艦艇池或 seeds 給的位址，不指向堆疊帶與 A0000 到 BFFFF；池基底 ds:CFB4..CFB7 列入 freeze（隨機突變會改到視訊區或堆疊帶，
 *   seeds 仍可造別名）。基底 fcmd-s3-base.state。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern void f_2BE1_0059(Mem *m, uint16_t f_off, uint16_t f_seg, uint16_t s_off, uint16_t s_seg);

/* 段內 word 讀取：位移各自 16 位元回繞（seg:FFFF 的高位元組在同一段的位移 0），與 dosgolem 一致 */
static uint16_t w2BE1_0000_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

void f_2BE1_0000(Mem *m, uint16_t f_off, uint16_t f_seg)
{
    uint8_t i = 0;
    uint16_t ids_off = (uint16_t)(f_off + 0x10u);

    while (i < rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x0Fu)))) {
        uint16_t idx = w2BE1_0000_rd16s(m, f_seg, ids_off);
        uint16_t ship = (uint16_t)(w2BE1_0000_rd16s(m, GIN3_DS, 0xCFB4u) + (uint16_t)(idx * 0x4Eu));
        uint16_t seg = w2BE1_0000_rd16s(m, GIN3_DS, 0xCFB6u);

        f_2BE1_0059(m, f_off, f_seg, ship, seg);
        i = (uint8_t)(i + 1u);
        ids_off = (uint16_t)(ids_off + 4u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2BE1_0000(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2BE1_0000", 2, run)
