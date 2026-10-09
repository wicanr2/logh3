/*
 * GIN3PS.EXE 1930:1572  某星域：列出該星域內「陣營 ＝ 星域擁有者」的現役艦隊，逐支以星域存量補給（1930:15F4）
 *
 * 原版（130 bytes，far 呼叫，一個堆疊字 region；push bp／si、sub sp,2Ah，所以 region 在 [bp+32h]）：
 *   region_no ← region 的低位元組（mov al,[bp+32h]；xor ah,ah）
 *   星域 ← ds:(0xA79C ＋ region_no × 0x7E)         ; [bp+22h]／[bp+24h] 存這個 far 指標（位移、ds）
 *   ds:8F1D ← 星域[+0x12]（byte）                   ; 原版把「目前陣營」設成星域擁有者，讓下面的過濾用
 *   n ← 297B:131C(清單 ＝ SS:bp+2, region（整個 word）, 旗標 1)   ; AL ＝ 1（AH 殘值 ＝ region_no × 0x7E 的高位元組，131C 只取低位元組）
 *       旗標 1 ＝ 只留艦隊[+1] ＝ ds:8F1D 的艦隊；131C 只取 region 的低位元組當所在星域條件（0xFF 不過濾）
 *   對 i ＝ 0 .. n−1（byte、無號比較 jnb）：
 *     id ← 清單[i]；艦隊 ← ds:(0xC074 ＋ id × 0x7A)
 *     call 1930:15F4(艦隊位移, ds, 星域位移, ds)    ; 已 diff-tested
 *   ds:8F1D ← 0xFF
 *   無回傳值（描述檔 ret ＝ none）。
 * 呼叫者（台帳）：0374:0556（hw-dependent）。docs/spec/063-region-resupply.md 第 20、32 行「1930:1572(星域) ← 飛彈／艦載機補給（規則 0）：
 *   整步的最前面先跑，逐組把星域的 +0x68／+0x70 補進艦艇的 +0x0E／+0x12」、docs/internals/62-strategic-attrition.md 第 10 行、
 *   docs/re/349-1930-supply-family.md，與本檔一致。補充：過濾的是「陣營 ＝ 星域擁有者」且「所在星域 ＝ 參數」的艦隊（docs/spec/063
 *   沒寫這個過濾，語意是強推論），ds:8F1D 在函式結束時被設成 0xFF（呼叫者之後是否使用該值未知）。
 * 副作用：ds:8F1D 先寫星域擁有者、最後寫 0xFF；其餘來自 131C（只寫堆疊上的清單）與 15F4（艦艇池與星域存量）。
 * 堆疊：清單緩衝區位置 ＝ 進入點 SP − 0x32（參數 2、far 回位址 4、bp 2、si 2、區域變數 0x2A，再 + 2）；進入點 SP 以下 0x400 bytes 是
 *   堆疊帶，比對時排除。SS 0x63EB、進入點 SP 0x1F68 由向量檔 stack_hi ＝ 0x65E18 反推。
 * 輸入域：region 全部 16 位元值（region_no >= 32 時星域記錄指到星域表之後的記憶體，照原版運算）；艦隊表、艦艇池取 S3 基底。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint8_t f_297B_131C(Mem *m, uint16_t buf_off, uint16_t buf_seg, uint16_t v_word, uint16_t flags_word);
extern void f_1930_15F4(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t reg_off, uint16_t reg_seg);

#define W1572_SS 0x63EBu
#define W1572_SP 0x1F68u

void f_1930_1572(Mem *m, uint16_t region_word)
{
    uint16_t buf = (uint16_t)(W1572_SP - 0x32u); /* SS:bp+2 */
    uint16_t region = (uint16_t)(0xA79Cu + (uint16_t)((uint16_t)(region_word & 0xFFu) * 0x7Eu));
    uint8_t n;
    uint8_t i;

    wr8(m, DS(0x8F1Du), rd8(m, DS((uint16_t)(region + 0x12u))));
    n = f_297B_131C(m, buf, W1572_SS, region_word, 1u);
    for (i = 0; i < n; i++) {
        uint8_t id = rd8(m, LIN(W1572_SS, (uint16_t)(buf + i)));

        f_1930_15F4(m, (uint16_t)(0xC074u + (uint16_t)(id * 0x7Au)), GIN3_DS, region, GIN3_DS);
    }
    wr8(m, DS(0x8F1Du), 0xFFu);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1930_1572(m, a[0]);
    return 0;
}
GIN3_REGISTER("1930_1572", 1, run)
