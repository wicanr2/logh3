/*
 * GIN3PS.EXE 36A1:0000  把 seg129 內前四個緩衝區的遠指標存進 ds:F128..F137（無參數）
 *
 * 原版：far 呼叫，無參數（118 bytes；序言 push bp／si／di、sub sp,4、mov bp,sp，結尾對稱），逐指令：
 *   call far 3703:000D                    ; AX ＝ seg129 的段值（dosgolem 下 0x3D22，載入位置相依）
 *   暫存器搬運後：[bp+0] ＝ 段、[bp+2] ＝ 0（區域變數，之後每段都從這兩格重載）
 *   對 (位移 o, 目的) 依序 (0x079E, ds:F128)、(0x2F9E, ds:F12C)、(0x551E, ds:F130)、(0x5B1E, ds:F134)：
 *       bx ← [bp+2] | o；ax ← [bp+0] | 0；  word 目的+2 ← ax（段）；word 目的 ← bx（位移）
 *   即四個遠指標（位移在前、段在後）：ds:F128 ＝ 3D22:079E、ds:F12C ＝ 3D22:2F9E、ds:F130 ＝ 3D22:551E、ds:F134 ＝ 3D22:5B1E。
 *   每組都是先寫段 word、後寫位移 word。回傳 AX ＝ 最後一次 `or ax,cx` 的結果 ＝ 段（暫存器殘值，本版仍比對）。
 *   區域變數 [bp+0..3] 在堆疊上，探針把堆疊帶的寫入排除在寫入集合外，C 版不模型化。
 * DS：進入時的 DS（呼叫者 341B:0004 沿用 DGROUP），函式不改 DS。
 * 呼叫者（台帳）：341B:0004（game-logic）；同一呼叫者另呼叫 36A1:0076（寫第五個緩衝區指標 ds:F138）。
 *
 * 語意（強推論）：seg129 內五塊緩衝區的起點位移 0x079E、0x2F9E、0x551E、0x5B1E、0x60BE（第五個由 36A1:0076 寫），
 *   相鄰起點的差 0x2800、0x2580、0x0600、0x05A0 是各緩衝區的大小上限（假說）；0x079E 起是 2376 模組的圖形表
 *   （docs/internals/67 的 2376 一節）。其餘緩衝區的用途：未知。
 * 輸入域：無輸入。函式不讀記憶體（除指令本身），差異測試的鑑別力全部來自 seeds：先把 ds:F128..F13B 預填成別的值。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；緩衝區大小與用途為假說／未知。
 */
#include "gin3_reg.h"

extern uint16_t f_3703_000D(void);

uint16_t f_36A1_0000(Mem *m)
{
    uint16_t seg = f_3703_000D();

    wr16(m, DS(0xF12Au), seg);
    wr16(m, DS(0xF128u), 0x079Eu);
    wr16(m, DS(0xF12Eu), seg);
    wr16(m, DS(0xF12Cu), 0x2F9Eu);
    wr16(m, DS(0xF132u), seg);
    wr16(m, DS(0xF130u), 0x551Eu);
    wr16(m, DS(0xF136u), seg);
    wr16(m, DS(0xF134u), 0x5B1Eu);
    return seg;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_36A1_0000(m);
}
GIN3_REGISTER("36A1_0000", 0, run)
