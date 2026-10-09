/*
 * GIN3PS.EXE 2703:0CED  行星經濟力的成長：經濟力（32 位元）加上 經濟力 × 成長率（千分率）÷ 1000
 *
 * 原版：far 呼叫，一個參數＝行星記錄（82 bytes）的 far 指標（位移、段）。序言 push bp／si／di、sub sp,4，所以指標在 [bp+0Eh]（位移）
 *   與 [bp+10h]（段）。無回傳值（描述檔 ret = none；AX 是殘值，唯一呼叫者 2703:0C26 之後不讀它）。
 *
 * 語意（已證實：與反組譯逐條對過；欄位名稱為強推論，見下）：
 *   e ← 行星[+0x1A..+0x1D] 的 32 位元值（`mov ax,es:[bx+2]; mov bx,es:[bx]`，先讀後呼叫）。
 *   r ← 2703:0D47(行星)（成長率，千分率；先把 e 的低字放在堆疊上保存，`pop bx` 取回）。
 *   p ← 3CC9:0009(e, r)：AX:BX × CX:DX，其中 CX = 0、DX = r，取低 32 位元（`xor cx,cx` 之後 DX 為 r）。
 *   q ← 3CC9:0020(p, 1000)：AX:BX ÷ CX:DX，CX = 0、DX = 0x3E8，無號除法取商（AX:BX）。
 *   行星[+0x1A] 低字 ← 行星[+0x1A] 低字 + q 低字（`add es:[si], bx`）；
 *   行星[+0x1C] 高字 ← 行星[+0x1C] 高字 + q 高字 + 進位（`adc es:[si+2], ax`）。
 *   也就是 e ← e + (e × r 的低 32 位元) ÷ 1000，結果截成 32 位元。乘積溢位時（e × r >= 2^32）商是截斷後的乘積除以 1000，
 *   與數學式不同；e + q 超過 2^32 時回繞。
 *
 * 欄位語意（強推論）：行星[+0x1A] 是經濟力（32 位元；docs/spec/039、2703:0D47 的檔頭同），2703:0D47 是行星成長率（千分率）。
 *   這支是「每回合經濟力成長」（假說，docs/re/43 有成長率表，沒有這支的逐位元組記載）。
 *
 * 輸入域：行星指標取行星表任一筆（81 筆，基底 ds:5336／5338 指向）；行星記錄的 +0x16、+0x1A..+0x1D 與 ds:4A68 起的門檻表由突變與種子涵蓋。
 *   word 存取按線性位址連續處理（行星記錄在段內低位移，不涵蓋跨越位移 0xFFFF，與 2703:0D47 同）。
 *   2703:0D47 在門檻表被改壞而一路走下去時原版會超出步數上限，探針丟棄該案。
 * 推論等級：已證實（反組譯逐指令＋差異測試，0D47、3CC9:0009、3CC9:0020 各有差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_2703_0D47(Mem *m, uint16_t planet_off, uint16_t planet_seg);
extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint64_t f_3CC9_0020(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

void f_2703_0CED(Mem *m, uint16_t planet_off, uint16_t planet_seg)
{
    uint16_t field = (uint16_t)(planet_off + 0x1Au);
    uint16_t lo = rd16(m, LIN(planet_seg, field));
    uint16_t hi = rd16(m, LIN(planet_seg, (uint16_t)(field + 2u)));
    uint16_t rate = f_2703_0D47(m, planet_off, planet_seg);
    uint32_t prod = f_3CC9_0009(hi, lo, 0, rate);
    uint32_t quot = (uint32_t)(f_3CC9_0020((uint16_t)(prod >> 16), (uint16_t)prod, 0, 0x3E8u) >> 32);
    uint32_t sum = (uint32_t)rd16(m, LIN(planet_seg, field)) + (uint16_t)quot; /* add es:[si], bx */

    wr16(m, LIN(planet_seg, field), (uint16_t)sum);
    wr16(m, LIN(planet_seg, (uint16_t)(field + 2u)),
         (uint16_t)(rd16(m, LIN(planet_seg, (uint16_t)(field + 2u))) + (uint16_t)(quot >> 16) + (sum >> 16))); /* adc es:[si+2], ax */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2703_0CED(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2703_0CED", 2, run)
