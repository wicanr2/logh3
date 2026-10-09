/*
 * 2376:05D6  圖塊（X 不是 16 的倍數時帶位移）複製進對齊緩衝區（near，暫存器傳參 SI，DS = 模組資料段 3D22）
 *
 * 輸入：DS:SI = 來源圖形（呼叫者 2376:0538 在 `mov si, es:[si+0Eh]` 後 `call 2376:05D6`）；
 *   ds:786 = X、ds:788 = Y、ds:78E = w（每列字數）、ds:790 = R（列數）、ds:782 = 05FC 的位移次數快取。
 *   方向旗標 DF = 0（2376:0538 開頭有 `cld`）。
 * 原版：
 *   X & 0x0F = 0（`test ax,0Fh` 為 0）：直接 `call 2376:06CF`（整字對齊搬運），返回。
 *   否則：
 *     push si；call 2376:05FC（依 X & 7 改寫程式碼段 2376 的兩段位移指令）；pop si。
 *     k ← X & 7；DH ← 0xFF >> k；DL ← ~DH（左右遮罩）；`call loc_30928`（= 2376:0658）。
 *   2376:0658 是 IDA 沒有切成函式的一段（位元組 8C D8 … C3，共 0x77 bytes，台帳沒有這一項）：它的兩段位移指令
 *   （程式碼段 cs:0691..069A 與 cs:06A5..06AE，各 5 個 2-byte 指令）正是 2376:05FC 改寫的位置。
 *   0658 的行為（逐指令）：
 *     1. rep stosw：ES(= DS):0 起 0x3C0 個 word 填 0（1920 bytes）。
 *     2. DI ← 0x20 × (Y mod 20)（無號 div）；X & 8 有設則 DI ← DI + 1。
 *     3. CX ← R；外圈 R 次（`loop`，R = 0 會繞 65536 圈，不涵蓋）：push CX／SI／DI；CX ← w；
 *        AX ← lodsw；AX ← 位移指令A（5 個，見下）；BH ← AL；AL ← AL & DH；
 *        w = 0（jcxz）跳到收尾；否則 BH ← BH & DL；內圈 w 次（`loop`）：
 *            stosw（ES:DI ← AX，DI += 2）；AX ← lodsw；AX ← 位移指令B（5 個）；BL ← AL；AL ← AL & DH；AL ← AL | BH；
 *            BH ← BL；BH ← BH & DL
 *        收尾：AL ← AL & DL；byte ES:[DI] ← AL；pop DI／SI／CX；SI ← SI + 2 × word(ds:78E)（重讀）；DI ← DI + 8。
 *   位移指令的每一格是 2-byte 指令，本版從程式碼段讀出這 2 個 byte 解碼，因為它們是自我修改的：
 *     89 C0／8B C0 = mov ax,ax（不動）；D1 C8 = ror ax,1；D1 C0 = rol ax,1；86 C4 = xchg ah,al。
 *   05FC 在快取命中（ds:782 = X & 7）時不改寫，這時執行的是記憶體裡既有的位元組，所以必須從記憶體解碼，
 *   不能由 X & 7 直接推（兩者可以不一致，例如快取被突變）。其他位元組不會出現（探針只突變資料讀取集；
 *   程式碼取指不在讀取集內；種子只放這五種編碼）。
 * 輸出：記憶體寫入（緩衝區、05FC 的程式碼段改寫、ds:782 快取）。暫存器是暫用值，呼叫者 2376:0538 之後
 *   只 `call 2376:070E` 並還原，描述檔 ret = none。
 * 位移 0xFFFF 的 word 存取：高位元組取同一段的位移 0（dosgolem 行為，w05D6_rd16／wr16）。
 *
 * 語意（強推論）：把 SI 指的來源（每列 w 個字、共 R 列）以「向右位移 X & 7 位元」的方式複製到 ds:0 起的緩衝區，
 *   列距 8 bytes、起點依 Y mod 20 取（每列 0x20 bytes）。與 docs/re/407-fleet-blocks.md 第 30 行把 2376:08A0 記成
 *   「依 X mod 8 改寫位移指令」是同一型；2376:05D6 與 087A 是 X 對齊 16／8 兩種版本（假說）。
 * 輸入域：SI 全域；ds:786、ds:788、ds:782 的全部 16 位元值；ds:78E（w）與 ds:790（R）凍結，由種子給定
 *   （w ∈ {0,1,2,3,4,8,12}，R ∈ {1,2,3,5,10,20,64,250,300}），不涵蓋 R = 0。DF = 1 與緩衝區蓋到呼叫堆疊不涵蓋。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；位移指令語意（ror／rol／xchg）來自位元組解碼；用途為強推論。
 */
#include "gin3_reg.h"

extern void f_2376_05FC(Mem *m, uint16_t ds, uint16_t cs);
extern void f_2376_06CF(Mem *m, uint16_t ds, uint16_t si);

static uint16_t w05D6_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w05D6_wr16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

/* 執行程式碼段 cs:off 起的 5 個 2-byte 位移指令（自我修改的位置），輸入輸出是 AX */
static uint16_t w05D6_slot(const Mem *m, uint16_t cs, uint16_t off, uint16_t ax)
{
    int i;

    for (i = 0; i < 5; i++) {
        switch (w05D6_rd16(m, cs, (uint16_t)(off + 2u * (uint16_t)i))) {
        case 0xC8D1u: /* D1 C8  ror ax,1 */
            ax = (uint16_t)((ax >> 1) | (ax << 15));
            break;
        case 0xC0D1u: /* D1 C0  rol ax,1 */
            ax = (uint16_t)((ax << 1) | (ax >> 15));
            break;
        case 0xC486u: /* 86 C4  xchg ah,al */
            ax = (uint16_t)((ax << 8) | (ax >> 8));
            break;
        default: /* 89 C0／8B C0  mov ax,ax */
            break;
        }
    }
    return ax;
}

/* loc_30928（2376:0658）：dh、dl 是左右遮罩 */
static void w05D6_tail(Mem *m, uint16_t ds, uint16_t cs, uint16_t si, uint8_t dh, uint8_t dl)
{
    uint16_t di, cx, ax, i;
    uint8_t bh, bl;

    for (i = 0; i < 0x3C0u; i++) /* rep stosw，AX = 0 */
        w05D6_wr16(m, ds, (uint16_t)(2u * i), 0);
    di = (uint16_t)(0x20u * (uint16_t)(w05D6_rd16(m, ds, 0x788u) % 20u));
    if ((w05D6_rd16(m, ds, 0x786u) & 8u) != 0)
        di = (uint16_t)(di + 1u);
    cx = w05D6_rd16(m, ds, 0x790u);
    do {
        uint16_t row_cx = cx, row_si = si, row_di = di;
        uint16_t w = w05D6_rd16(m, ds, 0x78Eu); /* mov cx, ds:78Eh */

        ax = w05D6_rd16(m, ds, si); /* lodsw */
        si = (uint16_t)(si + 2u);
        ax = w05D6_slot(m, cs, 0x0691u, ax);
        bh = (uint8_t)ax;
        ax = (uint16_t)((ax & 0xFF00u) | (uint8_t)((uint8_t)ax & dh));
        if (w != 0) {
            bh = (uint8_t)(bh & dl);
            do {
                w05D6_wr16(m, ds, di, ax); /* stosw */
                di = (uint16_t)(di + 2u);
                ax = w05D6_rd16(m, ds, si); /* lodsw */
                si = (uint16_t)(si + 2u);
                ax = w05D6_slot(m, cs, 0x06A5u, ax);
                bl = (uint8_t)ax;
                ax = (uint16_t)((ax & 0xFF00u) | (uint8_t)((uint8_t)ax & dh));
                ax = (uint16_t)((ax & 0xFF00u) | (uint8_t)((uint8_t)ax | bh));
                bh = (uint8_t)(bl & dl);
            } while ((uint16_t)(--w) != 0);
        }
        ax = (uint16_t)((ax & 0xFF00u) | (uint8_t)((uint8_t)ax & dl));
        wr8(m, LIN(ds, di), (uint8_t)ax);
        di = row_di;
        si = (uint16_t)(row_si + 2u * w05D6_rd16(m, ds, 0x78Eu));
        di = (uint16_t)(di + 8u);
        cx = row_cx;
    } while ((uint16_t)(--cx) != 0);
}

void f_2376_05D6(Mem *m, uint16_t ds, uint16_t cs, uint16_t si)
{
    uint16_t k;
    uint8_t dh;

    if ((w05D6_rd16(m, ds, 0x786u) & 0x0Fu) == 0) {
        f_2376_06CF(m, ds, si);
        return;
    }
    f_2376_05FC(m, ds, cs); /* push si；call；pop si：SI 不變 */
    k = (uint16_t)(w05D6_rd16(m, ds, 0x786u) & 7u);
    dh = (uint8_t)(0xFFu >> k);
    w05D6_tail(m, ds, cs, si, dh, (uint8_t)~dh);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    /* ds：描述檔 ds_seg（呼叫者 2376:0538 載入的模組資料段）；cs：dosgolem 載入位置下的 2376 程式段 */
    f_2376_05D6(m, 0x3D22u, 0x2376u, a[0]);
    return 0;
}
GIN3_REGISTER("2376_05D6", 1, run)
