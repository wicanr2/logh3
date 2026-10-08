/*
 * 35AA:0006  以物件欄位算出緩衝區大小、配置，並把結果指標存回物件
 * 呼叫：far，一個 far 指標參數 obj（位移在前、段在後），無回傳值。
 *   +03 byte  旗標：非 0 時寬高各減 2
 *   +07 byte  寬、+08 byte  高（旗標非 0 時先各減 2，結果存到 +0B、+0C）
 *   +0B、+0C  算出來的寬與高
 *   +0E word  配置結果的位移（BX），+10 word  配置結果的段（AX）
 * 配置大小 ＝ (+0B × +0C × 4) 取低 16 位元。
 * 呼叫 3ABF:0000（描述檔 stubs）：參數＝大小，回傳 AX＝段、BX＝位移。該函式是失敗就停機的配置器，由替身取代。
 * 欄位語意：只證實算術與欄位位移，物件種類未知。推論等級：欄位位移已證實，用途（影像緩衝區）假說。
 */
#include "gin3_reg.h"

void f_35AA_0006(Mem *m, uint16_t off, uint16_t seg)
{
    uint8_t w, h;
    uint16_t size;
    uint16_t a[1];
    uint32_t r;

    if (rd8(m, LIN(seg, off + 3)) != 0) {
        w = (uint8_t)(rd8(m, LIN(seg, off + 7)) + 0xFE);
        wr8(m, LIN(seg, off + 0xB), w);
        h = (uint8_t)(rd8(m, LIN(seg, off + 8)) + 0xFE);
        wr8(m, LIN(seg, off + 0xC), h);
    } else {
        w = rd8(m, LIN(seg, off + 7));
        wr8(m, LIN(seg, off + 0xB), w);
        h = rd8(m, LIN(seg, off + 8));
        wr8(m, LIN(seg, off + 0xC), h);
    }
    w = rd8(m, LIN(seg, off + 0xB));
    h = rd8(m, LIN(seg, off + 0xC));
    size = (uint16_t)(w * h * 4u);
    a[0] = size;
    r = gin3_stub(m, "3ABF_0000", 1, a); /* AX<<16 | BX */
    wr16(m, LIN(seg, off + 0x10), (uint16_t)(r >> 16));
    wr16(m, LIN(seg, off + 0xE), (uint16_t)r);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_35AA_0006(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("35AA_0006", 2, run)
