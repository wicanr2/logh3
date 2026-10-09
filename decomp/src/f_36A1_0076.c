/*
 * GIN3PS.EXE 36A1:0076  把 seg129 內第 5 個緩衝區的遠指標 3D22:60BE 存進 ds:F138／F13A（無參數）
 *
 * 原版：far 呼叫，無參數、沒有 push bp（36 bytes，逐指令）：
 *   call far 3703:000D                   ; AX ＝ seg129 的段值（dosgolem 下 0x3D22，載入位置相依）
 *   mov bx,ax；xor ax,ax；mov ax,bx；xor bx,bx；mov si,60BEh；xor cx,cx   ; 暫存器搬運，結果：AX ＝ 段、BX ＝ 0、SI ＝ 0x60BE、CX ＝ 0
 *   or bx,si；or ax,cx
 *   mov ds:F13Ah,ax                      ; word ds:F13A ← 段（0x3D22）
 *   mov ds:F138h,bx                      ; word ds:F138 ← 位移（0x60BE）
 *   retf                                 ; AX ＝ 段（SI、DI 由 push／pop 保存）
 *   即遠指標（位移 ds:F138、段 ds:F13A）＝ 3D22:60BE。回傳 AX ＝ 0x3D22（暫存器殘值，非有意義的回傳值，本版仍比對）。
 * DS：進入時的 DS（呼叫者 341B:0004 沿用 DGROUP），函式不改 DS。
 * 呼叫者（台帳）：341B:0004（game-logic）。呼叫順序：同一呼叫者還呼叫 36A1:0000（寫前四個緩衝區指標）。
 *
 * 語意（強推論）：seg129 內有 5 塊緩衝區，起點位移 0x079E、0x2F9E、0x551E、0x5B1E、0x60BE（36A1:0000 寫前四個，本函式寫第五個）；
 *   0x079E 起每筆 160 bytes 是 2376 模組的圖形表（docs/internals/67 的 2376 一節）；其餘緩衝區用途：未知。
 *   3703:0011 另寫 3D22:DDBE／DDC0，基底狀態兩者與本函式寫的值一致（見 docs/internals/67 的 3703 列）。
 * 輸入域：無輸入。函式不讀記憶體（除指令本身），差異測試的鑑別力全部來自 seeds：先把 ds:F138..F13B 預填成別的值，才看得出少寫、寫偏與寬度。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；緩衝區用途為強推論／未知。
 */
#include "gin3_reg.h"

extern uint16_t f_3703_000D(void);

uint16_t f_36A1_0076(Mem *m)
{
    uint16_t seg = f_3703_000D();

    wr16(m, DS(0xF13Au), seg);
    wr16(m, DS(0xF138u), 0x60BEu);
    return seg;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_36A1_0076(m);
}
GIN3_REGISTER("36A1_0076", 0, run)
