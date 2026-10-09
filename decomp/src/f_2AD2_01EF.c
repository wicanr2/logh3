/*
 * GIN3PS.EXE 2AD2:01EF  作廢一筆記錄：若 +0x38 的 far 指標非 0，就從顯示物件清單移除該物件並把指標清 0；無論如何 +0x01 ← 0xFF
 *
 * 原版（72 bytes，far 呼叫，一個 far 指標參數 p；序言 push bp／si、sub sp,6、mov bp,sp，所以 p 的位移在 [bp+0Eh]、段在 [bp+10h]），無有意義的回傳值
 *   （描述檔 ret ＝ none；AX 是 35D1:0102 的回傳或 p 的段殘值，呼叫者不用）。
 * 過程（已證實：反組譯逐指令；35D1:0102 由其差異測試證實）：
 *   q ← p ＋ 0x38；(seg ← word q[+2]，off ← word q[0])
 *   (seg | off) ≠ 0（兩個 word 都看）→
 *     35D1:0102(off, seg)                                   ; far 指標，位移在前（push seg、push off），呼叫端 mov sp,bp 清參數
 *     word q[0] ← 0；word q[+2] ← 0                         ; q 是先前存在 [bp+2]／[bp+4] 的值（p ＋ 0x38），在呼叫之後才寫
 *   byte p[+1] ← 0xFF                                        ; 再 les 一次 p，無條件
 * 語意（強推論）：docs/internals/64-ship-pool.md 第 52 行「`2AD2:01EF(艦艇)`：`+0x38` 非 0 就先釋放、`+0x01 ← FF`」、
 *   docs/internals/62-strategic-attrition.md 第 31 行、docs/re/329-fleet-disband-is-garrison.md 第 108 行（艦艇遠指標作廢，沒有 2703:155B）、
 *   docs/spec/060 第 20 行；與本檔一致：本函式只從清單移除並清指標，不釋放記憶體（35D1:0102 本身不釋放）。
 *   p ＝ 艦艇池記錄（78 bytes，+0x38 ＝ 戰場圖示物件的 far 指標）是呼叫者 0374:073F／0374:0896／185C:01ED 的用法；本函式只看 +0x01 與 +0x38。
 * 副作用：35D1:0102 的全部寫入（物件 +16..+27、+1E..+22，格表 ds:D81E 起，清單前移與 ds:EADE 減一），q 的 4 bytes 清 0，p[+1]。
 *   寫入線性 0xF0000 以上（ROM 區）在 dosgolem 被安靜丟棄，本檔按 byte 模型化（p 可能在 ROM 區）。
 *
 * 機器模型：C 版沒有堆疊；原版把 q 與 ES 存在堆疊上，沒有別的堆疊資料。35D1:0102 內部的堆疊區域變數在探針的堆疊帶內（比對時排除）。
 *   word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 的 read16／write16），用本檔的 w01EF_rd16s／w01EF_wr16s。
 * 輸入域：p 指到暫存區、DGROUP 或 ROM 區；p＋0x38 的 4 bytes（物件指標）由描述檔 seeds 給定並列入 freeze，因為指標被隨機突變會指向程式碼、堆疊帶或
 *   視訊記憶體區，那些案例原版與 C 版不可比；物件本身（35D1:0102／2376:0354／2376:0364 的輸入域）與清單取基底與突變。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_35D1_0102(Mem *m, uint16_t obj_off, uint16_t obj_seg);

static uint16_t w01EF_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效 */
static void w01EF_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w01EF_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w01EF_put8(m, LIN(seg, off), (uint8_t)v);
    w01EF_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_2AD2_01EF(Mem *m, uint16_t p_off, uint16_t p_seg)
{
    uint16_t q = (uint16_t)(p_off + 0x38u);
    uint16_t seg = w01EF_rd16s(m, p_seg, (uint16_t)(q + 2u));
    uint16_t off = w01EF_rd16s(m, p_seg, q);

    if ((uint16_t)(seg | off) != 0) {
        (void)f_35D1_0102(m, off, seg);
        w01EF_wr16s(m, p_seg, q, 0);
        w01EF_wr16s(m, p_seg, (uint16_t)(q + 2u), 0);
    }
    w01EF_put8(m, LIN(p_seg, (uint16_t)(p_off + 1u)), 0xFFu);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2AD2_01EF(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2AD2_01EF", 2, run)
