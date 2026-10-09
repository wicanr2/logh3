/*
 * GIN3PS.EXE 37A9:004C  把一筆元件記錄（按鈕、頁籤、箭頭）畫進視窗的內容緩衝區：寫字碼、顏色與裝飾旗標
 * 呼叫：far，兩個 far 指標參數（序言 push bp／si／di、sub sp,0Eh，第一個參數在 [bp+18h]）：
 *   win  [bp+18h]（位移）、[bp+1Ah]（段）：視窗物件
 *   rec  [bp+1Ch]（位移）、[bp+1Eh]（段）：元件記錄（12 bytes，docs/internals/23、docs/re/138）
 * 無回傳值（AX 是殘值，描述檔 ret ＝ none）。呼叫者 37A9:000F（依元件清單逐筆呼叫）。
 *
 * 元件記錄：+0 x（內容格）、+1 y、+2 w、+3 h、+4 啟用旗標、+5 動作碼（本函式不讀）、+6 種類、+8 word／+0A word ＝ 字串 far 指標的位移與段（字串是雙位元組字元）。
 * 視窗物件：+0x0B byte ＝ 內容緩衝區每列格數（stride）；+0x0E／+0x10 word ＝ 內容緩衝區 far 指標。內容格 4 bytes：+0 word 字碼、+2 byte 顏色、+3 byte 旗標。
 * 顏色表：ds:8BE4[種類]（基底狀態 87 87 87 87 87 87 97 00）。
 *
 * 算法（asm 逐條，算術 16 位元回繞，byte 計數器 8 位元；順序也照做，因為緩衝區可能蓋到記錄或字串）：
 *   種類 ＝ 4，或啟用旗標 ＝ 0：什麼都不做。
 *   color ＝ ds:8BE4[種類]；字串指標 s ＝ rec[+8..+0B]。
 *   種類 ＝ 6（箭頭鈕）：格位址 d ＝ 緩衝區 + (rec.y × stride) × 4 + rec.x × 4；
 *       i ＝ 0 起，i < rec.w（每次重讀）：word(d) ← (s[0] << 8) + s[1]、byte(d+2) ← color；d ← d + 4、s ← s + 2。
 *       不寫旗標（d+3）。字串在這裡是 w 個雙位元組字元，**每格兩個 byte**，不是 docs/re/138 第 179 行寫的「一個位元組一個位元組」。
 *   其他種類：j ＝ 0 起，j < rec.h（每次重讀）：
 *       d ＝ 緩衝區 + ((rec.y + j) × stride) × 4 + rec.x × 4 − 4（即記錄 x 前一格）。
 *       種類 ≠ 5：byte(d+3) |= 0x10（左端）；i ＝ 0。
 *       種類 ＝ 5（頁籤）：d ← d + 4；word(d) ← 0xFF16；i ＝ 1。
 *       d ← d + 4。
 *       i < rec.w（每次重讀）的每格：word(d) ← (s[0] << 8) + s[1]、byte(d+2) ← color、byte(d+3) ← 0x08；s ← s + 2、d ← d + 4、i ← i + 1。
 *       字串指標 s 跨列延續，不在每列重設。
 *       byte(d+3) |= 0x20（右端）；j ← j + 1。
 *   原版的區域變數 [bp+5]（旗標，恆為 0）參與 `or al,8; or al,[bp+5]`，值恆為 0x08。
 *
 * 機器模型：寫入落在線性 0xF0000 以上（BIOS ROM）時 dosgolem 安靜丟棄（internal/machine/bios_rom.go）；word 的兩個位元組位移各自在段內回繞。
 *   本檔以 put8／put16 表達。平面視訊記憶體的寫入由探針丟棄。
 * 輸入域：視窗取 ds:9518 與 ds:962C（基底狀態兩個都開著）；記錄取 ds:1CCE 起 6 筆（種類 4、4、4、5、5、5）加種子造的種類 0、6、7 與停用記錄，
 *   記錄與字串指標可被突變成任意值。兩個窗的緩衝區段凍結（種子仍可改）：緩衝區若蓋到呼叫堆疊（線性 0x65A3A 到 0x65E3A）會改寫函式自己的區域變數，
 *   原版行為隨之改變，C 版沒有堆疊無法重現，屬域外。w × h 很大、超過探針步數上限的案例被丟棄。不涵蓋線性位址超過 1 MB（A20）的位址。
 * 推論等級：位移、算術、值與順序已證實（asm 與差異測試）；視窗與記錄欄位、種類名稱為強推論（docs/re/156、138、internals/23）。
 */
#include "gin3_reg.h"

static void f_37A9_004C_put8(Mem *m, uint16_t seg, uint16_t off, uint8_t v)
{
    uint32_t lin = LIN(seg, off);

    if (lin < 0xF0000u)
        wr8(m, lin, v);
}

static void f_37A9_004C_put16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    f_37A9_004C_put8(m, seg, off, (uint8_t)v);
    f_37A9_004C_put8(m, seg, (uint16_t)(off + 1u), (uint8_t)(v >> 8));
}

static uint16_t f_37A9_004C_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 取字串的一個雙位元組字元：(s[0] << 8) + s[1] */
static uint16_t f_37A9_004C_pair(const Mem *m, uint16_t seg, uint16_t off)
{
    uint16_t hi = rd8(m, LIN(seg, off));
    uint16_t lo = rd8(m, LIN(seg, (uint16_t)(off + 1u)));

    return (uint16_t)((hi << 8) + lo);
}

void f_37A9_004C(Mem *m, uint16_t win_off, uint16_t win_seg, uint16_t rec_off, uint16_t rec_seg)
{
    uint8_t kind, color, i, j;
    uint16_t str_off, str_seg, d_off, d_seg;

    if (rd8(m, LIN(rec_seg, (uint16_t)(rec_off + 6u))) == 4 || rd8(m, LIN(rec_seg, (uint16_t)(rec_off + 4u))) == 0)
        return;
    kind = rd8(m, LIN(rec_seg, (uint16_t)(rec_off + 6u)));
    color = rd8(m, DS(0x8BE4 + kind));
    str_seg = f_37A9_004C_rd16s(m, rec_seg, (uint16_t)(rec_off + 0x0Au));
    str_off = f_37A9_004C_rd16s(m, rec_seg, (uint16_t)(rec_off + 8u));

    if (kind == 6) {
        uint16_t stride = rd8(m, LIN(win_seg, (uint16_t)(win_off + 0x0Bu)));
        uint16_t rowoff = (uint16_t)((uint16_t)(rd8(m, LIN(rec_seg, (uint16_t)(rec_off + 1u))) * stride) << 2);
        uint16_t x4 = (uint16_t)(rd8(m, LIN(rec_seg, rec_off)) << 2);

        d_seg = f_37A9_004C_rd16s(m, win_seg, (uint16_t)(win_off + 0x10u));
        d_off = f_37A9_004C_rd16s(m, win_seg, (uint16_t)(win_off + 0x0Eu));
        d_off = (uint16_t)(d_off + x4);
        d_off = (uint16_t)(d_off + rowoff);
        for (i = 0; i < rd8(m, LIN(rec_seg, (uint16_t)(rec_off + 2u))); i++) {
            f_37A9_004C_put16(m, d_seg, d_off, f_37A9_004C_pair(m, str_seg, str_off));
            f_37A9_004C_put8(m, d_seg, (uint16_t)(d_off + 2u), color);
            d_off = (uint16_t)(d_off + 4u);
            str_off = (uint16_t)(str_off + 2u);
        }
        return;
    }

    for (j = 0; j < rd8(m, LIN(rec_seg, (uint16_t)(rec_off + 3u))); j++) {
        uint16_t row = (uint16_t)(rd8(m, LIN(rec_seg, (uint16_t)(rec_off + 1u))) + j);
        uint16_t stride = rd8(m, LIN(win_seg, (uint16_t)(win_off + 0x0Bu)));
        uint16_t rowoff = (uint16_t)((uint16_t)(row * stride) << 2);
        uint16_t x4 = (uint16_t)(rd8(m, LIN(rec_seg, rec_off)) << 2);

        d_seg = f_37A9_004C_rd16s(m, win_seg, (uint16_t)(win_off + 0x10u));
        d_off = f_37A9_004C_rd16s(m, win_seg, (uint16_t)(win_off + 0x0Eu));
        d_off = (uint16_t)(d_off + x4);
        d_off = (uint16_t)(d_off + 0xFFFCu);
        d_off = (uint16_t)(d_off + rowoff);
        if (rd8(m, LIN(rec_seg, (uint16_t)(rec_off + 6u))) != 5) {
            f_37A9_004C_put8(m, d_seg, (uint16_t)(d_off + 3u), (uint8_t)(rd8(m, LIN(d_seg, (uint16_t)(d_off + 3u))) | 0x10));
            i = 0;
        } else {
            d_off = (uint16_t)(d_off + 4u);
            f_37A9_004C_put16(m, d_seg, d_off, 0xFF16);
            i = 1;
        }
        d_off = (uint16_t)(d_off + 4u);
        while (i < rd8(m, LIN(rec_seg, (uint16_t)(rec_off + 2u)))) {
            f_37A9_004C_put16(m, d_seg, d_off, f_37A9_004C_pair(m, str_seg, str_off));
            f_37A9_004C_put8(m, d_seg, (uint16_t)(d_off + 2u), color);
            f_37A9_004C_put8(m, d_seg, (uint16_t)(d_off + 3u), 0x08);
            str_off = (uint16_t)(str_off + 2u);
            d_off = (uint16_t)(d_off + 4u);
            i++;
        }
        f_37A9_004C_put8(m, d_seg, (uint16_t)(d_off + 3u), (uint8_t)(rd8(m, LIN(d_seg, (uint16_t)(d_off + 3u))) | 0x20));
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_37A9_004C(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("37A9_004C", 4, run)
