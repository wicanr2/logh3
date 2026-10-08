/*
 * GIN3PS.EXE 297B:083F  艦隊機動力（百分率）
 *
 * 原版：far 呼叫，兩個堆疊參數：
 *   args[0..1] ＝ 艦隊記錄的 far 指標（位移、段）
 *   args[2]    ＝ 編制組數 word，當索引查 ds:4E2E 的 byte 表（docs 記為表[組數]）
 *              兩個呼叫者（0FF4:04F4、1DD6:0421）都先把一個 byte 零延伸再推入，實際值 0..255。
 * 回傳：真正的結果在 AL，AH 是 idiv 商的高位元組（沒清掉就 retf）。兩個呼叫者推完都 xor ah,ah。
 *   差異測試比較整個 AX，所以這裡忠實回傳 (商 & 0xFF00) | AL。
 * 語意（已證實：反組譯逐指令；docs/re/317、docs/internals/56 的公式一致）：
 *   提督 1 ＝ ds:B77C + 艦隊[+4] × 28；提督 2（副官）＝ 艦隊[+5] == 0xFF ? 無 : ds:B77C + 艦隊[+5] × 28
 *   v   ＝ 提督 1 的 +0x17，有副官且 提督 1 的 +0x17 < 副官的 +0x17（byte 無號比較）時取副官的（即取較大者）
 *   t   ＝ ds:4E2E[args[2]]（byte，位移 16 位元回繞）
 *   p   ＝ ((270 - 2v) 取 16 位元) × t，只留低 16 位元，再當有號 word（cwd 符號延伸）
 *   q   ＝ p ÷ 100（有號，商向零取整）
 *   AL  ＝ (100 - q) 的低位元組
 * 注意：「(270-2v)×t 只留低 16 位元再有號除」在 v、t 夠大時與數學上的結果不同；docs 的公式寫成整數運算，
 *   沒記這個截斷。資料裡 v、t 是否走到這個範圍未驗。
 * 艦隊 ds:C074 32 × 122；提督 ds:B77C 82 × 28（docs/internals/33、08）。
 * 副作用：無。輸入域：除數是常數 100，不會 #DE。args[2] 的呼叫端實值是零延伸的 byte（0..255）；
 *   描述檔實際取樣 0..13、31、32、255，其餘值只是讀 ds:4E2E 之後的鄰近記憶體，C 版用同一份記憶體模型，未逐一驗證。
 */
#include "gin3_reg.h"

uint16_t f_297B_083F(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t slots)
{
    uint8_t adm1_index = rd8(m, LIN(fleet_seg, fleet_off + 4));
    uint8_t adm2_index = rd8(m, LIN(fleet_seg, fleet_off + 5));
    uint16_t adm1 = (uint16_t)(0xB77Cu + (uint16_t)(adm1_index * 0x1Cu));
    uint16_t adm2 = (uint16_t)(0xB77Cu + (uint16_t)(adm2_index * 0x1Cu));
    uint8_t ability = rd8(m, DS((uint16_t)(adm1 + 0x17)));
    uint32_t product;
    uint16_t scaled;
    int16_t quotient;
    uint8_t low;

    if (adm2_index != 0xFF) {
        uint8_t other = rd8(m, DS((uint16_t)(adm2 + 0x17)));

        if (ability < other)
            ability = other;
    }
    product = (uint32_t)(uint16_t)(270u - 2u * ability) * rd8(m, DS((uint16_t)(0x4E2Eu + slots)));
    scaled = (uint16_t)product;
    quotient = (int16_t)((int16_t)scaled / 100);
    low = (uint8_t)(100u - (uint8_t)quotient);
    return (uint16_t)(((uint16_t)quotient & 0xFF00u) | low);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_297B_083F(m, a[0], a[1], a[2]); }
GIN3_REGISTER("297B_083F", 3, run)
