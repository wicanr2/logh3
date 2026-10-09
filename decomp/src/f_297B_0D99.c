/*
 * GIN3PS.EXE 297B:0D99  艦隊 F 的一個分隊艦艇 S 的站位：從 S 的隊形槽目標 (S[+28], S[+2A]) 出發，站不下就朝艦隊中心 (F[+5C], F[+5E]) 每次挪 4 像素，
 *   距中心各軸 4 以內就直接放中心；最後把結果寫進 S 的現在位置 (S[+34], S[+36])
 *
 * 原版（242 bytes，far 呼叫，兩個 far 指標參數；序言 push bp、sub sp,0Ah、mov bp,sp，沒有 push si／di，所以 F 的位移在 [bp+10h]、段在 [bp+12h]，S 的位移在 [bp+14h]、段在 [bp+16h]），
 *   無有意義的回傳值（描述檔 ret ＝ none）。區域變數：X ＝ [bp+2]、Y ＝ [bp+4]（word，有號比較用）、[bp+0] ＝ 角度 byte（只存不讀）、
 *   [bp+6]／[bp+8] ＝ F ＋ 0x5C 的位移／段（迴圈裡用它重讀 F[+5C]）。
 * 過程（已證實：反組譯逐指令；297B:0C74、2BE1:011D、2BE1:026F、3C82:000F 各由其差異測試證實）：
 *   X ← S[+28]；Y ← S[+2A]
 *   重複：
 *     297B:0C74(F, X, Y) ≠ 0（這一點站得下）→ 離開
 *     X ＝ F[+5C] 且 Y ＝ F[+5E] → 離開                                    ; 已在中心
 *     |F[+5C] − X| > 4（有號比較；3C82:000F 是 abs，0x8000 取 abs 仍是 0x8000，有號為 −32768，不大於 4）或 |F[+5E] − Y| > 4 →
 *       a ← 2BE1:011D(F[+5C] − X, F[+5E] − Y)                              ; byte 角度（16 位元相減）
 *       2BE1:026F(X, Y, a, 4, far &X, far &Y)                              ; 沿角度前進 4，夾限後寫回 X、Y（堆疊上）
 *     否則（兩軸都在 4 以內）：X ← F[+5C]；Y ← F[+5E]
 *   S[+34] ← X；S[+36] ← Y
 * 語意（強推論）：docs/internals/53-strategic-planet-assault.md 第 50 行（站不下就朝艦隊中心每次挪 4 像素，距中心 4 以內直接放中心，297B:0D99）、
 *   docs/spec/058-strategic-planet-assault.md 第 24 行、docs/re/330-strategic-planet-assault.md 第 278 行（各分隊：297B:0D99(f, s)，位置 [+34/+36]）、
 *   docs/re/335-ai-order-command-kind.md 第 83 行（pos ← s[+28/+2A]，集結點，2BE1:0000），與本檔一致。
 * 副作用：S[+34]、S[+36]（word）；X、Y 在堆疊上；0C74 → 1588:04EA 與 026F 沒有全域寫入。
 *
 * 機器模型：C 版沒有堆疊。X、Y 放在原版的實際位置（基底狀態 fcmd-standby.state：SS 固定 0x63EB，進入點 SP ＝ 1E76，探針推 4 個參數字 8 bytes、返回位址 4、
 *   push bp 2、sub sp,0Ah，所以 bp ＝ 1E5E，X 在 SS:1E60、Y 在 SS:1E62），交給 f_2BE1_026F（用 far 指標寫兩個 word）就地更新，再從那裡讀回。
 *   這些位址在探針的堆疊帶（SS:1A76..1E75）內，寫入集合比對時雙方都排除，與原版把 X、Y 放在堆疊一致。
 * 輸入域：F 取艦隊表（ds:C074 起 0x7A 一筆）、S 取艦艇池，欄位全部 16 位元值；地圖與池取記憶體，基底 fcmd-standby.state（0C74 → 1588:04EA 需要戰場地圖）。
 *   迴圈次數取決於距離（每次 4 像素），超過探針步數上限的案例由探針丟棄。原版的迴圈沒有上限（X 被夾限卡在邊界而 0C74 一直回 0 時不終止）；C 版加 4096 圈的上界，
 *   原版每圈（含 0C74 與 04EA）至少數百步，在探針的步數上限內走不到 4096 圈，所以上界只影響突變版本，不影響收錄的案例。S 不在 ROM 區（寫入不模型化 ROM）。word 存取在位移 0xFFFF 時高位元組取同一段的位移 0
 *   （dosgolem 的 read16），用本檔的 w0D99_rd16s／w0D99_wr16s。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_297B_0C74(Mem *m, uint16_t f_off, uint16_t f_seg, uint16_t x, uint16_t y);
extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_026F(Mem *m, uint16_t x, uint16_t y, uint16_t angle_word, uint16_t dist, uint16_t ox_off, uint16_t ox_seg,
                            uint16_t oy_off, uint16_t oy_seg);
extern uint16_t f_3C82_000F(uint16_t x);

#define F0D99_SS 0x63EBu /* fcmd-standby.state 的 SS（diffprobe 無法設定） */
#define F0D99_X 0x1E60u  /* 進入點 SP 1E76 − 參數 8 − 返回位址 4 − push bp 2 − sub sp,0Ah ＝ bp(1E5E)；X 在 bp+2 */
#define F0D99_Y 0x1E62u  /* Y 在 bp+4 */

static uint16_t w0D99_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w0D99_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_297B_0D99(Mem *m, uint16_t f_off, uint16_t f_seg, uint16_t s_off, uint16_t s_seg)
{
    unsigned guard;

    wr16(m, LIN(F0D99_SS, F0D99_X), w0D99_rd16s(m, s_seg, (uint16_t)(s_off + 0x28u)));
    wr16(m, LIN(F0D99_SS, F0D99_Y), w0D99_rd16s(m, s_seg, (uint16_t)(s_off + 0x2Au)));
    for (guard = 0; guard < 4096u; guard++) {
        uint16_t x = rd16(m, LIN(F0D99_SS, F0D99_X));
        uint16_t y = rd16(m, LIN(F0D99_SS, F0D99_Y));
        uint16_t cx_, cy_;
        int far_away;

        if (f_297B_0C74(m, f_off, f_seg, x, y) != 0)
            break;
        cx_ = w0D99_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu));
        cy_ = w0D99_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu));
        if (x == cx_ && y == cy_)
            break;
        far_away = (int16_t)f_3C82_000F((uint16_t)(cx_ - x)) > 4;
        if (!far_away)
            far_away = (int16_t)f_3C82_000F((uint16_t)(cy_ - y)) > 4;
        if (far_away) {
            uint8_t a = f_2BE1_011D(m, (uint16_t)(cx_ - x), (uint16_t)(cy_ - y));

            (void)f_2BE1_026F(m, x, y, a, 4, F0D99_X, F0D99_SS, F0D99_Y, F0D99_SS);
        } else {
            wr16(m, LIN(F0D99_SS, F0D99_X), cx_);
            wr16(m, LIN(F0D99_SS, F0D99_Y), cy_);
        }
    }
    w0D99_wr16s(m, s_seg, (uint16_t)(s_off + 0x34u), rd16(m, LIN(F0D99_SS, F0D99_X)));
    w0D99_wr16s(m, s_seg, (uint16_t)(s_off + 0x36u), rd16(m, LIN(F0D99_SS, F0D99_Y)));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_297B_0D99(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("297B_0D99", 4, run)
