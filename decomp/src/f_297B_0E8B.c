/*
 * GIN3PS.EXE 297B:0E8B  艦隊的各分隊依艦種表與艦隊能力折算三項速度，再取各項最小值寫進艦隊 +0x4D..+0x50，並設定索敵分遣隊的三項速度
 *
 * 原版：far 呼叫，一個 far 指標參數＝艦隊記錄 F（122 bytes 一筆）（序言 push bp／si／di、sub sp,24h、mov bp,sp，所以參數在 [bp+2Eh]／[bp+30h]）。
 *   回傳 AX ＝ 1261:032B(F) 的 AX（殘值，呼叫者 297B:0546 不取用；描述檔 ret ＝ ax，差異測試照比）。
 *   區域變數：[bp+0] ＝ i（byte）、[bp+1..4] ＝ 四個最小值 min1..min4（byte，初值 0xFF）、[bp+6] ＝ 折扣 q（word），其餘是指標暫存。
 * 過程（已證實：反組譯逐指令；1261:032B 由其自己的差異測試證實）。每一步的讀寫順序與反組譯一致（F、池記錄、艦種表列別名時，前面的寫入
 *   影響後面的重讀；C 版不快取任何欄位）：
 *   對 i ＝ 0..F[+0x0F]−1（byte，jb 無號；每圈重讀 F[+0x0F]），組 ＝ F + 0x10 + 4i：
 *     P ← (ds:CFB6):(ds:CFB4 + low16(word 組[0] × 0x4E))                  ; 該組的池記錄（78 bytes 一筆）
 *     R ← ds:(0x4E3A + P[+0] × 0xD0 + P[+4] × 0x1A)                       ; 艦種表的列（P[+0] ＝ 陣營、P[+4] ＝ 艦種；16 位元算式）
 *     P[+0x1F] ← low8( R[+8] × F[+0x43] ÷ 100 )                             ; 無號 16 位元乘（≤ 65025）、除
 *     P[+0x21] ← low8( R[+6] × F[+0x42] ÷ 80 )
 *     P[+0x22] ← low8( R[+7] × F[+0x42] ÷ 60 )                              ; F[+0x42] 讀兩次
 *     q ← ( low16((0x10E − 2 × F[+0x46]) × ds:(0x4E2E + F[+0x0F]) ) ) ÷ 100  ; 16 位元相減、無號 16 位元相乘取低字，再當有號數 cwd／idiv（向零取整）
 *                                                                              ; F[+0x0F] 在此重讀一次
 *     P[+0x1F] −= low8( low16( P[+0x1F] × q ) ÷ 100 )                         ; q 當無號 16 位元；乘積取低字、除法無號
 *     P[+0x21] −= low8( low16( P[+0x21] × q ) ÷ 100 )
 *     P[+0x22] −= low8( low16( P[+0x22] × q ) ÷ 100 )
 *     min1 ≥ P[+0x1F] → min1 ← P[+0x1F]；min2 ≥ P[+0x21] → min2 ← P[+0x21]；min3 ≥ P[+0x22]；min4 ≥ P[+0x23]（byte 無號 jb 的反向）
 *   F[+0x3C] ← 0x64；F[+0x4D] ← min1；F[+0x4E] ← min2；F[+0x4F] ← min3；F[+0x50] ← min4
 *   索敵分遣隊：P2 ← (ds:CFB6):(ds:CFB4 + low16(word F[+0x38] × 0x4E))；R2 ← ds:(0x4E3A + P2[+0] × 0xD0 + 0xB6)       ; 艦種 7 的列
 *     P2[+0x1F] ← R2[+8]；P2[+0x21] ← R2[+6]；P2[+0x22] ← R2[+7]
 *   AX ← 1261:032B(F)
 * 語意（強推論）：docs/internals/58 第 58 行「十二格能力 艦隊 +0x3C–+0x47；三個速度 ＝ 各組 297B:0E8B(艦種, 能力) 的最小值，依組數照 ds:4E2E 打折」、
 *   docs/re/355 第 42 行「297B:0E8B 對每組查艦種表，基礎回轉／巡航／戰鬥速度分別是…」、docs/re/230 第 50 行、docs/re/235 第 14 行。
 *   P[+0x1F／0x21／0x22] 是三項速度（回轉、巡航、戰鬥為強推論），F[+0x42／0x43] 是艦隊能力折扣，ds:4E2E 是組數折扣表（基底前幾項 00 00 07 0E 15 1C…），
 *   F[+0x46] 與 0x10E 的組合為折扣基準（名稱未驗，位移與運算才是證據）。呼叫者 297B:0546（hw-dependent）。
 * 副作用：每組池記錄 P[+0x1F／0x21／0x22]；F[+0x3C]、F[+0x4D..0x50]；分遣隊池記錄 P2[+0x1F／0x21／0x22]；1261:032B 的寫入（F[+0]、F[+0x3C]）。
 *
 * 輸入域：F 取艦隊表或 seeds 的位址（位移 16 位元回繞，word 在位移 0xFFFF 時高位元組取同段位移 0）；池基底 ds:CFB4／CFB6 由描述檔 freeze 排除隨機
 *   突變（池段 668B，位移 16 位元回繞時線性位址最高 0x768AF，不落入 ROM 或視訊記憶體）；組數 0..255、組內池編號任意 word。
 *   基底狀態 fcmd-s3-base.state。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_1261_032B(Mem *m, uint16_t fleet_off, uint16_t fleet_seg);

static uint16_t w0E8B_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static uint16_t w0E8B_pool_off(const Mem *m, uint16_t w)
{
    return (uint16_t)(rd16(m, DS(0xCFB4u)) + (uint16_t)(w * 0x4Eu));
}

#define W0E8B_F(o) LIN(f_seg, (uint16_t)(f_off + (o)))

uint16_t f_297B_0E8B(Mem *m, uint16_t f_off, uint16_t f_seg)
{
    uint8_t i = 0, min1 = 0xFF, min2 = 0xFF, min3 = 0xFF, min4 = 0xFF;
    uint16_t grp = (uint16_t)(f_off + 0x10u);
    uint16_t seg;

    while (i < rd8(m, W0E8B_F(0x0Fu))) {
        uint16_t poff = w0E8B_pool_off(m, w0E8B_rd16s(m, f_seg, grp));
        uint16_t row, q16, t;
        uint8_t kind, side, v;

        seg = rd16(m, DS(0xCFB6u));
        kind = rd8(m, LIN(seg, (uint16_t)(poff + 4u)));
        side = rd8(m, LIN(seg, poff));
        row = (uint16_t)(0x4E3Au + (uint16_t)(side * 0xD0u) + (uint16_t)(kind * 0x1Au));

        t = (uint16_t)(rd8(m, DS((uint16_t)(row + 8u))) * rd8(m, W0E8B_F(0x43u)));
        wr8(m, LIN(seg, (uint16_t)(poff + 0x1Fu)), (uint8_t)(t / 100u));
        t = (uint16_t)(rd8(m, DS((uint16_t)(row + 6u))) * rd8(m, W0E8B_F(0x42u)));
        wr8(m, LIN(seg, (uint16_t)(poff + 0x21u)), (uint8_t)(t / 80u));
        t = (uint16_t)(rd8(m, DS((uint16_t)(row + 7u))) * rd8(m, W0E8B_F(0x42u)));
        wr8(m, LIN(seg, (uint16_t)(poff + 0x22u)), (uint8_t)(t / 60u));

        t = (uint16_t)((uint16_t)(0x10Eu - (uint16_t)(rd8(m, W0E8B_F(0x46u)) << 1))
                       * rd8(m, DS((uint16_t)(0x4E2Eu + rd8(m, W0E8B_F(0x0Fu))))));
        q16 = (uint16_t)((int16_t)t / 100); /* cwd；idiv：向零取整 */

        v = rd8(m, LIN(seg, (uint16_t)(poff + 0x1Fu)));
        wr8(m, LIN(seg, (uint16_t)(poff + 0x1Fu)),
            (uint8_t)(rd8(m, LIN(seg, (uint16_t)(poff + 0x1Fu))) - (uint8_t)((uint16_t)(v * q16) / 100u)));
        v = rd8(m, LIN(seg, (uint16_t)(poff + 0x21u)));
        wr8(m, LIN(seg, (uint16_t)(poff + 0x21u)),
            (uint8_t)(rd8(m, LIN(seg, (uint16_t)(poff + 0x21u))) - (uint8_t)((uint16_t)(v * q16) / 100u)));
        v = rd8(m, LIN(seg, (uint16_t)(poff + 0x22u)));
        wr8(m, LIN(seg, (uint16_t)(poff + 0x22u)),
            (uint8_t)(rd8(m, LIN(seg, (uint16_t)(poff + 0x22u))) - (uint8_t)((uint16_t)(v * q16) / 100u)));

        if (min1 >= rd8(m, LIN(seg, (uint16_t)(poff + 0x1Fu))))
            min1 = rd8(m, LIN(seg, (uint16_t)(poff + 0x1Fu)));
        if (min2 >= rd8(m, LIN(seg, (uint16_t)(poff + 0x21u))))
            min2 = rd8(m, LIN(seg, (uint16_t)(poff + 0x21u)));
        if (min3 >= rd8(m, LIN(seg, (uint16_t)(poff + 0x22u))))
            min3 = rd8(m, LIN(seg, (uint16_t)(poff + 0x22u)));
        if (min4 >= rd8(m, LIN(seg, (uint16_t)(poff + 0x23u))))
            min4 = rd8(m, LIN(seg, (uint16_t)(poff + 0x23u)));

        i++;
        grp = (uint16_t)(grp + 4u);
    }
    wr8(m, W0E8B_F(0x3Cu), 0x64u);
    wr8(m, W0E8B_F(0x4Du), min1);
    wr8(m, W0E8B_F(0x4Eu), min2);
    wr8(m, W0E8B_F(0x4Fu), min3);
    wr8(m, W0E8B_F(0x50u), min4);
    {
        uint16_t poff = w0E8B_pool_off(m, w0E8B_rd16s(m, f_seg, (uint16_t)(f_off + 0x38u)));
        uint16_t row;

        seg = rd16(m, DS(0xCFB6u));
        row = (uint16_t)(0x4E3Au + (uint16_t)(rd8(m, LIN(seg, poff)) * 0xD0u) + 0xB6u);
        wr8(m, LIN(seg, (uint16_t)(poff + 0x1Fu)), rd8(m, DS((uint16_t)(row + 8u))));
        wr8(m, LIN(seg, (uint16_t)(poff + 0x21u)), rd8(m, DS((uint16_t)(row + 6u))));
        wr8(m, LIN(seg, (uint16_t)(poff + 0x22u)), rd8(m, DS((uint16_t)(row + 7u))));
    }
    return f_1261_032B(m, f_off, f_seg);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_297B_0E8B(m, a[0], a[1]); }
GIN3_REGISTER("297B_0E8B", 2, run)
