/*
 * GIN3PS.EXE 042F:0142  依裝置遮罩 ds:934E 設定音源預設值與設定畫面的三個「可選」旗標
 *
 * 原版：far 呼叫，沒有參數。唯一讀的記憶體是 ds:934E（byte，裝置遮罩）：
 *   bit0 ＝ MPU-401 埠有效且初始化成功（有 MIDI）、bit1 ＝ 偵測到 OPL（有 FM）
 *   （docs/internals/17 第 217 到 219 行）。其餘位元不看。
 * 寫入（全是 byte，無條件寫）：
 *   ds:8F33 ← bit0 設 ? 0 : (bit1 設 ? 1 : 2)       音源預設值：0 ＝ MIDI、1 ＝ FM、2 ＝ 關閉
 *   ds:03DC ← bit0 設 ? 1 : 0                        設定畫面熱區清單 ds:03D8 第 0 筆的 +4（可選旗標）
 *   ds:03E8 ← bit1 設 ? 1 : 0                        第 1 筆的 +4
 *   ds:03F4 ← 1                                      第 2 筆的 +4
 *   ds:8F34 ← 0、ds:8F35 ← 0                         平時配樂與戰鬥配樂選項歸零
 *   後半（xor ax,ax；mov ds:8F34,al；mov ds:8F35,al）的位元組簽章是 31 C0 A2 34 8F A2 35 8F，
 *   docs/internals/17 第 184 行把這個簽章記為「開新遊戲時 8F34／8F35 都設 0」；
 *   該簽章其實是本函式的尾巴，同一個函式還寫 ds:8F33 與三個可選旗標，文件沒有指出是這支函式。
 * 熱區清單的位置：ds:03D8 起每筆 12 bytes（docs/internals/23 第 43 行：+4 ＝ 啟用旗標，0 ＝ 跳過），
 *   ds:03DC、03E8、03F4 正好是第 0、1、2 筆的 +4（docs/re/285 第 65 到 67 行：出貨的 MUSIC.INI 沒有 MIDI，
 *   ds:03DC 被寫成 0，那一格點不下去、框也不畫；03E8／03F4 是 1）。
 * 回傳：沒有回傳值。離開時 AX ＝ 0（xor ax,ax 在最後一個 test 之後）、CX ＝ 3DC 寫入的那個 0 或 1。
 *   唯一的呼叫者是 042F:0006 內的 call far，不使用。差異測試仍比較 AX:CX（pack:ax,cx），忠實表達殘值。
 *
 * 輸入域：ds:934E 任意 byte（256 種，其中只有 bit0／bit1 有作用，其餘位元用來驗證被忽略）。
 *   第二次讀 ds:934E 是原版的做法（mov al,ds:934Eh 重讀一次），中間只有對 ds:8F33 的寫入，不會別名。
 * 推論等級：寫入位址、值、條件已證實（反組譯逐指令＋差異測試）；ds:03DC／03E8／03F4 是熱區清單 +4
 *   的對應是強推論（位移與 docs/internals/23、docs/re/285 相符）；音源 0／1／2 的名稱來自 docs/internals/17。
 */
#include "gin3_reg.h"

uint32_t f_042F_0142(Mem *m)
{
    uint8_t mask = rd8(m, DS(0x934E));
    uint8_t al;
    uint16_t cx, ax;

    if (mask & 1u)
        wr8(m, DS(0x8F33), 0);
    else if (mask & 2u)
        wr8(m, DS(0x8F33), 1);
    else
        wr8(m, DS(0x8F33), 2);

    al = rd8(m, DS(0x934E));
    cx = (al & 1u) ? 1u : 0u;
    wr8(m, DS(0x03DC), (uint8_t)cx);
    ax = (al & 2u) ? 1u : 0u;
    wr8(m, DS(0x03E8), (uint8_t)ax);
    wr8(m, DS(0x03F4), 1);
    ax = 0; /* xor ax,ax */
    wr8(m, DS(0x8F34), (uint8_t)ax);
    wr8(m, DS(0x8F35), (uint8_t)ax);
    return ((uint32_t)ax << 16) | cx; /* AX<<16 | CX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_042F_0142(m);
}
GIN3_REGISTER("042F_0142", 0, run)
