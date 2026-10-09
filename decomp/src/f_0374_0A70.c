/*
 * GIN3PS.EXE 0374:0A70  從艦隊的編制組清單移除一組：後面的組往前搬一格，組數減一（far，兩個參數，無回傳值）
 *
 * 原版：push bp、push si、push di、sub sp,0Ah，所以參數是
 *   [bp+14h]／[bp+16h] ＝ 艦隊記錄的 far 指標（位移、段），[bp+18h] ＝ 組序 idx（word，只取低位元組：mov dl,[bp+18h]）。
 * 過程：
 *   艦隊[+0x0F] 減一（dec byte，8 位元回繞：組數原為 0 時變 0xFF）
 *   dst ＝ 艦隊位移 ＋ 0x10 ＋ idx × 4，src ＝ dst ＋ 4（位移都是 16 位元回繞，段固定是艦隊的段）
 *   k ＝ idx
 *   當 k < 艦隊[+0x0F]（byte 無號比較，減一「之後」的組數，每一輪重讀）：
 *     (dst) 的 word ← (src) 的 word；(dst+2) ← (src+2) byte；(dst+3) ← (src+3) byte      （一組 4 bytes，依這個順序）
 *     dst += 4、src += 4（16 位元）、k += 1（byte）
 * 不清最後一格：搬完之後原來的最後一組（舊組數 − 1 那格）內容原樣留著，只是組數少了一。
 * 組序 idx ≥ 新組數時迴圈不執行，只有組數減一。組數為 0 時減成 0xFF，迴圈從 idx 搬到第 254 組（讀寫到艦隊記錄之後的記憶體），
 *   原版同樣如此（輸入域內，差異測試涵蓋）。
 * 欄位：艦隊[+0x0F] ＝ 現有組數、[+0x10] 起每組 4 bytes（[0..1] 艦艇池索引、[2] 艦種、[3] 隊形槽），docs/internals/33。
 *   docs/internals/62 第 36 行：呼叫者 0374:0896 對戰後艦艇 [+0x08] 為 0 的組呼叫 `0374:0A70(f, i)` 移除，
 *   「後面往前搬、不清最後一格」，與本檔一致。
 * 呼叫者：0374:0896（台帳 callers＝1）。
 *
 * 模型限制：
 *   - 字的存取逐位元組各自回繞位移（8086 行為，docs 與 dosgolem cpu.read16 註解一致）：位移 0xFFFF 的 word 取 0xFFFF 與 0x0000。
 *   - 位址 0xF0000 以上是 BIOS ROM，寫入無效（dosgolem 的 ROMBase ＝ 0xF0000）。正常呼叫者的艦隊指標在 DS 內，不會到 ROM，
 *     但本版把艦隊指標當任意 far 指標處理，所以寫入走 w0A70_put8。
 *   - 不涵蓋線性位址超過 1 MB（A20）的位址。
 * 輸入域：艦隊指標取艦隊表任一筆（32 筆）及 seeds 造的任意位址；idx 高位元組任意（原版忽略）；組數 0..255。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位名稱是 docs 的結論，位移已證實。
 */
#include "gin3_reg.h"

static void w0A70_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

/* 8086 的 word 存取：兩個 byte 各自回繞位移 */
static uint16_t w0A70_rdw(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1))) << 8));
}

static void w0A70_wrw(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w0A70_put8(m, LIN(seg, off), (uint8_t)v);
    w0A70_put8(m, LIN(seg, (uint16_t)(off + 1)), (uint8_t)(v >> 8));
}

void f_0374_0A70(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t idx_w)
{
    uint8_t idx = (uint8_t)idx_w;
    uint16_t dst, src;
    uint8_t k;

    w0A70_put8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x0F)),
               (uint8_t)(rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x0F))) - 1));
    dst = (uint16_t)(fleet_off + 0x10 + (uint16_t)(idx * 4u));
    src = (uint16_t)(dst + 4);
    k = idx;
    while (k < rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x0F)))) {
        w0A70_wrw(m, fleet_seg, dst, w0A70_rdw(m, fleet_seg, src));
        w0A70_put8(m, LIN(fleet_seg, (uint16_t)(dst + 2)), rd8(m, LIN(fleet_seg, (uint16_t)(src + 2))));
        w0A70_put8(m, LIN(fleet_seg, (uint16_t)(dst + 3)), rd8(m, LIN(fleet_seg, (uint16_t)(src + 3))));
        dst = (uint16_t)(dst + 4);
        src = (uint16_t)(src + 4);
        k++;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_0374_0A70(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("0374_0A70", 3, run)
