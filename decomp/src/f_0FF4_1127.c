/*
 * GIN3PS.EXE 0FF4:1127  艦隊編成視窗「中止」：把編成時從艦艇池拿出去的艦艇記錄作廢（退回存量），再把進來時備份的艦艇記錄與星域存量寫回原處
 *
 * 原版（205 bytes，far 呼叫，一個 far 指標參數「艦隊」；序言 push bp／si／di、sub sp,0Ah、mov bp,sp，所以艦隊指標的位移在 [bp+14h]、段在 [bp+16h]）。
 *   無有意義的回傳值（描述檔 ret ＝ none）。區域變數：[bp+0] ＝ 迴圈計數 i（byte）、[bp+2]／[bp+4] ＝ 讀指標的位移／段、[bp+6]／[bp+8] ＝ 備份來源的位移／段。
 * 過程（已證實：反組譯逐指令；2AD2:018C、3CA9:0002 各由其差異測試證實）：
 *   迴圈一：[bp+2]:[bp+4] ← ds:981E（位移）:DS（段）；對 i ＝ 0..byte ds:984B−1（無號 byte 比較，jnb 結束）：
 *     2AD2:018C(星域 ＝ ds:9872 : ds:9874, 艦艇索引 ＝ word ds:981E)                           ; 每圈重讀 ds:9872／9874
 *     **讀指標 [bp+2] 在迴圈裡沒有被推進**：每圈都傳 ds:981E 的第 0 個 word，不是 ds:981E[i]（舊文件 docs/re/317 第 62 行、docs/re/344 第 289 行寫成 [i]；
 *     asm 在 loc_1DBEB 到 jmp 之間沒有任何 [bp+2] 的寫入）。
 *   迴圈二：[bp+2]:[bp+4] ← (艦隊 + 0x10) : 艦隊段（位移 16 位元回繞）；[bp+6]:[bp+8] ← ds:987A : ds:987C（備份來源）；
 *     對 i ＝ 0..byte 艦隊[+0x0F]−1（無號 byte 比較，每圈重讀艦隊指標）：
 *       idx ← word [bp+2]:[bp+4]                                                          ; 艦隊 +0x10 起每 4 bytes 一筆，只取第一個 word
 *       3CA9:0002(dst ＝ (ds:CFB4 ＋ low16(idx × 0x4E)) : ds:CFB6, src ＝ [bp+6]:[bp+8], 0x4E)   ; 艦艇記錄 78 bytes 寫回池
 *       [bp+6] ← [bp+6] ＋ 0x4E；[bp+2] ← [bp+2] ＋ 4
 *   收尾：3CA9:0002(dst ＝ (ds:9872 ＋ 0x30 ＋ low16(byte ds:8F1D × 0x1C)) : ds:9874, src ＝ ds:987E : ds:9880, 0x1C)    ; 星域待命存量 28 bytes 寫回
 * 呼叫的方式（逐一對過 asm）：
 *   018C 三個堆疊字（由低位址到高位址）：星域位移 ＝ ds:9872、星域段 ＝ ds:9874、艦艇索引 ＝ word ds:981E。
 *   3CA9:0002 五個堆疊字：dst 位移、dst 段、src 位移、src 段、count。迴圈二的 dst 位移是 `di`（ds:CFB4 ＋ idx×0x4E）、dst 段是 `dx`（ds:CFB6）。
 *   mul 之後 DX 被 ds:CFB6 覆蓋，所以 idx × 0x4E 只留低 16 位元；收尾的 byte × 0x1C 同理。
 * 語意（強推論）：docs/re/317-fleet-composition-editor.md 第 19 行「『中止』（0FF4:1127）把進來時備份的存量……」、docs/re/316-fleet-change-menu.md 第 79 行、
 *   docs/re/344-ship-pool-fields.md 第 289 行；ds:981E 的欄位名與備份緩衝區（ds:987A、ds:987E）的內容是假說（只由用法推得）。
 * 副作用：018C 及其呼叫閉包（155B、35D1:0102、2376:01E6 等）、memmove 寫的艦艇池與星域記錄。
 *
 * 機器模型：C 版沒有堆疊（區域變數用 C 變數）。018C 的 C 版與其閉包（2376:01E6）把自己的區域變數放在探針堆疊帶內的寫死位置（見其檔頭），
 *   原版的框架較深，實際位置不同；兩處都在堆疊帶（SS:1B8A..1F89）內，比對時排除，C 版自寫自讀，不影響結果。
 * 輸入域：基底 fcmd-base.state（ds:984B ＝ 1、ds:981E 第 0 個 word ＝ 0、ds:9872 指到星域 0、艦隊表 32 筆）。為了避免 018C／memmove 的 C 版不涵蓋的情形
 *   （2376:01E6 不涵蓋 ROM 與蓋到堆疊帶；3CA9:0002 的 C 版用 wr8，沒有 ROM 區語意），描述檔把 ds:981E 第 0 個 word（艦艇索引）、ds:9872／9874（星域指標）、
 *   ds:CFB4..CFB7（池基底）與基底艦艇 0..3、857 的 [+0x38..+0x3B] 列入 freeze，這些值只由 seeds 改；其餘（艦隊內的索引、備份來源指標、ds:984B、艦隊 +0x0F、ds:8F1D）可被突變。
 *   艦隊指標的 word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（rd16s）。dst 位移 16 位元回繞，落在池段 ds:CFB6 內。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_2AD2_018C(Mem *m, uint16_t region_off, uint16_t region_seg, uint16_t idx);
extern uint32_t f_3CA9_0002(Mem *m, uint16_t doff, uint16_t dseg, uint16_t soff, uint16_t sseg, uint16_t count);

/* seg:FFFF 的 word 高位元組在 seg:0000（8086 位移各自回繞）。 */
static uint16_t w1127_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

void f_0FF4_1127(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint8_t i;
    uint16_t rd_off, src_off, src_seg;

    for (i = 0; i < rd8(m, DS(0x984Bu)); i++)
        (void)f_2AD2_018C(m, rd16(m, DS(0x9872u)), rd16(m, DS(0x9874u)), rd16(m, DS(0x981Eu)));

    rd_off = (uint16_t)(fleet_off + 0x10u);
    src_off = rd16(m, DS(0x987Au));
    src_seg = rd16(m, DS(0x987Cu));
    for (i = 0; i < rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x0Fu))); i++) {
        uint16_t idx = w1127_rd16s(m, fleet_seg, rd_off);
        uint16_t pool_seg = rd16(m, DS(0xCFB6u));
        uint16_t pool_off = (uint16_t)(rd16(m, DS(0xCFB4u)) + (uint16_t)(idx * 0x4Eu));

        (void)f_3CA9_0002(m, pool_off, pool_seg, src_off, src_seg, 0x4E);
        src_off = (uint16_t)(src_off + 0x4Eu);
        rd_off = (uint16_t)(rd_off + 4u);
    }

    (void)f_3CA9_0002(m, (uint16_t)(rd16(m, DS(0x9872u)) + 0x30u + (uint16_t)(rd8(m, DS(0x8F1Du)) * 0x1Cu)),
                      rd16(m, DS(0x9874u)), rd16(m, DS(0x987Eu)), rd16(m, DS(0x9880u)), 0x1C);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_0FF4_1127(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("0FF4_1127", 2, run)
