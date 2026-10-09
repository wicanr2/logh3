/*
 * GIN3PS.EXE 297B:0C74  戰場像素 (x, y) 這一點能不能站（給戰略層的艦隊 F 用）：回 1 ＝ 空著，0 ＝ 出界、地形擋住、或同星域別支艦隊有分隊在 ±16 內
 *
 * 原版（293 bytes，far 呼叫，四個堆疊字；序言 push bp／si／di、sub sp,10h，所以 F 的位移在 [bp+1Ah]、段在 [bp+1Ch]，x 在 [bp+1Eh]，y 在 [bp+20h]），回 AX。
 *   區域變數：[bp+0] ＝ 艦隊序號 i（byte）、[bp+1] ＝ 分隊序號 j（byte）、[bp+2] ＝ r、[bp+4]／[bp+6] ＝ 艦隊 G 的分隊清單指標（位移／段）、
 *   [bp+8]／[bp+0Ah] ＝ G 指標（位移／段）、[bp+0Ch]／[bp+0Eh] ＝ 艦艇指標。
 * 過程（已證實：反組譯逐指令；1588:04EA 由其差異測試證實）：
 *   x < 0 或 x >= 0x7EF 或 y < 0 或 y >= 0x7E7（有號 js／jge／jl）→ 回 0
 *   r ← 1588:04EA(x, y)                                              ; 16 × 16 範圍全可通行 → 1，否則 0
 *   G ← DS:0xC074；i ← 0（byte）
 *   重複：
 *     i >= 0x20（無號）→ 離開
 *     r ＝ 0（整個 word）→ 離開                                       ; 每支艦隊一開始才檢查，所以離開不影響結果（沒有寫入）
 *     G[+4] ＝ 0xFF（byte）→ 下一支                                    ; 不在場
 *     G[+6] ≠ F[+6]（byte，星域）→ 下一支
 *     G 與 F 是同一個 far 指標（段與位移都相等）→ 下一支               ; 跳過自己
 *     j ← 0；清單 ← G ＋ 0x10（位移 16 位元回繞，段 ＝ G 的段）
 *     j >= G[+0x0F]（無號 byte，G 每圈重讀）→ 下一支
 *       idx ← word 清單[0]；艦艇 ← (word ds:CFB4 ＋ low16(idx × 0x4E)) : word ds:CFB6   ; 池基底每圈重讀
 *       (x − 16 回繞後) < 艦艇[+0x34]（無號）且 (x ＋ 16) > 艦艇[+0x34]（無號）
 *         且 (y − 16) < 艦艇[+0x36] 且 (y ＋ 16) > 艦艇[+0x36]（皆無號 16 位元，回繞）→ r ← 0（不離開分隊迴圈）
 *       j++；清單 ＋= 4
 *     i++；G ＋= 0x7A
 *   回 r
 * 語意（強推論）：docs/re/335-ai-order-command-kind.md 第 78 行（`297B:0C74(f, x, y) x ∉ [0, 0x7EF) 或 y ∉ [0, 0x7E7) → 0；r ← 1588:04EA(x, y)`）、
 *   docs/re/330-strategic-planet-assault.md 第 271、284 行（`297B:0C74(f, x, y)：出界回 0；否則 r ← 1588:04EA（地圖可站），再掃同星域（+6 相同）、
 *   +4 != FF、不是自己的每一支艦隊的每個分隊，任一分隊 +0x34／+0x36 在 ±16 內 → 0`）、docs/internals/53 第 47 行，與本檔一致；
 *   補充：±16 是開區間 (x − 16, x ＋ 16) 在 16 位元回繞下的無號比較，x < 16 時 x − 16 回繞成大值，幾乎不可能命中；艦隊序號 i 的上限是 0x20 筆。
 *   ds:C074 的艦隊表 122 bytes 一筆、艦艇池 78 bytes 一筆見 docs/internals/33、38。
 * 副作用：無。
 * 輸入域：x、y 全部 16 位元值；F 指到艦隊表或任意位址（只讀 F[+6]）；地圖與池取自記憶體，基底 fcmd-standby.state（04EA 需要戰場地圖）。
 *   池基底 ds:CFB4..CFB7 列入 freeze（隨機突變會改到視訊區）。word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（dosgolem 的 read16），用 w0C74_rd16s。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_1588_04EA(Mem *m, uint16_t x, uint16_t y);

static uint16_t w0C74_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_297B_0C74(Mem *m, uint16_t f_off, uint16_t f_seg, uint16_t x, uint16_t y)
{
    uint16_t r;
    uint16_t g = 0xC074u;
    uint8_t i;

    if ((int16_t)x < 0 || (int16_t)x >= 0x7EF || (int16_t)y < 0 || (int16_t)y >= 0x7E7)
        return 0;
    r = f_1588_04EA(m, x, y);
    for (i = 0; i < 0x20u; i = (uint8_t)(i + 1u), g = (uint16_t)(g + 0x7Au)) {
        uint16_t ids;
        uint8_t j;

        if (r == 0)
            break;
        if (rd8(m, DS((uint16_t)(g + 4u))) == 0xFFu)
            continue;
        if (rd8(m, DS((uint16_t)(g + 6u))) != rd8(m, LIN(f_seg, (uint16_t)(f_off + 6u))))
            continue;
        if (GIN3_DS == f_seg && g == f_off)
            continue;
        ids = (uint16_t)(g + 0x10u);
        for (j = 0; j < rd8(m, DS((uint16_t)(g + 0x0Fu))); j = (uint8_t)(j + 1u), ids = (uint16_t)(ids + 4u)) {
            uint16_t idx = w0C74_rd16s(m, GIN3_DS, ids);
            uint16_t ship = (uint16_t)(w0C74_rd16s(m, GIN3_DS, 0xCFB4u) + (uint16_t)(idx * 0x4Eu));
            uint16_t seg = w0C74_rd16s(m, GIN3_DS, 0xCFB6u);
            uint16_t sx = w0C74_rd16s(m, seg, (uint16_t)(ship + 0x34u));
            uint16_t sy;

            if ((uint16_t)(x - 16u) >= sx)
                continue;
            if ((uint16_t)(x + 16u) <= sx)
                continue;
            sy = w0C74_rd16s(m, seg, (uint16_t)(ship + 0x36u));
            if ((uint16_t)(y - 16u) >= sy)
                continue;
            sy = w0C74_rd16s(m, seg, (uint16_t)(ship + 0x36u));
            if ((uint16_t)(y + 16u) <= sy)
                continue;
            r = 0;
        }
    }
    return r;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_297B_0C74(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("297B_0C74", 4, run)
