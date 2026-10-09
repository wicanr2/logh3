/*
 * 2376:07B7  把 20 列圖塊遮罩合成進目的緩衝區（near，暫存器傳參 BP、DI，DS = 模組資料段 3D22）
 *
 * 輸入：DS:BP = 來源（原版 `mov si, bp`；BP 是呼叫鏈上 2376:070E 的格位移：每欄加 2、每列加 0x280），
 *       DS:DI = 目的（呼叫者 2376:077C 由 `call 2376:0374` 算出並存進物件 `es:[si+4]`，或直接取回該快取）。
 *       來源與目的都用 DS 定址（`[si]`、`[di]` 不是字串指令，沒有 ES）。
 * 原版：進入時 push bx、cx、dx、si，結束時 pop 還原；`mov cx,14h`（20 列），每列：
 *     p0 = [si]；p1 = [si+8]；p2 = [si+10h]；p3 = [si+18h]            （四個 word，先全部讀進暫存器）
 *     mask = ~(p0 | p1 | p2 | p3)
 *     [di+2] = ([di+2] & mask) | p1
 *     [di+4] = ([di+4] & mask) | p2
 *     [di+6] = ([di+6] & mask) | p3
 *     [di]   = [di] & mask
 *     p0' = [si]                                                        （重新讀來源，不是沿用 p0）
 *     [di]   = [di] | p0'
 *     si += 0x20；di += 8                                               （push／pop cx 保存列計數，loop）
 *   每一次 `and mem,reg`、`or mem,reg` 都是獨立的讀改寫，順序照原版。來源與目的重疊時，
 *   [di] 的 and 可能改到 [si]，之後「重新讀來源」讀到的是改過的值；前面各列寫入的目的也可能蓋到後面各列的來源。
 *   本版逐步照原版順序讀寫，不預先複製來源。
 * 輸出：記憶體寫入。暫存器：BX、CX、DX、SI、BP 保存或不變；DI 結束時是 DI + 0xA0（原版不還原），AX 是暫用值。
 *   呼叫者 2376:077C 在呼叫後直接 `retn`，回到 2376:070E 的迴圈不使用 DI。描述檔 ret = none，
 *   只比對記憶體寫入。
 *
 * 語意（強推論）：[si]、[si+8]、[si+10h]、[si+18h] 是同一列 16 像素的 4 個平面（每個平面一個 word，平面間隔 8 bytes，
 *   一列來源 0x20 bytes）；每列把「4 個平面都為 0」的像素當成透明，目的先清掉有像素的位置（and mask），再 OR 上
 *   各平面。[di]、[di+2]、[di+4]、[di+6] 是目的的 4 個平面，一列 8 bytes。平面順序（[di+2] 對應 p1 等）已證實
 *   （反組譯）；「透明」的說法是對這個合成式的解讀，沒有文件記載。docs/re/407-fleet-blocks.md 第 31 行寫
 *   「圖塊零色號透明」，與此一致。
 *
 * 輸入域：BP、DI 的全部 16 位元值（位移超過 0xFFFF 時回繞；word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0，
 *   dosgolem 行為，w07B7_rd16／w07B7_wr16）；來源與目的內容任意。DS 段（3D22）的線性範圍是 0x3D220..0x4D21F，
 *   不含視訊記憶體與堆疊。
 * 推論等級：運算與順序已證實（反組譯逐指令＋差異測試，含來源與目的重疊的種子）；欄位用途為強推論。
 */
#include "gin3_reg.h"

/* word 存取：兩個 byte 的位移各自在段內回繞（seg:FFFF 的高位元組在 seg:0000） */
static uint16_t w07B7_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w07B7_wr16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_2376_07B7(Mem *m, uint16_t ds, uint16_t bp, uint16_t di)
{
    uint16_t si = bp;
    uint16_t cx;

    for (cx = 0x14u; cx != 0; cx--) {
        uint16_t p0 = w07B7_rd16(m, ds, si);
        uint16_t p1 = w07B7_rd16(m, ds, (uint16_t)(si + 8u));
        uint16_t p2 = w07B7_rd16(m, ds, (uint16_t)(si + 0x10u));
        uint16_t p3 = w07B7_rd16(m, ds, (uint16_t)(si + 0x18u));
        uint16_t mask = (uint16_t)~(p0 | p1 | p2 | p3);

        w07B7_wr16(m, ds, (uint16_t)(di + 2u), (uint16_t)(w07B7_rd16(m, ds, (uint16_t)(di + 2u)) & mask));
        w07B7_wr16(m, ds, (uint16_t)(di + 2u), (uint16_t)(w07B7_rd16(m, ds, (uint16_t)(di + 2u)) | p1));
        w07B7_wr16(m, ds, (uint16_t)(di + 4u), (uint16_t)(w07B7_rd16(m, ds, (uint16_t)(di + 4u)) & mask));
        w07B7_wr16(m, ds, (uint16_t)(di + 4u), (uint16_t)(w07B7_rd16(m, ds, (uint16_t)(di + 4u)) | p2));
        w07B7_wr16(m, ds, (uint16_t)(di + 6u), (uint16_t)(w07B7_rd16(m, ds, (uint16_t)(di + 6u)) & mask));
        w07B7_wr16(m, ds, (uint16_t)(di + 6u), (uint16_t)(w07B7_rd16(m, ds, (uint16_t)(di + 6u)) | p3));
        w07B7_wr16(m, ds, di, (uint16_t)(w07B7_rd16(m, ds, di) & mask));
        {
            uint16_t again = w07B7_rd16(m, ds, si); /* mov ax,[si]：重新讀來源 */

            w07B7_wr16(m, ds, di, (uint16_t)(w07B7_rd16(m, ds, di) | again));
        }
        si = (uint16_t)(si + 0x20u);
        di = (uint16_t)(di + 8u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    /* 描述檔 ds_seg：呼叫鏈上 2376:0538 載入的模組資料段（IDA 的 seg129）；a[0] = BP，a[1] = DI */
    f_2376_07B7(m, 0x3D22u, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2376_07B7", 2, run)
