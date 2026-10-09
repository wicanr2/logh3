/*
 * GIN3PS.EXE 307D:06F7  AI 估算一支艦隊 Q 對目標點 (x, y) 的戰力：Q 各分隊的 [+8] 加總 × Q[+3C] ÷ 1000，再依方位與距離乘除幾個倍率，最後除以陣營對照表的一項
 *
 * 原版（495 bytes，far 呼叫，七個 word 堆疊參數：P（far 指標，位移 [bp+1Ah]、段 [bp+1Ch]）、Q（far 指標，位移 [bp+1Eh]、段 [bp+20h]）、
 *   x（[bp+22h]）、y（[bp+24h]）、a（[bp+26h]）；序言 push bp／si／di、sub sp,10h、mov bp,sp）。回傳 AX。
 *   區域變數：[bp+0] ＝ i（byte）、[bp+1] ＝ 相對方位（byte）、[bp+2] ＝ V（word，累計值）、[bp+4]／[bp+6] ＝ 距離 D（低字、高字）、[bp+8]／[bp+0Ah] ＝ 分隊槽指標。
 * 過程（已證實：反組譯逐指令；2BE1:011D／024A、3310:0006、3CC9:0009／0020／00B9 由其各自的差異測試證實）：
 *   Q[+0] bit0 有設 → 回 0
 *   V ← 0；對 i ＝ 0 .. Q[+0F]−1（byte，無號 jnb）：idx ＝ word Q[+0x10 + 4i]；ship ＝ 池基底（ds:CFB4、ds:CFB6，每圈重讀）＋ idx × 0x4E（16 位元回繞）；V ← V ＋ word ship[+8]（16 位元）
 *   V ← (V × Q[+3C]) ÷ 1000 的低 16 位元                      ; 3CC9:0009（V 零延伸）、3CC9:0020（無號，除數 1000）
 *   rel1 ← 2BE1:024A(P[+4A], 2BE1:011D(x − P[+5C], y − P[+5E]))   ; 位置 P 朝 (x, y) 的方位與 P 朝向的無號 byte 差
 *     rel1 >= 0x60：V ← V << 1；>= 0x40：V ← ((V × 5) 的低字) ÷ 3；>= 0x20：V ← ((V << 2) 的低字) ÷ 3；否則不變          ; 全部 16 位元
 *   rel2 ← 2BE1:024A(a, 2BE1:011D(P[+5C] − x, P[+5E] − y))
 *     rel2 >= 0x60：V ← V >> 1；>= 0x40：V ← ((V << 2) 的低字) ÷ 3；>= 0x20：V ← ((V × 5) 的低字) ÷ 3；否則 V ← V << 1
 *   D ← 3310:0006(P[+5C], P[+5E], x, y)                       ; 16.16 距離（原始 32 位元值與整數 300、600 比，不是 300.0）
 *     D < 300（有號）：V ← V << 2；否則 D >= 600（有號）：V ← V >> 1；300 <= D < 600 不變（3310:0006 的非 0 結果 >= 1.0 ＝ 65536，所以這一段到不了）
 *   Q[+48] 的低 4 位元全 0 → V ← V >> 1
 *   V ← V ÷ byte ds:(0xD6EC ＋ Q[+3])                          ; 無號 16 位元 div；除數 0 時 #DE，屬輸入域外（探針丟棄，C 版回 0）
 *   回 V
 * 語意（假說）：docs 沒有 307D:06F7 的記載（grep 只有台帳）。呼叫者 307D:0521（AI 常式，台帳 class unknown）。Q ＝ 被評估的己方艦隊（艦艇池記錄 +8 ＝ 艦艇數？），
 *   P ＝ 對方艦隊，倍率與方位、距離有關；整體用途未知，本版只照原版運算。名稱不是證據。
 * 副作用：只有 3310:0006 內 39B5:005C 的暫存（ds:8EDC..8EED）。
 *
 * 輸入域：P、Q 指向艦隊表內（描述檔 tables）；x、y、a 全部 16 位元值；池基底 ds:CFB4..CFB7 列入 freeze（隨機突變會改到視訊區）；除數為 0 的案例（#DE）不涵蓋。
 *   池記錄的 word 讀取在位移 0xFFFF 時，dosgolem 的高位元組取同一段的位移 0（f06F7_rd16s）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為假說。
 */
#include "gin3_reg.h"

extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_024A(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint64_t f_3CC9_0020(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

#define F06F7_ZF 0x0040u
#define F06F7_SF 0x0080u
#define F06F7_OF 0x0800u

/* 段內 word 讀取：位移 0xFFFF 時高位元組在同一段的位移 0 */
static uint16_t f06F7_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static int f06F7_jge(uint16_t fl) { return ((fl & F06F7_SF) != 0) == ((fl & F06F7_OF) != 0); }

uint16_t f_307D_06F7(Mem *m, uint16_t p_off, uint16_t p_seg, uint16_t q_off, uint16_t q_seg, uint16_t x, uint16_t y, uint16_t a)
{
    uint16_t v = 0, slot = (uint16_t)(q_off + 0x10u), rel1, rel2, ang, fl, div;
    uint8_t i = 0;
    uint32_t prod, dist;

    if (rd8(m, LIN(q_seg, q_off)) & 1u)
        return 0;
    while (i < rd8(m, LIN(q_seg, (uint16_t)(q_off + 0x0Fu)))) {
        uint16_t idx = f06F7_rd16s(m, q_seg, slot);
        uint16_t pool_off = rd16(m, DS(0xCFB4));
        uint16_t pool_seg = rd16(m, DS(0xCFB6));
        uint16_t ship = (uint16_t)(pool_off + (uint16_t)(idx * 0x4Eu));

        v = (uint16_t)(v + f06F7_rd16s(m, pool_seg, (uint16_t)(ship + 8u)));
        i++;
        slot = (uint16_t)(slot + 4u);
    }
    prod = f_3CC9_0009(0, v, 0, rd8(m, LIN(q_seg, (uint16_t)(q_off + 0x3Cu))));
    v = (uint16_t)(f_3CC9_0020((uint16_t)(prod >> 16), (uint16_t)prod, 0, 1000) >> 32);

    ang = f_2BE1_011D(m, (uint16_t)(x - rd16(m, LIN(p_seg, (uint16_t)(p_off + 0x5Cu)))), (uint16_t)(y - rd16(m, LIN(p_seg, (uint16_t)(p_off + 0x5Eu)))));
    rel1 = (uint8_t)f_2BE1_024A(m, rd16(m, LIN(p_seg, (uint16_t)(p_off + 0x4Au))), ang);
    if (rel1 >= 0x60u)
        v = (uint16_t)(v << 1);
    else if (rel1 >= 0x40u)
        v = (uint16_t)((uint16_t)(v * 5u) / 3u);
    else if (rel1 >= 0x20u)
        v = (uint16_t)((uint16_t)(v << 2) / 3u);

    ang = f_2BE1_011D(m, (uint16_t)(rd16(m, LIN(p_seg, (uint16_t)(p_off + 0x5Cu))) - x), (uint16_t)(rd16(m, LIN(p_seg, (uint16_t)(p_off + 0x5Eu))) - y));
    rel2 = (uint8_t)f_2BE1_024A(m, a, ang);
    if (rel2 >= 0x60u)
        v = (uint16_t)(v >> 1);
    else if (rel2 >= 0x40u)
        v = (uint16_t)((uint16_t)(v << 2) / 3u);
    else if (rel2 >= 0x20u)
        v = (uint16_t)((uint16_t)(v * 5u) / 3u);
    else
        v = (uint16_t)(v << 1);

    dist = f_3310_0006(m, rd16(m, LIN(p_seg, (uint16_t)(p_off + 0x5Cu))), rd16(m, LIN(p_seg, (uint16_t)(p_off + 0x5Eu))), x, y);
    fl = f_3CC9_00B9((uint16_t)(dist >> 16), (uint16_t)dist, 0, 300);
    if (!f06F7_jge(fl)) {
        v = (uint16_t)(v << 2);
    } else {
        fl = f_3CC9_00B9((uint16_t)(dist >> 16), (uint16_t)dist, 0, 600);
        if (f06F7_jge(fl))
            v = (uint16_t)(v >> 1);
    }
    if ((rd8(m, LIN(q_seg, (uint16_t)(q_off + 0x48u))) & 0x0Fu) == 0)
        v = (uint16_t)(v >> 1);
    div = rd8(m, DS((uint16_t)(0xD6ECu + rd8(m, LIN(q_seg, (uint16_t)(q_off + 3u))))));
    if (div == 0)
        return 0; /* #DE：輸入域外 */
    return (uint16_t)(v / div);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_307D_06F7(m, a[0], a[1], a[2], a[3], a[4], a[5], a[6]);
}
GIN3_REGISTER("307D_06F7", 7, run)
