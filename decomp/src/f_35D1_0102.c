/*
 * GIN3PS.EXE 35D1:0102  從顯示物件清單 ds:EAE0 移除一個物件：先重算它的矩形並標記畫面格表，再在清單找到後把後面的項目前移一格
 *
 * 原版（146 bytes，far 呼叫，一個 far 指標參數 obj；序言 push bp／si／di、mov bp,sp，沒有 sub sp，所以 obj 的位移在 [bp+0Ah]、段在 [bp+0Ch]），
 *   回 AX：0 ＝ 已移除，0xFFFF ＝ 清單裡沒有。
 * 過程（已證實：反組譯逐指令；2376:0354、2376:0364 由各自的差異測試證實）：
 *   di ← 0
 *   2376:0354(obj)                                  ; 新舊矩形更新（01E6 的包裝；它把 obj[+1E] 抄進 [+23]，重算 [+1E]）
 *   obj[+0x23] ＝ 0（byte，呼叫 0354 之後才讀）→ 2376:0364(obj)      ; 以舊矩形標記格表
 *   搜尋：重複
 *     entry ← ds:(0xEAE0 ＋ 4 × di)：位移 word 在 +0、段 word 在 +2（`di << 2` 是 16 位元運算，位移 16 位元回繞）
 *     entry ＝ obj（段與位移都相等）→ 離開
 *     di >= word ds:EADE（無號 jnb，筆數每圈重讀）→ 離開
 *     di++
 *   di >= word ds:EADE → 回 0xFFFF                   ; 所以比對在檢查筆數之前：第 EADE 筆（清單外的下一格）若等於 obj 也是回 0xFFFF
 *   壓縮：重複 (word ds:EADE ＋ 0xFFFF) > di（無號，筆數每圈重讀，jbe 離開）時：
 *     entry[di] ← entry[di ＋ 1]（先讀段與位移再寫，各一個 word；`di ＋ 1` 與 `<< 2` 都是 16 位元運算）；di++
 *   word ds:EADE 減一（讀改寫）；回 0
 * 語意（強推論）：docs/internals/33-fleet-table.md 第 121 行（艦隊 +0x6C 的遠指標指到戰場圖示物件）、docs/internals/62-strategic-attrition.md 第 32 行、
 *   docs/re/329-fleet-disband-is-garrison.md 第 71、93、124 行（艦隊解散時「+0x6C 遠指標不為 0 → 35D1:0102 釋放」、指標本身不清 0）、
 *   docs/re/336 第 177 行；與本檔一致：本函式只從清單移除，不釋放記憶體。清單格式（位移 word、段 word）與筆數 ds:EADE 見 f_2376_0164.c 檔頭。
 * 副作用：全部來自 0354、0364（物件 +16..+27、+1E..+22，格表 ds:D81E 起），清單前移的寫入與 ds:EADE 減一。
 *
 * 機器模型：0354 包的 01E6 的區域變數在探針的堆疊帶內（比對時排除）；本函式沒有自己的堆疊資料。
 * 輸入域：obj 指到暫存區、DGROUP 或 seeds 給的位址，不指向堆疊帶與 A0000 到 BFFFF；清單與筆數 ds:EADE 取基底（空）與 seeds。
 *   清單項目 EAE0 起的位移隨 di 增長在 16 位元內回繞；筆數大到超過探針步數上限的案例由探針丟棄，C 版不涵蓋。word 存取在位移 0xFFFF 時
 *   高位元組取同一段的位移 0（dosgolem 的 read16），用本檔的 w0102_rd16s／w0102_wr16s。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_2376_0354(Mem *m, uint16_t obj_off, uint16_t obj_seg);
extern void f_2376_0364(Mem *m, uint16_t obj_off, uint16_t obj_seg);

static uint16_t w0102_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w0102_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_35D1_0102(Mem *m, uint16_t obj_off, uint16_t obj_seg)
{
    uint16_t di = 0;

    (void)f_2376_0354(m, obj_off, obj_seg);
    if (rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 0x23u))) == 0)
        f_2376_0364(m, obj_off, obj_seg);
    for (;;) {
        uint16_t b = (uint16_t)(di << 2);

        if (w0102_rd16s(m, GIN3_DS, (uint16_t)(b + 0xEAE2u)) == obj_seg
            && w0102_rd16s(m, GIN3_DS, (uint16_t)(b + 0xEAE0u)) == obj_off)
            break;
        if (di >= w0102_rd16s(m, GIN3_DS, 0xEADEu))
            break;
        di = (uint16_t)(di + 1u);
    }
    if (di >= w0102_rd16s(m, GIN3_DS, 0xEADEu))
        return 0xFFFFu;
    while ((uint16_t)(w0102_rd16s(m, GIN3_DS, 0xEADEu) - 1u) > di) {
        uint16_t b1 = (uint16_t)((uint16_t)(di + 1u) << 2);
        uint16_t b0 = (uint16_t)(di << 2);
        uint16_t seg1 = w0102_rd16s(m, GIN3_DS, (uint16_t)(b1 + 0xEAE2u));
        uint16_t off1 = w0102_rd16s(m, GIN3_DS, (uint16_t)(b1 + 0xEAE0u));

        w0102_wr16s(m, GIN3_DS, (uint16_t)(b0 + 0xEAE2u), seg1);
        w0102_wr16s(m, GIN3_DS, (uint16_t)(b0 + 0xEAE0u), off1);
        di = (uint16_t)(di + 1u);
    }
    w0102_wr16s(m, GIN3_DS, 0xEADEu, (uint16_t)(w0102_rd16s(m, GIN3_DS, 0xEADEu) - 1u));
    return 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_35D1_0102(m, a[0], a[1]); }
GIN3_REGISTER("35D1_0102", 2, run)
