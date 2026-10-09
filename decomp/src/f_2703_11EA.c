/*
 * GIN3PS.EXE 2703:11EA  星域內本方「狀態 5」艦隊的加總除以 10000：(Σ f[+0x0F] 的低位元組) × 最大 adm[+0x0B] × 最大 adm[+0x13] ÷ 10000 的低 16 位元
 *
 * 原版：far 呼叫，一個 far 指標參數 R（星域記錄，ds:A79C 起每筆 126 bytes）（序言 push bp／si／di、sub sp,30h、mov bp,sp，
 *   所以第一個參數在 [bp+3Ah]），回傳 AX（見下）。區域變數：[bp+0] ＝ i、[bp+1] ＝ 清單項數 n、[bp+2] ＝ 計數（只寫不讀）、
 *   [bp+4..] ＝ 清單、[bp+25h] ＝ 加總（byte）、[bp+26h]、[bp+27h] ＝ 兩個最大值（byte）。
 * 過程（已證實：反組譯逐指令；297B:131C、3CC9:0009、3CC9:0020 由其各自的差異測試證實）：
 *   n ← AL of 297B:131C(清單 ＝ SS:bp+4, word R[+0x11], 0)        ; 星域 ＝ R[+0x11] 低位元組 的所有現役艦隊槽號
 *   sum、max1、max2 ← 0
 *   對 i ＝ 0..n−1（byte，jnb 無號）：idx ＝ 清單[i]；f ＝ ds:(0xC074 + idx × 0x7A)（艦隊記錄 122 bytes）
 *     f[+1] == R[+0x12]（byte 比較）且 f[+8] == 5 才處理（兩者缺一就跳過）：
 *       sum ← sum + f[+0x0F]（byte 相加，回繞）
 *       adm ← ds:(0xB77C + f[+4] × 0x1C)（提督記錄 28 bytes，16 位元乘法）
 *       adm[+0x0B] ≥ max1（無號，jb 反向）→ max1 ← adm[+0x0B]
 *       adm[+0x13] ≥ max2 → max2 ← adm[+0x13]
 *   AX ← 低 16 位元( (sum × max1 低 32 位元) × max2 低 32 位元 ÷ 10000 )   ; 3CC9:0009 兩次、3CC9:0020（無號）一次，三個運算元都 ≤ 255
 * 語意（強推論）：docs/re/24 第 129 行「+0x0E 再除以 100、夾在 200 以內，接著加上 2703:10FC 與 2703:11EA」、docs/re/43 第 48 行
 *   「v += f6(R) ; 2703:11EA」、第 128 行「f6 2703:11EA：艦隊（被加）」：R 是星域，叛亂（不安）公式的一項；f[+1] ＝ 陣營、R[+0x12] ＝ 星域擁有者。
 *   f[+8] ＝ 5 的意義、sum/max1/max2 的語意為假說。呼叫者 2703:0DA1。
 * 副作用：沒有堆疊以外的寫入（清單與區域變數在堆疊）。
 *
 * 機器模型：C 版沒有堆疊。清單緩衝區放在原版的實際位置 SS:1F2E（基底狀態（fcmd-s3-base，stack_hi 0x65E18）SP ＝ 1F68，SS 固定 0x63EB，diffprobe 無法設定；
 *   進入點 SP ＝ 1F68 − 參數 4 − 返回位址 4 − push bp／si／di 6 − sub sp,30h ＝ 1F2A ＝ bp，清單在 bp+4），再從那裡讀回。這個位址在
 *   探針的堆疊帶（SS:1B68..1F67）內，寫入集合比對時雙方都排除，與原版把清單寫在堆疊一致。
 * 輸入域：R 為星域表內的指標或 seeds 給的位址；word R[+0x11] 在位移 0xFFFF 時按段內回繞。基底狀態 fcmd-s3-base.state。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

extern uint8_t f_297B_131C(Mem *m, uint16_t buf_off, uint16_t buf_seg, uint16_t v_word, uint16_t flags_word);

#define F11EA_SS 0x63EBu   /* fcmd-s3-base.state 的 SS（diffprobe 無法設定） */
#define F11EA_LIST 0x1F2Eu /* 進入點 SP 1F68 − 4 − 4 − 6 − 0x30 ＝ bp(1F2A)，清單在 bp+4 */

static uint16_t f11EA_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_2703_11EA(Mem *m, uint16_t r_off, uint16_t r_seg)
{
    uint8_t n = f_297B_131C(m, F11EA_LIST, F11EA_SS, f11EA_rd16s(m, r_seg, (uint16_t)(r_off + 0x11u)), 0);
    uint8_t sum = 0, max1 = 0, max2 = 0, i;
    uint32_t prod, q, r;

    for (i = 0; i < n; i++) {
        uint8_t idx = rd8(m, LIN(F11EA_SS, (uint16_t)(F11EA_LIST + i)));
        uint16_t f = (uint16_t)(0xC074u + (uint16_t)(idx * 0x7Au));

        if (rd8(m, DS((uint16_t)(f + 1u))) != rd8(m, LIN(r_seg, (uint16_t)(r_off + 0x12u))))
            continue;
        if (rd8(m, DS((uint16_t)(f + 8u))) != 5u)
            continue;
        sum = (uint8_t)(sum + rd8(m, DS((uint16_t)(f + 0x0Fu))));
        {
            uint16_t adm = (uint16_t)(0xB77Cu + (uint16_t)(rd8(m, DS((uint16_t)(f + 4u))) * 0x1Cu));
            uint8_t a = rd8(m, DS((uint16_t)(adm + 0x0Bu)));

            if (!(a < max1))
                max1 = a;
            a = rd8(m, DS((uint16_t)(adm + 0x13u)));
            if (!(a < max2))
                max2 = a;
        }
    }
    prod = gin3_lmul(gin3_lmul(sum, max1), max2);
    gin3_uldivmod(prod, 10000u, &q, &r);
    return (uint16_t)q;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2703_11EA(m, a[0], a[1]); }
GIN3_REGISTER("2703_11EA", 2, run)
