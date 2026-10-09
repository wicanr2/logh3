/*
 * 3617:04CE  文字視窗物件的游標指標：內容緩衝區起點 + (列 × 欄數 + 欄) × 4（far，純讀取）
 *
 * 原版：far 呼叫，一個 far 指標參數 obj（位移在前、段在後），回 AX:BX（AX ＝ 段，BX ＝ 位移；不是 DX:AX）。
 *   堆疊：序言 push bp／di、mov bp,sp，沒有區域變數，所以參數在 [bp+8]（位移）與 [bp+0Ah]（段）。
 *   +0B byte  內容的格寬（欄數），+0E word、+10 word  內容緩衝區的 far 指標（位移、段）
 *   +1C byte  游標欄，+1D byte  游標列
 * 位移 ＝ obj[+0E] + obj[+1C] × 4 + obj[+1D] × obj[+0B] × 4（全部 16 位元回繞：乘積取低 16 位元，
 *   obj[+1D] × obj[+0B] 最大 255 × 255 = 65025，再乘 4 會溢位被截斷）。段 ＝ obj[+10]。
 * 其他暫存器：CX ＝ obj[+1C] × 4，DX ＝ obj[+10]（與 AX 同），DI 由 push／pop 還原。呼叫者
 *   （3617:03BE 等）用 AX:BX 存回物件的 [+1E]／[+20]，CX、DX 是暫用值；描述檔 ret = pack:ax,bx,cx,dx，一併比對。
 * 不寫記憶體。
 *
 * 語意：與 docs/re/141 第 32 行、docs/internals/49 的欄位表「游標指標 ＝ [+0E] ＋ 列 × [+0B] × 4 ＋ 欄 × 4」
 *   一致（已證實：反組譯逐指令）；一格 4 bytes（字碼 word、顏色 byte、旗標 byte，docs/re/137）。
 * 輸入域：物件指標任意。欄位位移各自在段內 16 位元回繞（obj 位移接近 0xFFFF 時，`es:[bx+10h]` 繞回段首），
 *   位移 0xFFFF 的 word 存取，高位元組取同一段的位移 0（dosgolem 的 8086 語意，見 rd16s）。
 *   描述檔用暫存緩衝區（lin 0x9D000 起）當物件，內容隨機，另用種子造位移 0xFFF1 的回繞物件。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

/* 段內回繞的 word 讀取：seg:FFFF 的高位元組在 seg:0000（dosgolem internal/cpu/cpu.go），不是下一段。 */
static uint16_t f3617_04CE_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint64_t f_3617_04CE(Mem *m, uint16_t off, uint16_t seg)
{
    uint16_t ax, cx, dx, di;

    ax = (uint16_t)(rd8(m, LIN(seg, (uint16_t)(off + 0x1Du))) * rd8(m, LIN(seg, (uint16_t)(off + 0x0Bu)))); /* mul cx，取 AX */
    ax = (uint16_t)(ax << 2); /* 兩次 shl ax,1 */
    cx = (uint16_t)(rd8(m, LIN(seg, (uint16_t)(off + 0x1Cu))) << 2);
    dx = f3617_04CE_rd16s(m, seg, (uint16_t)(off + 0x10u));
    di = f3617_04CE_rd16s(m, seg, (uint16_t)(off + 0x0Eu));
    di = (uint16_t)(di + cx);
    di = (uint16_t)(di + ax);
    /* mov ax,dx；mov bx,di */
    return ((uint64_t)dx << 48) | ((uint64_t)di << 32) | ((uint64_t)cx << 16) | dx;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3617_04CE(m, a[0], a[1]);
}
GIN3_REGISTER("3617_04CE", 2, run)
