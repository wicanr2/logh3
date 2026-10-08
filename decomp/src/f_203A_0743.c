/*
 * GIN3PS.EXE 203A:0743  功績累計（提督與副官各加一筆）
 *
 * 原版：far 呼叫，參數 = 艦隊記錄 far 指標（位移、段）、n（16 位元字，零擴充）。沒有回傳值
 *   （AX 只是 mul 的殘值，兩個呼叫端都不用），描述檔 ret = none。
 *
 * 行為（已證實：docs/re/330 §3.5 收據逐值相符；docs/internals/08 +2…+5；本檔與反組譯逐條對過）：
 *   提督表 ＝ ds:B77C，一筆 28 bytes，「+2…+5」是功績累計（u32）。
 *   主 ＝ 艦隊[+4]（提督索引，不檢查 0xFF）→ 提督[主][+2..+5] += n（32 位元加，低字進位進高字）
 *   副 ＝ 艦隊[+5]；副 != 0xFF → 提督[副][+2..+5] += n >> 1（無號邏輯右移）
 *
 * 原版是兩個 16 位元加法（add 低字、adc 高字），每次都用 ds 當段，位移 16 位元回繞。
 * 副官索引在主官的寫入「之後」才從艦隊記錄讀出（les si,[bp+0Eh] 在第一次 add 之後）。
 * 索引沒有上限檢查（0..254 都照算），超過提督表範圍的索引寫進 ds 的其他位置，本版同樣照寫，
 * 包括提督表索引撞到艦隊表本身時主官的寫入影響副官索引的情形。
 *
 * 輸入域：艦隊指標落在艦隊表 ds:C074 的 32 筆；n 的全部 16 位元值；其餘由突變導向讀取集改動。
 */
#include "gin3_reg.h"

/* 提督[idx][+2..+5] += v：低字 add、高字 adc，位移 16 位元回繞（原版的 add/adc 一對） */
static void add_merit(Mem *m, uint8_t idx, uint16_t v)
{
    uint16_t rec = (uint16_t)(0xB77Cu + (uint16_t)(idx * 0x1Cu));
    uint32_t lo_lin = DS(rec + 2);
    uint32_t hi_lin = DS(rec + 4);
    uint32_t sum = (uint32_t)rd16(m, lo_lin) + v;

    wr16(m, lo_lin, (uint16_t)sum);
    wr16(m, hi_lin, (uint16_t)(rd16(m, hi_lin) + (sum >> 16)));
}

void f_203A_0743(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t n)
{
    uint8_t chief = rd8(m, LIN(fleet_seg, fleet_off + 4));
    uint8_t deputy;

    add_merit(m, chief, n);
    deputy = rd8(m, LIN(fleet_seg, fleet_off + 5));
    if (deputy != 0xFF) {
        add_merit(m, deputy, (uint16_t)(n >> 1));
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_203A_0743(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("203A_0743", 3, run)
