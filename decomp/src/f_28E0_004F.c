/*
 * GIN3PS.EXE 28E0:004F  提督的軍階字串指標：依提督記錄的陣營位元組選表，依能力值換成的等級索引取 far 指標
 *
 * 原版：far 呼叫，一個參數＝提督記錄（28 bytes）的 far 指標（位移、段）。序言 push bp、mov bp,sp，所以指標在 [bp+6]（位移）與
 *   [bp+8]（段）。回傳 far 指標：AX ＝ 段、BX ＝ 位移（描述檔 ret = pack:ax,bx）。23 個呼叫者。
 *
 * 語意（已證實：與反組譯逐條對過）：
 *   記錄[+1] == 0（byte）：表 A（位移 ds:6830 起，每筆 4 bytes）；否則表 B（位移 ds:689C 起，每筆 4 bytes）。
 *   v ← 28E0:0096(記錄[+6] 的低位元組)（0 到 11；`push word ptr es:[bx+6]`，被呼叫者只取低位元組）。
 *   idx ← v × 4（`mov bx,ax; xor bh,bh; shl bx,1; shl bx,1`，只取 AL）。
 *   回傳 BX ← word ds:(表 + idx)（位移）、AX ← word ds:(表 + idx + 2)（段）。
 *
 * 欄位語意（強推論）：記錄[+1] 是陣營（0 為帝國，1 為同盟；2703:0214 的檔頭稱艦隊 +0x01 為陣營，提督記錄的 +1 同樣用法為假說）；
 *   記錄[+6] 是能力值（28E0:0096 把它換成 0 到 11 的等級，docs/internals/08 第 217 行）；兩張表各 12 筆，是各軍階名稱字串的 far 指標
 *   （docs/internals/08 第 217 行：「用等級查字串表，只用 AL」）。
 *
 * 輸入域：提督指標取提督表任一筆（82 筆）；記錄[+1]、[+6] 與兩張表的內容由突變與種子涵蓋。
 * 推論等級：已證實（反組譯逐指令＋差異測試，0096 本身另有差異測試）。
 */
#include "gin3_reg.h"

extern uint8_t f_28E0_0096(Mem *m, uint16_t v_word);

uint32_t f_28E0_004F(Mem *m, uint16_t rec_off, uint16_t rec_seg)
{
    uint16_t table = (rd8(m, LIN(rec_seg, (uint16_t)(rec_off + 1u))) == 0) ? 0x6830u : 0x689Cu;
    uint8_t level = f_28E0_0096(m, rd8(m, LIN(rec_seg, (uint16_t)(rec_off + 6u))));
    uint16_t idx = (uint16_t)((uint16_t)level << 2);
    uint16_t seg = rd16(m, DS((uint16_t)(table + idx + 2u)));
    uint16_t off = rd16(m, DS((uint16_t)(table + idx)));

    return ((uint32_t)seg << 16) | off;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_28E0_004F(m, a[0], a[1]); }
GIN3_REGISTER("28E0_004F", 2, run)
