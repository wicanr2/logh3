/*
 * 2376:01E6  顯示物件的新舊矩形更新：算新的圖塊矩形，與物件圖塊範圍相交後存回（near，暫存器傳參 ES:SI）
 *
 * 輸入：ES:SI = 顯示物件。呼叫者 2376:0164（第一個迴圈）與 2376:0354（`les si,[bp+6]`）。
 *   沒有堆疊參數；函式用 `enter 8,0` 開 4 個有號字區域變數 [bp-2]、[bp-4]、[bp-6]、[bp-8]，並以 BP 把這個框架交給
 *   2376:02B0（它讀 SS:[bp-2..-8]）。本版把四個值寫進基底狀態的堆疊位置 SS:[bp-n]，再呼叫 f_2376_02B0。
 *   堆疊帶（進入點 SP 以下 0x400 bytes）在比對時排除，這些寫入不影響寫入集合。
 * 原版（逐指令，欄位為物件位移；word 皆 16 位元）：
 *   [+23] ← [+1E]（byte）；[+24] ← word[+1F]；[+26] ← word[+21]                  舊旗標與舊矩形
 *   [+16] ← word[+06] + word[+12]；v2 ← 有號右移 4（[+16]）                       新矩形左（圖塊欄）
 *   [+18] ← word[+08] + word[+14]；v4 ← 有號 [+18] ÷ 20（idiv，向零取整）          新矩形上（圖塊列）
 *   v6 ← 有號右移 4（byte[+0A] + word[+16] − 1）                                  新矩形右
 *   v8 ← 有號 （byte[+0B] + word[+18] − 1）÷ 20                                    新矩形下
 *   AL ← 2376:02B0（矩形與物件 +1A..+1D 完全分離為 1，否則 0）；[+1E] ← AL；AL ≠ 0 就返回。
 *   否則把 v2..v8 夾進物件 +1A..+1D 圈出的範圍（byte 以零擴展成字、有號比較）：
 *     v2 ← max(v2, byte[+1A])；v4 ← max(v4, byte[+1B])；v6 ← min(v6, byte[+1C])；v8 ← min(v8, byte[+1D])
 *   [+1F] ← v2 低位元組；[+20] ← v4 低位元組；[+21] ← v6 − v2 + 1（byte）；[+22] ← v8 − v4 + 1（byte）。
 * 回傳：AX。分離路徑 AX 是 02B0 的回傳（AL = 1，AH 是 cbw 殘值）；其餘路徑 AH = 0（`xor ah,ah`）、AL 是 [+22] 的值。
 *   呼叫者 0164 之後只看 es:[si+23h]、es:[si+1Eh]，不用 AX；本版仍回 AX 以便逐位元比對。
 * 所有 word 存取按段內位移各自回繞（seg:FFFF 的高位元組在 seg:0000，dosgolem 行為，w01E6_rd16／wr16）。
 *
 * 語意（強推論）：docs/re/334-annihilation-crash-is-executor.md 第 29 到 33 行把 01E6 記成「存舊可見旗標與舊矩形
 *   （p[+1E]→[+23]、[+1F]w→[+24]w、[+21]w→[+26]w），算新矩形」，與本版一致；+0x06／+0x08 是世界座標、+0x12／+0x14 是偏移、
 *   +0x0A／+0x0B 是寬高、+0x1A..+0x1D 是可見圖塊範圍、20 是圖塊列高（每列 20 像素）都是對用法的推論（假說）。
 * 輸入域：ES 取 DGROUP、模組資料段、暫存區；SI 取小值並由種子造出各個 word 欄位跨越位移 0xFFFF 的情形。物件不蓋到
 *   堆疊帶。堆疊段 SS 與進入點 SP 取基底狀態的值（SS = 0x63EB，進入點 SP = 0x1F8A，由向量檔 stack_hi = 0x65E3A 反推），
 *   與 2376:02B0 的 C 版相同做法。寫入落到 0xF0000 以上（ROM）的 ES 值不涵蓋。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。欄位語意為假說。
 */
#include "gin3_reg.h"

extern uint16_t f_2376_02B0(Mem *m, uint16_t es, uint16_t ss, uint16_t si, uint16_t bp);

/* fcmd-base.state 的 SS 與進入點 SP（diffprobe 無法設定這兩個暫存器；SP = stack_hi − SS × 16） */
#define W01E6_SS 0x63EBu
#define W01E6_SP 0x1F8Au

static uint16_t w01E6_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w01E6_wr16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_2376_01E6(Mem *m, uint16_t es, uint16_t si)
{
    /* 進入點 SP 之下：近呼叫哨兵 2 bytes，enter 推 BP 2 bytes，BP 指到那個 BP 副本 */
    uint16_t bp = (uint16_t)(W01E6_SP - 4u);
    int16_t v2, v4, v6, v8;
    uint16_t ax;

    wr8(m, LIN(es, (uint16_t)(si + 0x23u)), rd8(m, LIN(es, (uint16_t)(si + 0x1Eu))));
    w01E6_wr16(m, es, (uint16_t)(si + 0x24u), w01E6_rd16(m, es, (uint16_t)(si + 0x1Fu)));
    w01E6_wr16(m, es, (uint16_t)(si + 0x26u), w01E6_rd16(m, es, (uint16_t)(si + 0x21u)));

    ax = (uint16_t)(w01E6_rd16(m, es, (uint16_t)(si + 6u)) + w01E6_rd16(m, es, (uint16_t)(si + 0x12u)));
    w01E6_wr16(m, es, (uint16_t)(si + 0x16u), ax);
    v2 = (int16_t)((int16_t)ax >> 4);

    ax = (uint16_t)(w01E6_rd16(m, es, (uint16_t)(si + 8u)) + w01E6_rd16(m, es, (uint16_t)(si + 0x14u)));
    w01E6_wr16(m, es, (uint16_t)(si + 0x18u), ax);
    v4 = (int16_t)((int16_t)ax / 20);

    ax = (uint16_t)(rd8(m, LIN(es, (uint16_t)(si + 0x0Au))) + w01E6_rd16(m, es, (uint16_t)(si + 0x16u)));
    ax = (uint16_t)(ax - 1u);
    v6 = (int16_t)((int16_t)ax >> 4);

    ax = (uint16_t)(rd8(m, LIN(es, (uint16_t)(si + 0x0Bu))) + w01E6_rd16(m, es, (uint16_t)(si + 0x18u)));
    ax = (uint16_t)(ax - 1u);
    v8 = (int16_t)((int16_t)ax / 20);

    /* 區域變數放在堆疊上，02B0 從 SS:[bp-n] 讀 */
    w01E6_wr16(m, W01E6_SS, (uint16_t)(bp - 2u), (uint16_t)v2);
    w01E6_wr16(m, W01E6_SS, (uint16_t)(bp - 4u), (uint16_t)v4);
    w01E6_wr16(m, W01E6_SS, (uint16_t)(bp - 6u), (uint16_t)v6);
    w01E6_wr16(m, W01E6_SS, (uint16_t)(bp - 8u), (uint16_t)v8);
    ax = f_2376_02B0(m, es, W01E6_SS, si, bp);
    wr8(m, LIN(es, (uint16_t)(si + 0x1Eu)), (uint8_t)ax);
    if ((uint8_t)ax != 0)
        return ax;

    ax = rd8(m, LIN(es, (uint16_t)(si + 0x1Au))); /* xor ah,ah；mov al,… */
    if ((int16_t)ax > v2)
        v2 = (int16_t)ax;
    ax = rd8(m, LIN(es, (uint16_t)(si + 0x1Bu)));
    if ((int16_t)ax > v4)
        v4 = (int16_t)ax;
    ax = rd8(m, LIN(es, (uint16_t)(si + 0x1Cu)));
    if ((int16_t)ax < v6)
        v6 = (int16_t)ax;
    ax = rd8(m, LIN(es, (uint16_t)(si + 0x1Du)));
    if ((int16_t)ax < v8)
        v8 = (int16_t)ax;

    wr8(m, LIN(es, (uint16_t)(si + 0x1Fu)), (uint8_t)v2);
    wr8(m, LIN(es, (uint16_t)(si + 0x20u)), (uint8_t)v4);
    wr8(m, LIN(es, (uint16_t)(si + 0x21u)), (uint8_t)((uint8_t)v6 - (uint8_t)v2 + 1u));
    ax = (uint16_t)(uint8_t)((uint8_t)v8 - (uint8_t)v4 + 1u);
    wr8(m, LIN(es, (uint16_t)(si + 0x22u)), (uint8_t)ax);
    return ax;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2376_01E6(m, a[0], a[1]); }
GIN3_REGISTER("2376_01E6", 2, run)
