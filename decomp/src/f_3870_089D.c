/*
 * GIN3PS.EXE 3870:089D  依模式設定繪圖裁切矩形 ds:8D6A（xmin）、8D6C（ymin）、8D6E（xmax）、8D70（ymax）
 * 呼叫：far，一個 word 參數 mode（[bp+6]），只取低位元組（`mov cl,[bp+6]; mov al,cl; xor ah,ah`）。
 *   mode 0：矩形 ← (0x10, 0x14, 0x1CF, 0x17B)，即 (16,20)–(463,379)。
 *   mode 1：先寫 (0x08, 0xA8, 0x277, 0x187)，然後**落入** mode 2 的程式（loc_45B41 結尾沒有 jmp），
 *           所以最後的值和 mode 2 相同；前四個寫入會執行，但被 mode 2 的四個寫入覆蓋（被覆蓋的寫入，不是不可達的碼）。
 *   mode 2：矩形 ← (0, 0, 0x27F, 0x18F)，即 (0,0)–(639,399)（整個畫面）。
 *   mode >= 3：不寫。
 * 回傳 AX：mode 0 到 2 都是 0（mode 0 本來就是 0；mode 1、2 走 `xor ax,ax`），mode >= 3 是 mode 的低位元組。呼叫者不使用。
 * 四個欄位是 word，各自寫入，順序 8D6A、8D6C、8D6E、8D70。
 * 證據：asm 逐條；矩形四欄的語意見 docs/internals/52 第 93 到 95 行（xmin、ymin、xmax、ymax，兩端含）。
 *   mode 0 的數值是 (16,20)–(463,379)；docs/re/155 第 70 行、docs/internals/52 第 95 行寫「戰場那一組是 (16,60)–(463,419)」，兩者 y 都差 40，
 *   而畫面上方保留 40 列（2376:09AE 的 0xC80 ＝ 40 × 80）。假說：同一個矩形，一邊是畫布座標、一邊是畫面座標；未驗證，不判定哪份文件有誤。docs/re/324 第 22 行把 `3870:089D(1)` 註為「載 ANI_PNL1.A98 當底圖」，
 *   本函式只設矩形，沒有任何檔案或繪圖動作（該行的說明推測來自同一時段的其他呼叫）。
 * 輸入域：全部 16 位元的 mode。推論等級：已證實（asm 與差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_3870_089D(Mem *m, uint16_t mode)
{
    uint8_t sel = (uint8_t)mode;

    if (sel == 0) {
        wr16(m, DS(0x8D6A), 0x0010);
        wr16(m, DS(0x8D6C), 0x0014);
        wr16(m, DS(0x8D6E), 0x01CF);
        wr16(m, DS(0x8D70), 0x017B);
        return 0;
    }
    if (sel == 1) {
        wr16(m, DS(0x8D6A), 0x0008);
        wr16(m, DS(0x8D6C), 0x00A8);
        wr16(m, DS(0x8D6E), 0x0277);
        wr16(m, DS(0x8D70), 0x0187);
        /* 落入 mode 2 */
    }
    if (sel == 1 || sel == 2) {
        wr16(m, DS(0x8D6A), 0x0000);
        wr16(m, DS(0x8D6C), 0x0000);
        wr16(m, DS(0x8D6E), 0x027F);
        wr16(m, DS(0x8D70), 0x018F);
        return 0;
    }
    return sel;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3870_089D(m, a[0]);
}
GIN3_REGISTER("3870_089D", 1, run)
