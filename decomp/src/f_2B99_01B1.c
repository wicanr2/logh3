/*
 * GIN3PS.EXE 2B99:01B1  Floyd–Warshall：把距離矩陣就地鬆弛成最短路，中點矩陣記下每次鬆弛的中繼點
 *
 * 原版：far 呼叫，四個堆疊字（push bp、si、di、sub sp,0Ah，所以參數從 [bp+14h] 起），無回傳值
 *   （AX 是殘留值，呼叫者 2B99:0180 不用；描述檔 ret=none）：
 *   args[0..1] ＝ 距離矩陣的 far 指標（位移、段）：32 × 32 個 word，列優先，每列 64 bytes
 *   args[2..3] ＝ 中點矩陣的 far 指標（位移、段）：32 × 32 個 byte，每列 32 bytes
 * 過程：
 *   1. memset(中點矩陣, 0xFF, 0x400)（呼叫 3C9C:000A，位移 16 位元回繞）
 *   2. 對角線：距離[i][i] ← 0，i = 0 到 31（位移 ＝ 指標 ＋ i × 66，16 位元運算）
 *   3. k、i、j 三層（都是 0 到 31，無號比較）：
 *        s ← 距離[i][k] ＋ 距離[k][j]    （16 位元回繞相加；每個 j 都重讀兩個加數，所以就地寫入會影響後面的讀取）
 *        s < 距離[i][j]（無號比較，jnb 跳過）→ 距離[i][j] ← s；中點[i × 32 + j] ← k（byte）
 *   位址的位移都用 16 位元運算（超過 0xFFFF 在同一段內回繞）；字的存取在本檔按線性位址連續，dosgolem 實際是兩次 byte 存取、位移各自 16 位元回繞（cpu.go read16／write16）；word 存取落在位移 0xFFFF 的輸入不在比較範圍（seeds 的格位移都是偶數，沒有 word 落在位移 0xFFFF）。
 *   寫入落在線性 0xF0000 以上（ROM 區）時被 dosgolem 安靜丟棄，本檔按 byte 模型化（含 memset；讀仍讀 ROM 內容）。
 *
 * 比較是無號的，所以 0x7FFF 當作無限大不會誤鬆弛（兩個 0x7FFF 相加是 0xFFFE，仍大於任何真值，docs/re/243 第 100 到 102 行）。
 * 欄位語意（強推論）：距離矩陣與中點矩陣的用途見 docs/internals/26、docs/re/243 第 76 到 102 行；位移與常數已證實。
 * 輸入域：兩個矩陣指標任意（描述檔用暫存緩衝區；seeds 另造回繞、ROM、重疊）；矩陣內容任意 16 位元值。
 *   每案約 150 萬步，描述檔 step_budget 調高。讀 A0000 到 BFFFF 的案例由探針丟棄。
 *   指標與本函式自己的程式碼或堆疊重疊的案例不在比較範圍。
 * 證據：asm 逐指令；docs/re/243 的虛擬碼（先 memset 0xFF、對角線歸零、三層迴圈、無號比較）與本檔一致。
 * 推論等級：已證實（反組譯逐指令＋差異測試；memset 的 3C9C:000A 由其自己的差異測試證實，本檔另加 ROM 區語意）。
 */
#include "gin3_reg.h"

static void w01b1_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w01b1_put16(Mem *m, uint32_t lin, uint16_t v)
{
    w01b1_put8(m, lin, (uint8_t)v);
    w01b1_put8(m, lin + 1, (uint8_t)(v >> 8));
}

void f_2B99_01B1(Mem *m, uint16_t d_off, uint16_t d_seg, uint16_t m_off, uint16_t m_seg)
{
    uint16_t i;
    uint16_t j;
    uint16_t k;

    for (i = 0; i < 0x400; i++)
        w01b1_put8(m, LIN(m_seg, m_off + i), 0xFF);
    for (i = 0; i < 0x20; i++)
        w01b1_put16(m, LIN(d_seg, d_off + (uint16_t)(i << 6) + (uint16_t)(i << 1)), 0);
    for (k = 0; k < 0x20; k++) {
        for (i = 0; i < 0x20; i++) {
            for (j = 0; j < 0x20; j++) {
                uint16_t dik = rd16(m, LIN(d_seg, d_off + (uint16_t)(i << 6) + (uint16_t)(k << 1)));
                uint16_t dkj = rd16(m, LIN(d_seg, d_off + (uint16_t)(k << 6) + (uint16_t)(j << 1)));
                uint16_t s = (uint16_t)(dik + dkj);
                uint32_t cell = LIN(d_seg, d_off + (uint16_t)(i << 6) + (uint16_t)(j << 1));

                if (s < rd16(m, cell)) {
                    w01b1_put16(m, cell, s);
                    w01b1_put8(m, LIN(m_seg, m_off + (uint16_t)(i << 5) + j), (uint8_t)k);
                }
            }
        }
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2B99_01B1(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("2B99_01B1", 4, run)
