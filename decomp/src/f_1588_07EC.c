/*
 * GIN3PS.EXE 1588:07EC  戰場單位直線走到目標：每步 8 像素朝目標方位角走，每走一步問 2583:040C 能不能站；走到目標回 1，被擋回 0
 *
 * 原版（165 bytes，far 呼叫，五個 word 堆疊參數 (x, y, tx, ty, faction)；序言 push bp／si、sub sp,2、mov bp,sp，所以 x 在 [bp+0Ah]、y 在 [bp+0Ch]、tx 在 [bp+0Eh]、
 *   ty 在 [bp+10h]、faction 在 [bp+12h]），回傳 AX ＝ 1 或 0。區域變數 [bp+0] ＝ 方位角（byte，只存不讀）。x、y 就是堆疊上的參數字，函式就地改它們
 *   （2BE1:026F 的輸出指標指向 [bp+0Ah]、[bp+0Ch]，SS 段）。
 * 過程（已證實：反組譯逐指令；3310:0006、2BE1:011D、2BE1:026F、2583:040C、3CC9:00B9 各由其差異測試證實）：
 *   迴圈：
 *     (x, y) ＝ (tx, ty)（兩個都相等，16 位元）→ 回 1
 *     d ← 3310:0006(x, y, tx, ty)                          ; 16.16 距離
 *     3CC9:00B9(d, 8.0) 之後 jge（有號 32 位元 d >= 8.0，ZF ＝ 1 或 SF ＝ OF）：
 *       成立：ang ← 2BE1:011D(a ＝ tx − x, b ＝ ty − y)；2BE1:026F(x, y, ang, 8, →x, →y)    ; 沿 ang 走 8 像素並夾進戰場範圍，結果寫回 x、y
 *       不成立（d < 8.0，含 d 因座標差 >= 0x8000 變負）：(x, y) ← (tx, ty)
 *     2583:040C(x, y, faction) ≠ 0 → 回 0（被擋）；否則回到迴圈開頭
 * 語意（強推論）：docs/re/330-strategic-planet-assault.md、docs/re/102-battle-movement.md 的移動檢查，呼叫者 307D:0B41；
 *   8 像素是 026F 的 dist 常數；「d < 8.0 就直接放到目標」讓最後一步不會繞圈。
 * 副作用：被呼叫者的 ds:8EDC..8EED 暫存（3310:0006 內部的 39B5:005C）；2BE1:026F 寫的 x、y 在呼叫者堆疊上（探針堆疊帶內，比對時排除）。
 *
 * 機器模型：C 版沒有堆疊。x、y 這兩個「堆疊字」放在 SS:1F00、SS:1F02（基底 fcmd-base.state，SS 固定 0x63EB，探針堆疊帶 SS:1B8A..1F89 之內，
 *   寫入集合比對時雙方都排除）；2BE1:026F 的 C 版以遠指標寫它們，再讀回。原版的實際位置是 SS:1F80、1F82，與被呼叫者 39B5 C 版借用的 SS:1F82..1F89 重疊，
 *   所以不能放在原位。
 *   原版的迴圈沒有上界（目標在戰場範圍外時 026F 夾限使距離不再縮小，永遠走不到），C 版加 F07EC_MAX_ITER 的圈數上限，只讓被改壞的突變停得下來
 *   （原版在步數上限內的案例最多約 700 圈，不會觸及；觸及時回 0xDEAD，不可能與原版相同）。
 * 輸入域：x、y、tx、ty、faction 全部 16 位元值，但只收原版在步數上限內結束的案例（目標不可達的案例被探針丟棄）；
 *   基底 fcmd-base.state，地圖、圖塊遮罩表、在場分隊清單與艦艇池由 seeds 造（見描述檔）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_026F(Mem *m, uint16_t x, uint16_t y, uint16_t angle_word, uint16_t dist, uint16_t ox_off, uint16_t ox_seg, uint16_t oy_off, uint16_t oy_seg);
extern uint16_t f_2583_040C(Mem *m, uint16_t x, uint16_t y, uint16_t faction);

#define F07EC_SS 0x63EBu
#define F07EC_PX 0x1F00u
#define F07EC_PY 0x1F02u
#define F07EC_MAX_ITER 20000u

#define F07EC_ZF 0x0040u
#define F07EC_SF 0x0080u
#define F07EC_OF 0x0800u

uint16_t f_1588_07EC(Mem *m, uint16_t x, uint16_t y, uint16_t tx, uint16_t ty, uint16_t faction)
{
    unsigned iter;

    for (iter = 0; iter < F07EC_MAX_ITER; iter++) {
        uint32_t d;
        uint16_t fl;

        if (x == tx && y == ty)
            return 1;
        d = f_3310_0006(m, x, y, tx, ty);
        fl = f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, 0x0008u, 0x0000u);
        if (((fl & F07EC_SF) != 0) == ((fl & F07EC_OF) != 0)) { /* jge */
            uint8_t ang = f_2BE1_011D(m, (uint16_t)(tx - x), (uint16_t)(ty - y));

            (void)f_2BE1_026F(m, x, y, ang, 8, F07EC_PX, F07EC_SS, F07EC_PY, F07EC_SS);
            x = rd16(m, LIN(F07EC_SS, F07EC_PX));
            y = rd16(m, LIN(F07EC_SS, F07EC_PY));
        } else {
            x = tx;
            y = ty;
        }
        if (f_2583_040C(m, x, y, faction) != 0)
            return 0;
    }
    return 0xDEADu;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1588_07EC(m, a[0], a[1], a[2], a[3], a[4]); }
GIN3_REGISTER("1588_07EC", 5, run)
