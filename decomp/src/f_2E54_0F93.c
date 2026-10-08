/*
 * GIN3PS.EXE 2E54:0F93  費沙（星域 14）能不能當戰略 AI 的候選目標（無參數，far，回 AX = 0 或 1）
 *
 * 反組譯逐條對應（IDA sub_3C043）：
 *   本地 far 指標 = ds:AE02（星域表 ds:A79C 的第 13 筆，13 × 126 = 0x666；星域 13 = 伊謝爾倫）
 *   ds:8F48 != 0                       → 1（放行旗標）
 *   ds:8F1D == 0                       → (伊謝爾倫[+0x12] == 1 且 word ds:8F18 >= 0xEE ) ? 1 : 0
 *   ds:8F1D == 1                       → (伊謝爾倫[+0x12] == 0 且 word ds:8F18 >= 0x13E) ? 1 : 0
 *   其餘                               → 0
 * 比較是無號的 word 比較（jnb）。函式沒有寫入任何全域記憶體。
 *
 * 推論等級
 *   已證實（差異測試的位元組與回傳值）：上列分支結構與常數。
 *   強推論（docs/re/178、docs/internals/54）：ds:8F1D 是陣營（0 帝國、1 同盟）、ds:8F18 是回合計數器、
 *     星域記錄 +0x12 是擁有者。名稱不是證據，欄位語意以 docs 為準。
 *   未知：ds:8F48 的寫入者（docs/re/178 同樣記為未查）。
 *   與舊文件的差異：docs/re/178 的虛擬碼寫「owner == 1 且回合 < 238 回 0、其餘回 1」。
 *     原版位元組是「owner 不等於指定值一律回 0；等於指定值時回合 >= 門檻才回 1」，兩者在
 *     owner 不等於指定值時結果相反。以本檔（差異測試通過者）為準。
 *
 * 輸入域：全部輸入都是記憶體（無參數）。ds:8F1D 在 0..255 全域、星域記錄與回合計數器為任意值。
 */
#include "gin3_reg.h"

uint16_t f_2E54_0F93(Mem *m)
{
    uint32_t rec = DS(0xAE02); /* 星域 13 記錄 */
    uint8_t side;

    if (rd8(m, DS(0x8F48)) != 0)
        return 1;
    side = rd8(m, DS(0x8F1D));
    if (side == 0) {
        if (rd8(m, rec + 0x12) == 1 && rd16(m, DS(0x8F18)) >= 0xEEu)
            return 1;
        return 0;
    }
    if (side == 1) {
        if (rd8(m, rec + 0x12) == 0 && rd16(m, DS(0x8F18)) >= 0x13Eu)
            return 1;
        return 0;
    }
    return 0;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_2E54_0F93(m);
}
GIN3_REGISTER("2E54_0F93", 0, run)
