/*
 * GIN3PS.EXE 35D1:005D  依編號找視窗節點（341B:0353），把它的種類與格座標換成像素矩形填進另一個記錄
 *
 * 原版：far 呼叫，一個參數＝記錄 P 的 far 指標（位移在 [bp+12h]、段在 [bp+14h]；push bp／si／di、sub sp,8）。無回傳值。
 *   唯一呼叫者 35D1:0009（台帳 callers = 1）。
 * 過程（已證實：反組譯逐指令；341B:0353 由其自己的差異測試證實）：
 *   Q ← 341B:0353(word P[+2])        ; 以 P[+2] 的低位元組當編號找視窗節點，回 AX:BX ＝ 節點 seg:off（找不到 0:0）
 *   以下每個欄位寫入之後，後面的欄位才讀 Q（讀寫交錯；P 與 Q 重疊時前面的寫入會影響後面的讀取）：
 *   P[+0x00] ← Q[+0]（byte）
 *   P[+0x12] ← word( Q[+5] << 4 )      ; 左移 4：格 X × 16 像素
 *   P[+0x14] ← word( Q[+6] × 20 )      ; 格 Y × 20 像素
 *   P[+0x1A] ← Q[+5]（byte）
 *   P[+0x1B] ← Q[+6]
 *   P[+0x1C] ← (Q[+5] + Q[+7] − 1) 低位元組   ; Q[+7] ＝ 寬；`dec ax` 的借位不影響 AL
 *   P[+0x1D] ← (Q[+6] + Q[+8] − 1) 低位元組   ; Q[+8] ＝ 高
 *   Q 的欄位位移加法（+5、+6）在段內 16 位元回繞。
 * 欄位語意（強推論，docs/internals/47、docs/re/138：視窗節點 +5 格 x、+6 格 y、+7 格寬、+8 格高；格是 16 × 20 像素）：
 *   P[+0x12]／[+0x14] ＝ 像素 X／Y，P[+0x1A..0x1D] ＝ 格矩形的左、上、右、下（含）。名稱不是證據，位移與運算才是。
 * 副作用：寫 P 的 +0x00、+0x12..0x15、+0x1A..0x1D。線性 0xF0000 以上（ROM）的寫入在 dosgolem 被安靜丟棄，本檔按 byte 模型化。
 * 輸入域：P 取暫存區或 seeds（含 P 與節點重疊）；節點取自基底狀態的視窗串列（5 個節點）與其突變；找不到時 Q ＝ 0000:0000，
 *   讀 IVT 的 byte。串列有環的案例被探針丟棄。word 存取在位移 0xFFFF 時高位元組取同一段位移 0（rd16s）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint32_t f_341B_0353(Mem *m, uint16_t key);

static uint16_t w005D_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效 */
static void w005D_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w005D_wr8s(Mem *m, uint16_t seg, uint16_t off, uint8_t v)
{
    w005D_put8(m, LIN(seg, off), v);
}

static void w005D_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w005D_put8(m, LIN(seg, off), (uint8_t)v);
    w005D_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_35D1_005D(Mem *m, uint16_t p_off, uint16_t p_seg)
{
    uint32_t q = f_341B_0353(m, w005D_rd16s(m, p_seg, (uint16_t)(p_off + 2u)));
    uint16_t qseg = (uint16_t)(q >> 16);
    uint16_t qoff = (uint16_t)q;
    uint8_t a, b;

    w005D_wr8s(m, p_seg, p_off, rd8(m, LIN(qseg, qoff)));
    a = rd8(m, LIN(qseg, (uint16_t)(qoff + 5u)));
    w005D_wr16s(m, p_seg, (uint16_t)(p_off + 0x12u), (uint16_t)((uint16_t)a << 4));
    b = rd8(m, LIN(qseg, (uint16_t)(qoff + 6u)));
    w005D_wr16s(m, p_seg, (uint16_t)(p_off + 0x14u), (uint16_t)(b * 20u));
    w005D_wr8s(m, p_seg, (uint16_t)(p_off + 0x1Au), rd8(m, LIN(qseg, (uint16_t)(qoff + 5u))));
    w005D_wr8s(m, p_seg, (uint16_t)(p_off + 0x1Bu), rd8(m, LIN(qseg, (uint16_t)(qoff + 6u))));
    a = rd8(m, LIN(qseg, (uint16_t)(qoff + 5u)));
    a = (uint8_t)(a + rd8(m, LIN(qseg, (uint16_t)(qoff + 7u))) - 1u);
    w005D_wr8s(m, p_seg, (uint16_t)(p_off + 0x1Cu), a);
    b = rd8(m, LIN(qseg, (uint16_t)(qoff + 6u)));
    b = (uint8_t)(b + rd8(m, LIN(qseg, (uint16_t)(qoff + 8u))) - 1u);
    w005D_wr8s(m, p_seg, (uint16_t)(p_off + 0x1Du), b);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_35D1_005D(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("35D1_005D", 2, run)
