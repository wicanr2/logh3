/*
 * GIN3PS.EXE 2526:005A  戰場回合開頭的「發現程度」衰減：對一支艦隊的每個分隊套一次
 *
 * 原版：far 呼叫，一個參數＝艦隊記錄（122 bytes）的 far 指標，無回傳值（描述檔 ret = none）。
 *   堆疊：序言 push bp／si、sub sp,0Ah，所以第一個參數在 [bp+12h]（位移）與 [bp+14h]（段）。
 *   用到的全域變數：ds:CFB4（位移）／ds:CFB6（段）＝ 艦艇池的 far 指標，每筆 78（0x4E）bytes。
 *   呼叫者是 2526:000E（回合開頭，docs/internals/38-battle-units.md 第 304 行：12AA:0245 → 2526:000E → 2526:005A）。
 *
 * 語意（已證實：本檔與反組譯逐條對過，並與 docs/internals/38-battle-units.md 第 304 到 320 行的衰減虛擬碼相符）：
 *   對 k = 0 .. f[+0x0F] − 1（每圈重讀 f[+0x0F]，byte 無號比較）：
 *     s ← (ds:CFB6):(ds:CFB4 + word(f[+0x10 + 4k]) × 0x4E)   ; 乘積只取低 16 位元，每圈重讀 CFB4／CFB6
 *     若 s[+0x16] 的 bit0 沒設（上一輪沒被看到）：
 *       s[+0x17] == 3：s[+0x18] ← s[+0x18] − 1（byte 回繞，0 變 0xFF）；減完是 0 才把 s[+0x17] ← 2
 *       s[+0x17] == 2：s[+0x17] ← 1
 *       s[+0x17] == 1：s[+0x17] ← 0
 *       其他值（0、4 以上）：不動
 *     不論 bit0 有沒有設：s[+0x16] ← s[+0x16] & 0xFE
 *   s[+0x17] 是以 byte 取出再零擴展成 word 去比 3／2／1，s[+0x16] 的其餘位元不動。
 *
 * 欄位語意（強推論，來自 docs/internals/38 與 docs/spec/087）：s[+0x16] bit0 ＝ 這一輪被看到過，s[+0x17] ＝ 發現程度
 *   （0 到 3），s[+0x18] ＝ 程度 3 時還要撐幾輪才降成 2。「先清 bit0 再判斷」與本函式的順序（先判斷再清）等價，
 *   本檔照反組譯寫：先判斷、最後清。
 *
 * 讀寫順序：艦艇欄位都經過同一個 s 指標，艦隊的分隊數、分隊槽與 ds:CFB4／CFB6 都是每圈重讀，
 *   所以艦艇記錄若與艦隊記錄重疊（寫入蓋到分隊數或槽），後續圈會看到新值，本檔逐圈重讀以保持一致。
 *
 * 輸入域：艦隊指標取艦隊表任一筆（32 筆）或 seeds 指定的任意位址；分隊數可到 255，超出 12 時讀到艦隊記錄之後
 *   的記憶體當槽，本版照做。基底是戰略層（fcmd-base.state），艦艇欄位 +0x16 到 +0x18 全由突變與 seeds 供給。
 *   字的存取跨越位移 0xFFFF 時，本版與探針都按線性位址連續處理，不涵蓋真機上回繞到段首的行為。
 *   寫入的目標若落在 0xF0000 以上（ROM，dosgolem 的 ROMBase）不生效，本檔 put8 照該模型處理。
 */
#include "gin3_reg.h"

static void put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_2526_005A(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint8_t i = 0;
    uint16_t slot_off = (uint16_t)(fleet_off + 0x10);

    while (i < rd8(m, LIN(fleet_seg, fleet_off + 0x0F))) {
        uint16_t idx = rd16(m, LIN(fleet_seg, slot_off));
        uint16_t s_off = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(idx * 0x4Eu));
        uint16_t s_seg = rd16(m, DS(0xCFB6));

        if ((rd8(m, LIN(s_seg, s_off + 0x16)) & 1u) == 0) {
            uint16_t level = rd8(m, LIN(s_seg, s_off + 0x17));

            if (level == 3) {
                uint8_t hold = (uint8_t)(rd8(m, LIN(s_seg, s_off + 0x18)) - 1u);

                put8(m, LIN(s_seg, s_off + 0x18), hold);
                if (hold == 0)
                    put8(m, LIN(s_seg, s_off + 0x17), 2);
            } else if (level == 2) {
                put8(m, LIN(s_seg, s_off + 0x17), 1);
            } else if (level == 1) {
                put8(m, LIN(s_seg, s_off + 0x17), 0);
            }
        }
        put8(m, LIN(s_seg, s_off + 0x16), (uint8_t)(rd8(m, LIN(s_seg, s_off + 0x16)) & 0xFEu));
        i++;
        slot_off = (uint16_t)(slot_off + 4);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2526_005A(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2526_005A", 2, run)
