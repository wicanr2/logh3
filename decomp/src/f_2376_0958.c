/*
 * 2376:0958  圖塊對齊時的來源搬運（near，暫存器傳參 SI，DS = 模組資料段 3D22）
 *
 * 輸入：DS:SI = 來源（呼叫者 2376:087A 先 push si 再呼叫，SI 是物件的圖形位址）；ES = DS（原版開頭 `mov ax,ds; mov es,ax`）；
 *   方向旗標 DF = 0（本函式沒有 `cld`；2376:07FA 與它的三個呼叫者 321F:06FC、0897、0A08 都沒有 `cld`
 *   或 `std`，程式的其他地方是否改過 DF 未查證；本版與描述檔的基底狀態都是 DF = 0，不涵蓋 DF = 1）。
 * 讀：ds:790 = R（列數，也是 `loop` 的次數）、ds:78E = w（每列搬幾個 word）。
 * 原版：
 *   1. `rep stosw`：ES:0 起 0x3C0 個 word（1920 bytes）寫成 0。
 *   2. 迴圈 R 次（`loop`，cx 先載入 ds:790）：每次先 `push cx; push di`，cx ← ds:78E（每列重讀），
 *      `rep movsw`（ES:DI ← DS:SI，逐個 word 先讀後寫，SI、DI 各加 2），然後 `pop di; pop cx`，
 *      di ← di + 0x10，`loop`。所以目的每列前進 16 bytes，來源連續前進 w × 2 bytes。
 *   ds:78E 在每一列開頭重讀，目的列若蓋到 ds:78E，之後各列的 w 跟著變；逐 word 的先讀後寫使來源與目的
 *   重疊時後面讀到的是前面剛寫入的值。本版照同一順序逐 word 讀寫，不用 memcpy。
 * 輸出：記憶體寫入。暫存器（AX、CX、DX、SI、DI）在返回時是暫用值，呼叫者 2376:087A 之後 `pop si` 並重算，
 *   描述檔 ret = none，只比對記憶體寫入。
 *
 * 語意（強推論）：呼叫者 2376:087A 在 X 為 8 的倍數（`test ax,7` 為 0）時呼叫本函式，不需要位移，
 *   把來源圖塊原樣複製到 ds:0 起的緩衝區；X 不是 8 的倍數時走 2376:08A0 的位移路徑。ds:78E 與 ds:790 由
 *   2376:07FA 設定：78E = 第 3 個參數 >> 1（`shr ax,1` 後存 word_4A51E）、790 = 第 4 個參數 × 4（兩次
 *   `add ax,ax` 後存 word_4A520）。基底狀態的 78E = 1、790 = 0x40 相當於寬 16 像素、高 16 列的圖塊
 *   （「× 4」可能是 4 個平面，假說）。docs/re/407-fleet-blocks.md 與 docs/spec/095-fleet-blocks.md
 *   只記到 09AE／0A34／08A0，沒有 0958 的逐位元組記載。
 *
 * 輸入域：SI 的全部 16 位元值；ds:78E 與 ds:790 由描述檔 seeds 給定（凍結，不讓隨機突變改成會讓單案跑上億步的值）。
 *   ds:790 = 0 時 `loop` 會繞 65536 次，超過每案步數上限，不涵蓋。word 存取在位移 0xFFFF 時，高位元組取
 *   同一段的位移 0（dosgolem 行為，w0958_rd16／w0958_wr16）。
 * 推論等級：搬運順序、位移、次數已證實（反組譯逐指令＋差異測試）；78E、790 的來源為強推論，圖塊配置為假說。
 */
#include "gin3_reg.h"

/* word 存取：兩個 byte 的位移各自在段內回繞（seg:FFFF 的高位元組在 seg:0000） */
static uint16_t w0958_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w0958_wr16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_2376_0958(Mem *m, uint16_t ds, uint16_t si)
{
    uint16_t di = 0;
    uint16_t cx;
    uint16_t i;

    for (i = 0; i < 0x3C0u; i++) { /* rep stosw，ax = 0 */
        w0958_wr16(m, ds, di, 0);
        di = (uint16_t)(di + 2u);
    }
    cx = rd16(m, LIN(ds, 0x790u));
    di = 0;
    do {
        uint16_t row = di;
        uint16_t n = rd16(m, LIN(ds, 0x78Eu)); /* mov cx, ds:78Eh */

        while (n != 0) { /* rep movsw */
            w0958_wr16(m, ds, di, w0958_rd16(m, ds, si));
            si = (uint16_t)(si + 2u);
            di = (uint16_t)(di + 2u);
            n--;
        }
        di = (uint16_t)(row + 0x10u);
    } while (--cx != 0); /* loop */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    /* 描述檔 ds_seg：呼叫者 2376:07FA 載入的模組資料段（IDA 的 seg129） */
    f_2376_0958(m, 0x3D22u, a[0]);
    return 0;
}
GIN3_REGISTER("2376_0958", 1, run)
