/*
 * 348A:0281  COMANI 動畫的「下一格」前置步驟（near，暫存器傳參；碰計時器變數與程式段內的兩個狀態字）
 *
 * 原版：near 呼叫，無堆疊參數。呼叫者是 348A:0273 與 348A:027A（各自接著呼叫格種 0／1 的畫圖常式
 *   348A:02BA／036C，那兩支碰 VGA 與 I/O 埠，不在本批）。輸入暫存器：
 *     DS = 動畫資料（COMANI）所在的段  SI = 目前格的記錄位移 + 1（呼叫者已把格種 byte 讀過、si 加 1）
 *   輸出暫存器：SI、DI、AX、DX（BX 被設成常數 0x8F10，旗標被最後的 add 改過；呼叫者 02BA／036C 讀 SI 與 DI，
 *   不讀 AX／DX／旗標）。描述檔 ret = pack:si,di,ax,dx；旗標不納入比對。
 *
 * 動作（已證實：反組譯逐條對過）：
 *   1. 暫時把 DS 換成 DGROUP（push ds; mov ax,54B3h; mov ds,ax），等 ds:8EFE（word，動畫計時器）變成 0
 *      （cmp word ptr, 0; jnz 自旋）。
 *   2. xlat：AL = ds:[8F10h + ds:8F37]（byte 表 ds:8F10，索引是 ds:8F37；位移 16 位元回繞）；
 *      ds:8EFE ← AL（高位元組 0）。這就是「每一格之前等計時器到 0，再設成動畫速度的間隔」
 *      （docs/internals/20-comani-bmg.md 第 50 行；間隔表 ds:8F10 的內容 08／04／00 見該文件）。
 *   3. pop ds，回到呼叫者的 DS（動畫資料段）。以下用這個 DS 讀。
 *   4. SI ← SI + 1；AX ← word [SI]（記錄 +2：到下一格的距離）；SI ← SI + 2；AX ← AX + SI；
 *      cs:01D8 ← AX（下一格記錄的位移，存在程式段內的 word，IDA 名稱 word_415E8；
 *      348A:01DC 以 cs:word_415E8 為迴圈的記錄指標）。
 *   5. DI ← cs:01DA（IDA 名稱 word_415EA，348A:01DC 在進迴圈前存入的 VRAM 位移）。
 *   6. AX ← word [SI]（記錄 +4：VRAM 位移）；SI ← SI + 2；DX ← AX 的符號延伸（cwd）；DI ← DI + AX（16 位元）。
 *   cs 的段值是 348A（dosgolem 的載入位置，與函式位址同段）。cs:01D8、cs:01DA 在程式段內（線性 0x34A78 起），
 *   原版把「狀態」存在程式段裡，本版同樣寫該位址以對上寫入集合。
 *
 * 語意：欄位名稱與用途出自 docs/internals/20-comani-bmg.md 第 28 到 50 行（格的記錄 +2 距離、+4 VRAM 位移）；
 *   AX、DX 是 02BA／036C 用不到的暫存值（DX 是 cwd 的結果）。
 *
 * 輸入域與未驗證處：
 *   - ds:8EFE 非 0 時原版自旋，等計時器中斷（IRQ0 每 10 次減一次，docs/internals/20）把它減到 0。
 *     差異測試不跑中斷，步數超過上限的案例被探針丟棄，所以該情形不在驗證範圍。本版當作等待已經結束，
 *     照樣執行後面的步驟（假說：真機上等完計時器之後的行為就是這樣）。
 *   - word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（rd16s），不是下一段。
 *   - 程式段內的 cs:01D8／01DA 與資料位址重疊（DS = 348A 且 SI 指到 01D8 附近）時，寫入先於第二次 word 讀取，
 *     順序照原版；差異測試用 seeds 造這種別名。
 *   - 不涵蓋 DS 指到 A0000 到 BFFFF（視訊記憶體）的情形（探針丟棄）。
 */
#include "gin3_reg.h"

#define F348A0281_CS 0x348Au /* dosgolem 載入位置下，函式所在的程式段 */

static uint16_t f348A0281_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 回傳 pack:si,di,ax,dx（由高位到低位） */
uint64_t f_348A_0281(Mem *m, uint16_t ds, uint16_t si)
{
    uint16_t ax, dx, di;

    /* 等計時器（本版不模擬中斷，視為已到 0）；然後 ds:8EFE ← 速度間隔表 ds:8F10[ds:8F37] */
    wr16(m, DS(0x8EFE), rd8(m, DS((uint16_t)(0x8F10u + rd8(m, DS(0x8F37))))));

    si = (uint16_t)(si + 1u);
    ax = f348A0281_rd16s(m, ds, si);
    si = (uint16_t)(si + 2u);
    ax = (uint16_t)(ax + si);
    wr16(m, LIN(F348A0281_CS, 0x01D8u), ax);
    di = rd16(m, LIN(F348A0281_CS, 0x01DAu));
    ax = f348A0281_rd16s(m, ds, si);
    si = (uint16_t)(si + 2u);
    dx = (ax & 0x8000u) ? 0xFFFFu : 0u;
    di = (uint16_t)(di + ax);
    return ((uint64_t)si << 48) | ((uint64_t)di << 32) | ((uint64_t)ax << 16) | dx;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_348A_0281(m, a[0], a[1]);
}
GIN3_REGISTER("348A_0281", 2, run)
