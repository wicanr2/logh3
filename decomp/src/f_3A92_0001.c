/*
 * GIN3PS.EXE 3A92:0001  二維旋轉（指標版）：以 byte 角度旋轉 *x、*y，結果寫回原處
 *
 * 原版：far 呼叫，三個參數（push bp；mov bp,sp；push si；push di，所以第一個參數在 [bp+6]）：
 *   args[0..1] ＝ x 的 far 指標（位移、段；指向一個 word，有號）
 *   args[2..3] ＝ y 的 far 指標（位移、段；指向一個 word，有號）
 *   args[4]    ＝ 角度（word，只取低位元組：mov al,[bp+arg_8]; xor ah,ah）
 * 回傳：AX ＝ x'（最後 pop ax 取回 3A92:0062 的結果），沒有人保證呼叫者用它；描述檔 ret = ax 一併比對。
 *
 * 過程（已證實：反組譯逐指令；3A92:0062、0081 由其各自的差異測試證實）：
 *   BX ← *x；CX ← *y（讀 word；位移 0xFFFF 時高位元組取同一段的位移 0）
 *   SI ← 0x573A + 2a；DI ← 0x573A + 2 × ((a + 0x40) mod 256)
 *   x' ← 3A92:0062(SI, DI, BX, CX)；y' ← 3A92:0081(SI, DI, BX, CX)（兩者輸入相同，與 3A92:0042 同一個核心）
 *   先 *y ← y'，再 *x ← x'（兩個指標相同或重疊時，後寫的 x' 蓋掉 y'）
 * 副作用：寫兩個 word。線性 0xF0000 以上（ROM）的寫入在 dosgolem 被安靜丟棄，本檔按 byte 模型化；
 *   word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（rd16s／wr16s）。
 * 輸入域：全部 16 位元參數與記憶體內容；讀到 A0000 到 BFFFF（視訊記憶體）的案例由探針丟棄。
 *   指標指到本函式的堆疊或程式碼的案例不在比較範圍。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_3A92_0062(Mem *m, uint16_t si, uint16_t di, uint16_t bx, uint16_t cx);
extern uint16_t f_3A92_0081(Mem *m, uint16_t si, uint16_t di, uint16_t bx, uint16_t cx);

static uint16_t w0001_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效 */
static void w0001_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w0001_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w0001_put8(m, LIN(seg, off), (uint8_t)v);
    w0001_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_3A92_0001(Mem *m, uint16_t x_off, uint16_t x_seg, uint16_t y_off, uint16_t y_seg, uint16_t angle_word)
{
    uint16_t bx = w0001_rd16s(m, x_seg, x_off);
    uint16_t cx = w0001_rd16s(m, y_seg, y_off);
    uint8_t a = (uint8_t)angle_word;
    uint16_t si = (uint16_t)(0x573Au + a + a);
    uint8_t a2 = (uint8_t)(a + 0x40u);
    uint16_t di = (uint16_t)(0x573Au + a2 + a2);
    uint16_t x1 = f_3A92_0062(m, si, di, bx, cx);
    uint16_t y1 = f_3A92_0081(m, si, di, bx, cx);

    w0001_wr16s(m, y_seg, y_off, y1);
    w0001_wr16s(m, x_seg, x_off, x1);
    return x1;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_3A92_0001(m, a[0], a[1], a[2], a[3], a[4]); }
GIN3_REGISTER("3A92_0001", 5, run)
