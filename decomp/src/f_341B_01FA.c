/*
 * GIN3PS.EXE 341B:01FA  視窗鏈的「重畫標記」：對 [+0] ＝ 2 的視窗改成 1，依 [+1] 通知物件清單與畫面格表，再沿 [+0x16] 鏈走每個後繼視窗
 *
 * 原版（169 bytes，far 呼叫，一個 far 指標參數 w；序言 push bp、sub sp,4、mov bp,sp，所以 w 的位移在 [bp+0Ah]、段在 [bp+0Ch]；區域變數
 *   [bp+0]／[bp+2] ＝ 目前節點指標的位移／段），無有意義的回傳值（AX 是殘值，描述檔 ret ＝ none）：
 *   w[+0] ≠ 2（byte）→ 結束
 *   w[+0] ← 1
 *   w[+1] ＝ 0 → 35D1:01DE(word w[+4])                    ; 以 w[+4] 為鍵，把物件清單中 obj[+2] ＝ 鍵低位元組者的 obj[+0] 設成 1
 *   34CB:0370(w)                                          ; 清格（格表中擁有者符合者 ← 0xA0）
 *   節點 ← w
 *   重複：
 *     節點 ← (word 節點[+0x16], word 節點[+0x18])（位移、段；後繼視窗）
 *     節點位移 ＝ 0 且段 ＝ 0 → 離開
 *     節點[+0] ≠ 2 → 下一個
 *     節點[+1] ＝ 0xFF → 34CB:062C(節點)；下一個
 *     節點[+1] ≠ 0 → 34CB:050C(節點)；下一個
 *     否則（節點[+1] ＝ 0）→ 34CB:0413(節點)；下一個
 *   2376:0164()                                           ; 顯示物件清單的兩趟處理（沒有參數）
 * 過程（已證實：反組譯逐指令；35D1:01DE、34CB:0370／0413／050C／062C、2376:0164 各由其差異測試證實）。
 * 語意（強推論）：34CB 的 docs 註記「062C／0413／0370 的唯一呼叫者是 341B:01FA（sub_40F1A）」（docs/internals/67 視窗元件一節）；[+0] ＝ 2 對應
 *   「1 ＝ 已建立、2 ＝ 已顯示」的推論（docs/internals/47 第 74 行）；[+0x16] 是視窗鏈的後繼指標（基底狀態 ds:94F6 起的視窗 3 → 視窗 2 → 視窗 1 → 視窗 0 → 空，
 *   推論）。呼叫者 341B:01CF（台帳）。docs 沒有 341B:01FA 的專門記載。
 * 副作用：全部來自被呼叫者與 w[+0]（byte）的寫入。
 * 機器模型：C 版沒有堆疊，節點指標放在 C 區域變數；各被呼叫者的 C 版若有堆疊暫存位址，都落在探針的堆疊帶內（比對時排除）。
 * 輸入域：w 與鏈上節點指到視窗表（ds:94F6 起 4 筆，每筆 0x22 bytes）或 seeds 給的位址，不指向堆疊帶與 A0000 到 BFFFF；
 *   後繼指標 [+0x16..0x19]、物件清單 ds:EADE／EAE0 列入 freeze（隨機突變會造成環、ROM 或視訊區指標），seeds 仍可造。鏈有環或被呼叫者不終止的
 *   案例由探針丟棄（超過步數上限），C 版有 20000 圈的上界（原版每圈至少 12 步，200,000 步內走不到），只影響突變版本。基底 fcmd-base.state。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途與 [+1] 的語意為強推論／未知。
 */
#include "gin3_reg.h"

extern void f_35D1_01DE(Mem *m, uint16_t key);
extern void f_34CB_0370(Mem *m, uint16_t w_off, uint16_t w_seg);
extern void f_34CB_0413(Mem *m, uint16_t w_off, uint16_t w_seg);
extern void f_34CB_050C(Mem *m, uint16_t w_off, uint16_t w_seg);
extern void f_34CB_062C(Mem *m, uint16_t w_off, uint16_t w_seg);
extern void f_2376_0164(Mem *m);

/* 段內 word 讀取：位移各自 16 位元回繞（seg:FFFF 的高位元組在同一段的位移 0），與 dosgolem 一致 */
static uint16_t w01FA_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

void f_341B_01FA(Mem *m, uint16_t w_off, uint16_t w_seg)
{
    uint16_t cur_off = w_off, cur_seg = w_seg;
    unsigned guard;

    if (rd8(m, LIN(w_seg, w_off)) != 2)
        return;
    wr8(m, LIN(w_seg, w_off), 1);
    if (rd8(m, LIN(w_seg, (uint16_t)(w_off + 1u))) == 0)
        f_35D1_01DE(m, w01FA_rd16s(m, w_seg, (uint16_t)(w_off + 4u)));
    f_34CB_0370(m, w_off, w_seg);
    for (guard = 0; guard < 20000u; guard++) {
        uint16_t n_seg = w01FA_rd16s(m, cur_seg, (uint16_t)(cur_off + 0x18u));
        uint16_t n_off = w01FA_rd16s(m, cur_seg, (uint16_t)(cur_off + 0x16u));

        cur_seg = n_seg;
        cur_off = n_off;
        if ((uint16_t)(cur_seg | cur_off) == 0)
            break;
        if (rd8(m, LIN(cur_seg, cur_off)) != 2)
            continue;
        if (rd8(m, LIN(cur_seg, (uint16_t)(cur_off + 1u))) == 0xFFu) {
            f_34CB_062C(m, cur_off, cur_seg);
            continue;
        }
        if (rd8(m, LIN(cur_seg, (uint16_t)(cur_off + 1u))) != 0) {
            f_34CB_050C(m, cur_off, cur_seg);
            continue;
        }
        f_34CB_0413(m, cur_off, cur_seg);
    }
    f_2376_0164(m);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_341B_01FA(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("341B_01FA", 2, run)
