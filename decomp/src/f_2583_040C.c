/*
 * GIN3PS.EXE 2583:040C  戰場像素 (x, y) 這一點能不能站：地形不通回 0xFFFF；地形通則看被誰佔著（池索引 ＋ 1，沒人佔回 0）
 *
 * 原版（69 bytes，far 呼叫，三個堆疊字 x、y、faction；序言 push bp／si／di、sub sp,2，所以 x 在 [bp+0Ch]、y 在 [bp+0Eh]、
 *   faction 在 [bp+10h]），回 AX（區域變數 [bp+0] 暫存結果）：
 *   r ← 1588:04EA(x, y)                       ; 16 × 16 範圍全可通行 → 非 0
 *   r == 0（整個 AX 為 0）→ 回 0xFFFF
 *   否則 r ← 17A0:06D7(x, y, faction)          ; 在場分隊清單裡誰佔著這一點：池索引 ＋ 1，沒有回 0xFFFF
 *     r == 0xFFFF → r ← 0
 *   回 r
 * 過程（已證實：反組譯逐指令；1588:04EA、17A0:06D7 各由其差異測試證實）。06D7 回 0xFFFF 與「池索引 0xFFFE ＋ 1」不衝突：
 *   後者是 0xFFFF 本身（16 位元加法），同樣被改成 0，C 版逐步照做。
 * 語意（強推論）：docs/re/102-battle-movement.md 第 35、50 行「r ＝ 能不能站(nx, ny, 陣營) ; 2583:040C」、
 *   docs/spec/081-battle-fleet-step.md 第 45 行「地形不通回FFFF；地形通才掃敵分隊，略過同陣營、種類7、Action bit7」，與本檔一致；
 *   docs/re/128-fleet-budget-accounting.md 第 58 到 64 行「0 才可以走」：回 0 表示通行且無人佔，回 0xFFFF 是地形擋，
 *   其餘（池索引 ＋ 1）是被佔。
 * 副作用：無（04EA、06D7 都不寫記憶體）。
 * 輸入域：x、y、faction 全部 16 位元值；地圖、圖塊遮罩表、在場分隊清單 ds:98BA（筆數 word ds:9896）與艦艇池取自記憶體，
 *   基底 fcmd-standby.state（04EA 需要戰場地圖）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_1588_04EA(Mem *m, uint16_t x, uint16_t y);
extern uint16_t f_17A0_06D7(Mem *m, uint16_t x, uint16_t y, uint16_t faction);

uint16_t f_2583_040C(Mem *m, uint16_t x, uint16_t y, uint16_t faction)
{
    uint16_t r;

    if (f_1588_04EA(m, x, y) == 0)
        return 0xFFFFu;
    r = f_17A0_06D7(m, x, y, faction);
    if (r == 0xFFFFu)
        r = 0;
    return r;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2583_040C(m, a[0], a[1], a[2]); }
GIN3_REGISTER("2583_040C", 3, run)
