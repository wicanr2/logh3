/*
 * GIN3PS.EXE 2376:0164  顯示物件清單的兩趟處理：重算物件的新舊矩形並標記畫面格表，再查詢「尚未可見」的物件
 *
 * 原版（130 bytes，far 呼叫，沒有堆疊參數，回傳不使用；push si／di／es 保存暫存器）：
 *   清單：筆數 word ds:EADE，陣列 ds:EAE0 起每筆 4 bytes（位移 word、段 word，`les si,[di]` 後 di += 4）。
 *   第一趟（cx ← 筆數；cx = 0 就整個函式結束，連第二趟也不做）：
 *     每筆 ES:SI ← 清單項目；[SI+0] == 2 且 [SI+1] != 0 才處理（兩個條件都是 byte）：
 *       call 2376:01E6(ES:SI)                       ; 新舊矩形更新（已 diff-tested），它把 [+1E] 抄進 [+23]、重算 [+1E]（分離旗標）
 *       [SI+23h] == 0 → call 2376:02DE(ES:SI)        ; 以舊矩形（[+24..27]）標記格表
 *       [SI+1Eh] == 0 → call 2376:02F2、2376:0306    ; 以新矩形（[+1F..22]）標記、再計數
 *     cx 用 push／pop 包住處理，`loop` 遞減（cx 是 word）。
 *   第二趟（cx ← 重新讀筆數；沒有 jcxz，筆數被第一趟改成 0 時會繞 65536 圈，超過步數上限，不在比較範圍）：
 *     每筆 ES:SI ← 清單項目；[SI+1] == 0 且 [SI+0] == 2 且 [SI+1Eh] == 0 才處理：
 *       AH ← [SI+2]（鍵值）、BX ← xchg(word[SI+1Fh])、DX ← xchg(word[SI+21h])   ; 同 02F2／0306 的矩形取法
 *       AL ← 2376:0112(AH, BX, DX)（矩形內有沒有可見、同鍵值、[+1] bit0 已設的格）
 *       [SI+1] ← AL；AL != 0 → call 2376:0306
 * 被呼叫者保存 ES、SI、DI（02DE／02F2／0306／0112 內 push／pop si、di；01E6 不碰），所以 ES:SI 與清單位置在呼叫之間不變；
 *   AX 在 02DE／02F2／0306 的入口不使用 AL、覆蓋 AH（各自的差異測試涵蓋），所以本版傳 0。
 * 兩個 word 讀取（`les` 的兩個 word、[SI+1Fh]、[SI+21h]）按段內位移各自回繞（dosgolem 的 read16，w0164_rd16s）。
 *   清單位移恆為偶數（0xEAE0 加 4 的倍數），不會正好落在 0xFFFF。
 *
 * 語意（強推論）：docs/re/381-battle-spotting-display.md 第 49 行（35D1:026D→2376:0164→01E6）、docs/internals/38-battle-units.md 第 471 行
 *   「35D1:026D→2376:0164／01E6 使用物件 offset 與大小計算邊界」。物件欄位名稱（+1 可見、+1E 分離旗標、+1F..22 新矩形、+23..27 舊矩形、
 *   +2 鍵值）見 f_2376_01E6.c、f_2376_00B2.c，皆為假說；本檔以反組譯與差異測試為準。
 * 副作用：全部來自被呼叫者（物件 +16..+27、+1E..+22，格表 ds:D81E 起的 [0]／[1]／[3]），以及第二趟的 [SI+1]。
 *   01E6 的區域變數落在堆疊帶（進入點 SP 以下 0x400 bytes），比對時排除；本函式的真實框架更深，位置與 01E6 的 C 版不同，
 *   但同樣在堆疊帶內，不影響結果。
 * 輸入域：清單項目與物件指標取暫存區、DGROUP 或種子造的位置；清單與筆數在描述檔列入 freeze（隨機突變會把指標改到 ROM 或
 *   視訊區，被呼叫者的 C 版不涵蓋 ROM 寫入）；物件不蓋到堆疊帶；矩形的列數或每列格數為 0 時原版繞圈，超過步數上限被探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位語意為假說。
 */
#include "gin3_reg.h"

extern uint16_t f_2376_01E6(Mem *m, uint16_t es, uint16_t si);
extern void f_2376_02DE(Mem *m, uint16_t ax, uint16_t es, uint16_t si);
extern void f_2376_02F2(Mem *m, uint16_t ax, uint16_t es, uint16_t si);
extern void f_2376_0306(Mem *m, uint16_t ax, uint16_t es, uint16_t si);
extern uint16_t f_2376_0112(Mem *m, uint16_t ax, uint16_t bx, uint16_t dx);

/* 段內 word 讀取：seg:FFFF 的高位元組在 seg:0000 */
static uint16_t w0164_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static uint16_t w0164_xchg(uint16_t w) { return (uint16_t)((w >> 8) | (w << 8)); }

static uint8_t w0164_b(const Mem *m, uint16_t es, uint16_t si, uint16_t d)
{
    return rd8(m, LIN(es, (uint16_t)(si + d)));
}

void f_2376_0164(Mem *m)
{
    uint16_t di = 0xEAE0u;
    uint16_t cx = w0164_rd16s(m, GIN3_DS, 0xEADEu);

    if (cx == 0)
        return;
    do {
        uint16_t si = w0164_rd16s(m, GIN3_DS, di);
        uint16_t es = w0164_rd16s(m, GIN3_DS, (uint16_t)(di + 2u));

        di = (uint16_t)(di + 4u);
        if (w0164_b(m, es, si, 0) == 2 && w0164_b(m, es, si, 1) != 0) {
            f_2376_01E6(m, es, si);
            if (w0164_b(m, es, si, 0x23) == 0)
                f_2376_02DE(m, 0, es, si);
            if (w0164_b(m, es, si, 0x1E) == 0) {
                f_2376_02F2(m, 0, es, si);
                f_2376_0306(m, 0, es, si);
            }
        }
        cx = (uint16_t)(cx - 1u);
    } while (cx != 0);

    di = 0xEAE0u;
    cx = w0164_rd16s(m, GIN3_DS, 0xEADEu);
    do {
        uint16_t si = w0164_rd16s(m, GIN3_DS, di);
        uint16_t es = w0164_rd16s(m, GIN3_DS, (uint16_t)(di + 2u));

        di = (uint16_t)(di + 4u);
        if (w0164_b(m, es, si, 1) == 0 && w0164_b(m, es, si, 0) == 2 && w0164_b(m, es, si, 0x1E) == 0) {
            uint16_t ax = (uint16_t)((uint16_t)w0164_b(m, es, si, 2) << 8);
            uint16_t bx = w0164_xchg(w0164_rd16s(m, es, (uint16_t)(si + 0x1Fu)));
            uint16_t dx = w0164_xchg(w0164_rd16s(m, es, (uint16_t)(si + 0x21u)));
            uint8_t al = (uint8_t)f_2376_0112(m, ax, bx, dx);

            wr8(m, LIN(es, (uint16_t)(si + 1u)), al);
            if (al != 0)
                f_2376_0306(m, 0, es, si);
        }
        cx = (uint16_t)(cx - 1u);
    } while (cx != 0);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2376_0164(m);
    return 0;
}
GIN3_REGISTER("2376_0164", 0, run)
