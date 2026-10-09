/*
 * GIN3PS.EXE 185C:00B2  艦隊的第 0 組艦艇池記錄「現有艘數」是不是 0
 *
 * 原版：far 呼叫，一個參數＝艦隊記錄的 far 指標（位移在前、段在後），回 AX = 1 或 0，不寫記憶體。
 *   序言 push bp／si／di、sub sp,4，所以參數在 [bp+0Eh]（位移）、[bp+10h]（段）。
 *
 * 語意（已證實：docs/re/339 §3.1 "185C:00B2(f)：word 池[f[+0x10]][+0x08] == 0 → 1"；
 *   docs/internals/33 艦隊記錄 +0x10 起十組編制、每組 4 bytes 的第一個 word 是艦艇池索引；
 *   docs/internals/38 艦艇 +0x08 ＝ 現有艘數；本檔與反組譯逐條對過）：
 *   idx  ＝ word f[+0x10]                          ; 第 0 組的艦艇池索引
 *   ship ＝ ds:CFB4（位移）＋ idx × 0x4E 的低 16 位元 ; 艦艇池段在 ds:CFB6，每筆 78 bytes
 *   回 (word ship[+0x08] == 0) ? 1 : 0
 *   只看第 0 組，不看陣營、士氣與其他組（docs/re/339 同）。
 *
 * 原版的細節（本版照做）：
 *   - 池位移的乘法只留低 16 位元（mul cx 後只用 AX），相加也是 16 位元，超過 0xFFFF 回繞。
 *   - 艦隊記錄的位移 ＋0x10、池位移 ＋0x08 都是 16 位元回繞；word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0
 *     （rd16s），不是下一段的第一個 byte。
 *   - 序言裡的 `mov di,1` 是沒用到的局部，不影響結果。
 *
 * 輸入域：艦隊指標取 ds:C074 艦隊表 32 筆（段為 DS）與位移接近 0xFFFF 的邊緣值；艦艇池索引與池基底
 *   由突變導向讀取集改動，索引可到 0xFFFF（池位移回繞，讀池段之後 64 KB 內的記憶體）。
 */
#include "gin3_reg.h"

/*
 * gin3_mem.h 的 rd16 吃線性位址，高位元組取「線性位址＋1」；真正的 CPU 位移各自回繞，seg:FFFF 的 word
 * 高位元組在 seg:0000（dosgolem internal/cpu/cpu.go read16，8086 語意）。位移可能是 0xFFFF 時用這個
 * （gin3_mem.h 沒有，回報給主工作階段）。
 */
static uint16_t rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_185C_00B2(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint16_t idx = rd16s(m, fleet_seg, (uint16_t)(fleet_off + 0x10u));
    uint16_t pool_seg = rd16(m, DS(0xCFB6));
    uint16_t ship = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(idx * 0x4Eu));

    return rd16s(m, pool_seg, (uint16_t)(ship + 8u)) == 0 ? 1 : 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_185C_00B2(m, a[0], a[1]); }
GIN3_REGISTER("185C_00B2", 2, run)
