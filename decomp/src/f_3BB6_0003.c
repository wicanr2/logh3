/*
 * 3BB6:0003  重建 8 個聲道記錄與兩個全域值（near，無參數，無回傳值）
 *
 * 原版：DS ＝ seg130（dosgolem 下 0x4AFE，載入位置相依；呼叫者 3BB6:01FB、3BB6:0703 先 `mov ds,seg130`），
 *   描述檔用 ds_seg 4AFE，C 版把 ds 當參數。SI／DI／BX／CX／DX／AX 全被改寫，不比對（ret ＝ none）。
 *
 * 步驟（全部對過反組譯；偏移都是 DS 內位移）：
 *   1. S ← word 常數 0x13F2。ah ← byte [S+0x21]（0x1413）；ds:01A7 ← ah；
 *      ds:01A8 ← (ah × 256) ÷ 0x2D8 的商（無號 `div bx`，被除數 ＝ ah:00，最大 0xFF00，商最大 89，不會 #DE）；ds:01A9 ← 0xFF。
 *   2. 複製：對 k ＝ 0..7，ds:(0x11 ＋ 0x14 × k) ← byte [S ＋ 0x40 ＋ k]（來源 0x1432..0x1439，目的是聲道記錄 +0x0B）。
 *   3. si ← S ＋ 0x100 ＝ 0x14F2；掃描：重複 `al ← [si]；si ← si ＋ 4`，直到 al ＝ 0xFE。si 停在 0xFE 那一筆「之後」（＋4）。
 *   4. dl ← 1、ah ← 0。對聲道 k ＝ 0..7（記錄基底 R ＝ 0x06 ＋ 0x14 × k，緩衝區指標 bx ＝ 0xA6 ＋ 0x20 × k）：
 *        依序寫 [R+0] ← 0xFF、[R+3] ← 0xFF、[R+1] ← 1、[R+2] ← 0、word [R+0x0E] ← 0x8000、[R+5] ← 0x7F、[R+7] ← 0x80、
 *        word [R+0x12] ← si（軌道起點）、word [R+0x10] ← bx；
 *        al ← [ds:0BF1]；若 (al and dl) ＝ 0，或 byte [si] ＝ 0xFE（空軌道，第一筆就是結尾）：ah ← ah or dl、[R+0] ← 0；
 *        之後再掃描一次（同第 3 步，從目前 si 起）把 si 推到下一條軌道的起點；dl ← dl << 1。
 *   5. ds:0BEE ← not ah（ah ＝ 被停用／空軌道的聲道遮罩，bit k ＝ 聲道 k）。
 *
 * 語意（假說）：seg130 是音樂驅動的資料段（基底狀態：ds:13F2 起是 "M1E_BOLEOR" 字串後接歌曲標頭，+0x40 起是
 *   `01 02 03 04 05 06 FF FF`，ds:14F2 起是 4 bytes 一筆、以 0xFE 結尾的軌道資料）。本函式把 8 條軌道的起點填進 8 個記錄的
 *   +0x12，[R+0] ＝ 0xFF／0 可能是聲道啟用旗標，ds:0BF1 的各 bit 可能是呼叫者要求的聲道遮罩。ds:01A7／01A8 是歌曲標頭
 *   +0x21 的原值與其 ÷ 2.84 的整數商（速度？）。以上皆假說；欄位名稱與用途：未知。
 *
 * 推論等級：寫入的位址、順序、常數與條件已證實（逐指令對過）；欄位語意假說。
 *
 * 輸入域與本版不涵蓋的情形：
 *   - 掃描（第 3、4 步）依賴 0xFE 結尾終止。位移是 16 位元，掃描在 64 KB 的 DS 內繞圈（與 0x14F2 同餘 2 mod 4 的 16384 個位置）；
 *     整圈都沒有 0xFE 時原版不會終止（探針超過步數上限而丟棄）。本版的 f0003_scan 在 16385 次後停止，只為了不讓突變版當機，
 *     此情形不在輸入域內。
 *   - 掃描繞過 0xFFFF 後會讀到 ds:0002、0006、…，也就是本函式自己正在寫的聲道記錄；原版逐指令交錯「寫記錄、再掃描」，
 *     本版照原順序，描述檔的別名種子（掃描途中讀到剛寫入的位元組）驗證這點。
 *   - DS 固定 seg130（線性 0x4AFE0..0x5AFDF），不會讀寫 A0000 以上，也沒有 word 存取落在位移 0xFFFF。
 */
#include "gin3_reg.h"

/* 4 bytes 一筆的掃描：al ← [ds:si]；si ← si ＋ 4（16 位元回繞），直到 al ＝ 0xFE；回傳最後的 si */
static uint16_t f0003_scan(const Mem *m, uint16_t ds, uint16_t si)
{
    unsigned n;

    for (n = 0; n < 0x4001u; n++) {
        uint8_t al = rd8(m, LIN(ds, si));

        si = (uint16_t)(si + 4u);
        if (al == 0xFEu)
            break;
    }
    return si;
}

void f_3BB6_0003(Mem *m, uint16_t ds)
{
    uint16_t si = 0x13F2u;
    uint16_t di, bx;
    uint8_t ah, al, dl, k;

    ah = rd8(m, LIN(ds, (uint16_t)(si + 0x21u)));
    wr8(m, LIN(ds, 0x01A7u), ah);
    al = (uint8_t)(((uint32_t)ah << 8) / 0x2D8u);
    wr8(m, LIN(ds, 0x01A8u), al);
    wr8(m, LIN(ds, 0x01A9u), 0xFFu);

    di = 6u;
    bx = 0x40u;
    for (k = 0; k < 8u; k++) { /* 複製 8 個初值 */
        al = rd8(m, LIN(ds, (uint16_t)(bx + si)));
        wr8(m, LIN(ds, (uint16_t)(di + 0x0Bu)), al);
        bx = (uint16_t)(bx + 1u);
        di = (uint16_t)(di + 0x14u); /* 複製迴圈 */
    }

    si = (uint16_t)(si + 0x100u);
    si = f0003_scan(m, ds, si);

    dl = 1u;
    ah = 0u;
    di = 6u;
    bx = 0xA6u;
    for (k = 0; k < 8u; k++) { /* 8 個聲道 */
        wr8(m, LIN(ds, di), 0xFFu);
        wr8(m, LIN(ds, (uint16_t)(di + 3u)), 0xFFu);
        wr8(m, LIN(ds, (uint16_t)(di + 1u)), 1u);
        wr8(m, LIN(ds, (uint16_t)(di + 2u)), 0u);
        wr16(m, LIN(ds, (uint16_t)(di + 0x0Eu)), 0x8000u);
        wr8(m, LIN(ds, (uint16_t)(di + 5u)), 0x7Fu);
        wr8(m, LIN(ds, (uint16_t)(di + 7u)), 0x80u);
        wr16(m, LIN(ds, (uint16_t)(di + 0x12u)), si);
        wr16(m, LIN(ds, (uint16_t)(di + 0x10u)), bx);
        al = rd8(m, LIN(ds, 0x0BF1u));
        if ((al & dl) == 0u || rd8(m, LIN(ds, si)) == 0xFEu) {
            ah = (uint8_t)(ah | dl);
            wr8(m, LIN(ds, di), 0u);
        }
        si = f0003_scan(m, ds, si);
        dl = (uint8_t)(dl << 1);
        di = (uint16_t)(di + 0x14u); /* 聲道迴圈 */
        bx = (uint16_t)(bx + 0x20u);
    }
    ah = (uint8_t)~ah;
    wr8(m, LIN(ds, 0x0BEEu), ah);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_3BB6_0003(m, 0x4AFEu); /* 描述檔 ds_seg：呼叫者 3BB6:01FB、0703 載入的 seg130 */
    return 0;
}
GIN3_REGISTER("3BB6_0003", 0, run)
