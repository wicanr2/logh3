/*
 * GIN3PS.EXE 25DB:04E7  艦隊的繪製方向索引：依旗標把 艦隊[+0x6A] 設成 8 方向索引或 9
 *
 * 原版：far 呼叫，三個堆疊字：艦隊記錄的 far 指標（位移、段）與旗標 flag（word）。
 *   序言 push bp／si、mov bp,sp，所以指標在 [bp+8]（位移）與 [bp+0Ah]（段）、旗標在 [bp+0Ch]。無回傳值（描述檔 ret = none；
 *   AX 在返回時是殘值，呼叫者 25DB:03E9 之後不讀它）。
 *
 * 語意（已證實：與反組譯逐條對過；欄位名稱為強推論，見下）：
 *   flag（整個 word，`or si,si`）== 0：艦隊[+0x6A] ← 9（byte）。
 *   否則：d ← 3982:0022(艦隊[+0x4A] 的低位元組)（`push word es:[bx+4Ah]`，被呼叫者只取低位元組），艦隊[+0x6A] ← d（byte）。
 *   3982:0022 把 256 刻的角度換成 0..7 的方向索引。
 *
 * 欄位語意（強推論）：艦隊[+0x4A] 是朝向角度（256 刻一圈），艦隊[+0x6A] 是 8 方向的繪製索引（9 表示無方向，
 *   `25DB:0670` 在 [+0x6A] 為 9 時不覆寫它）。docs/internals/33-fleet-table.md 與 docs/re/326 沒有這支函式的逐位元組記載。
 *
 * 模型限制：
 *   - 位址 0xF0000 以上是 BIOS ROM，寫入無效（dosgolem 的 ROMBase ＝ 0xF0000），put8 不動作；讀仍照基底傾印。
 *   - 讀 艦隊[+0x4A] 的 word 在位移 0xFFFF 時高位元組取同一段的位移 0（dosgolem 行為）；被呼叫者不用高位元組，
 *     所以這裡只讀低位元組，結果相同。
 *
 * 輸入域：艦隊指標取艦隊表任一筆（32 筆）或種子指定的位址；flag 的全部 16 位元值。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_3982_0022(uint8_t angle);

static void put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_25DB_04E7(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t flag)
{
    if (flag == 0) {
        put8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x6Au)), 9);
        return;
    }
    put8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x6Au)),
         (uint8_t)f_3982_0022(rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x4Au)))));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_25DB_04E7(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("25DB_04E7", 3, run)
