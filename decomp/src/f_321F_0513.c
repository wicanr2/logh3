/*
 * GIN3PS.EXE 321F:0513  戰場艦艇的演出投影：對 n 艘艦艇（指標表 ds:D78E）算螢幕上的相對座標，旋轉、平移後寫回艦艇 +0x3E／+0x40／+0x44，並標記可見性 +0x42
 *
 * 原版：far 呼叫，一個 word 參數 n（序言 push bp／si／di、sub sp,18h，所以參數在 [bp+22h]），沒有有意義的回傳值（描述檔 ret ＝ none）。
 *   區域變數：[bp+0／+2] ＝ 指標表的游標（位移／段 ds）、[bp+4／+6] ＝ 艦艇 S 的 far 指標、[bp+8] ＝ i、[bp+0Ah／0Ch／0Eh／10h] ＝ 四個平移常數 A、B、C、D。
 * 過程（已證實：反組譯逐指令；3A92:0001 由其自己的差異測試證實）。讀寫的交錯順序與反組譯一致（S 與指標表、常數表 ds:D7EA..D7FC、
 *   自己的欄位別名時，前面的寫入影響後面的重讀；C 版不快取任何欄位）：
 *   對 i ＝ 0..n−1（word，jb 無號），游標 ＝ ds:(0xD78E + 4i)：
 *     S ← far 指標（word 游標[0] 位移、word 游標[2] 段）
 *     S[+0] ＝ 0（byte）：A ← ds:D7EE、B ← ds:D7F0、C ← ds:D7F6、D ← ds:D7F8；否則：A ← ds:D7F2、B ← ds:D7F4、C ← ds:D7FA、D ← ds:D7FC
 *     S[+0x3E] ← half3(S[+0x34] − A)；S[+0x40] ← half3(S[+0x36] − B)       ; half3(v)：v 為 16 位元有號差，乘三取低 16 位元，再 cwd；sub ax,dx；sar ax,1
 *                                                                          ; ＝ 低 16 位元當有號數除以 2、向零取整；乘積的高字被 cwd 蓋掉
 *     3A92:0001(&S[+0x3E], &S[+0x40], word ds:D7EA)                         ; 以 byte 角度 ds:D7EA 旋轉，結果寫回 S[+0x3E]、S[+0x40]
 *     S[+0x44] ← S[+0x40]
 *     q ← S[+0x40]（重讀）÷ 2（有號 idiv，向零取整）；S[+0x40] ← q
 *     S[+0x3E] ← S[+0x3E] − q；S[+0x3E] ← S[+0x3E] + C                        ; 兩次 16 位元讀改寫
 *     S[+0x40] ← S[+0x40] + D                                                ; 讀改寫
 *     x ← S[+0x3E]（重讀）、y ← S[+0x40]（重讀）；S[+0x42] ← (8 ≤ x ≤ 0x238 且 0xDC ≤ y ≤ 0x178（皆有號）) ? 1 : 0     ; word 寫入
 * 語意（強推論）：docs/internals/38 第 552 行「321F:0513 把 SHIP+3E／+40 寫為演出投影 XY，+44 為深度，+42 改成 0／1 的可見性」；
 *   docs/spec/097 第 16 行「依 SHIP+0 陣營取該側參考與基底；相對 XY 各乘 3 除 2，signed 向零捨去」、docs/spec/095 第 59 行「321F:0513 的可見方塊 X8..568、
 *   虛擬 Y220..376」（0x238 ＝ 568、0xDC ＝ 220、0x178 ＝ 376）、docs/re/408 第 72 行。y 的二分之一（×0.5 壓縮）是等角投影的縱向壓縮（假說）。
 *   呼叫者 321F:000A、321F:008A（unknown）。
 * 副作用：每艘艦艇 S 的 +0x3E、+0x40、+0x42、+0x44（共 8 bytes）。線性 0xF0000 以上（ROM）的寫入在 dosgolem 被安靜丟棄，本檔按 byte 模型化
 *   （重讀讀到 ROM 原內容）；word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（rd16s／wr16s）。
 * 輸入域：n 取小值（每艘約 170 步，探針步數上限 200000，n 太大的案例被丟棄）；指標表內容由突變與 seeds 決定，S 可落在任意位址，
 *   但不涵蓋指向堆疊帶、A0000 到 BFFFF 的案例（探針丟棄）。基底狀態 fcmd-base.state（指標表 ds:D78E 全 0，即 S ＝ 0000:0000，寫入落在中斷向量表）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_3A92_0001(Mem *m, uint16_t x_off, uint16_t x_seg, uint16_t y_off, uint16_t y_seg, uint16_t angle_word);

static uint16_t w0513_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效 */
static void w0513_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w0513_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w0513_put8(m, LIN(seg, off), (uint8_t)v);
    w0513_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

/* sub ax, A；mov cx,3；imul cx；cwd；sub ax,dx；sar ax,1 */
static uint16_t w0513_half3(uint16_t v)
{
    int16_t lo = (int16_t)(uint16_t)(int32_t)((int32_t)(int16_t)v * 3); /* imul 的低 16 位元 */
    int16_t dx = lo < 0 ? -1 : 0;                                      /* cwd 以低字的符號蓋掉高字 */
    int16_t ax = (int16_t)(uint16_t)(lo - dx);

    return (uint16_t)(ax >> 1);
}

void f_321F_0513(Mem *m, uint16_t n)
{
    uint16_t i = 0;
    uint16_t cur = 0xD78Eu;

    while (i < n) {
        uint16_t s_seg = rd16(m, DS((uint16_t)(cur + 2u)));
        uint16_t s_off = rd16(m, DS(cur));
        uint16_t ca, cb, cc, cd, q, ang;
        int16_t x, y;

        if (rd8(m, LIN(s_seg, s_off)) == 0) {
            ca = rd16(m, DS(0xD7EEu));
            cb = rd16(m, DS(0xD7F0u));
            cc = rd16(m, DS(0xD7F6u));
            cd = rd16(m, DS(0xD7F8u));
        } else {
            ca = rd16(m, DS(0xD7F2u));
            cb = rd16(m, DS(0xD7F4u));
            cc = rd16(m, DS(0xD7FAu));
            cd = rd16(m, DS(0xD7FCu));
        }
        w0513_wr16s(m, s_seg, (uint16_t)(s_off + 0x3Eu),
                    w0513_half3((uint16_t)(w0513_rd16s(m, s_seg, (uint16_t)(s_off + 0x34u)) - ca)));
        w0513_wr16s(m, s_seg, (uint16_t)(s_off + 0x40u),
                    w0513_half3((uint16_t)(w0513_rd16s(m, s_seg, (uint16_t)(s_off + 0x36u)) - cb)));
        ang = rd16(m, DS(0xD7EAu));
        f_3A92_0001(m, (uint16_t)(s_off + 0x3Eu), s_seg, (uint16_t)(s_off + 0x40u), s_seg, ang);
        w0513_wr16s(m, s_seg, (uint16_t)(s_off + 0x44u), w0513_rd16s(m, s_seg, (uint16_t)(s_off + 0x40u)));
        q = (uint16_t)((int16_t)w0513_rd16s(m, s_seg, (uint16_t)(s_off + 0x40u)) / 2);
        w0513_wr16s(m, s_seg, (uint16_t)(s_off + 0x40u), q);
        w0513_wr16s(m, s_seg, (uint16_t)(s_off + 0x3Eu), (uint16_t)(w0513_rd16s(m, s_seg, (uint16_t)(s_off + 0x3Eu)) - q));
        w0513_wr16s(m, s_seg, (uint16_t)(s_off + 0x3Eu), (uint16_t)(w0513_rd16s(m, s_seg, (uint16_t)(s_off + 0x3Eu)) + cc));
        w0513_wr16s(m, s_seg, (uint16_t)(s_off + 0x40u), (uint16_t)(w0513_rd16s(m, s_seg, (uint16_t)(s_off + 0x40u)) + cd));
        y = (int16_t)w0513_rd16s(m, s_seg, (uint16_t)(s_off + 0x40u));
        x = (int16_t)w0513_rd16s(m, s_seg, (uint16_t)(s_off + 0x3Eu));
        w0513_wr16s(m, s_seg, (uint16_t)(s_off + 0x42u), (x >= 8 && x <= 0x238 && y >= 0xDC && y <= 0x178) ? 1u : 0u);
        i++;
        cur = (uint16_t)(cur + 4u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_321F_0513(m, a[0]);
    return 0;
}
GIN3_REGISTER("321F_0513", 1, run)
