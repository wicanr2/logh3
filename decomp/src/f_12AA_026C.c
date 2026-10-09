/*
 * GIN3PS.EXE 12AA:026C  依星域重建戰場艦隊與艦艇池索引清單，再為兩個座位設鏡頭，最後把 ds:9B7A 設成 0xFFFE
 *
 * 原版（26 bytes，far 呼叫，一個堆疊字 star；序言 push bp、mov bp,sp，所以 star 在 [bp+6]），回 AX：
 *   push star ; call 12AA:03BC ; mov sp,bp          ; 清單重建（03BC 只用 star 的低位元組）
 *   call 1588:0930                                  ; 座位 0、座位 1 的鏡頭（兩次 1588:0945）
 *   mov word ds:9B7A, 0xFFFE
 *   retf                                            ; AX 是 0930 的回傳（＝ 座位 1 那次 0945 的 AX）；最後的 mov 不動 AX
 * 過程（已證實：反組譯逐指令；12AA:03BC、1588:0930 各由其差異測試證實）：
 *   12AA:03BC(star) → 1588:0930() → word ds:9B7A ← 0xFFFE，回 0930 的回傳。
 *   03BC 寫的 ds:9894／9898／9896／98BA 清單，0930 內的 0945 會讀（byte 筆數 ds:9894、byte 清單 ds:9898），所以順序不可對調。
 * 語意（強推論）：呼叫者是 IDA 交叉參照 seg028:017E（即 12AA:017E 的 call；台帳 callers 欄為 1，但 ledger.sh show 的「被呼叫」欄是空的，兩者不一致）；docs/re/339 第 314 行寫 `ds:9B7A ← 0xFFF4` 的用途 U 與 12AA:026C 寫 0xFFFE 並列，
 *   ds:9B7A 的語意未知。
 * 副作用：全部來自 03BC（ds:9894、9896、9898 起、98BA 起）、0930（ds:8F3A 起 16 bytes、借用 0x9D000 淨影響為零）與
 *   ds:9B7A 一個 word。
 * 輸入域：star 任意 word（03BC 取低位元組）；基底 fcmd-standby.state（0930 的鏡頭座標需要戰場狀態），艦隊表由探針突變與 seeds 造。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；ds:9B7A 與清單的用途為強推論。
 */
#include "gin3_reg.h"

extern void f_12AA_03BC(Mem *m, uint16_t side_word);
extern uint16_t f_1588_0930(Mem *m);

uint16_t f_12AA_026C(Mem *m, uint16_t star_word)
{
    uint16_t ax;

    f_12AA_03BC(m, star_word);
    ax = f_1588_0930(m);
    wr16(m, DS(0x9B7Au), 0xFFFEu);
    return ax;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_12AA_026C(m, a[0]); }
GIN3_REGISTER("12AA_026C", 1, run)
