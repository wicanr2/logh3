/*
 * GIN3PS.EXE 2B99:007D  建立星域圖的初始距離矩陣（某一陣營視角）：32 × 32 個 word，寫進該陣營的距離表
 *
 * 原版：far 呼叫，一個堆疊字 faction（push bp、si、di、sub sp,16h，所以參數在 [bp+20h]；只取低位元組：cmp byte ptr [bp+20h]）。
 *   無回傳值（AX 是殘留值，呼叫者 2B99:0068 不用；描述檔 ret=none）。
 * 輸出表：faction == 0（低位元組）→ far 指標在 ds:D062（位移）／ds:D064（段）；否則在 ds:D066／ds:D068。
 *   指標在函式一開始讀一次（複製到堆疊），之後表寫入蓋到指標格自己也不影響這次的位址。
 * 過程（i 與 j 都是 byte 計數器，0 到 31，無號 byte 比較）：
 *   w ← ds:[533A + i × 32 + j]（邊權表，byte）
 *   w == 0xFF（不相鄰）→ 表[i × 64 + j × 2]（word）← 0x7FFF
 *   否則 表[i × 64 + j × 2] ← w（零延伸成 word）；
 *        再看星域記錄 ds:[A79C + i × 0x7E + 0x12] 與 ds:[A79C + j × 0x7E + 0x12]（byte，擁有者）：
 *        任何一個 != faction（低位元組）→ 該 word += 0x400（word 讀改寫，16 位元回繞）
 * 表位移 ＝ 指標位移 ＋ (i << 6) ＋ (j << 1)，16 位元運算；字的存取是兩次 byte 存取、位移各自 16 位元回繞（DS:FFFF 的 word 取 FFFF 與 0000）；
 * 寫入落在線性 0xF0000 以上（ROM 區）時被 dosgolem 安靜丟棄，本檔按 byte 模型化（加 0x400 的讀仍讀 ROM 內容）。
 * 寫入表之後才讀星域記錄，所以表與星域記錄或邊權表重疊時，後面的讀取看到新值。
 *
 * 欄位語意（強推論）：邊權表 ds:533A 是 32 × 32 byte（0xFF ＝ 不相鄰），星域記錄 126 bytes、+0x12 是擁有者
 *   （docs/internals/26 第 63 到 70 行、docs/re/243 第 43 到 74 行）。位移與常數已證實。
 * 輸入域：faction 全部 16 位元值（只取低位元組）；表指標與星域、邊權資料任意（來自基底狀態、突變與 seeds）。
 *   讀 A0000 到 BFFFF 的案例由探針丟棄；表蓋到本函式自己的程式碼上的案例（突變把段改到那裡）不在比較範圍。
 * 證據：asm 逐指令；docs/re/243 的虛擬碼（0x7FFF、+0x400、擁有者條件）與本檔一致。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

static void w007d_put8(Mem *m, uint16_t seg, uint16_t off, uint8_t v)
{
    uint32_t lin = LIN(seg, off);

    if (lin < 0xF0000u)
        wr8(m, lin, v);
}

/* 字的存取是兩次 byte 存取，位移各自 16 位元回繞（dosgolem cpu.read16／write16） */
static uint16_t w007d_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, off + 1)) << 8));
}

static void w007d_put16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w007d_put8(m, seg, off, (uint8_t)v);
    w007d_put8(m, seg, (uint16_t)(off + 1), (uint8_t)(v >> 8));
}

void f_2B99_007D(Mem *m, uint16_t faction_word)
{
    uint8_t faction = (uint8_t)faction_word;
    uint16_t tbl_seg;
    uint16_t tbl_off;
    uint16_t rec_i = 0xA79C;
    uint8_t i;
    uint8_t j;

    if (faction == 0) {
        tbl_seg = rd16(m, DS(0xD064));
        tbl_off = rd16(m, DS(0xD062));
    } else {
        tbl_seg = rd16(m, DS(0xD068));
        tbl_off = rd16(m, DS(0xD066));
    }
    for (i = 0; i < 0x20; i++) {
        uint16_t rec_j = 0xA79C;

        for (j = 0; j < 0x20; j++) {
            uint8_t w = rd8(m, DS(0x533A + (i << 5) + j));
            uint16_t cell = (uint16_t)(tbl_off + (uint16_t)(i << 6) + (uint16_t)(j << 1));

            if (w == 0xFFu) {
                w007d_put16(m, tbl_seg, cell, 0x7FFF);
            } else {
                w007d_put16(m, tbl_seg, cell, w);
                if (rd8(m, DS(rec_i + 0x12)) != faction || rd8(m, DS(rec_j + 0x12)) != faction)
                    w007d_put16(m, tbl_seg, cell, (uint16_t)(w007d_rd16(m, tbl_seg, cell) + 0x400u));
            }
            rec_j = (uint16_t)(rec_j + 0x7E);
        }
        rec_i = (uint16_t)(rec_i + 0x7E);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2B99_007D(m, a[0]);
    return 0;
}
GIN3_REGISTER("2B99_007D", 1, run)
