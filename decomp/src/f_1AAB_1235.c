/*
 * GIN3PS.EXE 1AAB:1235  列出（可選星域內的）全部現役艦隊，逐支呼叫 1AAB:1286（修護艦重設並補缺額）
 *
 * 原版（81 bytes，far 呼叫，一個堆疊字 v；push bp／si、sub sp,22h，所以 v 在 [bp+2Ah]）：
 *   區域變數（堆疊）：[bp+0] ＝ i（byte）、[bp+1] ＝ n（byte）、[bp+2 .. bp+0x21] ＝ 艦隊槽號清單（32 bytes）。
 *   n ← 297B:131C(清單 ＝ SS:bp+2, v, 旗標 0)        ; push 0（旗標 word ＝ 0：不過濾陣營）、push v、push ss、push bp+2；已 diff-tested
 *   對 i ＝ 0 .. n−1（byte、無號比較 jnb）：
 *     id ← 清單[i]；艦隊 ← ds:(0xC074 ＋ id × 0x7A)
 *     call 1AAB:1286(艦隊位移, ds)                    ; 已 diff-tested
 *   無回傳值（描述檔 ret ＝ none）。
 * 參數：args[0] ＝ v（131C 只取低位元組：0xFF 不過濾，其餘只留所在星域 ＝ v 的現役艦隊）。
 * 呼叫者（台帳）：1AAB:0FAF（hw-dependent）。docs/spec/063-region-resupply.md 第 34 行「先把停在這一格的所有陣營艦隊裡的修護艦
 *   （艦種 6），飛彈那一格重設成現有艘數：1AAB:1235 → 1286」、docs/re/346-region-resupply.md 第 40、75 行
 *   「1AAB:1235(星域編號) ① 修護艦重置（不分陣營）；傳旗標 ＝ 0，所以掃這一格所有陣營的艦隊」，與本檔一致。
 * 堆疊：清單緩衝區在呼叫者的堆疊上，位置 ＝ 進入點 SP − 0x2A（參數 2、far 回位址 4、bp 2、si 2、區域變數 0x22 後 bp，再 + 2）。
 *   描述檔的基底 fcmd-s3-base.state 進入點 SP 以下 0x400 bytes 是堆疊帶，比對時排除；本版把清單寫在同一個線性位置，
 *   131C 寫、本函式讀回。SS 0x63EB、進入點 SP 0x1F68 由向量檔 stack_hi ＝ 0x65E18 反推（與 2376:01E6 同一個 SS）。
 * 輸入域：v 全部 16 位元值；艦隊表、艦艇池取 S3 基底；清單不蓋到艦隊表（在堆疊帶內）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint8_t f_297B_131C(Mem *m, uint16_t buf_off, uint16_t buf_seg, uint16_t v_word, uint16_t flags_word);
extern void f_1AAB_1286(Mem *m, uint16_t fleet_off, uint16_t fleet_seg);

/* fcmd-s3-base.state 的 SS 與進入點 SP（SP ＝ stack_hi − SS × 16） */
#define W1235_SS 0x63EBu
#define W1235_SP 0x1F68u

void f_1AAB_1235(Mem *m, uint16_t v)
{
    uint16_t buf = (uint16_t)(W1235_SP - 0x2Au); /* SS:bp+2 */
    uint8_t n = f_297B_131C(m, buf, W1235_SS, v, 0u);
    uint8_t i;

    for (i = 0; i < n; i++) {
        uint8_t id = rd8(m, LIN(W1235_SS, (uint16_t)(buf + i)));

        f_1AAB_1286(m, (uint16_t)(0xC074u + (uint16_t)(id * 0x7Au)), GIN3_DS);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1AAB_1235(m, a[0]);
    return 0;
}
GIN3_REGISTER("1AAB_1235", 1, run)
