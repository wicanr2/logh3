/*
 * GIN3PS.EXE 297B:116C  艦隊維持費（同時寫回艦隊 +0x0A 的維持費與 +0x0C 的移動經費）
 *
 * 原版：far 呼叫，一個參數＝艦隊記錄的 far 指標（args[0]＝位移、args[1]＝段），回 AX:BX（32 位元）：
 *   對 i ＝ 0..艦隊[+0x0F]-1（艦隊[+0x0F] 是 byte，每輪重讀，無號比較 jb；現有組數）：
 *     艦艇索引 ＝ word(艦隊 + 0x10 + 4 × i)（編制槽，每槽 4 bytes，只用前兩個 byte）
 *     艦艇 ＝ 艦艇池 (ds:CFB4 位移、ds:CFB6 段) + 索引 × 0x4E（16 位元乘法與加法）
 *     列 ＝ ds:4E3A + 艦艇[+0] × 0xD0 + 艦艇[+4] × 0x1A（16 位元；艦種表，每陣營 0xD0、每艦種 0x1A）
 *     組艘數 ＝ word(艦艇 + 8)
 *     A += 列[+4] × 組艘數；B += 列[+5] × 組艘數（byte 零延伸；32 位元乘法與累加，溢位回繞）
 *   A ← A ÷ 1000；B ← B ÷ 1000（無號 32 位元，3CC9:0020）
 *   f ＝ 0xC8 - 艦隊[+0x47]（16 位元相減，再符號延伸成 32 位元：艦隊[+0x47] > 200 時是負數）
 *   艦隊[+0x0A]（word）← 低 16 位元( (A × f) ÷ 100 )   ；乘法 32 位元回繞，除法無號 32 位元
 *   艦隊[+0x0C]（word）← 低 16 位元( (B × f) ÷ 100 )
 *   回傳 ← 艦隊[+0x0A] 的 word 符號延伸成 32 位元（AX＝0 或 0xFFFF、BX＝word）
 * 語意見 docs/re/211、docs/re/239、docs/internals/27／33、docs/re/393（維持費 ＝ Σ(列[+4] × 艘數) ÷ 1000 × (200 − 艦隊[+0x47]) ÷ 100，
 * 已證實）。docs 沒有記的細節（反組譯可見）：(200 - 艦隊[+0x47]) 以 16 位元相減再符號延伸；兩個 ÷ 100 是 32 位元無號除法
 * 對 32 位元回繞後的乘積；回傳值是低 16 位元的符號延伸而不是完整商（只有低 word 被存下與回傳）。
 * 艦隊記錄 ds:C074 起 32 筆 × 122 bytes、艦艇池 78 bytes 一筆（docs/internals/33、64）。
 * 呼叫：3CC9:0009（乘）、3CC9:0020（除）；以 gin3_lmul／gin3_uldiv 表達（兩者已 diff-tested）。
 * 副作用：寫艦隊 +0x0A、+0x0C 兩個 word。輸入域：任意 far 指標與記憶體內容（除數是常數，不會 #DE）；
 *   例外：艦艇池遠指標（ds:CFB4／CFB6）加上索引位移後落在 0xA0000..0xAFFFF（VGA 視訊記憶體區）時，原版與基底傾印不一致，
 *   本版與差異測試都不涵蓋。機制是強推論：與 docs/re/445「限制」（平面視訊記憶體讀取不經 WatchReads）一致，未直接觀察探針讀到什麼（框架限制）。
 *   實測：另一個 RNG 種子的批次出現 4 到 5 案池段被突變到 0x998B 等值而比對失敗，診斷批次證實池段在 0xA000 時失敗、0x8000／0x9000 時通過；
 *   收據批次（種子 20261011）的向量檔已稽核，沒有任何一案改動池遠指標（ds:CFB4..CFB7）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

uint32_t f_297B_116C(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint32_t upkeep = 0;
    uint32_t move = 0;
    uint16_t slot = (uint16_t)(fleet_off + 0x10);
    uint8_t i = 0;
    uint32_t factor;
    uint32_t value;

    while (i < rd8(m, LIN(fleet_seg, fleet_off + 0x0F))) {
        uint16_t index = rd16(m, LIN(fleet_seg, slot));
        uint16_t pool_seg = rd16(m, DS(0xCFB6));
        uint16_t ship = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(index * 0x4Eu));
        uint16_t row = (uint16_t)(0x4E3Au + (uint16_t)(rd8(m, LIN(pool_seg, ship)) * 0xD0u) +
                                  (uint16_t)(rd8(m, LIN(pool_seg, ship + 4)) * 0x1Au));
        uint16_t count = rd16(m, LIN(pool_seg, ship + 8));

        upkeep += gin3_lmul(rd8(m, DS((uint16_t)(row + 4))), count);
        move += gin3_lmul(rd8(m, DS((uint16_t)(row + 5))), count);
        i++;
        slot = (uint16_t)(slot + 4);
    }
    upkeep = gin3_uldiv(upkeep, 1000);
    move = gin3_uldiv(move, 1000);
    factor = (uint32_t)(int32_t)(int16_t)(uint16_t)(0xC8u - rd8(m, LIN(fleet_seg, fleet_off + 0x47)));
    value = gin3_uldiv(gin3_lmul(upkeep, factor), 100);
    wr16(m, LIN(fleet_seg, fleet_off + 0x0A), (uint16_t)value);
    value = gin3_uldiv(gin3_lmul(move, factor), 100);
    wr16(m, LIN(fleet_seg, fleet_off + 0x0C), (uint16_t)value);
    return (uint32_t)(int32_t)(int16_t)rd16(m, LIN(fleet_seg, fleet_off + 0x0A));
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_297B_116C(m, a[0], a[1]); }
GIN3_REGISTER("297B_116C", 2, run)
