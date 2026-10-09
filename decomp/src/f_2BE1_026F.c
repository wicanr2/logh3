/*
 * GIN3PS.EXE 2BE1:026F  沿 byte 角度前進一段距離：(x, y) += 正弦表 × 距離（四捨五入）後夾進戰場範圍，結果存到兩個 word 指標
 *
 * 原版：far 呼叫，六個參數（push bp、si、di、sub sp,4，所以參數從 [bp+0Eh] 起，共 8 個堆疊字）：
 *   args[0] ＝ x（word）
 *   args[1] ＝ y（word）
 *   args[2] ＝ angle（word，只取低位元組：mov bl,[bp+12h]）
 *   args[3] ＝ dist（word，當無號數）
 *   args[4..5] ＝ out_x 的 far 指標（位移、段）
 *   args[6..7] ＝ out_y 的 far 指標（位移、段）
 * 回傳：AX ＝ 夾限後的 y（殘留值，呼叫者以輸出指標取結果；描述檔用 ret=ax 逐位元比對）。
 *
 * 過程（表 T 是 ds:573A 起 256 筆有號 word，docs/internals/41 稱為正弦表，值 ≈ sin × 1024；idx 是 byte）：
 *   x ← x + step((angle + 0x40) mod 256)      ; 余弦分量（sin 表相位 +64）
 *   y ← y + step(angle)                      ; 正弦分量
 *   step(idx) ＝ 低 16 位元( 算術右移10( T[idx] 符號延伸成 32 位元 × dist（無號零延伸）的低 32 位元 ＋ 0x200 ) )
 *     乘法走 3CC9:0009（已 diff-tested），加 0x200 是 32 位元加（add bx,200h；adc ax,0），右移是 10 次 sar ax,1／rcr bx,1。
 *   x、y 都是 16 位元回繞相加，之後以有號 word 夾限：x 夾進 [0, 0x7EF]，y 夾進 [0, 0x7E7]（jns／jle，有號）。
 *   寫入：*out_x ← x（word）、*out_y ← y（word），先 x 後 y（兩個指標相同或重疊時，後寫的蓋掉先寫的）。
 *   寫入落在線性 0xF0000 以上（ROM 區）時被 dosgolem 安靜丟棄，本檔按 byte 模型化；字的存取在本檔按線性位址連續，dosgolem 實際是兩次 byte 存取、位移各自 16 位元回繞（cpu.go read16／write16）；word 存取落在位移 0xFFFF 的輸入不在比較範圍（seeds 的輸出指標位移皆遠離 0xFFFF）。
 * 兩個夾限值與正弦表來源：docs/internals/41 第 32 到 40 行、docs/re/128 第 33 行（朝向 163 走一步位移 (−5,−6)）、
 *   docs/re/130 第 60 行（另一處夾 0x7FF／0x7F7，「捨去與夾值一律看呼叫點」，docs/internals/46 第 28 行）。
 *
 * 輸入域：全部 16 位元參數；表內容取自記憶體（基底狀態與突變）；輸出指標任意。純計算加兩個 word 寫入。
 * 推論等級：已證實（反組譯逐指令＋差異測試；3CC9:0009 由其自己的差異測試證實）。座標與表的名稱是強推論，不使用。
 */
#include "gin3_reg.h"

extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

static void w026f_put16(Mem *m, uint32_t lin, uint16_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, (uint8_t)v);
    if (((lin + 1) & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin + 1, (uint8_t)(v >> 8));
}

/* 一個分量的位移（回傳低 16 位元） */
static uint16_t w026f_step(Mem *m, uint8_t idx, uint16_t dist)
{
    uint16_t t = rd16(m, DS(0x573A + 2 * idx));
    uint16_t sign = (t & 0x8000u) ? 0xFFFFu : 0u;           /* cwd */
    uint32_t prod = f_3CC9_0009(sign, t, 0, dist);          /* AX:BX × CX:DX，CX = 0 */

    prod += 0x200u;                                         /* add bx,200h；adc ax,0 */
    return (uint16_t)(uint32_t)((int32_t)prod >> 10);       /* sar ax,1／rcr bx,1 × 10，只留 BX（只留低 16 位元，所以算術或邏輯右移結果相同；本檔照原版寫算術右移） */
}

uint16_t f_2BE1_026F(Mem *m, uint16_t x, uint16_t y, uint16_t angle_word, uint16_t dist,
                     uint16_t ox_off, uint16_t ox_seg, uint16_t oy_off, uint16_t oy_seg)
{
    uint8_t angle = (uint8_t)angle_word;

    x = (uint16_t)(x + w026f_step(m, (uint8_t)(angle + 0x40u), dist));
    y = (uint16_t)(y + w026f_step(m, angle, dist));
    if ((int16_t)x < 0)
        x = 0;
    if ((int16_t)x > 0x7EF)
        x = 0x7EF;
    if ((int16_t)y < 0)
        y = 0;
    if ((int16_t)y > 0x7E7)
        y = 0x7E7;
    w026f_put16(m, LIN(ox_seg, ox_off), x);
    w026f_put16(m, LIN(oy_seg, oy_off), y);
    return y;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_2BE1_026F(m, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7]);
}
GIN3_REGISTER("2BE1_026F", 8, run)
