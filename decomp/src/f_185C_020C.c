/*
 * GIN3PS.EXE 185C:020C  艦隊記錄移除第 i 組編制：組數減 1，第 i+1 組以後各往前搬一組
 *
 * 原版：far 呼叫，兩個參數：艦隊記錄的 far 指標（位移在前、段在後）與組序號 i（一個 word，原版只取低位元組），
 *   無回傳值（描述檔 ret = none）。序言 push bp／si／di、sub sp,0Ah，所以 f 在 [bp+14h]（位移）、
 *   [bp+16h]（段），i 在 [bp+18h]（`mov dl,[bp+18h]` 只讀低位元組）。
 *
 * 語意（已證實：docs/re/339 §3.1 "185C:0177(f)：……185C:020C(f, i)（移除第 i 組、後面往前搬，i 不遞增）"；
 *   docs/internals/33 艦隊記錄 +0x0F 現有組數、+0x10 起十組編制每組 4 bytes（[0..1] word 艦艇池索引、
 *   [2] 艦種、[3] 隊形槽）；本檔與反組譯逐條對過）：
 *   byte f[+0x0F] ← f[+0x0F] − 1                  ; 先減，從 0 會繞到 0xFF
 *   dst ＝ f + 0x10 + i × 4、src ＝ dst + 4         ; i 取低位元組，位移運算 16 位元回繞
 *   for j ＝ i; j < f[+0x0F]（byte、無號，每圈重讀）; j++：
 *       word dst[+0] ← src[+0]、byte dst[+2] ← src[+2]、byte dst[+3] ← src[+3]
 *       dst += 4、src += 4
 *   即把減過之後的組數範圍內，第 i+1 組起的記錄逐組往前搬一格。最後一組的舊內容留在原處，不清除。
 *
 * 原版的細節（本版照做）：
 *   - 組數已經是 0 時 dec 會變 0xFF，迴圈從 i 跑到 254，搬到最多 1020 bytes。
 *   - 迴圈上限每圈重新從記憶體讀（les bx,[bp+14h]；mov cl,es:[bx+0Fh]），搬動期間若覆寫到它就生效。
 *   - 先讀 src 再寫 dst，逐欄（word、byte、byte）進行，順序與反組譯相同。
 *   - word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（rd16s／wr16s_rom），不是下一段的第一個 byte。
 *   - 艦隊記錄段被改到 0xF000 以上時，寫入落在 BIOS ROM 區（線性 0xF0000 起），機器忽略該寫入
 *     （dosgolem internal/machine/bios_rom.go）；本版用 wr8_skip_rom／wr16s_rom 表達。平面視訊記憶體的寫入由探針丟棄。
 *
 * 輸入域：f 取艦隊表 ds:C074 的 32 筆（段為 DS）、位移接近 0xFFFF 的邊緣值與種子指定的位置；i 是 16 位元全域
 *   （高位元組應被忽略）；組數 0..255 由突變與種子決定。
 */
#include "gin3_reg.h"

/*
 * 段內存取的輔助。gin3_mem.h 的 rd16／wr16 吃線性位址，高位元組取「線性位址＋1」；真正的 CPU 位移各自回繞，
 * seg:FFFF 的 word 高位元組在 seg:0000（dosgolem internal/cpu/cpu.go read16／write16，8086 語意）。
 * 位移可能是 0xFFFF 時用 rd16s／wr16s_rom。寫入另外表達 ROM：0xF0000 以上寫入無效
 * （dosgolem internal/machine/bios_rom.go），用本檔的 wr8_skip_rom。這兩點 gin3_mem.h 都沒有（回報給主工作階段）。
 */
static uint16_t rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go，真機同）。gin3_mem.h 的 wr8 沒有這個語意（回報給主工作階段）。 */
static void wr8_skip_rom(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void wr16s_rom(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8_skip_rom(m, LIN(seg, off), (uint8_t)v);
    wr8_skip_rom(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_185C_020C(Mem *m, uint16_t f_off, uint16_t f_seg, uint16_t group)
{
    uint8_t i = (uint8_t)group;
    uint8_t j;
    uint16_t dst, src;

    wr8_skip_rom(m, LIN(f_seg, f_off + 0x0Fu), (uint8_t)(rd8(m, LIN(f_seg, f_off + 0x0Fu)) - 1u));
    dst = (uint16_t)(f_off + 0x10u + (uint16_t)(i * 4u));
    src = (uint16_t)(dst + 4u);
    j = i;
    while (j < rd8(m, LIN(f_seg, f_off + 0x0Fu))) {
        wr16s_rom(m, f_seg, dst, rd16s(m, f_seg, src));
        wr8_skip_rom(m, LIN(f_seg, dst + 2u), rd8(m, LIN(f_seg, src + 2u)));
        wr8_skip_rom(m, LIN(f_seg, dst + 3u), rd8(m, LIN(f_seg, src + 3u)));
        dst = (uint16_t)(dst + 4u);
        src = (uint16_t)(src + 4u);
        j++;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_185C_020C(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("185C_020C", 3, run)
