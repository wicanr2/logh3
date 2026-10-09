/*
 * 2376:087A  圖塊（X 不是 8 的倍數時帶位移）複製進緩衝區（near，暫存器傳參 SI，DS = 模組資料段 3D22）
 *
 * 輸入：DS:SI = 來源圖形（呼叫者 2376:07FA 的呼叫鏈在 `call 2376:087A` 前準備）；
 *   ds:786 = X、ds:78E = w（每列字數）、ds:790 = R（列數）、ds:784 = 08A0 的位移次數快取。
 *   方向旗標 DF = 0。
 * 原版：
 *   X & 7 = 0（`test ax,7` 為 0）：直接 `call 2376:0958`（整字對齊搬運），返回。
 *   否則：
 *     push si；call 2376:08A0（依 X & 7 改寫程式碼段 2376 的兩段位移指令）；pop si。
 *     k ← X & 7；DH ← 0xFF >> k；DL ← ~DH（左右遮罩）；`call loc_30BCD`（= 2376:08FD）。
 *   2376:08FD 是 IDA 沒有切成函式的一段（位元組 8C D8 … C3，共 0x5B bytes，台帳沒有這一項）。與 2376:05D6 呼叫的
 *   2376:0658 同一型，差別：沒有「依 Y mod 20 的起點」與「X & 8 加 1」（DI 起點 0）、列距 0x10 bytes（0658 是 8），
 *   兩段位移指令在程式碼段 cs:091A..0923 與 cs:092E..0937（2376:08A0 改寫的位置）。逐指令：
 *     1. rep stosw：ES(= DS):0 起 0x3C0 個 word 填 0（1920 bytes）；DI ← 0。
 *     2. CX ← R；外圈 R 次（`loop`）：push CX／SI／DI；CX ← w；AX ← lodsw；AX ← 位移指令A（5 個）；BH ← AL；AL ← AL & DH；
 *        w = 0（jcxz）跳到收尾；否則 BH ← BH & DL；內圈 w 次：
 *            stosw；AX ← lodsw；AX ← 位移指令B（5 個）；BL ← AL；AL ← AL & DH；AL ← AL | BH；BH ← BL；BH ← BH & DL
 *        收尾：AL ← AL & DL；byte ES:[DI] ← AL；pop DI／SI／CX；SI ← SI + 2 × word(ds:78E)（重讀）；DI ← DI + 0x10。
 *   位移指令的 2-byte 編碼與解法同 2376:05D6（自我修改，從程式碼段讀出來解碼）：
 *     89 C0／8B C0 = mov ax,ax；D1 C8 = ror ax,1；D1 C0 = rol ax,1；86 C4 = xchg ah,al。
 * 輸出：記憶體寫入。暫存器是暫用值，呼叫者 2376:07FA 之後 `pop` 還原；描述檔 ret = none。
 * 位移 0xFFFF 的 word 存取：高位元組取同一段的位移 0（dosgolem 行為）。
 *
 * 語意（強推論）：與 2376:05D6 平行，是 X 對齊 8（而非 16）的版本；緩衝區列距 16 bytes（0958 同）。docs/re/407-fleet-blocks.md
 *   與 spec/095 沒有 087A／08FD 的逐位元組記載。
 * 輸入域：SI 全域；ds:786 與 ds:784 的全部 16 位元值；ds:78E（w）與 ds:790（R）凍結，由種子給定
 *   （w ∈ {0,1,2,3,4,8,12}，R ∈ {1,2,3,5,10,20,64,250,300}），不涵蓋 R = 0。DF = 1 與緩衝區蓋到呼叫堆疊不涵蓋。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；位移指令語意來自位元組解碼；用途為強推論。
 */
#include "gin3_reg.h"

extern void f_2376_08A0(Mem *m, uint16_t ds, uint16_t cs);
extern void f_2376_0958(Mem *m, uint16_t ds, uint16_t si);

static uint16_t w087A_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w087A_wr16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

/* 執行程式碼段 cs:off 起的 5 個 2-byte 位移指令（自我修改的位置），輸入輸出是 AX */
static uint16_t w087A_slot(const Mem *m, uint16_t cs, uint16_t off, uint16_t ax)
{
    int i;

    for (i = 0; i < 5; i++) {
        switch (w087A_rd16(m, cs, (uint16_t)(off + 2u * (uint16_t)i))) {
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

/* loc_30BCD（2376:08FD）：dh、dl 是左右遮罩 */
static void w087A_tail(Mem *m, uint16_t ds, uint16_t cs, uint16_t si, uint8_t dh, uint8_t dl)
{
    uint16_t di = 0, cx, ax, i;
    uint8_t bh, bl;

    for (i = 0; i < 0x3C0u; i++) /* rep stosw，AX = 0 */
        w087A_wr16(m, ds, (uint16_t)(2u * i), 0);
    cx = w087A_rd16(m, ds, 0x790u);
    do {
        uint16_t row_cx = cx, row_si = si, row_di = di;
        uint16_t w = w087A_rd16(m, ds, 0x78Eu); /* mov cx, ds:78Eh */

        ax = w087A_rd16(m, ds, si); /* lodsw */
        si = (uint16_t)(si + 2u);
        ax = w087A_slot(m, cs, 0x091Au, ax);
        bh = (uint8_t)ax;
        ax = (uint16_t)((ax & 0xFF00u) | (uint8_t)((uint8_t)ax & dh));
        if (w != 0) {
            bh = (uint8_t)(bh & dl);
            do {
                w087A_wr16(m, ds, di, ax); /* stosw */
                di = (uint16_t)(di + 2u);
                ax = w087A_rd16(m, ds, si); /* lodsw */
                si = (uint16_t)(si + 2u);
                ax = w087A_slot(m, cs, 0x092Eu, ax);
                bl = (uint8_t)ax;
                ax = (uint16_t)((ax & 0xFF00u) | (uint8_t)((uint8_t)ax & dh));
                ax = (uint16_t)((ax & 0xFF00u) | (uint8_t)((uint8_t)ax | bh));
                bh = (uint8_t)(bl & dl);
            } while ((uint16_t)(--w) != 0);
        }
        ax = (uint16_t)((ax & 0xFF00u) | (uint8_t)((uint8_t)ax & dl));
        wr8(m, LIN(ds, di), (uint8_t)ax);
        di = row_di;
        si = (uint16_t)(row_si + 2u * w087A_rd16(m, ds, 0x78Eu));
        di = (uint16_t)(di + 0x10u);
        cx = row_cx;
    } while ((uint16_t)(--cx) != 0);
}

void f_2376_087A(Mem *m, uint16_t ds, uint16_t cs, uint16_t si)
{
    uint16_t k;
    uint8_t dh;

    if ((w087A_rd16(m, ds, 0x786u) & 7u) == 0) {
        f_2376_0958(m, ds, si);
        return;
    }
    f_2376_08A0(m, ds, cs); /* push si；call；pop si：SI 不變 */
    k = (uint16_t)(w087A_rd16(m, ds, 0x786u) & 7u);
    dh = (uint8_t)(0xFFu >> k);
    w087A_tail(m, ds, cs, si, dh, (uint8_t)~dh);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    /* ds：描述檔 ds_seg（呼叫者 2376:07FA 載入的模組資料段）；cs：dosgolem 載入位置下的 2376 程式段 */
    f_2376_087A(m, 0x3D22u, 0x2376u, a[0]);
    return 0;
}
GIN3_REGISTER("2376_087A", 1, run)
