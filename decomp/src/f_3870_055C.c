/*
 * GIN3PS.EXE 3870:055C  畫封閉折線：物件 obj 的點陣列（每點 x、y 兩個 word）平移 (dx, dy) 後，依序連線，最後從最後一點連回第一點，每段用 3870:0833（異或線）畫
 *
 * 原版（184 bytes，far 呼叫，四個堆疊字；序言 push bp／si／di、sub sp,12h、mov bp,sp，所以 dx 在 [bp+1Ch]、dy 在 [bp+1Eh]、
 *   物件 far 指標的位移在 [bp+20h]、段在 [bp+22h]）。沒有有意義的回傳（描述檔 ret ＝ none）。
 *   區域變數：[bp+0] ＝ i、[bp+2]／[bp+4] ＝ 下一點 x'／y'、[bp+6]／[bp+8] ＝ 起點 x／y、[bp+0Ah]／[bp+0Ch] ＝ 目前點 x／y、[bp+0Eh]／[bp+10h] ＝ 點指標（位移／段）。
 * 過程（已證實：反組譯逐指令；3870:0833 由其差異測試證實）：
 *   P ← far 指標 obj[+0Ah]（位移）:obj[+0Ch]（段）；di ← dx
 *   cur ← (word P[0] ＋ dx, word P[2] ＋ dy)（16 位元）；start ← cur；P ← P ＋ 4（位移 16 位元回繞）；i ← 1
 *   重複：
 *     i >= word obj[+8]（無號 jnb，obj 每圈重讀）→ 離開
 *     next ← (word P[0] ＋ dx, word P[2] ＋ dy)；P ← P ＋ 4
 *     3870:0833(cur.x, cur.y, next.x, next.y, 色 word ← word obj[+5], 最後 word ← 0xFFFF)
 *     cur ← next；i++
 *   3870:0833(cur.x, cur.y, start.x, start.y, 色 word ← word obj[+5], 最後 word ← 0xFFFF)            ; 收尾：最後一點連回第一點
 *   所以點數 n ＝ 0 或 1 時沒有迴圈，只畫一段 (P0)-(P0) 的退化線；n >= 2 畫 n − 1 段加一段收尾，共 n 段。
 *   色 word 是 obj[+5]、obj[+6] 兩個 byte 組成（+6 是物件種類 5），0833 只用低位元組，所以 +6 不影響畫線。
 * 語意（強推論）：docs/re/135-manoeuvre-boundary.md 第 96 行「[+6] ＝ 5 那一支（3870:055C）是折線」、第 226 行「5 ｜ 3870:055C ｜ 折線 ｜ [+8] 點數、[+0A] 點陣列、[+5] 色」，與本檔一致；
 *   原文說「折線」，asm 顯示收尾會連回第一點，所以是封閉折線（多邊形外框）。
 * 呼叫的函式：3870:0833 內呼叫畫線 39D7:0936（IDA 沒切成函式，台帳 callees 看不到），差異測試用描述檔 stubs 取代，
 *   **畫線本身的像素行為不在驗證範圍**。
 * 副作用：只經由 0833（寫本函式堆疊上的端點副本與替身）；本函式不寫全域記憶體。
 *
 * 機器模型：C 版沒有堆疊。f_3870_0833 把端點放在原版的 SS:1F7E..1F85（見其檔頭），落在探針的堆疊帶內，寫入集合比對時雙方都排除。
 * 輸入域：dx、dy 全部 16 位元值；物件放在 ds:F400，點數 obj[+8] 與點陣列指標 obj[+0A..+0D] 列 freeze（隨機突變會改成上萬點或指到視訊區），由 seeds 給；
 *   物件與點陣列不指向堆疊帶與 A0000 到 BFFFF；0833 內 08FF 在原版不終止的案例由探針丟棄。
 *   點陣列的 word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（dosgolem 的 read16），用本檔的 rd16s。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern void f_3870_0833(Mem *m, uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1, uint16_t color_word, uint16_t last_word);

static uint16_t w055C_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

void f_3870_055C(Mem *m, uint16_t dx, uint16_t dy, uint16_t obj_off, uint16_t obj_seg)
{
    uint16_t p_seg = w055C_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x0Cu));
    uint16_t p_off = w055C_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x0Au));
    uint16_t cur_x = (uint16_t)(w055C_rd16s(m, p_seg, p_off) + dx);
    uint16_t cur_y = (uint16_t)(w055C_rd16s(m, p_seg, (uint16_t)(p_off + 2u)) + dy);
    uint16_t start_x = cur_x;
    uint16_t start_y = cur_y;
    uint16_t i;

    p_off = (uint16_t)(p_off + 4u);
    for (i = 1; i < w055C_rd16s(m, obj_seg, (uint16_t)(obj_off + 8u)); i++) {
        uint16_t nx = (uint16_t)(w055C_rd16s(m, p_seg, p_off) + dx);
        uint16_t ny = (uint16_t)(w055C_rd16s(m, p_seg, (uint16_t)(p_off + 2u)) + dy);

        p_off = (uint16_t)(p_off + 4u);
        f_3870_0833(m, cur_x, cur_y, nx, ny, w055C_rd16s(m, obj_seg, (uint16_t)(obj_off + 5u)), 0xFFFFu);
        cur_x = nx;
        cur_y = ny;
    }
    f_3870_0833(m, cur_x, cur_y, start_x, start_y, w055C_rd16s(m, obj_seg, (uint16_t)(obj_off + 5u)), 0xFFFFu);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_3870_055C(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("3870_055C", 4, run)
