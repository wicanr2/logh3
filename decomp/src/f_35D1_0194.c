/*
 * 35D1:0194  戰場圖示物件清單中，鍵值等於參數的物件，把它的 [+0] 位元組設成 2（far）
 *
 * 原版：far 呼叫，一個參數 key（word，只讀低位元組 dl），無回傳值。
 *   堆疊：序言 push bp／si、sub sp,8、mov bp,sp，所以參數在 [bp+10h]。
 *   清單：筆數 ds:EADE（word），陣列 ds:EAE0 起每筆 4 bytes＝far 指標（位移 word、段 word）。
 *   for (si = 0; si < ds:EADE; si++)：p ＝ ds:EAE0 + si × 4（位移 16 位元回繞）；
 *       obj ＝ far 指標（ds:[p] 位移、ds:[p+2] 段）；
 *       若 obj[+2] == dl，obj[+0] ← 2（byte 寫入）。
 *   筆數 ds:EADE 每一圈重讀（cmp ax,ds:0EADEh），陣列項每一圈才讀；寫入若落在清單或筆數上
 *   （物件指標指到 ds:EADE、ds:EAE0 起），會影響後面的圈數與項目，本版照原版順序重讀。
 *   si 是 16 位元暫存器、無號比較，筆數 0xFFFF 時最多跑 0xFFFF 圈；p 的位移每圈 +4，繞過 0xFFFF 會回到 ds:0。
 * 相鄰函式：35D1:01DE 是同型函式，只有寫入值不同（2 與 1）。
 * 呼叫者：341B:0145（sub_40E65+0x6D，台帳分類 hw-dependent）。docs 沒有這支的專門記載（docs/internals/data/battle-object-keys.json 只把 35D1:0194、01DE
 *   的位元組標成「未審查」）。清單 ds:EADE／EAE0 的身分見 docs/re/376、docs/re/394（圖示物件清單）。
 * 語意：obj[+2] 當鍵值、obj[+0] 的 1／2 兩個值的語意都未知；本檔只證實讀寫位移、比較與寫入值
 *   （推論等級：位移與值已證實，欄位語意未知）。
 *
 * 寫入落在線性 0xF0000 以上（BIOS ROM 區）時，dosgolem 安靜丟棄；本版用 put8 表達。寫入或讀取落在
 * 0xA0000 到 0xBFFFF（視訊記憶體區）的案例由探針丟棄。
 *
 * 輸入域：key 的全部 16 位元值；清單與物件由基底狀態、突變與 seeds 構造（基底：筆數 0、陣列預填同一個指標
 *   0644:79A9）。步數超過 200,000（筆數約 > 11,000）的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

/* 線性 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go）。 */
static void f35D1_0194_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_35D1_0194(Mem *m, uint16_t key)
{
    uint8_t dl = (uint8_t)key;
    uint16_t si = 0;
    uint16_t p = 0xEAE0u;

    while (si < rd16(m, DS(0xEADEu))) {
        uint16_t seg = rd16(m, DS((uint16_t)(p + 2u)));
        uint16_t off = rd16(m, DS(p));

        if (rd8(m, LIN(seg, (uint16_t)(off + 2u))) == dl)
            f35D1_0194_put8(m, LIN(seg, off), 2);
        si++;
        p = (uint16_t)(p + 4u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_35D1_0194(m, a[0]);
    return 0;
}
GIN3_REGISTER("35D1_0194", 1, run)
