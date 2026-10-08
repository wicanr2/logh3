/*
 * GIN3PS.EXE 2703:1789  叛亂判定：這一格這一回合會不會叛亂（回 1 或 0）
 *
 * 原版：far 呼叫，兩個參數：星域編號（word，只讀低位元組）、已失格數陣列的 far 指標（位移在前、段在後）。
 * 回 AX ＝ 1（叛亂）或 0。
 * 語意（已證實：反組譯逐指令；docs/re/292、docs/re/245、spec/041 一致）：
 *   星域 ＝ ds:A79C + 編號 × 126
 *   編號 == 0 或 == 31                        → 回 0
 *   擁有者 ＝ 星域[+0x12] == 2（無主）         → 回 0   （反組譯連續檢查兩次；第二次永遠不成立，是死碼）
 *   星域[+0x28 + 擁有者] != 0（該陣營有駐軍） → 回 0
 *   參謀 ＝ ds:8F4D[擁有者]；提督 ＝ ds:B77C + 參謀 × 28
 *   r1 ＝ rand() % 400；rand() 再抽 4 次丟掉；r2 ＝ rand() % 400；rand() 再抽 4 次丟掉（共 10 次）
 *   u ＝ 星域[+0x7D]（byte）
 *   t ＝ u < 50 ? u ÷ 3 : u < 100 ? u ÷ 2 : u < 150 ? u × 3 ÷ 4（先乘後右移 2）: u × 4 ÷ 3     （16 位元無號）
 *   thr ＝ t × 100 ÷ 提督[+0x0E]                  （無號 16 位元；除數 0 時原版 #DE）
 *   thr ＝ thr ÷ (已失格數[擁有者] + 1)           （已失格數是 byte，除數 1..256）
 *   回傳 ＝ (r1 + r2) < thr（16 位元無號比較）
 * 欄位名稱：「未安定度」「提督+0x0E 營運」「已失格數」「情報專任參謀 ds:8F4D」來自 docs/re/292、
 *   docs/internals/30（confirmed）；亂數是 3CD5:000C（已 diff-tested，會寫 ds:94DC 的種子）。
 * 副作用：過了前四道閘門後才抽亂數，種子 ds:94DC 被寫 10 次；提早回 0 的路徑不碰亂數。
 *   thr 的除數為 0（提督營運為 0）時原版 #DE，在該步之前已抽完十次亂數。
 * 輸入域：編號當 byte 使用；測試涵蓋 0..31 與 16 位元高位元組不為 0 的值。提督營運為 0 的輸入原版
 *   觸發 #DE（INT 0），差異測試會丟棄；C 版此時回 0 只為避免 SIGFPE，該路徑不在比較範圍內。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_3CD5_000C(Mem *m);

uint16_t f_2703_1789(Mem *m, uint16_t region_no, uint16_t counts_off, uint16_t counts_seg)
{
    uint8_t no = (uint8_t)region_no;
    uint16_t region = (uint16_t)(0xA79C + (uint16_t)(no * 0x7E));
    uint8_t owner;
    uint16_t officer, r1, r2, u, t, ops, thr, lost;
    int i;

    if (no == 0 || no == 0x1F)
        return 0;
    if (rd8(m, DS((uint16_t)(region + 0x12))) == 2)
        return 0;
    owner = rd8(m, DS((uint16_t)(region + 0x12)));
    if (owner == 2) /* 反組譯裡的第二次檢查（死碼） */
        return 0;
    if (rd8(m, DS((uint16_t)(region + 0x28 + owner))) != 0)
        return 0;

    officer = (uint16_t)(0xB77C + (uint16_t)(rd8(m, DS((uint16_t)(0x8F4D + owner))) * 0x1C));

    r1 = (uint16_t)(f_3CD5_000C(m) % 400);
    for (i = 0; i < 4; i++)
        f_3CD5_000C(m);
    r2 = (uint16_t)(f_3CD5_000C(m) % 400);
    for (i = 0; i < 4; i++)
        f_3CD5_000C(m);

    u = rd8(m, DS((uint16_t)(region + 0x7D)));
    if (u < 0x32)
        t = (uint16_t)(u / 3u);
    else if (u < 0x64)
        t = (uint16_t)(u >> 1);
    else if (u < 0x96)
        t = (uint16_t)((u * 3u) >> 2);
    else
        t = (uint16_t)((uint16_t)(u << 2) / 3u);

    ops = rd8(m, DS((uint16_t)(officer + 0x0E)));
    if (ops == 0)
        return 0;
    thr = (uint16_t)((uint16_t)(t * 100u) / ops);
    lost = rd8(m, LIN(counts_seg, (uint16_t)(counts_off + owner)));
    thr = (uint16_t)(thr / (uint16_t)(lost + 1u));
    return (uint16_t)((uint16_t)(r1 + r2) < thr);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2703_1789(m, a[0], a[1], a[2]); }
GIN3_REGISTER("2703_1789", 3, run)
