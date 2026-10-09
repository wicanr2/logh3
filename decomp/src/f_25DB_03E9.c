/*
 * GIN3PS.EXE 25DB:03E9  戰場單位物件（艦艇記錄，78 bytes）的圖示索引：依旗標與物件欄位算出 obj[+0x2F]
 *
 * 原版（254 bytes，far 呼叫，三個堆疊字：物件 far 指標（位移、段）與旗標 force（word）；push bp／si／di、sub sp,6，
 *   所以物件在 [bp+10h]／[bp+12h]、force 在 [bp+14h]；沒有回傳值，描述檔 ret ＝ none）：
 *   base ← (obj[+0] == 0) ? 0x14 : 0x32                    ; 存 [bp+0]（堆疊區域變數，byte）
 *   force != 0（整個 word，`or di,di`）或 obj[+0x17] == 3：
 *     obj[+3] == 0：
 *         艦隊 ← ds:(0xC074 ＋ obj[+1] × 0x7A)              ; mul cx（cx ＝ 7Ah），obj[+1] <= 255 所以乘積不超過 16 位元
 *         call 25DB:04E7(艦隊, ds, 1)                       ; 已 diff-tested：艦隊[+0x6A] ← 3982:0022(艦隊[+0x4A])
 *         obj[+0x2F] ← 艦隊[+2] ＋ base − 1                  ; 呼叫之後才讀艦隊[+2]；add al 與 dec ax 只存 AL，等於 byte 運算
 *     否則：obj[+0x2F] ← obj[+4] ＋ 0x10 ＋ base            ; byte
 *   否則（force ＝ 0 且 obj[+0x17] != 3）：
 *     obj[+3] == 0：call 25DB:04E7(艦隊 ＝ ds:(0xC074 ＋ obj[+1] × 0x7A), ds, 0)  ; 艦隊[+0x6A] ← 9
 *     之後（obj[+3] 非 0 時直接落到這裡，沒有 jmp）：重讀 obj[+4]、obj[+0x17]：
 *       obj[+4] == 7：obj[+0x2F] ← 9
 *       否則依 obj[+0x17]：0 → obj[+0x2F] ← 9；1 → ← base ＋ 0x18；2 → ← obj[+4] ＋ 0x10 ＋ base；其他值不寫。
 * 每一步都重新由 `les bx,[bp+10h]` 載入物件指標並讀欄位，所以物件與艦隊記錄別名時，04E7 的寫入會被後面的讀取看到（本版逐步照做）。
 * 副作用：obj[+0x2F]（byte）、艦隊[+0x6A]（04E7 寫）；base 存在堆疊上（帶內，比對時排除）。
 *
 * 呼叫者（台帳）：2526:033B（hw-dependent）等四處。docs/spec/018-battle-resolution.md 第 57 行「己方走 25DB:03E9 的完整分支」、
 *   docs/spec/087-battle-spotting-phase.md 第 95 行「以 force0 呼叫 25DB:03E9 重算圖示」、docs/internals/38-battle-units.md
 *   第 248、509 行「FLEET+3A 為 FF 仍以 force0 呼叫 25DB:03E9；艦種 7 且 Detail0 於 0498 寫 SHIP+2F 為 9」。本檔把 obj[+4] ＝ 7 寫 9 的分支
 *   與「艦種 7」對上（obj[+4] ＝ 艦種，obj[+0x17] ＝ 偵察層級），欄位語意為強推論，位移與運算是證據。
 * 模型限制：線性位址 0xF0000 以上是 ROM，寫入無效（dosgolem 的 ROMBase）；本檔 put8 模型化。obj 在 ROM 區時 obj[+0x2F] 的寫入被丟棄。
 * 輸入域：物件指標取暫存區、跨位移 0xFFFF 的位置、ROM；force 全部 16 位元值；艦隊編號 obj[+1] 0..255（艦隊表只有 32 筆，
 *   更大的編號指到表之後的記憶體，照原版運算）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位語意為強推論。
 */
#include "gin3_reg.h"

extern void f_25DB_04E7(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t flag);

static void w03E9_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_25DB_03E9(Mem *m, uint16_t obj_off, uint16_t obj_seg, uint16_t force)
{
#define OB(d) LIN(obj_seg, (uint16_t)(obj_off + (d)))
    uint8_t base = rd8(m, OB(0)) == 0 ? 0x14u : 0x32u;

    if (force != 0 || rd8(m, OB(0x17)) == 3) {
        if (rd8(m, OB(3)) == 0) {
            uint16_t fleet = (uint16_t)(0xC074u + (uint16_t)(rd8(m, OB(1)) * 0x7Au));

            f_25DB_04E7(m, fleet, GIN3_DS, 1u);
            w03E9_put8(m, OB(0x2F), (uint8_t)(rd8(m, DS((uint16_t)(fleet + 2u))) + base - 1u));
        } else {
            w03E9_put8(m, OB(0x2F), (uint8_t)(rd8(m, OB(4)) + 0x10u + base));
        }
        return;
    }
    if (rd8(m, OB(3)) == 0) {
        uint16_t fleet = (uint16_t)(0xC074u + (uint16_t)(rd8(m, OB(1)) * 0x7Au));

        f_25DB_04E7(m, fleet, GIN3_DS, 0u);
    }
    if (rd8(m, OB(4)) == 7) {
        w03E9_put8(m, OB(0x2F), 9);
        return;
    }
    switch (rd8(m, OB(0x17))) {
    case 0:
        w03E9_put8(m, OB(0x2F), 9);
        break;
    case 1:
        w03E9_put8(m, OB(0x2F), (uint8_t)(base + 0x18u));
        break;
    case 2:
        w03E9_put8(m, OB(0x2F), (uint8_t)(rd8(m, OB(4)) + 0x10u + base));
        break;
    default:
        break;
    }
#undef OB
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_25DB_03E9(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("25DB_03E9", 3, run)
