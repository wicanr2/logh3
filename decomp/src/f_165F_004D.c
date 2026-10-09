/*
 * GIN3PS.EXE 165F:004D  艦隊層「移動／戰鬥」選點的外形：用回轉速度與巡航距離積出一圈點表（ds:9C1A），鏡像、依艦隊朝向旋轉
 *
 * 原版（770 bytes，far 呼叫，四個堆疊參數，沒有回傳值（描述檔 ret ＝ none）；push bp／si／di、sub sp,34h，所以
 *   out 在 [bp+3Eh]／[bp+40h]、艦隊在 [bp+42h]／[bp+44h]、A 在 [bp+46h]、B 在 [bp+48h]）：
 *   呼叫者 2583:0003（docs/re/120-battle-move-range.md「165F:004D(→[bp+18], 艦隊, di, 速度A)」：di ＝ 速度B × 8、速度A ＝ 艦隊 +0x4D 回轉；asm 的推入順序 `push [bp+3Ch]` 先、`push di` 後，所以 [bp+46h] ＝ di、[bp+48h] ＝ 速度A）。
 *   所以 args[0..1] ＝ out（描述結構，far 指標，位移在前）、args[2..3] ＝ 艦隊、args[4] ＝ A（本檔的名稱：[bp+46h] ＝ di ＝ 速度B × 8，與 Q 比較的上限）、args[5] ＝ B（[bp+48h] ＝ 速度A ＝ 回轉速度，除數）。
 *   以下只寫反組譯逐指令對出來的運算；變數名稱是本檔自取。
 *
 * 1. 描述結構（out，word 存取按段內位移各自回繞）：
 *      out[+0] ← 艦隊[+0x5C] ＋ 8 ；out[+2] ← 艦隊[+0x5E] ＋ 8 ；out[+4] ← 0 ；out[+5] ← 1 ；out[+6] ← 5 ；out[+7] ← 1 ；
 *      out[+0x0A] ← 0x9C1A ；out[+0x0C] ← DS（0x54B3）。寫完 out[+0] 才讀艦隊[+0x5E]（out 與艦隊別名時看得到前一筆寫入）。
 *      n ← 0（點數）、i ← 0（word）、結束旗標 ← 0。
 * 2. 取樣迴圈（直到結束旗標）：
 *      V ← (有號 16 位元 (i << 1)) 符號延伸成 32 位元後 << 3；Q ← V ÷ B（3CC9:0070，有號 32 位元，B 零延伸成 32 位元）
 *      Q > A（3CC9:00B9 比 Q 與 0:A，`jle` 不成立）：x ← 0、y ← 0、結束旗標 ← 1
 *      否則：
 *        acc1 ← 0、acc2 ← 0、k ← 0、mm ← i                                  ; 全是 32 位元／word 區域變數
 *        do { acc1 += T[(k + 0x40) & 0xFF] << 3；acc2 += T[k & 0xFF] << 3     ; T ＝ word 表 ds:573A（符號延伸成 32 位元再 << 3）
 *             k += B；mm −= B } while (mm >= B)                              ; 16 位元、無號比較（jnb）；do-while，先做一次
 *        D ← ((有號 16 位元)mm << 3 的 32 位元) ÷ B（有號）
 *        T1 ← T[(i + 0x40) & 0xFF]、T2 ← T[i & 0xFF]（符號延伸）
 *        acc1 += D × T1；acc2 += D × T2                                       ; 3CC9:0009，低 32 位元
 *        R ← (0:A) − (V ÷ B)                                                  ; 32 位元借位回繞；V ÷ B 再算一次（有號）
 *        x ← acc1 + R × T1；y ← acc2 + R × T2
 *      y < 0（3CC9:00B9 比 y 與 0，SF）→ 結束旗標 ← 1（不加點）
 *      否則：ds:(0x9C1A ＋ 4n) ← (x >> 10)（算術右移 10，取低 word）；ds:(0x9C1A ＋ 4n ＋ 2) ← (y >> 10)（同）；i += 8；n++
 *    （Q > A 的那一圈：x ＝ y ＝ 0 不是負數，所以仍加一個 (0, 0) 點，之後因結束旗標離開。）
 * 3. 鏡像：for j ＝ 0 .. n−1（有號 word 比較 jge）：點[2n−1−j].x ← 點[j].x；點[2n−1−j].y ← −點[j].y（16 位元取負）
 *      點位址 ＝ ds:(0x9C1A ＋ 4 × 序號)，16 位元回繞；先寫 x、再讀 y。
 * 4. 旋轉：for j ＝ 0 .. 2n−1（有號 word 比較 `2n <= j` 結束）：3A92:0001(&點[j].x, &點[j].y, 艦隊[+0x4A] 的 word)；
 *      每圈重讀艦隊[+0x4A]；3A92:0001 只取角度的低位元組。
 * 5. out[+8] ← 2n（word）。
 * 副作用：out 的 8 個欄位（共 +0..+7、+0x0A、+0x0C 與 +8）、ds:9C1A 起 4 × 2n bytes 的點表；區域變數在堆疊（帶內，比對時排除）。
 *
 * 語意（假說）：docs/re/120 第 71 行「165F:004D 用 ds:9C1A 那張點表生出來的多邊形」、docs/re/121 第 46 行「迴圈用 ds:573A 那張正弦表
 *   （一整圈 256、值 ＝ sin × 1024）」、docs/spec/027 第 57 行「外形是照回轉速率積出來的可達邊界（165F:004D）」。x、y 是相對艦隊的點，
 *   旋轉後以 out[+0]、out[+2] 為圓心；2583:0003 之後交給繪圖。本檔不判斷幾何意義，只照反組譯寫運算。i 每圈加 8、迴圈量為 16 · i ÷ B。
 *   A 是 2583:0003 的 di（速度B × 8，docs/re/120 的例子 30 × 8 ＝ 240），B 是回轉速度（艦隊 +0x4D）；Q ＝ 16 · i ÷ B 與 A 的比較，語意為假說（未驗證）。
 * 呼叫的函式：3CC9:0009／0070／00B9（runtime，已 diff-tested，這裡用 gin3_lmul、gin3_ldivmod 與 f_3CC9_00B9 的旗標）、3A92:0001（已 diff-tested）。
 * 注意：內層 do-while 在 i < B 時 mm ＝ i − B 會 16 位元回繞成接近 65536，迴圈要跑約 65536 ÷ B 圈才離開（原版如此）。
 *   所以一次呼叫需要約 30 萬步（與 B 無關：(B ÷ 8) 圈 × (65536 ÷ B) 圈 × 約 37 條指令）；描述檔的 step_budget 調高。B ＝ 0 時 mm 不變、
 *   迴圈不終止，不在輸入域內。
 * 輸入域：out 在暫存區（不涵蓋 ROM 區）；艦隊取艦隊表或種子指定的位置；A、B 為 16 位元 word（B 零除外：迴圈不終止）；點表長度由步數上限限制；
 *   讀到 A0000 到 BFFFF 的案例由探針丟棄；點表（ds:9C1A 起）不蓋到本函式讀的表（ds:573A）；點表長到蓋過艦隊表時，後面的讀取看到的是
 *   前面寫入的值（本版逐步照做）。i 超過 0x4000（點數超過 2048）時 (i << 1) 的符號延伸才與零延伸不同，記錄下來的案例沒有走到
 *   （A < 8 時 Q > A 在 i < 0x4000 就結束；A >= 8 只靠 y < 0 結束，記錄案例的點數都在幾百以內），所以「V 零延伸」不是可被分辨的突變，
 *   屬輸入域邊緣（未驗證）。同理 mm >= 0x8000（B > 32768 且 i >= 232）由種子以全為 +1000 的表造出（D 夠大才在右移 10 位元後看得出差別），涵蓋；
 *   (x >> 10) 取低 16 位元時，算術與邏輯右移的差別只在被丟掉的高 10 位元，所以兩者等價（x、y 都是）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；幾何語意為假說。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint16_t f_3A92_0001(Mem *m, uint16_t x_off, uint16_t x_seg, uint16_t y_off, uint16_t y_seg, uint16_t angle_word);

#define F_SF 0x0080u
#define F_OF 0x0800u
#define F_ZF 0x0040u

/* 點數的安全上限。原版每個點約 200 到 400 步（含鏡像與旋轉共約 500 步），200 萬步的上限下最多約 4000 點，所以記錄下來的案例都遠低於此；
 * 這個上限只讓「結束條件被改壞」的突變版停得下來，不影響輸入域內的行為。 */
#define W004D_MAX_POINTS 16384u

static uint16_t w004D_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w004D_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

/* 符號延伸成 32 位元後左移 3 位（cwd；shl／rcl ×3） */
static uint32_t w004D_sx3(uint16_t w) { return (uint32_t)(int32_t)(int16_t)w << 3; }

/* word 表 ds:573A 的符號延伸值 */
static uint32_t w004D_tab(const Mem *m, uint16_t idx) { return (uint32_t)(int32_t)(int16_t)rd16(m, DS((uint16_t)(0x573Au + 2u * (idx & 0xFFu)))); }

void f_165F_004D(Mem *m, uint16_t out_off, uint16_t out_seg, uint16_t fl_off, uint16_t fl_seg, uint16_t A, uint16_t B)
{
    uint16_t n = 0;
    uint16_t i = 0;
    uint16_t j;
    int done = 0;

    w004D_wr16s(m, out_seg, out_off, (uint16_t)(w004D_rd16s(m, fl_seg, (uint16_t)(fl_off + 0x5Cu)) + 8u));
    w004D_wr16s(m, out_seg, (uint16_t)(out_off + 2u), (uint16_t)(w004D_rd16s(m, fl_seg, (uint16_t)(fl_off + 0x5Eu)) + 8u));
    wr8(m, LIN(out_seg, (uint16_t)(out_off + 4u)), 0);
    wr8(m, LIN(out_seg, (uint16_t)(out_off + 5u)), 1);
    wr8(m, LIN(out_seg, (uint16_t)(out_off + 6u)), 5);
    wr8(m, LIN(out_seg, (uint16_t)(out_off + 7u)), 1);
    w004D_wr16s(m, out_seg, (uint16_t)(out_off + 0x0Au), 0x9C1Au);
    w004D_wr16s(m, out_seg, (uint16_t)(out_off + 0x0Cu), GIN3_DS);

    while (!done) {
        uint32_t V = w004D_sx3((uint16_t)(i << 1));
        uint32_t Q, rem, x, y;
        uint16_t fl;

        gin3_ldivmod(V, (uint32_t)B, &Q, &rem);
        fl = f_3CC9_00B9((uint16_t)(Q >> 16), (uint16_t)Q, 0, A);
        if (!(((fl & F_ZF) != 0) || (((fl & F_SF) != 0) != ((fl & F_OF) != 0)))) {
            x = 0;
            y = 0;
            done = 1;
        } else {
            uint32_t acc1 = 0, acc2 = 0, M, D, Q2, R;
            uint32_t T1, T2;
            uint16_t k = 0, mm = i;

            do {
                acc1 += w004D_tab(m, (uint16_t)(k + 0x40u)) << 3;
                acc2 += w004D_tab(m, k) << 3;
                k = (uint16_t)(k + B);
                mm = (uint16_t)(mm - B);
            } while (mm >= B);
            M = w004D_sx3(mm);
            gin3_ldivmod(M, (uint32_t)B, &D, &rem);
            T1 = w004D_tab(m, (uint16_t)(i + 0x40u));
            acc1 += gin3_lmul(D, T1);
            T2 = w004D_tab(m, i);
            acc2 += gin3_lmul(D, T2);
            gin3_ldivmod(w004D_sx3((uint16_t)(i << 1)), (uint32_t)B, &Q2, &rem);
            R = (uint32_t)A - Q2;
            x = acc1 + gin3_lmul(R, T1);
            y = acc2 + gin3_lmul(R, T2);
        }
        fl = f_3CC9_00B9((uint16_t)(y >> 16), (uint16_t)y, 0, 0);
        if ((fl & F_SF) != 0) {
            done = 1;
        } else {
            uint16_t at = (uint16_t)(0x9C1Au + (uint16_t)(n << 2));

            w004D_wr16s(m, GIN3_DS, at, (uint16_t)((int32_t)x >> 10));
            w004D_wr16s(m, GIN3_DS, (uint16_t)(at + 2u), (uint16_t)((int32_t)y >> 10));
            i = (uint16_t)(i + 8u);
            n = (uint16_t)(n + 1u);
            if (n >= W004D_MAX_POINTS)
                done = 1; /* 輸入域外的安全上限：原版在 200 萬步內到不了這麼多點（見檔頭），突變版才會停不下來 */
        }
    }

    for (j = 0; (int16_t)j < (int16_t)n; j++) {
        uint16_t src = (uint16_t)(j << 2);
        uint16_t dst = (uint16_t)((uint16_t)(n + n - 1u - j) << 2);

        w004D_wr16s(m, GIN3_DS, (uint16_t)(0x9C1Au + dst), w004D_rd16s(m, GIN3_DS, (uint16_t)(0x9C1Au + src)));
        w004D_wr16s(m, GIN3_DS, (uint16_t)(0x9C1Au + dst + 2u),
                    (uint16_t)(0u - w004D_rd16s(m, GIN3_DS, (uint16_t)(0x9C1Au + src + 2u))));
    }
    for (j = 0; (int16_t)(uint16_t)(n << 1) > (int16_t)j; j++) {
        uint16_t p = (uint16_t)(0x9C1Au + (uint16_t)(j << 2));

        f_3A92_0001(m, p, GIN3_DS, (uint16_t)(p + 2u), GIN3_DS, w004D_rd16s(m, fl_seg, (uint16_t)(fl_off + 0x4Au)));
    }
    w004D_wr16s(m, out_seg, (uint16_t)(out_off + 8u), (uint16_t)(n << 1));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_165F_004D(m, a[0], a[1], a[2], a[3], a[4], a[5]);
    return 0;
}
GIN3_REGISTER("165F_004D", 6, run)
