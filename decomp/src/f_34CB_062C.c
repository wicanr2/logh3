/*
 * 34CB:062C  視窗範圍內「已鎖住」的畫面格，標成髒並記上擁有者（與 34CB:02C4 只差鎖住位元的判斷方向）
 *
 * 原版：far 呼叫，參數是視窗結構的 far 指標（位移 [bp+1Ah]、段 [bp+1Ch]；序言 push bp／si／di、sub sp,10h）。
 *   唯一呼叫者是 341B:01FA（sub_40F1A，+0x70）。與 34CB:02C4 的 asm 逐指令相同，只有
 *   `test byte ptr es:[bx], 20h` 之後的 jnz 換成 jz。回傳值沒人用，描述檔 ret = none。
 *
 * 讀視窗結構（docs/internals/47-window-frame-tiles.md「結構位移」；名稱是本檔取的）：
 *   +4 槽位編號  +5 格 x  +6 格 y  +7 格寬  +8 格高（皆 byte）。
 *   +E／+10（內容緩衝區遠指標）進函式時讀進區域變數，之後每格位移加 4，但從來不拿去讀寫：
 *   本函式不使用內容緩衝區（對輸出沒有影響，本版不複製這段讀取）。
 *
 * 動作（已證實：反組譯逐條對過）：
 *   格起點 = D81E + (+6) × 0xF0 + (+5) × 6（16 位元回繞），段 = 呼叫時的 DS（GIN3_DS）。
 *   w、h 在進函式時各讀一次（(+7)、(+8) 快取在區域變數，與 34CB:0370 等重讀的函式不同）。
 *   外層 i = 0 起，i < h：列起點 = 目前格位移；格位移 += 0xF0；
 *       內層 j = 0 起，j < w：若 (格[+0] & 0x20) != 0：格[+0] = (+4) | 0x80（每次重讀 +4）；格位移 += 6
 *   其他格欄位（[+1]、[+2]、[+3]）不動。寫入的 (+4) | 0x80 沒有 0x20 位元，所以被改寫的格子不再鎖住。
 *   每格重讀 +4 對結果沒有影響（結構與畫面格表重疊使 +4 被這支函式改寫時，改寫只會多設 0x80，
 *   而寫入的值本來就帶 0x80），本版仍照原版每格重讀。
 *
 * 語意（假說）：0x20 是鎖住位元（docs/re/138、docs/internals/47）。34CB:0370 清格時寫 0xA0（髒 ＋ 鎖住 ＋ 擁有者 0），
 *   本函式隨後把範圍內仍鎖住的格子改記成本視窗擁有且為髒（使之重畫）。與 341B:01FA 的整體流程相符，但沒有驗證：未知。
 * 輸入域：視窗結構任意 byte 值，位移全部 16 位元回繞。寬 255、高 255 時超過探針步數上限的案例被丟棄；
 *   視窗指標段指到 A0000 到 BFFFF（視訊記憶體）的案例被探針丟棄。
 */
#include "gin3_reg.h"

static uint8_t f34CB062C_wb(const Mem *m, uint16_t seg, uint16_t off, uint16_t n)
{
    return rd8(m, LIN(seg, (uint16_t)(off + n)));
}

void f_34CB_062C(Mem *m, uint16_t w_off, uint16_t w_seg)
{
    const uint16_t ds = GIN3_DS;
    uint16_t cur = (uint16_t)(0xD81Eu + (uint16_t)(f34CB062C_wb(m, w_seg, w_off, 6) * 0xF0u)
                              + (uint16_t)(f34CB062C_wb(m, w_seg, w_off, 5) * 6u));
    const uint8_t w = f34CB062C_wb(m, w_seg, w_off, 7);
    const uint8_t h = f34CB062C_wb(m, w_seg, w_off, 8);
    uint8_t i = 0;

    while (i < h) {
        uint16_t cell = cur;
        uint8_t j = 0;

        cur = (uint16_t)(cur + 0xF0u);
        while (j < w) {
            if ((rd8(m, LIN(ds, cell)) & 0x20u) != 0)
                wr8(m, LIN(ds, cell), (uint8_t)(f34CB062C_wb(m, w_seg, w_off, 4) | 0x80u));
            cell = (uint16_t)(cell + 6u);
            j++;
        }
        i++;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_34CB_062C(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("34CB_062C", 2, run)
