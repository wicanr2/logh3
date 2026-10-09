/*
 * GIN3PS.EXE 2AD2:018C  作廢一筆艦艇池記錄：艘數還給星域待命存量（艦種 7 不還）、移除它的顯示物件、把記錄的 [+1] 標成 0xFF
 *
 * 原版（99 bytes，far 呼叫，三個堆疊字；序言 push bp／si／di、sub sp,4、mov bp,sp，所以星域記錄 far 指標的位移在 [bp+0Eh]、段在 [bp+10h]，
 *   艦艇索引 idx 在 [bp+12h]）。區域變數 [bp+0]／[bp+2] ＝ 艦艇記錄的位移／段。
 *   回傳 AX 沒有被呼叫端使用（描述檔仍用 ret ＝ ax 逐位元比對）：艦艇 [+38h]／[+3Ah] 皆 0 時 AX ＝ 0，否則 AX ＝ 35D1:0102 的回傳（0 或 0xFFFF）。
 * 過程（已證實：反組譯逐指令；2703:155B、35D1:0102 各由其差異測試證實）：
 *   艦艇 ← (word ds:CFB4 ＋ low16(idx × 0x4E)) : word ds:CFB6        ; 池基底先讀（cx ＝ ds:CFB6、si ＝ ds:CFB4），之後 155B 即使改到 ds:CFB4 也不影響本函式
 *   byte 艦艇[+4] ≠ 7（比 byte，`cmp byte ptr es:[si+4],7`）→ 2703:155B(星域 far, 艦艇 far)     ; 艘數還回星域的待命存量
 *   ax ← word 艦艇[+3Ah]（段）；bx ← word 艦艇[+38h]（位移）                                    ; 在 155B 之後才讀，155B 若寫到這兩格會被看到
 *   (ax | bx) ≠ 0 → 35D1:0102(obj ＝ bx:ax)                                                 ; 把顯示物件從清單 ds:EAE0 移除
 *   byte 艦艇[+1] ← 0xFF                                                                      ; 在 0102 之後寫
 * 語意（強推論）：docs/internals/64-ship-pool.md 第 51 行「作廢：2AD2:018C(星域記錄, 艦艇索引) 退回存量（艦種 7 不退）」、
 *   docs/re/329-fleet-disband-is-garrison.md 第 56、57、69 行（艦隊解散時對每組呼叫，+0x38 是索敵分遣隊那一筆）、
 *   docs/re/234-overwrite-branch-receipt.md 第 39、45 行（018C 標成空槽之後 2AD2:0029 拿到同一格），與本檔一致。
 *   艦艇 [+0x38]／[+0x3A] 是顯示物件的遠指標（位移在前、段在後）：只由本函式把它交給 35D1:0102 的用法推得（假說）；
 *   docs/internals/33 第 121 行稱艦隊 +0x6C 的遠指標指到戰場圖示物件，是艦隊記錄的同型欄位，與艦艇池記錄的對應未驗證。
 * 副作用：155B（寫星域記錄內一個 dword）、0102（物件 +16..+27、+1E..+22，格表 ds:D81E 起，清單前移與 ds:EADE 減一）與艦艇 [+1]。
 *
 * 機器模型：C 版沒有堆疊（區域變數 [bp+0]／[bp+2] 即艦艇記錄指標，C 版用區域變數）。0102 經 2376:0354 呼叫的 01E6 把自己的區域變數放在堆疊帶內（見其檔頭）。
 * 輸入域：基底 fcmd-base.state；星域指標取星域表，艦艇索引取少數幾個值，這些艦艇的 [+38h]..[+3Bh] 列入 freeze（隨機突變會把它改成 ROM 區或堆疊帶的
 *   物件位址，而 2376:01E6 的 C 版不涵蓋 ROM 與蓋到堆疊帶），遠指標值由 seeds 給；池基底 ds:CFB4..CFB7 列入 freeze。
 *   idx × 0x4E 只留低 16 位元；艦艇指標與 [+38h] 的 word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（dosgolem 的 read16，用本檔的 rd16s）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern void f_2703_155B(Mem *m, uint16_t region_off, uint16_t region_seg, uint16_t ship_off, uint16_t ship_seg);
extern uint16_t f_35D1_0102(Mem *m, uint16_t obj_off, uint16_t obj_seg);

static uint16_t w018C_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_2AD2_018C(Mem *m, uint16_t region_off, uint16_t region_seg, uint16_t idx)
{
    uint16_t ship_seg = rd16(m, DS(0xCFB6u));
    uint16_t ship_off = (uint16_t)(rd16(m, DS(0xCFB4u)) + (uint16_t)(idx * 0x4Eu));
    uint16_t obj_seg, obj_off, ret;

    if (rd8(m, LIN(ship_seg, (uint16_t)(ship_off + 4u))) != 7)
        f_2703_155B(m, region_off, region_seg, ship_off, ship_seg);
    obj_seg = w018C_rd16s(m, ship_seg, (uint16_t)(ship_off + 0x3Au));
    obj_off = w018C_rd16s(m, ship_seg, (uint16_t)(ship_off + 0x38u));
    if ((uint16_t)(obj_seg | obj_off) != 0)
        ret = f_35D1_0102(m, obj_off, obj_seg);
    else
        ret = obj_seg;
    wr8(m, LIN(ship_seg, (uint16_t)(ship_off + 1u)), 0xFFu);
    return ret;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2AD2_018C(m, a[0], a[1], a[2]); }
GIN3_REGISTER("2AD2_018C", 3, run)
