/*
 * GIN3PS.EXE 2E54:16F5  戰略 AI：一支艦隊所在的星域，我方戰力與敵方戰力相比，這支艦隊是不是「有仗可打／該留下」（回 1 或 0）
 *
 * 原版：far 呼叫，一個 far 指標參數 F（艦隊記錄，122 bytes 一筆）（序言 push bp／si／di、sub sp,58h，所以參數在 [bp+62h]／[bp+64h]），
 *   回傳 AX ＝ 1 或 0。區域變數：[bp+0／+2] ＝ 星域記錄 R 的位址、[bp+4] ＝ n2（byte）、[bp+5] ＝ n1（byte）、[bp+6..25h] ＝ 清單 B（32 bytes）、
 *   [bp+26h..45h] ＝ 清單 A（32 bytes）、[bp+46h..4Dh] ＝ 兩個 32 位元戰力 V2（[bp+48h]:[bp+46h]）與 V1（[bp+4Ch]:[bp+4Ah]）。
 * 過程（已證實：反組譯逐指令；297B:131C、2E54:1876、3982:0250 由其各自的差異測試證實）：
 *   r ← F[+6]（星域編號，byte）；R ← ds:(0xA79C + r × 0x7E)                                  ; 星域記錄 126 bytes 一筆
 *   F[+1] ＝ 0 且 r ＝ 0 → 回 1
 *   F[+1] ＝ 1 且 F[+6] ＝ 0x1F → 回 1
 *   n1 ← AL of 297B:131C(清單 A, word F[+6], 1)                  ; 該星域內、陣營 ＝ ds:8F1D 的現役艦隊
 *   n2 ← AL of 297B:131C(清單 B, word F[+6], 2)                  ; 該星域內、陣營 ≠ ds:8F1D 的現役艦隊
 *   V1 ← 2E54:1876(word F[+6], 清單 A, n1, word F[+1])          ; 清單 A 的戰力總和（32 位元，AX:BX ＝ 高:低）
 *   V2 ← 2E54:1876(word F[+6], 清單 B, n2, 3982:0250(word F[+1]))
 *   F[+1] ＝ R[+0x12]（同陣營擁有）：  回 (V1 × 4 ≥ V2) 或 (F[+0x3C] ＝ 0x64) ？ 1 : 0   ; V1 × 4 取 32 位元（回繞）
 *   否則 R[+0x12] ＝ 2（無主）：        回 (V1 × 2 ＞ V2) ？ 1 : 0                    ; 32 位元無號比較
 *   否則：                              回 (V1 ＞ V2) ？ 1 : 0
 * 呼叫的函式只讀不寫（131C 寫堆疊上的清單，1876 與 3982:0250 不寫記憶體）；本函式沒有堆疊以外的寫入。1876 的 region、faction 兩個參數它不讀，
 *   n 只取低位元組，所以 C 版給 n1／n2 即可（原版推入的 n 字高位元組是另一個區域變數的殘值）。
 * 語意（強推論）：docs/spec/029 第 93 行「有戰鬥時 2E54:16F5 回非 0 → +0x79 ＝ 3、ds:D5CE[星域]++」、第 423 行（無收據）、docs/internals/54 第 110 到 124 行
 *   「決定：2E54:16F5(艦隊)」、docs/re/196。F[+0x3C] ＝ 0x64 是「完整度 100」（強推論，1261:032B 寫的欄位）。
 *   「我方」是指 ds:8F1D 的陣營，不是 F 自己的陣營：清單 A 篩 ds:8F1D 相等者，清單 B 篩不等者。呼叫者 2E54:005D（hw-dependent）。
 *
 * 機器模型：C 版沒有堆疊。清單 B 放在原版的實際位置 SS:1F08、清單 A 放在 SS:1F28（基底狀態（fcmd-s3-base，stack_hi 0x65E18）SP ＝ 1F68，SS 固定 0x63EB，diffprobe 無法設定；
 *   進入點 SP ＝ 1F68 − 參數 4 − 返回位址 4 − push bp／si／di 6 − sub sp,58h ＝ 1F02 ＝ bp，清單 B 在 bp+6、清單 A 在 bp+0x26）。這個位址在探針的
 *   堆疊帶（SS:1B68..1F67）內，寫入集合比對時雙方都排除，與原版把清單寫在堆疊一致。
 * 輸入域：F 取艦隊表或 seeds 的位址；池指標 ds:CFB4..CFB7 由描述檔 freeze 排除隨機突變；基底狀態 fcmd-s3-base.state（艦隊表有 18 支現役艦隊）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint8_t f_297B_131C(Mem *m, uint16_t buf_off, uint16_t buf_seg, uint16_t v_word, uint16_t flags_word);
extern uint32_t f_2E54_1876(Mem *m, uint16_t region, uint16_t list_off, uint16_t list_seg, uint16_t count_w, uint16_t faction);
extern uint16_t f_3982_0250(uint8_t faction);

#define F16F5_SS 0x63EBu    /* fcmd-s3-base.state 的 SS（diffprobe 無法設定） */
#define F16F5_BP 0x1F02u    /* 進入點 SP 1F68 − 4 − 4 − 6 − 0x58 */
#define F16F5_LIST_B ((uint16_t)(F16F5_BP + 6u))
#define F16F5_LIST_A ((uint16_t)(F16F5_BP + 0x26u))

static uint16_t f16F5_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_2E54_16F5(Mem *m, uint16_t f_off, uint16_t f_seg)
{
    uint8_t r = rd8(m, LIN(f_seg, (uint16_t)(f_off + 6u)));
    uint16_t reg = (uint16_t)(0xA79Cu + (uint16_t)(r * 0x7Eu));
    uint8_t n1, n2;
    uint32_t v1, v2;
    uint8_t side, owner;

    if (rd8(m, LIN(f_seg, (uint16_t)(f_off + 1u))) == 0 && r == 0)
        return 1;
    if (rd8(m, LIN(f_seg, (uint16_t)(f_off + 1u))) == 1 && rd8(m, LIN(f_seg, (uint16_t)(f_off + 6u))) == 0x1Fu)
        return 1;

    n1 = f_297B_131C(m, F16F5_LIST_A, F16F5_SS, f16F5_rd16s(m, f_seg, (uint16_t)(f_off + 6u)), 1);
    n2 = f_297B_131C(m, F16F5_LIST_B, F16F5_SS, f16F5_rd16s(m, f_seg, (uint16_t)(f_off + 6u)), 2);
    v1 = f_2E54_1876(m, f16F5_rd16s(m, f_seg, (uint16_t)(f_off + 6u)), F16F5_LIST_A, F16F5_SS, n1,
                     f16F5_rd16s(m, f_seg, (uint16_t)(f_off + 1u)));
    v2 = f_2E54_1876(m, f16F5_rd16s(m, f_seg, (uint16_t)(f_off + 6u)), F16F5_LIST_B, F16F5_SS, n2,
                     f_3982_0250((uint8_t)f16F5_rd16s(m, f_seg, (uint16_t)(f_off + 1u))));

    side = rd8(m, LIN(f_seg, (uint16_t)(f_off + 1u)));
    owner = rd8(m, DS((uint16_t)(reg + 0x12u)));
    if (side == owner) {
        if ((uint32_t)(v1 << 2) >= v2 || rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x3Cu))) == 0x64u)
            return 1;
        return 0;
    }
    if (owner == 2)
        return (uint32_t)(v1 << 1) > v2 ? 1 : 0;
    return v1 > v2 ? 1 : 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2E54_16F5(m, a[0], a[1]); }
GIN3_REGISTER("2E54_16F5", 2, run)
