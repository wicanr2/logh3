/*
 * GIN3PS.EXE 20B4:0091  行星攻擊的範圍標記：在場艦隊清單裡每支對方陣營、艦艇池有編制的艦隊，依與行星 P 的 16.16 距離 d 設旗標 +0x70 的 bit0／bit1，並把行星索引存進 +0x5A（參數 side 在程式裡的名稱沿用寫檔當時的誤稱；docs/spec/019 第 54 行與 docs/internals/33 第 114 行：+0x5A ＝ 正在攻打哪一顆行星的索引，所以第一個參數是行星索引，強推論）
 *
 * 原版（304 bytes，far 呼叫，堆疊參數 side（word，只用低位元組）、行星遠指標 P；序言 push bp／si／di、sub sp,16h、mov bp,sp，所以 side 在 [bp+20h]、
 *   P 的位移在 [bp+22h]、段在 [bp+24h]）。沒有回傳值（AX 是殘值，描述檔 ret ＝ none）。
 *   區域變數：[bp+0] ＝ i（byte）、[bp+2]／[bp+4] ＝ 清單指標（位移從 ds:9898 起）、[bp+6]／[bp+8] ＝ 門檻 T（低字 0、高字 ＝ 2AF5:01F7 的回傳）、
 *   [bp+0Ah]／[bp+0Ch] ＝ 距離 d（低字／高字）、[bp+0Eh]／[bp+10h] ＝ 目前艦隊指標、[bp+12h]／[bp+14h] ＝ 艦艇指標。
 * 過程（已證實：反組譯逐指令；2AF5:01F7、3310:0006、3CC9:00B9 各由其差異測試證實）：
 *   T ← 2AF5:01F7(P) << 16                                     ; 16.16 的「查詢半徑」（強推論：行星防衛力夾限後除以 25）
 *   i ← 0；當 i < byte ds:9894（無號，每圈重讀）：
 *     f ← 0xC074 + byte ds:(9898 + i) × 0x7A                    ; 16 位元
 *     ship ← (word ds:CFB6):(word ds:CFB4 ＋ low16(word f[+0x10] × 0x4E))
 *     ship[+8]（word）== 0 → 下一個
 *     byte P[+0x11] == byte f[+1]（同陣營）→ 下一個
 *     d ← 3310:0006(P[+0x46], P[+0x48], f[+0x5C], f[+0x5E])      ; 16.16 距離（P 的座標在 +0x46／+0x48）
 *     d <= 150.0（3CC9:00B9 後 jg 不成立，有號 32 位元）：
 *       byte f[+0] 的 bit0 為 0 且 (byte f[+0x48] & 0x0F) == 1：f[+0x70] |= 1；f[+0x5A] ← side 低位元組
 *     d <= T（jg 不成立）：
 *       P[+0x12] == 0，或 P[+0x45] != 3：f[+0x70] |= 2；f[+0x5A] ← side 低位元組
 *     （第一段與第二段是兩次獨立的判斷，不是 if／else；第二段的 P 欄位在第一段寫入之後才讀。）
 *     i 加一、清單指標加一
 * 語意（強推論）：docs/spec/019-planet-assault.md 第 50 行「距離 ＝ 3310:0006(行星 +0x46/+0x48, 艦隊 +0x5C/+0x5E) ≤ 150.0（16.16）」、
 *   docs/internals/33-fleet-table.md 的艦隊 +0x70 位元旗標；150.0 與 T（行星防衛力決定的查詢半徑）是兩個圈，bit0 ＝ 內圈且艦隊狀態 1、bit1 ＝ 外圈；
 *   P[+0x12]／P[+0x45] 的語意未查（強推論只到「某種行星狀態」）。
 * 副作用：f[+0x70]、f[+0x5A]；3310:0006 內部的 ds:8EDC..8EED 暫存由被呼叫者的 C 版寫入。
 *
 * 輸入域：基底 fcmd-base.state（被呼叫的 39B5 C 版寫死進入點 SP ＝ 1F8A 的堆疊位址）；P 取行星表（ptr ds:5336／5338，82 bytes 一筆）或 seeds 給定的任意位置；
 *   清單、艦隊表、艦艇池取自記憶體。word 讀取在位移 0xFFFF 時 dosgolem 的高位元組取同一段的位移 0（rd16s）。
 *   2AF5:01F7 的 C 版用線性位址讀 P[+0x14] 的 word，P 的位移不落在 0xFFEB 一帶（行星表與 seeds 都滿足）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_2AF5_01F7(Mem *m, uint16_t planet_off, uint16_t planet_seg);
extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

static uint16_t f20B4_0091_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

#define F0091_ZF 0x0040u
#define F0091_SF 0x0080u
#define F0091_OF 0x0800u

/* jg：ZF ＝ 0 且 SF ＝ OF */
static int f20B4_0091_jg(uint16_t fl)
{
    return !(fl & F0091_ZF) && (((fl & F0091_SF) != 0) == ((fl & F0091_OF) != 0));
}

void f_20B4_0091(Mem *m, uint16_t side, uint16_t p_off, uint16_t p_seg)
{
    uint32_t t = (uint32_t)f_2AF5_01F7(m, p_off, p_seg) << 16;
    uint8_t i = 0;
    uint16_t list = 0x9898u;

    while (i < rd8(m, DS(0x9894u))) {
        uint16_t f = (uint16_t)(0xC074u + (uint16_t)(rd8(m, DS(list)) * 0x7Au));
        uint16_t ship = (uint16_t)(f20B4_0091_rd16s(m, GIN3_DS, 0xCFB4u)
                                   + (uint16_t)(f20B4_0091_rd16s(m, GIN3_DS, (uint16_t)(f + 0x10u)) * 0x4Eu));
        uint16_t ship_seg = f20B4_0091_rd16s(m, GIN3_DS, 0xCFB6u);

        if (f20B4_0091_rd16s(m, ship_seg, (uint16_t)(ship + 8u)) != 0
            && rd8(m, LIN(p_seg, (uint16_t)(p_off + 0x11u))) != rd8(m, DS((uint16_t)(f + 1u)))) {
            uint32_t d = f_3310_0006(m,
                                     f20B4_0091_rd16s(m, p_seg, (uint16_t)(p_off + 0x46u)),
                                     f20B4_0091_rd16s(m, p_seg, (uint16_t)(p_off + 0x48u)),
                                     f20B4_0091_rd16s(m, GIN3_DS, (uint16_t)(f + 0x5Cu)),
                                     f20B4_0091_rd16s(m, GIN3_DS, (uint16_t)(f + 0x5Eu)));

            if (!f20B4_0091_jg(f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, 0x0096u, 0x0000u))) {
                if ((rd8(m, DS(f)) & 1u) == 0 && (rd8(m, DS((uint16_t)(f + 0x48u))) & 0x0Fu) == 1) {
                    wr8(m, DS((uint16_t)(f + 0x70u)), (uint8_t)(rd8(m, DS((uint16_t)(f + 0x70u))) | 1u));
                    wr8(m, DS((uint16_t)(f + 0x5Au)), (uint8_t)side);
                }
            }
            if (!f20B4_0091_jg(f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, (uint16_t)(t >> 16), (uint16_t)t))) {
                if (rd8(m, LIN(p_seg, (uint16_t)(p_off + 0x12u))) == 0
                    || rd8(m, LIN(p_seg, (uint16_t)(p_off + 0x45u))) != 3) {
                    wr8(m, DS((uint16_t)(f + 0x70u)), (uint8_t)(rd8(m, DS((uint16_t)(f + 0x70u))) | 2u));
                    wr8(m, DS((uint16_t)(f + 0x5Au)), (uint8_t)side);
                }
            }
        }
        i = (uint8_t)(i + 1u);
        list = (uint16_t)(list + 1u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_20B4_0091(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("20B4_0091", 3, run)
