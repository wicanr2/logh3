/*
 * GIN3PS.EXE 20B4:0E2C  行星換手判定：攻擊力追上剩餘防衛力了嗎
 *
 * 原版：far 呼叫，參數 = 行星記錄 far 指標（位移、段）、A（16 位元字，攻擊力）。回 AX：
 *   A >= 行星[+0x14]（無號 16 位元比較，cmp ax,cx ＋ jnb）→ 1，否則 0。
 * 沒有寫入。
 *
 * 語意（已證實：docs/re/330 §3.2 "回 (A >= 行星[+0x14]) ? 1 : 0（無號）"；docs/re/151、docs/spec/019
 *   同結論，呼叫端在扣完防衛力之後用 A = 203A:0424(艦隊) 呼叫；+0x14 是行星防衛力，docs/internals/29）。
 *   本檔與反組譯逐條對過。
 *
 * 輸入域：行星指標落在行星表 81 筆內（ds:5336/5338 起，一筆 82 bytes）；A 的全部 16 位元值。
 */
#include "gin3_reg.h"

uint16_t f_20B4_0E2C(Mem *m, uint16_t planet_off, uint16_t planet_seg, uint16_t a)
{
    return a >= rd16(m, LIN(planet_seg, planet_off + 0x14)) ? 1 : 0;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_20B4_0E2C(m, a[0], a[1], a[2]);
}
GIN3_REGISTER("20B4_0E2C", 3, run)
