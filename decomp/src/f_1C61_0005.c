/*
 * GIN3PS.EXE 1C61:0005  戰場移動階段開頭：艦隊每個分隊（含啟用的索敵分遣隊）的「這一回合剩下的行程」初始化
 *
 * 原版：far 呼叫，一個參數＝艦隊記錄的 far 指標（位移在前、段在後），無回傳值（描述檔 ret = none）。
 * 艦艇池基底在 ds:CFB4（位移）與 ds:CFB6（段），每次都重讀。
 *
 * 逐條對應反組譯（IDA sub_29185；對一艘艦艇做的事稱為「初始化行程」）：
 *   for i ← 0；i < 艦隊[+0x0F](byte，每圈重讀)；i++：
 *     初始化行程(池[艦隊[+0x10 + 4i](word)])
 *   若 艦隊[+0x3A](byte) != 0xFF：
 *     初始化行程(池[艦隊[+0x38](word)])              ; 索敵分遣隊
 *
 *   初始化行程(艦艇)：
 *     add ← (32 位元)艦艇[+0x21](byte) << 19            ; ＝ 速度 × 8 的 16.16 定點；低 16 位元恆為 0
 *     cur ← (有號 32 位元)艦艇[+0x26:+0x24]
 *     若 cur >= 0x00080000（8.0）：艦艇[+0x26:+0x24] ← add               ; 覆蓋
 *     否則若 cur <= 0：          艦艇[+0x26:+0x24] ← add               ; 覆蓋
 *     否則（0 < cur < 8.0）：     艦艇[+0x24] += add 的低字；[+0x26] += add 的高字＋進位   ; 累加
 *   兩次比較走 3CC9:00B9（有號 32 位元比較，輸出旗標），呼叫端接 jge、jle；該函式不改 BX／CX／DX／ES，
 *   旗標組合對 jge／jle 等價於真正的有號 32 位元比較，所以這裡直接用有號比較表達。
 *   覆蓋與「cur ＝ 0 時累加」得到同一個值；cur 的正負號在這支函式裡不會造成可觀察的差異
 *   （負數與超過 8.0 的值都覆蓋），但比較仍照原版寫成有號。
 *
 * 推論等級
 *   已證實（差異測試的位元組與寫入集合）：分支條件 0 < cur < 0x80000 才累加、其餘覆蓋，覆蓋值與累加量都是 速度<<19；
 *     索敵分遣隊只在艦隊 +0x3A != 0xFF 時初始化。
 *   強推論（docs/re/371-battle-movement-phase.md 的「已證實的額度與停止條件」、docs/internals/64-ship-pool.md 的
 *     「移動階段額度與停止條件」、docs/spec/080 第 53 行）：艦艇 +0x21 是巡航速度、+0x24／+0x26 是這回合剩下的行程。
 *     這些文件對本函式的描述與本檔逐項相符。
 *   與文件的差異：docs/internals/38-battle-units.md 欄位表第 122 行寫「每回合開頭從 +0x21（巡航速度）左移 19 位補滿」，
 *     實際只有 cur 不在 (0, 8.0) 時才是覆蓋（補滿），在區間內是累加；docs/re/124 第 38–45 行的虛擬碼寫到
 *     「與 8 比、與 0 比 …」就停了。見回報。
 *   未知：呼叫者 1C03:0006 的時機（本函式只有一個呼叫者）。
 *
 * 呼叫的函式：3CC9:00B9 已 diff-tested；其輸出旗標的語意見 decomp/src/f_3CC9_00B9.c。
 *
 * 輸入域：艦隊指標取自艦隊表（描述檔 table），艦隊 +0x0F 到 255、池索引任意；
 *   池索引超出 352 筆時讀寫池之後的記憶體，本版同樣照做（平面視訊記憶體的讀取除外，見 docs/re/445 限制）。
 */
#include "gin3_reg.h"

static void init_move_budget(Mem *m, uint16_t idx)
{
    uint16_t seg = rd16(m, DS(0xCFB6));
    uint16_t ship = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(idx * 0x4Eu));
    uint32_t add = (uint32_t)rd8(m, LIN(seg, ship + 0x21)) << 19;
    int32_t cur = (int32_t)(((uint32_t)rd16(m, LIN(seg, ship + 0x26)) << 16) | rd16(m, LIN(seg, ship + 0x24)));

    if (cur >= 0x80000 || cur <= 0) {
        wr16(m, LIN(seg, ship + 0x26), (uint16_t)(add >> 16));
        wr16(m, LIN(seg, ship + 0x24), (uint16_t)add);
    } else {
        uint32_t lo = (uint32_t)rd16(m, LIN(seg, ship + 0x24)) + (uint16_t)add;

        wr16(m, LIN(seg, ship + 0x24), (uint16_t)lo);
        wr16(m, LIN(seg, ship + 0x26),
             (uint16_t)(rd16(m, LIN(seg, ship + 0x26)) + (uint16_t)(add >> 16) + (lo >> 16)));
    }
}

void f_1C61_0005(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint8_t i = 0;
    uint16_t slot = (uint16_t)(fleet_off + 0x10);

    while (i < rd8(m, LIN(fleet_seg, fleet_off + 0x0F))) {
        init_move_budget(m, rd16(m, LIN(fleet_seg, slot)));
        i++;
        slot = (uint16_t)(slot + 4);
    }
    if (rd8(m, LIN(fleet_seg, fleet_off + 0x3A)) != 0xFF)
        init_move_budget(m, rd16(m, LIN(fleet_seg, fleet_off + 0x38)));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1C61_0005(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("1C61_0005", 2, run)
