/*
 * GIN3PS.EXE 1F6F:0C11  戰場像素 (x, y) 這一點「被擋」：地形不通，或落在清單內某艘艦艇的 16 × 16 範圍裡（回 1）；旗標低 3 位非 0 回 0
 *
 * 原版（161 bytes，far 呼叫，六個堆疊字；序言 push bp／si／di、sub sp,6，所以第一個參數在 [bp+10h]），回 AX：
 *   x ＝ [bp+10h]（放進 DI）  y ＝ [bp+12h]  flags ＝ [bp+14h]（只測低位元組）
 *   list ＝ far 指標（位移 [bp+16h]、段 [bp+18h]；函式裡把位移每次加 2）  n ＝ [bp+1Ah]（word）
 * 過程（已證實：反組譯逐指令；1588:0571 由其差異測試證實）：
 *   (flags 的低位元組 & 7) ≠ 0 → 回 0                         ; test byte [bp+14h],7
 *   1588:0571(x, y) == 0（地形擋住）→ 回 1
 *   i ← 0（word）；i < n（無號 jnb）時：
 *     idx ← word list[i]（段 list_seg、位移 list_off ＋ 2i，16 位元回繞）
 *     艦艇 ← (word ds:CFB4 ＋ low16(idx × 0x4E)) : word ds:CFB6               ; 池基底每圈重讀
 *     sx ← word 艦艇[+0x34]；sx > x（無號）→ 下一個
 *     sy ← word 艦艇[+0x36]；sy > y（無號）→ 下一個
 *     low16(sx ＋ 0x10) <= x（無號 jbe）→ 下一個
 *     low16(sy ＋ 0x10) <= y（無號 jbe）→ 下一個
 *     否則回 1
 *   回 0
 *   所以艦艇的範圍是半開區間 [sx, sx ＋ 16) × [sy, sy ＋ 16)，全部是無號 16 位元比較（sx ＋ 0x10 超過 0xFFFF 回繞）。
 * 語意（強推論）：呼叫者 1F6F:091C（台帳，兩處）；艦艇 +0x34／+0x36 是戰場座標（同 17A0:06D7 的距離判定）。清單的來源與 flags 的語意
 *   未知（markdown 文件沒有 1F6F:0C11 的記載，只有 docs/internals/data/weapon-candidates.json 第 75842、81846 行列出它與呼叫點）。
 * 副作用：無（0571 不寫記憶體）。
 * 輸入域：x、y、flags、n 任意 16 位元值；清單與池取自記憶體，基底 fcmd-standby.state（0571 需要戰場地圖）。n 大到超過探針步數上限的案例
 *   由探針丟棄。word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（dosgolem 的 read16），用本檔的 w0C11_rd16s。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為未知。
 */
#include "gin3_reg.h"

extern uint16_t f_1588_0571(Mem *m, uint16_t x, uint16_t y);

/* 段內 word 讀取：位移各自 16 位元回繞（seg:FFFF 的高位元組在同一段的位移 0），與 dosgolem 一致 */
static uint16_t w0C11_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_1F6F_0C11(Mem *m, uint16_t x, uint16_t y, uint16_t flags, uint16_t list_off, uint16_t list_seg, uint16_t n)
{
    uint16_t i;

    if (((uint8_t)flags & 7u) != 0)
        return 0;
    if (f_1588_0571(m, x, y) == 0)
        return 1;
    for (i = 0; i < n; i++) {
        uint16_t idx = w0C11_rd16s(m, list_seg, list_off);
        uint16_t ship = (uint16_t)(w0C11_rd16s(m, GIN3_DS, 0xCFB4u) + (uint16_t)(idx * 0x4Eu));
        uint16_t seg = w0C11_rd16s(m, GIN3_DS, 0xCFB6u);
        uint16_t sx, sy;

        list_off = (uint16_t)(list_off + 2u);
        sx = w0C11_rd16s(m, seg, (uint16_t)(ship + 0x34u));
        if (sx > x)
            continue;
        sy = w0C11_rd16s(m, seg, (uint16_t)(ship + 0x36u));
        if (sy > y)
            continue;
        if ((uint16_t)(sx + 0x10u) <= x)
            continue;
        if ((uint16_t)(sy + 0x10u) <= y)
            continue;
        return 1;
    }
    return 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1F6F_0C11(m, a[0], a[1], a[2], a[3], a[4], a[5]); }
GIN3_REGISTER("1F6F_0C11", 6, run)
