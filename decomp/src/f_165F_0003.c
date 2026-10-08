/*
 * GIN3PS.EXE 165F:0003  建「內圈」的圖形描述結構（戰場移動範圍圈的圓心與半徑）
 *
 * 原版：far 呼叫，三個參數（由低位址起）：描述結構指標 dst（far）、艦隊記錄指標 src（far）、半徑參數 w（word）。
 * 無回傳值；AX 的殘值是 w ＋ 8（最後一個 mov ax,cx / add ax,8），呼叫端 2583:0003 不使用，本版仍回傳它並納入比對。
 * 唯一呼叫者 2583:0003（docs/re/120 §「艦隊層的選點」：165F:0003(→[bp+8], 艦隊, ax)）。
 *
 * 語意（已證實：docs/re/120 §「內圈」，與反組譯逐條對過；收據量到像素 s3cmdrange 圓心 (298,409)、半徑 88）：
 *   dst[+0] (word) ← src[+0x5C] ＋ 8      圓心 x（艦隊現在位置 x ＋ 8，8 ＝ 圖塊中心）
 *   dst[+2] (word) ← src[+0x5E] ＋ 8      圓心 y
 *   dst[+4] (byte) ← 0
 *   dst[+5] (byte) ← 5
 *   dst[+6] (byte) ← 1
 *   dst[+7] (byte) ← 1
 *   dst[+8] (word) ← w ＋ 8               半徑
 *   加法都是 16 位元回繞。寫入次序如上；dst 與 src 重疊時次序有影響，本版照做。
 *   各格在描述結構裡的意義（+4／+5／+6／+7 的 0／5／1／1）：未知（docs/re/120 只量到圓心與半徑；
 *   docs/re/135 說 +6 在另一個圖形裡是 5，同一套描述格式）。
 *
 * 輸入域：src 取艦隊表 32 筆（122 bytes 一筆）；dst 指向 16 bytes 的可寫緩衝區；w 取 16 位元全域。
 */
#include "gin3_reg.h"

uint16_t f_165F_0003(Mem *m, uint16_t dst_off, uint16_t dst_seg, uint16_t src_off, uint16_t src_seg, uint16_t w)
{
    uint16_t r = (uint16_t)(w + 8u);

    wr16(m, LIN(dst_seg, dst_off), (uint16_t)(rd16(m, LIN(src_seg, src_off + 0x5Cu)) + 8u));
    wr16(m, LIN(dst_seg, dst_off + 2u), (uint16_t)(rd16(m, LIN(src_seg, src_off + 0x5Eu)) + 8u));
    wr8(m, LIN(dst_seg, dst_off + 4u), 0);
    wr8(m, LIN(dst_seg, dst_off + 5u), 5);
    wr8(m, LIN(dst_seg, dst_off + 6u), 1);
    wr8(m, LIN(dst_seg, dst_off + 7u), 1);
    wr16(m, LIN(dst_seg, dst_off + 8u), r);
    return r;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_165F_0003(m, a[0], a[1], a[2], a[3], a[4]); }
GIN3_REGISTER("165F_0003", 5, run)
