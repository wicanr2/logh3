/*
 * 2376:03F6  圖形表一筆（160 bytes）搬進目的緩衝區（near，暫存器傳參 ES:SI、DI；DS = 模組資料段 3D22）
 *
 * 輸入：ES:SI = 顯示物件（讀 +2 的 byte 當編號）；DI = 目的位移（DS 內）；DS = 模組資料段（IDA 的 seg129，
 *   dosgolem 段值 0x3D22；唯一呼叫者 2376:077C 的呼叫鏈上由 2376:0538 載入）。方向旗標 DF = 0（呼叫鏈 2376:0538 有 `cld`，
 *   本函式自己也有 `cld`）。
 * 原版：
 *   1. push si／di／es；al ← es:[si+2]（byte）；call 2376:051F 得 SI ← 0x79E + 160 × al（16 位元）。
 *   2. mov ax,ds；mov es,ax；cx ← 0x50；cld；rep movsw：ES:DI ← DS:SI，80 個 word（160 bytes），
 *      逐 word 先讀後寫，SI、DI 各加 2（來源與目的重疊時，後面讀到的是前面剛寫入的值）。
 *   3. pop es／di／si 後返回。SI、DI、ES 還原；AX、CX 與旗標是暫用值，呼叫者（2376:077C 內 `push cx … pop cx`）
 *      之後不依賴它們；描述檔 ret = none，只比對記憶體寫入。
 * 所有位址都是段內 16 位元位移。word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 行為，w03F6_rd16／wr16）。
 *
 * 語意（強推論）：0x79E 起每筆 160 bytes 是 2376 模組的圖形表（2376:051F 檔頭；2376:0486 讀它 4 平面各 20 列 × 2 bytes）；
 *   本函式把第 al 筆複製到 DI 指的緩衝區（呼叫者 2376:077C 以 `mov es:[si+4],di` 記下目的位移）。名稱為假說。
 * 輸入域：ES、SI 的全部 16 位元值（物件只讀 1 byte）；DI 的全部 16 位元值。來源的最大位移 0x79E + 160 × 255 + 159 = 0xA0FD
 *   不會跨越 0xFFFF。緩衝區位址蓋到呼叫堆疊的情形不涵蓋（DS 段與堆疊段沒有重疊）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_2376_051F(uint16_t ax);

/* word 存取：兩個 byte 的位移各自在段內回繞（seg:FFFF 的高位元組在 seg:0000） */
static uint16_t w03F6_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w03F6_wr16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_2376_03F6(Mem *m, uint16_t ds, uint16_t es, uint16_t si, uint16_t di)
{
    uint16_t src = f_2376_051F(rd8(m, LIN(es, (uint16_t)(si + 2u))));
    uint16_t n;

    for (n = 0x50u; n != 0; n--) { /* rep movsw，ES = DS */
        w03F6_wr16(m, ds, di, w03F6_rd16(m, ds, src));
        src = (uint16_t)(src + 2u);
        di = (uint16_t)(di + 2u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    /* 描述檔 ds_seg：呼叫者 2376:0538 載入的模組資料段（IDA 的 seg129） */
    f_2376_03F6(m, 0x3D22u, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("2376_03F6", 3, run)
