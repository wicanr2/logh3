/*
 * GIN3PS.EXE 3310:014B  16.16 定點開平方：位移湊初值，再跑牛頓法 g ← (n ÷ g ＋ g) ÷ 2，等到不再變小就回上一次的值
 *
 * 原版（160 bytes，far 呼叫，一個 dword 參數＝兩個堆疊字；序言 push bp／si、sub sp,8，所以 n 的低字在 [bp+10h]、高字在 [bp+12h]），
 *   回傳 AX ＝ 高字、BX ＝ 低字（描述檔 ret ＝ pack:ax,bx）。區域變數：[bp+4]／[bp+6] ＝ G（低字／高字）、[bp+0]／[bp+2] ＝ T（低字／高字）。
 * 過程（已證實：反組譯逐指令；39B5:005C、3CC9:00B9 各由其差異測試證實）：
 *   n ＝ 0（高字 | 低字 ＝ 0）→ 回 0:0
 *   G ← 0x00010000（1.0）；T ← n
 *   迴圈一（對齊）：G >= T（32 位元有號，3CC9:00B9 後接 jge）→ 離開；否則 G ← G << 1（32 位元）、T ← T >> 1（算術右移 32 位元）重複
 *   迴圈二（牛頓）：
 *     T ← G                                                       ; 記下上一次的值
 *     Q ← 39B5:005C(n, G)                                          ; (n << 16) ÷ G 的低 32 位元（有號；G ＝ 0 回 0）
 *     G ← (Q ＋ G) >> 1                                            ; add／adc 後 sar／rcr：32 位元和（進位丟掉）算術右移一位
 *     G < T（有號 jl）→ 重複；否則離開
 *   回 T（也就是最後一次沒有再變小之前的 G）
 *   3CC9:00B9 只看旗標，且 AX、BX、CX、DX 原樣返回（原版 `mov ax,cx; mov bx,si` 取回的是 T），32 位元有號比較等同 (int32_t) 比較。
 * 語意（強推論）：docs/re/94-attack-coefficient-is-a-conic.md 第 44 行「3310:014B：16.16 定點開平方，位移湊初值，再跑牛頓法 x ← (x ＋ n÷x) ÷ 2，
 *   等到不再變小就回上一次的值」、docs/re/147-fortress-gun-damage.md 第 13 行「3310:014B(v)：16.16 開方」，與本檔一致；
 *   docs/re/124-fleet-move-remaining-pieces.md 第 34 行「3310:014B 的捨去還沒讀」：回傳是牛頓序列開始不下降前的那一項，
 *   不是最後算出的一項（那一項被丟掉）。
 * 副作用：全部來自 39B5:005C（ds:8EDC..8EED 共 18 bytes，R／D／Q 暫存的最終值），本函式自己不寫記憶體。
 *
 * 機器模型：被呼叫的 f_39B5_005C 把參數放在 SS:1F82..1F89（基底 fcmd-base.state，進入點 SP ＝ 1F8A 的框架）。從本函式呼叫時原版的框架更深，
 *   實際位置不同，但兩處都在探針的堆疊帶（SS:1B8A..1F89）內，寫入集合比對時排除，C 版自寫自讀，不影響結果。
 * 輸入域：n 任意 32 位元值。迴圈在 C 版有上界 4096 圈（超過回 0xFFFFFFFF）；原版每圈（含 005C 的 48 位元移位相減）至少數千步，
 *   在探針的 200,000 步上限內走不到 4096 圈，所以上界只影響突變版本，不影響收錄的案例。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_39B5_005C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi);

/* 32 位元算術右移一位（不依賴編譯器對負數 >> 的實作定義行為） */
static uint32_t w3310_sar1(uint32_t v) { return (v >> 1) | (v & 0x80000000u); }

uint32_t f_3310_014B(Mem *m, uint16_t lo, uint16_t hi)
{
    uint32_t g = 0x00010000u;
    uint32_t t = ((uint32_t)hi << 16) | lo;
    unsigned guard;

    if ((uint16_t)(hi | lo) == 0)
        return 0;
    for (guard = 0; guard < 4096u; guard++) {
        if ((int32_t)g >= (int32_t)t)
            break;
        g = g << 1;
        t = w3310_sar1(t);
    }
    if (guard == 4096u)
        return 0xFFFFFFFFu;
    for (guard = 0; guard < 4096u; guard++) {
        uint32_t q;

        t = g;
        q = f_39B5_005C(m, lo, hi, (uint16_t)g, (uint16_t)(g >> 16));
        g = w3310_sar1(q + g);
        if (!((int32_t)g < (int32_t)t))
            return t;
    }
    return 0xFFFFFFFFu;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_3310_014B(m, a[0], a[1]); }
GIN3_REGISTER("3310_014B", 2, run)
