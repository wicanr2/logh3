/*
 * GIN3PS.EXE 3870:0A30  Cohen–Sutherland 線段裁切的一步：把落在裁切矩形外的端點沿線段推到矩形邊界上，回傳新位置的 outcode
 *
 * 原版：far 呼叫，七個 word 的堆疊參數（序言 push bp／si／di、sub sp,10h、mov bp,sp，所以第一個參數在 [bp+1Ah]），由低位址到高位址：
 *   code [bp+1Ah]（只測低位元組的 bit 3／2／0／1，`test byte ptr [bp+1Ah], n`）、
 *   P1 位移 [bp+1Ch]、P1 段 [bp+1Eh]（遠指標，指到端點的 x，word）、P2 位移 [bp+20h]、P2 段 [bp+22h]（指到端點的 y，word）、
 *   c [bp+24h]（線段另一端的 x）、d [bp+26h]（另一端的 y）。回傳 AX ＝ 3870:09F1(*P1, *P2)（新端點的 outcode，AH 是 y 的高位元組殘值）。
 *   x ＝ *P1、y ＝ *P2（有號 word，符號延伸成 32 位元）；矩形 xmin ＝ ds:8D6A、ymin ＝ ds:8D6C、xmax ＝ ds:8D6E、ymax ＝ ds:8D70
 *   （3870:09F1 的檔頭），邊界值當**無號** word 零延伸成 32 位元（`xor ax,ax`／`xor cx,cx` 當高字）。
 * 動作（已證實：反組譯逐指令；bit 的優先順序是 8、4、1、2，只執行第一個符合的）：
 *   code&8（上）：B ＝ ds:8D6C；  x ＋＝ low16( ((B − y) × (x − c)) ÷ (y − d) )；y ＝ B（再讀一次 ds:8D6C）
 *   code&4（下）：B ＝ ds:8D70；  同上
 *   code&1（左）：B ＝ ds:8D6A；  y ＋＝ low16( ((y − d) × (B − x)) ÷ (x − c) )；x ＝ B（再讀一次 ds:8D6A）
 *   code&2（右）：B ＝ ds:8D6E；  同上
 *   都不符合就什麼也不改。最後無論如何呼叫 09F1(x ＝ *P1 重讀, y ＝ *P2 重讀) 並回傳它的 AX。
 *   乘法與除法是 runtime 的 32 位元 long 運算（3CC9:0009 低 32 位元乘、3CC9:0070 有號除、商向零取整、除數 0 時商 0xFFFFFFFF 不引發 #DE），
 *   運算元都是 32 位元的差，只有商的低 16 位元（BX）加進原座標；「((B − y) × (x − c))」乘積只留低 32 位元。
 *   32 位元差：(B − y) ＝ 零延伸的 B 減符號延伸的 y（16 位元高字 ＝ 0 − 符號 − 借位）；(x − c)、(y − d) 是兩個符號延伸 word 的 32 位元差。
 *   `add es:[si], bx` 是對記憶體的讀改寫（重讀 *P1／*P2）；之後的 `mov es:[bx], ax` 用 ds 重讀邊界值。P1／P2 與矩形、彼此別名時，
 *   後面的讀取看得到前面的寫入，本版逐步照做。
 * 語意（強推論）：這是用相似三角形求線段與裁切邊的交點（線性內插）；3870:08FF 反覆呼叫它直到兩端 outcode 為 0（台帳 callers）。
 *   `((B − y) × (x − c)) ÷ (y − d)` 整除（向零取整）的截斷與 16 位元溢位照原版行為，不保證幾何上精確。
 * 呼叫的函式：3CC9:0009（lmul）、3CC9:0070（ldiv）與 3870:09F1（outcode），三者皆已 diff-tested（lmul／ldiv 以 gin3_rt.h 的數學結果表達）。
 * 呼叫者（台帳）：3870:08FF（兩處）。
 *
 * 機器模型：*P1、*P2 是 word 存取，位移 0xFFFF 時高位元組取同段位移 0（dosgolem 的 read16／write16），用本檔的 rd16s／wr16s；
 *   寫入線性 0xF0000 以上（ROM 區）被安靜丟棄，本檔的 wr16s 模型化（之後的重讀讀到 ROM 原內容）。P1／P2 指向堆疊帶、A0000 到 BFFFF 的案例不涵蓋。
 * 輸入域：全部 16 位元值。除數為 0（y ＝ d 或 x ＝ c）原版不引發 #DE（3CC9:0070 在 0020 內處理）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

extern uint16_t f_3870_09F1(Mem *m, uint16_t x, uint16_t y);

/* 段內 word 讀取：seg:FFFF 的高位元組在 seg:0000 */
static uint16_t f3870_0A30_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go，真機同） */
static void f3870_0A30_wr8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void f3870_0A30_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    f3870_0A30_wr8(m, LIN(seg, off), (uint8_t)v);
    f3870_0A30_wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

static uint32_t f3870_0A30_sx(uint16_t w) { return (uint32_t)(int32_t)(int16_t)w; }

static uint16_t f3870_0A30_quot(uint32_t num, uint32_t den1, uint32_t den2)
{
    uint32_t q, r;

    gin3_ldivmod(gin3_lmul(num, den1), den2, &q, &r);
    return (uint16_t)q;
}

uint16_t f_3870_0A30(Mem *m, uint16_t code, uint16_t p1_off, uint16_t p1_seg, uint16_t p2_off, uint16_t p2_seg,
                     uint16_t c, uint16_t d)
{
    if (code & 8u) { /* 上邊界 ds:8D6C，y 被推到邊界，x 依比例調整 */
        uint32_t b = rd16(m, DS(0x8D6C));
        uint16_t y = f3870_0A30_rd16s(m, p2_seg, p2_off);
        uint16_t x = f3870_0A30_rd16s(m, p1_seg, p1_off);
        uint16_t q = f3870_0A30_quot(b - f3870_0A30_sx(y), f3870_0A30_sx(x) - f3870_0A30_sx(c),
                                     f3870_0A30_sx(y) - f3870_0A30_sx(d));

        f3870_0A30_wr16s(m, p1_seg, p1_off, (uint16_t)(f3870_0A30_rd16s(m, p1_seg, p1_off) + q));
        f3870_0A30_wr16s(m, p2_seg, p2_off, rd16(m, DS(0x8D6C)));
    } else if (code & 4u) { /* 下邊界 ds:8D70 */
        uint32_t b = rd16(m, DS(0x8D70));
        uint16_t y = f3870_0A30_rd16s(m, p2_seg, p2_off);
        uint16_t x = f3870_0A30_rd16s(m, p1_seg, p1_off);
        uint16_t q = f3870_0A30_quot(b - f3870_0A30_sx(y), f3870_0A30_sx(x) - f3870_0A30_sx(c),
                                     f3870_0A30_sx(y) - f3870_0A30_sx(d));

        f3870_0A30_wr16s(m, p1_seg, p1_off, (uint16_t)(f3870_0A30_rd16s(m, p1_seg, p1_off) + q));
        f3870_0A30_wr16s(m, p2_seg, p2_off, rd16(m, DS(0x8D70)));
    } else if (code & 1u) { /* 左邊界 ds:8D6A，x 被推到邊界，y 依比例調整 */
        uint32_t b = rd16(m, DS(0x8D6A));
        uint16_t y = f3870_0A30_rd16s(m, p2_seg, p2_off);
        uint16_t x = f3870_0A30_rd16s(m, p1_seg, p1_off);
        uint16_t q = f3870_0A30_quot(f3870_0A30_sx(y) - f3870_0A30_sx(d), b - f3870_0A30_sx(x),
                                     f3870_0A30_sx(x) - f3870_0A30_sx(c));

        f3870_0A30_wr16s(m, p2_seg, p2_off, (uint16_t)(f3870_0A30_rd16s(m, p2_seg, p2_off) + q));
        f3870_0A30_wr16s(m, p1_seg, p1_off, rd16(m, DS(0x8D6A)));
    } else if (code & 2u) { /* 右邊界 ds:8D6E */
        uint32_t b = rd16(m, DS(0x8D6E));
        uint16_t y = f3870_0A30_rd16s(m, p2_seg, p2_off);
        uint16_t x = f3870_0A30_rd16s(m, p1_seg, p1_off);
        uint16_t q = f3870_0A30_quot(f3870_0A30_sx(y) - f3870_0A30_sx(d), b - f3870_0A30_sx(x),
                                     f3870_0A30_sx(x) - f3870_0A30_sx(c));

        f3870_0A30_wr16s(m, p2_seg, p2_off, (uint16_t)(f3870_0A30_rd16s(m, p2_seg, p2_off) + q));
        f3870_0A30_wr16s(m, p1_seg, p1_off, rd16(m, DS(0x8D6E)));
    }
    return f_3870_09F1(m, f3870_0A30_rd16s(m, p1_seg, p1_off), f3870_0A30_rd16s(m, p2_seg, p2_off));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3870_0A30(m, a[0], a[1], a[2], a[3], a[4], a[5], a[6]);
}
GIN3_REGISTER("3870_0A30", 7, run)
