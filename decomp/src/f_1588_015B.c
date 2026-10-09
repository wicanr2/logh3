/*
 * GIN3PS.EXE 1588:015B  戰場地圖緩衝區的參考計數減一，歸零時釋放並清空指標（無參數、無回傳值）
 *
 * 原版：far 呼叫，無參數、沒有 push bp、不用暫存器傳入（40 bytes）：
 *   dec byte ds:3C19；jnz 返回                          ; 計數減一（byte，0 減 1 變 0xFF，不等於 0）
 *   push word ds:3C1C；push word ds:3C1A；call far 3ABF:003B；add sp,4   ; free(ds:3C1C:ds:3C1A)（位移 3C1A、段 3C1C）
 *   word ds:3C1A ← 0；word ds:3C1C ← 0；byte ds:3C18 ← 0xFF
 *   返回；AX 是殘值（計數不歸零時是進入值，歸零時是 3ABF:003B 的回傳），描述檔 ret = none，不比對。
 * 用 ds:3C1A／3C1C 指標指向的是戰場地圖緩衝區（13,056 bytes，1588:06C0 的目的地，docs/re/314 §四；
 *   基底 fcmd-standby.state 裡 ds:3C1A:3C1C ＝ 79B0:0010）。ds:3C19 在 standby 狀態是 1，ds:3C18 是 0。
 *   ds:3C18 寫 0xFF 與 3C19 的名稱和語意：未知。計數／旗標的名稱屬假說，這裡只照寫入的位址與值做。
 * 呼叫者（台帳）：073F:060A、0DEA:1320、113A:03D5、20B4:1497、297B:08D8、297B:0ABF。
 * 呼叫的函式：3ABF:003B（free 的 far 包裝，已 diff-tested）。釋放的實際行為（自由串列的節點與 rover 的寫入）
 *   全部來自 3CD9:01FC，本函式不碰。
 *
 * 輸入域：ds:3C18..3C1C 任意；指標指向的區塊與自由串列要是良構（損壞的串列原版會無窮迴圈，被探針以步數上限丟棄）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；計數與旗標的用途為假說。
 */
#include "gin3_reg.h"

extern uint16_t f_3ABF_003B(Mem *m, uint16_t off, uint16_t seg);

void f_1588_015B(Mem *m)
{
    uint8_t n = (uint8_t)(rd8(m, DS(0x3C19)) - 1u);

    wr8(m, DS(0x3C19), n);
    if (n != 0)
        return;
    (void)f_3ABF_003B(m, rd16(m, DS(0x3C1A)), rd16(m, DS(0x3C1C)));
    wr16(m, DS(0x3C1A), 0);
    wr16(m, DS(0x3C1C), 0);
    wr8(m, DS(0x3C18), 0xFFu);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_1588_015B(m);
    return 0;
}
GIN3_REGISTER("1588_015B", 0, run)
