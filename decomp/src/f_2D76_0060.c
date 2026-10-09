/*
 * GIN3PS.EXE 2D76:0060  戰略 AI「能不能建艦隊」的閘門：回 0xFFFF 表示不能建，否則回本方現有艦隊數（0 到 255）
 *
 * 原版（188 bytes，far 呼叫，沒有參數；序言 push bp／si／di、sub sp,8，區域變數 [bp+0] ＝ 艦隊數 n、[bp+2]／[bp+4] ＝ V 的低字／高字、[bp+6] ＝ 陣營），
 *   回 AX。side ＝ byte ds:8F1D（每次都重讀 `mov bl,ds:8F1Dh`；011C 不寫它，所以重讀與快取等價，本版照原版重讀）。
 * 過程（已證實：反組譯逐指令；2D76:011C 由其差異測試證實；3CC9:0009／0070／00B9 的語意由各自的差異測試證實）：
 *   byte ds:(0x8F49 ＋ side) >= 0x10（無號 jb 反向）→ 回 0xFFFF
 *   n ← AL of 2D76:011C()                                          ; 數本方各類別現役艦隊並填 ds:D506，回總數
 *   n < 3（無號）→ 回 n（AH ＝ 0）
 *   V ← dword ds:(0x8F6A ＋ 4 × side)                               ; 低字在低位址
 *   byte ds:(0x8FD4 ＋ side) >= 3（無號）→ V ← (V × 3 的低 32 位元) ÷ 4（3CC9:0070 有號除法，向零取整）
 *   W ← dword ds:(0x8F72 ＋ 4 × side)
 *   (W × 20 的低 32 位元) >= V（3CC9:00B9 有號 32 位元比較後接 jge）→ 回 0xFFFF
 *   否則回 n
 * 語意（強推論）：docs/internals/54-strategic-ai-decision.md 第 526、558 行（`n = 2D76:0060() ; 閘門，0FFFFh 就結束`）、docs/spec/029 第 319 行
 *   「2D76:0060()：能不能建」、docs/re/206-ds8fd4-is-a-five-step-level.md 第 3 行（`ds:8FD4[陣營] >= 3` 就把預算打七五折）、docs/internals/55 第 42 行，
 *   與本檔一致。ds:8F49[陣營] 的 0x10 門檻、V（預算）與 W（維持費？）的欄位名稱為假說；本檔只照 asm。
 * 副作用：全部來自 2D76:011C（ds:D506 起 8 bytes）。
 *
 * 機器模型：被呼叫的 f_2D76_011C 把清單放在 SS:1F3C（基底 fcmd-s3-base.state，進入點 SP ＝ 1F68 的框架）。從本函式呼叫時原版的框架更深（清單在
 *   SS:1F26 起），位置不同，但兩處都在探針的堆疊帶（SS:1B68..1F67）內，比對時排除，C 版自寫自讀，不影響結果。
 * 輸入域：側別、門檻、V、W 取記憶體任意值（ds:8F1D 任意 byte；side 大時 ds 位移在 16 位元內）。基底 fcmd-s3-base.state。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位語意為假說。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

extern uint8_t f_2D76_011C(Mem *m);

uint16_t f_2D76_0060(Mem *m)
{
    uint8_t side = rd8(m, DS(0x8F1Du));
    uint8_t n;
    uint32_t v, w, q, r;

    if (rd8(m, DS((uint16_t)(0x8F49u + side))) >= 0x10u)
        return 0xFFFFu;
    n = f_2D76_011C(m);
    if (n < 3u)
        return n;
    side = rd8(m, DS(0x8F1Du));
    v = rd32(m, DS((uint16_t)(0x8F6Au + 4u * side)));
    if (rd8(m, DS((uint16_t)(0x8FD4u + side))) >= 3u) {
        gin3_ldivmod(gin3_lmul(v, 3u), 4u, &q, &r);
        v = q;
    }
    side = rd8(m, DS(0x8F1Du));
    w = rd32(m, DS((uint16_t)(0x8F72u + 4u * side)));
    if ((int32_t)gin3_lmul(w, 20u) >= (int32_t)v)
        return 0xFFFFu;
    return n;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_2D76_0060(m);
}
GIN3_REGISTER("2D76_0060", 0, run)
