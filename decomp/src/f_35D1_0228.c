/*
 * GIN3PS.EXE 35D1:0228  從顯示物件清單 ds:EAE0 移除所有 obj[+2] 等於 key 低位元組的物件（逐一呼叫 35D1:0102）
 *
 * 原版（69 bytes，far 呼叫，一個 word 參數 key；序言 push bp／si、sub sp,4、mov bp,sp，所以 key 在 [bp+0Ch]），無有意義的回傳值（描述檔 ret ＝ none）。
 *   區域變數 [bp+0]／[bp+2] ＝ 目前項目的位移／段（只存不讀）；si ＝ 項目序號。
 * 過程（已證實：反組譯逐指令；35D1:0102 由其差異測試證實）：
 *   si ← 0
 *   重複：
 *     si >= word ds:EADE（無號 jnb，筆數每圈重讀）→ 結束
 *     entry ← ds:(0xEAE0 ＋ 4 × si)：位移 word 在 +0、段 word 在 +2（`si << 2` 是 16 位元運算）；段先讀、位移後讀（互不相干）
 *     byte entry 指到的物件[+2] ＝ key 的低位元組（cmp cl,[bp+0Ch]，byte 比較，key 的高位元組不看）→
 *       35D1:0102(entry)（push 段、push 位移）；si 不加（0102 把後面的項目前移，這一格換成下一筆，所以同一格要再看一次）
 *     否則 si ← si ＋ 1
 * 語意（強推論）：呼叫者 341B:00D9（台帳）；docs/internals/data/battle-object-keys.json 約第 7903、8222 行（35D1:0194／01DE 同樣以 obj[+2] 當 key 掃 ds:EAE0 清單）。
 *   obj[+2] 是物件的 key（byte），docs 沒有 35D1:0228 本身的記載；本函式只做「移除所有 key 相同的物件」。
 * 副作用：全部來自 35D1:0102（物件 +16..+27、+1E..+22，格表 ds:D81E 起，清單前移與 ds:EADE 減一）；本函式自己不寫記憶體。
 *
 * 機器模型：C 版沒有堆疊；原版的 [bp+0]／[bp+2] 只寫不讀。35D1:0102 內部的堆疊區域變數在探針的堆疊帶內（比對時排除）。
 *   清單的位移 `ds:(0xEAE0 + 4 × si)` 恆為偶數，word 存取不會落在位移 0xFFFF，直接用 rd16。
 * 輸入域：清單項目（ds:EAE0 起 16 筆）與筆數 ds:EADE 由描述檔 seeds 給定並列入 freeze：項目會被取指標讀 obj[+2]，再交給 35D1:0102 改寫物件，
 *   隨機突變會讓它指向程式碼、堆疊帶或視訊記憶體區，那些案例原版與 C 版不可比。物件本身（obj[+2] 與 35D1:0102 讀的欄位）與 key 取突變。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_35D1_0102(Mem *m, uint16_t obj_off, uint16_t obj_seg);

void f_35D1_0228(Mem *m, uint16_t key)
{
    uint16_t si = 0;

    while (si < rd16(m, DS(0xEADEu))) {
        uint16_t b = (uint16_t)(si << 2);
        uint16_t seg = rd16(m, DS((uint16_t)(b + 0xEAE2u)));
        uint16_t off = rd16(m, DS((uint16_t)(b + 0xEAE0u)));

        if (rd8(m, LIN(seg, (uint16_t)(off + 2u))) == (uint8_t)key)
            (void)f_35D1_0102(m, off, seg);
        else
            si = (uint16_t)(si + 1u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_35D1_0228(m, a[0]);
    return 0;
}
GIN3_REGISTER("35D1_0228", 1, run)
