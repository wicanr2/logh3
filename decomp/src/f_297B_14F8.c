/*
 * GIN3PS.EXE 297B:14F8  數一個星域裡的現役艦隊：陣營 0 的個數寫到 *out1、其餘陣營的個數寫到 *out2（跳過 +0x00 bit0 有設的艦隊）
 *
 * 原版：far 呼叫，三個 far 指標參數（序言 push bp／si、sub sp,26h、mov bp,sp，所以第一個參數在 [bp+2Eh]），沒有有意義的回傳值
 *   （AX 是殘值，描述檔 ret ＝ none）：
 *   args[0..1] ＝ 星域記錄（region，每筆 126 bytes，ds:A79C 起）  args[2..3] ＝ out1（byte 的位址）  args[4..5] ＝ out2（byte 的位址）
 * 過程（已證實：反組譯逐指令；297B:131C 由其自己的差異測試證實）：
 *   n ← AL of 297B:131C(清單緩衝區 ＝ SS:bp（區域變數 [bp+0..1Fh]）, word region[+0x11], 0)   ; 星域 ＝ region[+0x11] 低位元組的現役艦隊槽號清單
 *   *out2 ← 0；*out1 ← 0                                                              ; 先 out2 後 out1
 *   對 i ＝ 0..n−1（byte，jnb 無號比較）：idx ＝ 清單[i]；f ＝ ds:(0xC074 + idx × 0x7A)（艦隊記錄 122 bytes）
 *     f[+0] bit0 有設 → 跳過
 *     否則 f[+1] ＝ 0 → ++*out1（byte 讀改寫），f[+1] ≠ 0 → ++*out2
 *   每次寫 out 都重新從堆疊參數取指標；out1、out2 彼此別名、或指到艦隊記錄時，前面的寫入影響後面的讀取（逐步照做）。
 * 語意（強推論）：docs/internals/22 第 176 行「bit2 ＝ 這一格雙方都有艦隊：297B:14F8 分別數帝國與同盟在這一格的艦隊（跳過艦隊 +0x00 bit0 有設的）」；
 *   docs/internals/53 第 32 行、docs/spec/032 第 46 行、docs/re/245 第 32 行同；欄位 f[+1] ＝ 陣營（297B:142A／1491 用 ds:8F1D 比較它）。
 *   呼叫者（台帳）：056A:0469、056A:059E、20B4:1497、2703:0933。
 * 副作用：*out1、*out2 兩個 byte。線性 0xF0000 以上（ROM）的寫入在 dosgolem 被安靜丟棄，本檔按 byte 模型化（讀改寫的讀仍讀到 ROM 原內容）。
 *
 * 機器模型：C 版沒有堆疊。清單緩衝區放在原版的實際位置 SS:1F2A（基底狀態（fcmd-s3-base，stack_hi 0x65E18）SP ＝ 1F68，SS 固定 0x63EB，diffprobe 無法設定；
 *   進入點 SP ＝ 1F68 − 參數 12 − 返回位址 4 − push bp／si 4 − sub sp,26h ＝ 1F2A ＝ bp），再從那裡讀回。這個位址在探針的堆疊帶
 *   （SS:1B68..1F67）內，寫入集合比對時雙方都排除，與原版把清單寫在堆疊一致。
 * 輸入域：region 與 out 指標指到暫存區、DGROUP 或 seeds 給的位址，不指向堆疊帶；region[+0x11] 的 word 讀取在位移 0xFFFF 時按段內回繞。
 *   基底狀態 fcmd-s3-base.state（艦隊表有 18 支現役艦隊）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint8_t f_297B_131C(Mem *m, uint16_t buf_off, uint16_t buf_seg, uint16_t v_word, uint16_t flags_word);

#define F14F8_SS 0x63EBu /* fcmd-s3-base.state 的 SS（diffprobe 無法設定） */
#define F14F8_BP 0x1F2Au /* 進入點 SP 1F68 − 12 − 4 − 4 − 0x26 */

static uint16_t f14F8_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效 */
static void f14F8_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_297B_14F8(Mem *m, uint16_t reg_off, uint16_t reg_seg, uint16_t o1_off, uint16_t o1_seg, uint16_t o2_off,
                 uint16_t o2_seg)
{
    uint8_t n = f_297B_131C(m, F14F8_BP, F14F8_SS, f14F8_rd16s(m, reg_seg, (uint16_t)(reg_off + 0x11u)), 0);
    uint8_t i;

    f14F8_put8(m, LIN(o2_seg, o2_off), 0);
    f14F8_put8(m, LIN(o1_seg, o1_off), 0);
    for (i = 0; i < n; i++) {
        uint8_t idx = rd8(m, LIN(F14F8_SS, (uint16_t)(F14F8_BP + i)));
        uint16_t f = (uint16_t)(0xC074u + (uint16_t)(idx * 0x7Au));

        if (rd8(m, DS(f)) & 1u)
            continue;
        if (rd8(m, DS((uint16_t)(f + 1u))) == 0)
            f14F8_put8(m, LIN(o1_seg, o1_off), (uint8_t)(rd8(m, LIN(o1_seg, o1_off)) + 1u));
        else
            f14F8_put8(m, LIN(o2_seg, o2_off), (uint8_t)(rd8(m, LIN(o2_seg, o2_off)) + 1u));
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_297B_14F8(m, a[0], a[1], a[2], a[3], a[4], a[5]);
    return 0;
}
GIN3_REGISTER("297B_14F8", 6, run)
