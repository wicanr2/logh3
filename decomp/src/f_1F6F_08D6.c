/*
 * GIN3PS.EXE 1F6F:08D6  依艦種調整艦艇記錄 +0x46 的值（word）
 *
 * 原版：far 呼叫，一個參數＝艦艇池記錄的 far 指標（位移在前、段在後），無回傳值（描述檔 ret = none）。
 *
 * 逐條對應反組譯（IDA sub_2CB36；本函式沒有 push si／di，也沒有區域變數）：
 *   若 艦艇[+0x03](byte) == 0：             艦艇[+0x46] ← low16(3 × 艦艇[+0x46])
 *   否則若 艦艇[+0x04](byte) == 6：         艦艇[+0x46] ← low16(艦艇[+0x46] << 1)         ; shl word ptr，進位丟掉
 *   否則若 艦艇[+0x04](byte) == 5：         艦艇[+0x46] ← low16(艦艇[+0x46] × 3) >> 1      ; mul 的 AX 再 shr，DX 丟掉
 *   其餘不動。
 * 乘法是 16 位元的 mul（結果取 AX），所以 3 × v 超過 0xFFFF 時先截成 16 位元再做；艦種 5 的右移也是對截斷後的值。
 *
 * 推論等級
 *   已證實（差異測試的位元組與寫入集合）：三個分支的條件順序、運算、欄位位移。
 *   強推論（docs/re/412-fleet-auto-command.md 第 32 行、docs/spec/101-fleet-auto-command.md 第 32 行）：
 *     呼叫者 1F6F:061F（反組譯 sub_2C87F）只在 3982:022D(攻方記錄[+0x00] 的 word) == 1 時呼叫本函式，傳入的是
 *     061F 的第二個參數（目標艦艇池記錄的 far 指標；061F 自己對同一筆清 +0x48／+0x4A／+0x4C 並測 +0x3C 的位元）。
 *     docs/re/412 與 docs/spec/101 把本函式描述成對目標權重（score）的修正：旗艦 ×3、艦種 6 ×2、
 *     艦種 5 取低 word ×3 後除 2，與本檔逐項相符。+0x03 在 docs/internals/64 是組序（0 ＝ 旗艦那一組），+0x04 是艦種。
 *   未解的文件衝突：docs/internals/38-battle-units.md 欄位表與 docs/re/96 寫「艦艇記錄 +0x46／+0x48 整場都是 0xFFFF，
 *     戰鬥期間沒有任何寫入」。本函式會寫 +0x46（呼叫端傳入的指標若是艦艇池記錄就直接矛盾）；docs/re/96 的證據只來自
 *     一條 s3fight 序列。本檔不裁決，差異測試只證明函式本身對這個位移做了什麼。
 *   3 倍、2 倍、1.5 倍各對應什麼戰術意義沒有證據，不推論。
 *
 * 輸入域：艦艇指標任意（描述檔從池取，並由突變導向讀取集改 +0x03／+0x04／+0x46）。
 */
#include "gin3_reg.h"

void f_1F6F_08D6(Mem *m, uint16_t ship_off, uint16_t ship_seg)
{
    if (rd8(m, LIN(ship_seg, ship_off + 0x03)) == 0) {
        uint16_t v = rd16(m, LIN(ship_seg, ship_off + 0x46));

        wr16(m, LIN(ship_seg, ship_off + 0x46), (uint16_t)(3u * v));
    } else if (rd8(m, LIN(ship_seg, ship_off + 0x04)) == 6) {
        uint16_t v = rd16(m, LIN(ship_seg, ship_off + 0x46));

        wr16(m, LIN(ship_seg, ship_off + 0x46), (uint16_t)(v << 1));
    } else if (rd8(m, LIN(ship_seg, ship_off + 0x04)) == 5) {
        uint16_t v = rd16(m, LIN(ship_seg, ship_off + 0x46));

        wr16(m, LIN(ship_seg, ship_off + 0x46), (uint16_t)((uint16_t)(v * 3u) >> 1));
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1F6F_08D6(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("1F6F_08D6", 2, run)
