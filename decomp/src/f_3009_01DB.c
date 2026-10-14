/*
 * GIN3PS.EXE 3009:01DB  對在場艦隊清單裡每支「+1 不等於 ds:8F1D」的艦隊呼叫 3009:0235（艦隊朝目的地走一步）
 *
 * 原版（90 bytes，far 呼叫，沒有堆疊參數；序言 push bp／si、sub sp,0Ah、mov bp,sp）。沒有有意義的回傳值（AX 是殘值，描述檔 ret ＝ none）。
 *   區域變數：[bp+0] ＝ i（byte）、[bp+2]／[bp+4] ＝ 清單指標（位移從 ds:9898 起、段 ＝ DS）、[bp+6]／[bp+8] ＝ 艦隊指標（寫了不再讀）。
 * 過程（已證實：反組譯逐指令；3009:0235 由其差異測試證實）：
 *   i ← 0；當 i < byte ds:9894（無號，每圈重讀）：
 *     f ← 0xC074 + byte ds:(9898 + i) × 0x7A                       ; 16 位元
 *     byte f[+1] ≠ byte ds:8F1D → 3009:0235(f, DS)                  ; 兩個堆疊字（位移、段），add sp,4
 *     i 加一、清單指標加一
 * 語意（強推論）：docs/internals/33-fleet-table.md 的艦隊表（ds:C074 起 122 bytes 一筆）；ds:9894／9898 是「在場艦隊清單」（20B4:0091 檔頭同一組欄位）；
 *   f[+1] 為陣營（20B4:0091 以 P[+0x11] == f[+1] 判同陣營），所以本函式 ＝ 對「陣營不是 ds:8F1D」的艦隊各走一步（用途為強推論，ds:8F1D 的語意未查）。
 * 副作用：全部來自 3009:0235（f[+0x74]、f[+0x76]、f[+0x78]）與其被呼叫者（3310:0006 的 ds:8EDC..8EED 暫存）；本函式自己不寫記憶體。
 *
 * 機器模型：C 版沒有堆疊。被呼叫者 C 版（含 39B5）放在進入點 SP ＝ 1F8A 的框架，原版從本函式呼叫時框架更深，兩處都在探針的堆疊帶內，比對時排除。
 * 輸入域：清單、艦隊表取自記憶體（清單的每個 byte 0..255，count 0..255）。被呼叫者可以改寫的位置（f + 0x74..0x78 與 8EDC..8EED）不含 ds:9894、
 *   ds:9898..9997、ds:8F1D，所以「每圈重讀 count」與「只讀一次」在輸入域內等價；本版仍照原版每圈重讀。
 *   步數：count 很大且每支艦隊都走 0／1 狀態時單案可能超出步數上限，探針會丟棄（描述檔 step_budget 已調高）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern void f_3009_0235(Mem *m, uint16_t f_off, uint16_t f_seg);

void f_3009_01DB(Mem *m)
{
    uint8_t i = 0;
    uint16_t list = 0x9898u;

    while (i < rd8(m, DS(0x9894u))) {
        uint16_t f = (uint16_t)(0xC074u + (uint16_t)(rd8(m, DS(list)) * 0x7Au));

        if (rd8(m, DS((uint16_t)(f + 1u))) != rd8(m, DS(0x8F1Du)))
            f_3009_0235(m, f, GIN3_DS);
        i = (uint8_t)(i + 1u);
        list = (uint16_t)(list + 1u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_3009_01DB(m);
    return 0;
}
GIN3_REGISTER("3009_01DB", 0, run)
