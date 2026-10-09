/*
 * GIN3PS.EXE 056A:0AA5  各星域的民心往擁有者的目標值走一格（far，無參數，無回傳值）
 *
 * 原版：對星域表 ds:A79C 起 32 筆（每筆 126 bytes）逐筆：
 *   目標 ＝ 星域[+0x12] 為 0 ? 0 : (星域[+0x12] 為 2 ? 100 : 200)
 *   民心 ＝ 星域[+0x7C]（byte，零延伸成 16 位元）
 *   差 ＝ 民心 − 目標（16 位元有號）
 *   差 < 0 → 星域[+0x7C] 加一；差 > 0 → 星域[+0x7C] 減一；差 ＝ 0 → 不寫。
 * 輸入域：民心 0..255、目標 ∈ {0,100,200}，差落在 −200..255，沒有溢位；加一只發生在民心 < 目標 ≤ 200，
 *   減一只發生在民心 > 目標 ≥ 0，所以 byte 加減都不回繞。擁有者 +0x12 任意 byte：0 → 目標 0、2 → 100、其餘（含 1、3..255）→ 200。
 * 欄位：星域[+0x12] ＝ 擁有者（0 帝國、1 同盟、2 無主）、[+0x7C] ＝ 民心 0..200（高的偏帝國），
 *   docs/internals/22 第 151 行、docs/re/295 第 62 到 100 行、docs/spec/040 都與本檔一致。
 *   「擁有者 1 的目標是 200、2 的目標是 100」這個對應在 asm 裡只有 0／2／其他三段，沒有判斷 1。
 * 副作用：只寫各星域 +0x7C 的一個 byte（讀改寫）。回傳：無；離開時 AX 等殘值依最後一筆而定，呼叫者 056A:000C 不使用，不比較。
 * 呼叫者：056A:000C（每按一次「實行」命令階段處理那一串的第五支，docs/re/295）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位名稱（擁有者、民心）是 docs 的結論，位移已證實。
 */
#include "gin3_reg.h"

void f_056A_0AA5(Mem *m)
{
    uint16_t region = 0xA79C;
    uint8_t i;

    for (i = 0; i < 0x20; i++) {
        uint16_t target;
        int16_t diff;

        if (rd8(m, DS((uint16_t)(region + 0x12))) == 0)
            target = 0;
        else if (rd8(m, DS((uint16_t)(region + 0x12))) == 2)
            target = 100;
        else
            target = 200;

        diff = (int16_t)((uint16_t)rd8(m, DS((uint16_t)(region + 0x7C))) - target);
        if (diff < 0)
            wr8(m, DS((uint16_t)(region + 0x7C)), (uint8_t)(rd8(m, DS((uint16_t)(region + 0x7C))) + 1));
        else if (diff > 0)
            wr8(m, DS((uint16_t)(region + 0x7C)), (uint8_t)(rd8(m, DS((uint16_t)(region + 0x7C))) - 1));
        region = (uint16_t)(region + 0x7E);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_056A_0AA5(m);
    return 0;
}
GIN3_REGISTER("056A_0AA5", 0, run)
