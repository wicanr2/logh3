/*
 * GIN3PS.EXE 341B:0004  視窗系統的初始化：清窗口鏈結串列的頭尾指標與計數、依序呼叫六個初始化步驟、清空 ds:D80E 起 16 bytes
 *
 * 原版：far 呼叫，沒有參數（80 bytes，沒有 push bp 序言），沒有回傳值（描述檔 ret ＝ none）。逐指令：
 *   word ds:D804 ← 0；word ds:D806 ← 0；word ds:D808 ← 0；word ds:D80A ← 0；byte ds:D80C ← 0
 *   2376:0082()        ; 清空畫面格表 ds:D81E 起 4800 bytes
 *   3703:0011()        ; seg129 的兩個 word 設初值（3D22:DDBE ← 0、3D22:DDC0 ← 0x60BE）
 *   2376:0092()        ; 畫面格表的旗標整理（緊接著上一步清空之後，表全為 0，這一步不改任何 byte）
 *   35D1:0002()        ; 戰場圖示物件清單的筆數 ds:EADE ← 0
 *   36A1:0000()        ; seg129 內前四塊緩衝區的遠指標存進 ds:F128..F137
 *   36A1:0076()        ; 第五塊緩衝區的遠指標存進 ds:F138..F13B
 *   3C9C:000A(ds:D80E, 0, 0x10)   ; 16 bytes 清 0（memset）
 * 語意（強推論）：docs/internals/20 與台帳：0374:0191（unknown，開機初始化序列）唯一的呼叫者；ds:D804..D80C 與 ds:D80E.. 是窗口系統的串列頭尾、
 *   計數與緩衝（341B:039D、0353、057F、0677 走這份串列：docs/internals/67「艦隊完整度、選項縮放與視窗格座標」一節）。欄位名稱未驗，位移與運算才是證據。
 *   2376:0092 在 2376:0082 之後呼叫時無可觀察效果（表已全為 0，and 後仍為 0），C 版仍照原順序呼叫，突變清單不含「略過 0092」（等價突變）。
 * 副作用：ds:D804..D80C 共 9 bytes 清 0（D80D 不動）、ds:D80E..D81D 共 16 bytes 清 0、ds:D81E..EADD 清 0、ds:EADE 清 0（word）、ds:F128..F13B 的緩衝區指標、
 *   seg129 的兩個 word。各步驟寫入的範圍互不重疊，順序不影響結果（0082 與 0092 除外，見上）。
 * 輸入域：沒有輸入（函式與各步驟都不讀記憶體）；探針沒有讀取集可突變，案例全靠 seeds：先把寫入範圍內外都填成非最終值，才看得出少寫、多寫與位址偏差。
 * 推論等級：已證實（反組譯逐指令＋差異測試；被呼叫者各自由差異測試證實）；用途為強推論。
 */
#include "gin3_reg.h"

extern void f_2376_0082(Mem *m);
extern uint16_t f_3703_0011(Mem *m);
extern void f_2376_0092(Mem *m);
extern void f_35D1_0002(Mem *m);
extern uint16_t f_36A1_0000(Mem *m);
extern uint16_t f_36A1_0076(Mem *m);
extern uint32_t f_3C9C_000A(Mem *m, uint16_t off, uint16_t seg, uint16_t val, uint16_t count);

void f_341B_0004(Mem *m)
{
    wr16(m, DS(0xD804u), 0);
    wr16(m, DS(0xD806u), 0);
    wr16(m, DS(0xD808u), 0);
    wr16(m, DS(0xD80Au), 0);
    wr8(m, DS(0xD80Cu), 0);
    f_2376_0082(m);
    f_3703_0011(m);
    f_2376_0092(m);
    f_35D1_0002(m);
    f_36A1_0000(m);
    f_36A1_0076(m);
    f_3C9C_000A(m, 0xD80Eu, GIN3_DS, 0, 0x10u);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_341B_0004(m);
    return 0;
}
GIN3_REGISTER("341B_0004", 0, run)
