/*
 * GIN3PS.EXE 2D76:0434  AI 挑提督用的分數：依艦隊類別乘上提督的一項能力，結果寫回提督記錄 +0x1A，並回傳
 *
 * 原版：far 呼叫，兩個參數（push bp、push si、沒有 sub sp，所以參數從 [bp+8] 起）：
 *   args[0..1] ＝ 提督記錄的 far 指標（位移、段）；記錄 28 bytes，表在 ds:B77C
 *   args[2]    ＝ kind（word，只取低位元組：mov bl,[bp+0Ch]）＝ 艦隊類別 0 到 3
 * 回傳：AX ＝ 分數（16 位元），同時寫進 提督[+0x1A]（word）。
 *
 * 過程（所有乘法只留低 16 位元，除法是無號 16 位元）：
 *   lvl ← 提督[+6]（byte，等級）
 *   side ← ds:8F1D（byte）；side == 1 時看 funds ＝ dword ds:8F6E（低字 ds:8F6E、高字 ds:8F70）：
 *     funds < 0x4E20（20000，高字為 0 且低字 < 0x4E20；無號）→ si ← word ds:[5FCA + lvl × 2]
 *   其餘情形（side != 1，或 funds >= 20000）→ si ← lvl × 2
 *   si ← (si × 提督[+0x0B]（byte）的低 16 位元) ÷ 100          （mul 之後 xor dx,dx，DX 的高半部被丟掉）
 *   kind 的低位元組 <= 3 時查跳躍表 ds:[6022 + kind × 2]（本檔讀記憶體裡的位址，對應到四個 case 的程式位移）：
 *     0x049F → si ← (提督[+0x14] × si) 低 16 位元       （kind 0）
 *     0x04AE → si ← (提督[+0x13] × si) 低 16 位元       （kind 1）
 *     0x04BD → si ← (提督[+0x12] × si) 低 16 位元       （kind 2）
 *     0x04CC → si ← (提督[+0x17] × si) 低 16 位元       （kind 3）
 *     0x04D9 → 不乘（default 標籤）
 *   kind 的低位元組 > 3 → 不乘（default）。
 *   (提督[+8] & 0x0F) == 1 → si ← ((si × 3) 的低 16 位元) >> 1（邏輯右移）
 *   提督[+0x1A] ← si；回傳 si。
 * 寫入落在線性 0xF0000 以上（ROM 區）時被 dosgolem 安靜丟棄，本檔按 byte 模型化；字的寫入是兩次 byte 存取，位移各自 16 位元回繞。
 *
 * 欄位語意（強推論）：[+6] ＝ 等級、[+0x0B] ＝ 統率力、[+0x14]／[+0x13]／[+0x12]／[+0x17] ＝ 四項能力、ds:8F1D ＝ 陣營、
 *   funds ＝ 資金（docs/internals/54 第 536 到 551 行：同盟且錢 < 20000 用 ds:5FCA[等級]，其餘 等級 × 2）；位移與常數已證實。
 * 輸入域：提督指標取提督表前 100 筆或任意位址（seeds 另造）；跳躍表 ds:6022 起四筆沿用基底值（049F、04AE、04BD、04CC）
 *   或換成表內另一個 case 的位址（seeds 與突變）。跳躍位址被突變成其他值時，原版跳到不相干的程式位置，本版視同不乘；
 *   這類案例若被探針收錄就是輸入域外（探針會因 INT、步數上限或讀視訊記憶體丟棄絕大多數）。
 * 證據：asm 逐指令。跳躍表位址與 case 本體的位移由函式位元組核對（insns.tsv 的指令位移 +0x6B、+0x7A、+0x89、+0x98、default +0xA5）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

static void w0434_put8(Mem *m, uint16_t seg, uint16_t off, uint8_t v)
{
    uint32_t lin = LIN(seg, off);

    if (lin < 0xF0000u)
        wr8(m, lin, v);
}

/* 字的存取是兩次 byte 存取，位移各自 16 位元回繞（dosgolem cpu.write16） */
static void w0434_put16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w0434_put8(m, seg, off, (uint8_t)v);
    w0434_put8(m, seg, (uint16_t)(off + 1), (uint8_t)(v >> 8));
}

/* 跳躍表的目標位址（程式位移）→ 要乘的欄位位移；-1 表示不乘 */
static int w0434_field(uint16_t target)
{
    switch (target) {
    case 0x049F: return 0x14;
    case 0x04AE: return 0x13;
    case 0x04BD: return 0x12;
    case 0x04CC: return 0x17;
    default: return -1;
    }
}

uint16_t f_2D76_0434(Mem *m, uint16_t rec_off, uint16_t rec_seg, uint16_t kind_word)
{
    uint8_t side = rd8(m, DS(0x8F1D));
    uint8_t kind = (uint8_t)kind_word;
    uint8_t lvl;
    uint16_t si;
    int use_table = 0;

    if (side == 1) {
        uint16_t funds_hi = rd16(m, DS(0x8F70));
        uint16_t funds_lo = rd16(m, DS(0x8F6E));

        use_table = (funds_hi == 0 && funds_lo < 0x4E20u);
    }
    lvl = rd8(m, LIN(rec_seg, rec_off + 6));
    if (use_table)
        si = rd16(m, DS(0x5FCA + lvl * 2));
    else
        si = (uint16_t)(lvl * 2);
    si = (uint16_t)((uint16_t)(si * rd8(m, LIN(rec_seg, rec_off + 0x0B))) / 100u);
    if (kind <= 3) {
        int field = w0434_field(rd16(m, DS(0x6022 + kind * 2)));

        if (field >= 0)
            si = (uint16_t)(rd8(m, LIN(rec_seg, rec_off + field)) * si);
    }
    if ((rd8(m, LIN(rec_seg, rec_off + 8)) & 0x0Fu) == 1)
        si = (uint16_t)((uint16_t)(si * 3u) >> 1);
    w0434_put16(m, rec_seg, (uint16_t)(rec_off + 0x1A), si);
    return si;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2D76_0434(m, a[0], a[1], a[2]); }
GIN3_REGISTER("2D76_0434", 3, run)
