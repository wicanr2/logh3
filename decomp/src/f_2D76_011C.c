/*
 * GIN3PS.EXE 2D76:011C  數各類別的現役艦隊：ds:D506 起四個 word 清 0 後，每支現役、bit0 未設的艦隊在其類別的 word 加一，回總數
 *
 * 原版：far 呼叫，沒有參數（push bp／si、sub sp,28h；區域變數 [bp+0] ＝ 迴圈 i、[bp+1] ＝ 清單項數 n、[bp+2] ＝ 計數、[bp+4..] ＝ 清單）。
 *   回傳 AL ＝ 計數（byte）；AH 是殘值（最後一次 `mul` 的 AH，或項數為 0 時 131C 的殘值），描述檔 ret ＝ pack:al。
 * 過程（已證實：反組譯逐指令；3C9C:000A、297B:131C 由其各自的差異測試證實）：
 *   3C9C:000A(ds:D506, 0, 8)                                   ; ds:D506..D50D 共 8 bytes 清 0
 *   n ← AL of 297B:131C(清單 ＝ SS:bp+4, v ＝ 0xFF（不過濾星域）, flags ＝ 1（只留本方 f[+1] == ds:8F1D 者）)
 *   count ← 0
 *   對 i ＝ 0..n−1（byte，jnb 無號）：idx ＝ 清單[i]；f ＝ ds:(0xC074 + idx × 0x7A)
 *     f[+0] bit0 有設 → 跳過
 *     否則 t ＝ f[+9]；k ＝ byte ds:(0x5F52 + t)；word ds:(0xD506 + 2k) 加一（16 位元讀改寫）；count 加一
 *   回 count
 * 語意（強推論）：docs/internals/54 第 503 行「2D76:011C 數各類別現有幾支放進 ds:D506 的四個 word」、第 560 行、docs/re/202 第 14 行
 *   「現有艦隊數，順便填 ds:D506」、docs/spec/029 第 327 行；f[+9] ＝ 艦隊類別碼，ds:5F52 ＝ 類別碼到 0..3 的對照表（基底狀態前 16 項為
 *   00 00 00 00 00 02 00 00 00 00 01 01 02 02 03 03，其餘為別的資料）。呼叫者 2D76:0060。
 *   注意 131C 的 flags ＝ 1 是只留本方（陣營 ＝ ds:8F1D），所以計的是本方的艦隊。
 * 副作用：ds:D506 起 8 bytes 清 0 並累加；k 超過 3 時（表項被改）word 寫到 ds:D50E 以後，照原版寫。
 *
 * 機器模型：C 版沒有堆疊。清單緩衝區放在原版的實際位置 SS:1F3C（基底狀態（fcmd-s3-base，stack_hi 0x65E18）SP ＝ 1F68，SS 固定 0x63EB，diffprobe 無法設定；
 *   進入點 SP ＝ 1F68 − 返回位址 4 − push bp／si 4 − sub sp,28h ＝ 1F38 ＝ bp，清單在 bp+4），再從那裡讀回。這個位址在探針的堆疊帶
 *   （SS:1B68..1F67）內，寫入集合比對時雙方都排除，與原版把清單寫在堆疊一致。
 * 輸入域：基底狀態 fcmd-s3-base.state（艦隊表有 18 支現役艦隊）；其餘由探針突變讀取集。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_3C9C_000A(Mem *m, uint16_t off, uint16_t seg, uint16_t val, uint16_t count);
extern uint8_t f_297B_131C(Mem *m, uint16_t buf_off, uint16_t buf_seg, uint16_t v_word, uint16_t flags_word);

#define F011C_SS 0x63EBu   /* fcmd-s3-base.state 的 SS（diffprobe 無法設定） */
#define F011C_LIST 0x1F3Cu /* 進入點 SP 1F68 − 4 − 4 − 0x28 ＝ bp(1F38)，清單在 bp+4 */

uint8_t f_2D76_011C(Mem *m)
{
    uint8_t n, count = 0, i;

    f_3C9C_000A(m, 0xD506u, GIN3_DS, 0, 8);
    n = f_297B_131C(m, F011C_LIST, F011C_SS, 0xFFu, 1);
    for (i = 0; i < n; i++) {
        uint8_t idx = rd8(m, LIN(F011C_SS, (uint16_t)(F011C_LIST + i)));
        uint16_t f = (uint16_t)(0xC074u + (uint16_t)(idx * 0x7Au));

        if (rd8(m, DS(f)) & 1u)
            continue;
        {
            uint8_t t = rd8(m, DS((uint16_t)(f + 9u)));
            uint8_t k = rd8(m, DS((uint16_t)(0x5F52u + t)));
            uint16_t w = (uint16_t)(0xD506u + 2u * k);

            wr16(m, DS(w), (uint16_t)(rd16(m, DS(w)) + 1u));
        }
        count++;
    }
    return count;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_2D76_011C(m);
}
GIN3_REGISTER("2D76_011C", 0, run)
