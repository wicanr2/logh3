/*
 * GIN3PS.EXE 3982:0198  把選單／視窗的左上角格座標夾進戰場可視範圍（格 x 在 1..0x1D−寬，格 y 在 1..0x13−高）
 *
 * 呼叫：far，五個堆疊參數（push bp; mov bp,sp，沒有 push si／di、沒有 sub sp）：
 *   [bp+06h, 08h]  px  far 指標，指向一個 byte：格 x（欄）
 *   [bp+0Ah, 0Ch]  py  far 指標，指向一個 byte：格 y（列）
 *   [bp+0Eh]       w   一個字，只讀低位元組（mov dl,[bp+0Eh]）：寬（格）
 *   [bp+10h]       h   一個字，只讀低位元組（byte ptr）：高（格）
 *   [bp+12h]       f   一個字，只讀低位元組（cmp byte ptr [bp+12h],0）：旗標
 * 無回傳值（AX 為殘值，不屬於語意輸出）。副作用：只寫 *px 與 *py（最多各寫多次），暫存器除 AX、CX、DX、BX、ES、旗標外不動。
 * 行為（反組譯逐指令，指令順序保留，因為 px 與 py 可能指到同一個位元組，每一步都重讀記憶體）：
 *   1. f 的低位元組非 0：dl ← w + 5（8 位元回繞）；h ← h + 2（8 位元，改寫的是堆疊上的參數，後面都用改過的值）；
 *      *px 大於等於 1 就減 1；*py 大於等於 1 就減 1（含 −(1,1)：游標所在格 − (1,1)）。
 *   2. *px 小於 1 就設成 1；*py 小於 1 就設成 1。（比較都是無號 byte）
 *      步驟 1 的「大於等於 1」守衛與步驟 2 的最小值夾限合起來，從外面看不出守衛是 ≥1 還是 >1
 *      （*px 為 1 時，減成 0 再被夾回 1，記憶體最終相同）；本版照反組譯寫 ≥1。守衛整個拿掉則可觀測（0 會變 0xFF）。
 *   3. 若 *px + dl（16 位元，不回繞）> 0x1D，*px ← 0x1D − dl（8 位元回繞）。
 *   4. 若 *py + h（16 位元）> 0x13，*py ← 0x13 − h（8 位元回繞）。
 * 與 docs/re/117 第 108 行、docs/re/289 第 101 行（「格＝游標所在格 − (1,1) 再夾進可視範圍」）相符，但兩份文件沒寫：
 *   「−(1,1)」與寬加 5、高加 2 只在 f 的低位元組非 0 時發生；「最小夾到 1」這一步不分 f 恆做。
 *   上限 0x1D、0x13 與 docs/re/117 的「格y ≤ 19 − 項目數」一致（0x13 ＝ 19）。w 超過 0x1D 或 h 超過 0x13 時
 *   0x1D − dl 會 8 位元回繞成大值，原版照算不報錯。
 * 呼叫者（台帳 6 個）真實呼叫時 px、py 是呼叫者堆疊上相鄰的兩個 byte（push ss; lea bx,[bp+1]; …），
 *   差異測試以 RAM 內的指標涵蓋；指標落在探針堆疊區（SS:1B8A..1F89）的案例不涵蓋（探針排除該區的寫入，
 *   而 call 推入的返回位址又會改變原版讀到的值）。
 * 輸入域：px、py 指向 RAM（不含 A0000 到 BFFFF 視訊記憶體區，那種案例由探針丟棄；也不含 F0000 以上 ROM 區，
 *   描述檔的指標段只取 9D00 與 8000，所以不需要 ROM 寫入模型）。px 與 py 可相同（別名），w、h、f 為全部 16 位元值。
 * 欄位語意：格 x／格 y 與寬高的語意是強推論（見上，來自 docs/re/117、289）；0x1D、0x13 是畫面可視格範圍的常數，來源同。
 * 推論等級：行為已證實（反組譯逐指令＋差異測試）；用途是強推論。
 */
#include "gin3_reg.h"

void f_3982_0198(Mem *m, uint16_t px_off, uint16_t px_seg, uint16_t py_off, uint16_t py_seg,
                 uint16_t w, uint16_t h, uint16_t f)
{
    uint32_t la = LIN(px_seg, px_off), lb = LIN(py_seg, py_off);
    uint8_t dl = (uint8_t)w;
    uint8_t hh = (uint8_t)h;

    if ((uint8_t)f != 0) {
        dl = (uint8_t)(dl + 5);
        hh = (uint8_t)(hh + 2);
        if (rd8(m, la) >= 1)
            wr8(m, la, (uint8_t)(rd8(m, la) - 1));
        if (rd8(m, lb) >= 1)
            wr8(m, lb, (uint8_t)(rd8(m, lb) - 1));
    }
    if (rd8(m, la) < 1)
        wr8(m, la, 1);
    if (rd8(m, lb) < 1)
        wr8(m, lb, 1);
    if ((uint16_t)(rd8(m, la) + dl) > 0x1D)
        wr8(m, la, (uint8_t)(0x1D - dl));
    if ((uint16_t)(rd8(m, lb) + hh) > 0x13)
        wr8(m, lb, (uint8_t)(0x13 - hh));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_3982_0198(m, a[0], a[1], a[2], a[3], a[4], a[5], a[6]);
    return 0;
}
GIN3_REGISTER("3982_0198", 7, run)
