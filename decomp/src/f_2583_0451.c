/*
 * GIN3PS.EXE 2583:0451  戰場單位移動的避障搜尋：沿「朝目標的方位角 ± 16 的倍數」依序試走一步 dist，第一個能站的落點寫回兩個 word 指標；全部不能站就還原成原位
 *
 * 原版（219 bytes，far 呼叫，八個堆疊字；序言 push bp／si／di、sub sp,4、mov bp,sp，所以 x 在 [bp+0Eh]、y 在 [bp+10h]、
 *   far 指標 A（x 座標，位移 [bp+12h]、段 [bp+14h]）、far 指標 B（y 座標，位移 [bp+16h]、段 [bp+18h]）、dist 在 [bp+1Ah]、faction 在 [bp+1Ch]）。
 *   區域變數：[bp+0] ＝ 基準角 base（byte）、[bp+1] ＝ 試走角（byte，只存不讀）、[bp+2] ＝ i（byte）。
 *   回傳 AX：有落點時 0（最後一次 2583:040C 的回傳），全部不能站時 y（`mov ax,di`）。呼叫者 1C03:0185 等只看寫回的指標，不用回傳值（描述檔 ret ＝ ax 逐位元比對）。
 * 過程（已證實：反組譯逐指令；2BE1:011D、2BE1:026F、2583:040C 各由其差異測試證實）：
 *   base ← 2BE1:011D(a ＝ word *A − x, b ＝ word *B − y)                       ; 目標 (*A, *B) 相對 (x, y) 的方位角，位移是 16 位元相減
 *   i ← 0；當 i <= 0x80（byte，無號 jbe）：
 *     2BE1:026F(x, y, 角 ＝ base ＋ i（8 位元）, dist, A, B)                      ; 從 (x, y) 沿該角走 dist，夾進戰場範圍，結果寫進 *A、*B
 *     2583:040C(*A, *B, faction) ＝ 0 → 回 0                                      ; 能站（地形通且沒人佔）
 *     2BE1:026F(x, y, 角 ＝ base − i（8 位元）, dist, A, B)
 *     2583:040C(*A, *B, faction) ＝ 0 → 回 0
 *     i ← i ＋ 0x10
 *   *A ← x；*B ← y（先 A 後 B）；回 y                                           ; 沒有能站的落點，還原
 *   i ＝ 0 時 ＋ 與 − 試的是同一個角，所以同一個落點會問兩次（docs/re/372 第 69 行「包含重複 0／128 候選」）；i ＝ 0x80 時 ＋ 與 − 也是同一個角。
 *   026F 的 x、y 參數每次都是進入時的 x、y（不是上一次寫進指標的值）；040C 的座標則是剛寫進 *A、*B 的值（指標別名時兩者互相蓋寫）。
 * 語意（強推論）：docs/re/372-battle-step-and-collision.md 第 69 行「2583:0451 依候選方位 ± 16 交替搜尋，包含重複 0／128 候選；全部受阻恢復舊座標」、
 *   docs/spec/081-battle-fleet-step.md 第 47 行（遇地形 FFFF 才呼叫 2583:0451）與本檔一致。
 *   docs/re/129-fleet-budget-exact.md 第 25 行、docs/re/294-short-move-strafe-and-arrival-facing.md 第 84 行把參數寫成 `2583:0451(&+0x5C, &+0x5E, 8, 陣營)` 四個，
 *   實際是八個堆疊字：x、y 的值在前（呼叫端 1C03:0185 推入 ＝ 艦隊 [+5C]、[+5E] 的值，同時 A、B 就是這兩格的位址），所以那次呼叫的 a ＝ b ＝ 0，
 *   base 是 011D(0, 0) 的結果（A ＝ 0 時比值取 0xFFFF 再查表）。
 * 副作用：026F 對 *A、*B 的寫入（每次試走都寫，不論能不能站）與全部不能站時的還原寫入。
 *
 * 輸入域：基底 fcmd-standby.state（040C 內 1588:04EA 需要戰場地圖）；x、y、dist、faction 全部 16 位元值（dist 與 faction 以 word 傳給 026F、040C）；
 *   指標 A、B 指到 DGROUP 或暫存區（描述檔放在 ds:F400／F410）。寫入落在線性 0xF0000 以上（ROM 區）時被 dosgolem 安靜丟棄，
 *   還原寫入在本檔按 byte 模型化（026F 自己也模型化）；word 存取在位移 0xFFFF 時 dosgolem 的高位元組取同一段的位移 0（rd16s），
 *   但 026F 的寫入按線性位址連續處理（見 f_2BE1_026F.c 檔頭，輸出指標位移 0xFFFF 的輸入不在比較範圍），所以描述檔與 seeds 的指標位移皆遠離 0xFFFF。
 *   指標蓋到呼叫堆疊（探針堆疊帶）的情形不涵蓋；池基底 ds:CFB4..CFB7 列入 freeze（隨機突變會改到視訊區）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_026F(Mem *m, uint16_t x, uint16_t y, uint16_t angle_word, uint16_t dist, uint16_t ox_off, uint16_t ox_seg, uint16_t oy_off, uint16_t oy_seg);
extern uint16_t f_2583_040C(Mem *m, uint16_t x, uint16_t y, uint16_t faction);

/* word 讀取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 的 read16） */
static uint16_t w0451_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 還原寫入：線性 0xF0000 以上（ROM 區）的 byte 被 dosgolem 安靜丟棄 */
static void w0451_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w0451_wr16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w0451_put8(m, LIN(seg, off), (uint8_t)v);
    w0451_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_2583_0451(Mem *m, uint16_t x, uint16_t y, uint16_t a_off, uint16_t a_seg, uint16_t b_off, uint16_t b_seg,
                     uint16_t dist, uint16_t faction)
{
    uint16_t diff_y = (uint16_t)(w0451_rd16s(m, b_seg, b_off) - y);
    uint16_t diff_x = (uint16_t)(w0451_rd16s(m, a_seg, a_off) - x);
    uint8_t base;
    uint8_t i;

    /* 原版先算 B 的差（推入 b），再算 A 的差（推入 a）；兩者只讀不寫，順序無關 */
    base = f_2BE1_011D(m, diff_x, diff_y);
    for (i = 0; i <= 0x80u; i = (uint8_t)(i + 0x10u)) {
        (void)f_2BE1_026F(m, x, y, (uint8_t)(base + i), dist, a_off, a_seg, b_off, b_seg);
        if (f_2583_040C(m, w0451_rd16s(m, a_seg, a_off), w0451_rd16s(m, b_seg, b_off), faction) == 0)
            return 0;
        (void)f_2BE1_026F(m, x, y, (uint8_t)(base - i), dist, a_off, a_seg, b_off, b_seg);
        if (f_2583_040C(m, w0451_rd16s(m, a_seg, a_off), w0451_rd16s(m, b_seg, b_off), faction) == 0)
            return 0;
    }
    w0451_wr16(m, a_seg, a_off, x);
    w0451_wr16(m, b_seg, b_off, y);
    return y;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2583_0451(m, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7]); }
GIN3_REGISTER("2583_0451", 8, run)
