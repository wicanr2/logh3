/*
 * GIN3PS.EXE 090A:0DC0  無主星域（擁有者 2）的稅率重算（寫各星域 +0x26，無回傳值）
 *
 * 原版：far 呼叫，一個參數 level（word，序言 push bp／si／di、sub sp,0Ch，所以在 [bp+16h]）。沒有回傳值。
 *
 * 反組譯逐條對應：
 *   對星域 i ＝ 0 到 31（word 有號比較，記錄 ds:A79C ＋ i × 0x7E，16 位元位移）：
 *     星域[+0x12] != 2（byte）→ 下一個
 *     base ← dword ds:[ (level << 2) ＋ 0x147E ]（低字）／ds:[ (level << 2) ＋ 0x1480 ]（高字），位移 16 位元回繞
 *     g ← word 星域[+0x1E]（有號）
 *     g <= 0：v ← base × 5                       （32 位元）
 *     g > 0 ：v ← sext32(g) × base               （32 位元，低 32 位元）
 *     q ← v ÷ 100（有號 32 位元除法，商向零取整，3CC9:0070）
 *     星域[+0x26] ← q 的低位元組
 *   沒有 200／1 的夾限（090A:0B93 有，這支沒有；docs/spec/038 第 63 行寫「還沒讀」，此為解答）。
 *
 * 推論等級
 *   已證實（差異測試）：上列結構、常數、有號性、位移回繞、寫入寬度（只寫低位元組）。
 *   強推論（docs/spec/038、docs/re/223、docs/internals/55）：星域 +0x12 是擁有者（2 ＝ 無主）、+0x1E 是成長率
 *     （有號）、+0x26 是稅率；ds:147E 起是 32 位元基價表（50 75 100 150 200），level 是難度等級
 *     （呼叫者 090A:0B5E 傳 2）。
 *   未知：成長率 <= 0 時基價 × 5 的設計意圖（相對於 g > 0 時的 g × base）。
 *
 * 輸入域：level 任意 word（表外是 DGROUP 的鄰近資料）；星域記錄與表內容任意。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

void f_090A_0DC0(Mem *m, uint16_t level)
{
    uint16_t reg = 0xA79C;
    int16_t i;

    for (i = 0; i < 0x20; i++) {
        if (rd8(m, DS(reg + 0x12)) == 2) {
            uint16_t at = (uint16_t)(level << 2);
            uint32_t base = ((uint32_t)rd16(m, DS((uint16_t)(at + 0x1480u))) << 16) | rd16(m, DS((uint16_t)(at + 0x147Eu)));
            int16_t g = (int16_t)rd16(m, DS(reg + 0x1E));
            uint32_t v, q, r;

            if (g <= 0)
                v = gin3_lmul(base, 5u);
            else
                v = gin3_lmul((uint32_t)(int32_t)g, base);
            gin3_ldivmod(v, 100u, &q, &r);
            wr8(m, DS(reg + 0x26), (uint8_t)q);
        }
        reg = (uint16_t)(reg + 0x7E);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_090A_0DC0(m, a[0]);
    return 0;
}
GIN3_REGISTER("090A_0DC0", 1, run)
