/*
 * GIN3PS.EXE 0FF4:0E5A  指定星域與艦種，算這個艦種在該星域的存量夠編幾組（先把 ds:9872／9874 設成該星域）
 *
 * 原版：far 呼叫，兩個堆疊字（序言 push bp／si、沒有 sub sp，所以參數在 [bp+8] 與 [bp+0Ah]），回 AX。
 *   region_no ← [bp+8] 的低位元組（mov al,[bp+8]；xor ah,ah）
 *   kind      ← [bp+0Ah]（原樣再推入 0E84，0E84 只用低位元組）
 *   星域位移  ← 0xA79C ＋ low16(region_no × 0x7E)
 *   ds:9874 ← DS（段，先寫）；ds:9872 ← 星域位移
 *   回 AX ← 0FF4:0E84(kind)，AX 原封不動（= 存量 ÷ 每組艘數 的低 16 位元，0E84 的語意見 f_0FF4_0E84.c）
 *
 * 副作用：寫 ds:9872／9874 兩個 word（0E84 稍後從那裡讀星域指標）；其餘寫入都在 0E84（它沒有寫入）。
 * 呼叫者：2D76:0A44（編成流程的模板試算，docs/re/229、231）。
 *
 * 推論等級
 *   已證實（差異測試）：星域記錄大小 0x7E 與基底 0xA79C、region_no 只取低位元組、兩格指標的值、回傳原樣轉出。
 *   強推論（docs/re/229、231）：ds:9872／9874 是「目前星域」的 far 指標，供 0E84 取存量。
 * 輸入域：region_no、kind 任意 word（各只取低位元組）；region_no >= 32 時指到星域表之後的記憶體，照原版運算。
 */
#include "gin3_reg.h"

extern uint16_t f_0FF4_0E84(Mem *m, uint16_t kind_word);

uint16_t f_0FF4_0E5A(Mem *m, uint16_t region_no, uint16_t kind)
{
    uint16_t region = (uint16_t)(0xA79Cu + (uint16_t)((uint16_t)(region_no & 0xFFu) * 0x7Eu));

    wr16(m, DS(0x9874), GIN3_DS);
    wr16(m, DS(0x9872), region);
    return f_0FF4_0E84(m, kind);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_0FF4_0E5A(m, a[0], a[1]); }
GIN3_REGISTER("0FF4_0E5A", 2, run)
