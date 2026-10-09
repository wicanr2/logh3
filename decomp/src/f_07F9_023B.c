/*
 * GIN3PS.EXE 07F9:023B  提督的艦種欄換一個艦種試試看：現在的艦種存量不夠，就換成另一個（0 ↔ 非 0）（回 AX = 0 或 1）
 *
 * 原版：far 呼叫，一個 far 指標參數＝提督記錄（序言 push bp／si、sub sp,6，所以在 [bp+0Eh]／[bp+10h]）。
 * 提督記錄欄位（相對提督記錄，與 07F9:033A 同表）：[+0x01] 陣營（word 讀 [+1][+2]，被呼叫者只取低位元組）、
 *   [+0x08] 的低 4 位元＝艦種、高 4 位元不動、[+0x0A] 星域編號。
 *
 * 逐條對應反組譯：
 *   region ← ds:(0xA79C ＋ low16(adm[+0x0A] × 0x7E))        ; 星域記錄，段固定為 DS
 *   type   ← adm[+0x08] & 0x0F
 *   07F9:02CB(region, 陣營 word = adm[+1..+2], 艦種 word) ≠ 0 → 回 1（不寫記憶體）
 *     艦種 word 的高位元組是 mul 的殘值（adm[+0x0A] × 0x7E 的高位元組），被呼叫者只取低位元組，所以不影響結果
 *   next ← (type == 0) ? 1 : 0
 *   07F9:02CB(region, 陣營 word（重讀 adm[+1..+2]）, next) ≠ 0：
 *       adm[+0x08] ← (adm[+0x08] & 0xF0) | next（先 and 後 or，兩步各自讀寫記憶體）；回 1
 *   否則回 0（不寫記憶體）
 * 也就是「這個艦種存量不夠就改成 0 號艦種（現在是 0 號則改成 1 號艦種），那個夠才改」。
 * 02CB 是唯讀函式（已 diff-tested，語意見 f_07F9_02CB.c：存量 >= 每組艘數門檻）。
 *
 * 推論等級
 *   已證實（差異測試）：欄位位移、常數 0x7E／0xA79C、低 4 位元遮罩、next 的算法、寫回的 and／or 兩步、兩次呼叫的參數。
 *   強推論（docs/re/182、07F9:02CB 檔頭）：adm[+1] 是陣營、adm[+8] 低 4 位元是艦種、adm[+0x0A] 是所在星域。
 *   未知：為什麼只在 0 與非 0 之間切換（docs/ 沒有描述 07F9:023B）。
 * 呼叫者：07F9:00D5、2D76:0611（兩者都在挑艦隊槽附近，用途為強推論）。
 * 輸入域：提督記錄 far 指標任意（含位移接近 0xFFFF 的邊緣：欄位位移 16 位元回繞）；
 *   adm[+0x0A] 為 0..255（星域 >= 32 時指到星域表之後的記憶體，照原版運算）。
 */
#include "gin3_reg.h"

extern uint16_t f_07F9_02CB(Mem *m, uint16_t region_off, uint16_t region_seg, uint16_t faction_w, uint16_t type_w);

/* word 讀取：位移各自 16 位元回繞（seg:FFFF 的高位元組在同一段的位移 0），與 dosgolem 一致 */
static uint16_t f07F9_023B_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_07F9_023B(Mem *m, uint16_t adm_off, uint16_t adm_seg)
{
    uint16_t prod = (uint16_t)(rd8(m, LIN(adm_seg, (uint16_t)(adm_off + 0x0Au))) * 0x7Eu);
    uint16_t region = (uint16_t)(0xA79Cu + prod);
    uint8_t type = (uint8_t)(rd8(m, LIN(adm_seg, (uint16_t)(adm_off + 8u))) & 0x0Fu);
    uint8_t next;

    if (f_07F9_02CB(m, region, GIN3_DS, f07F9_023B_rd16s(m, adm_seg, (uint16_t)(adm_off + 1u)),
                    (uint16_t)((prod & 0xFF00u) | type)) != 0)
        return 1;

    next = (uint8_t)(type == 0 ? 1 : 0);
    if (f_07F9_02CB(m, region, GIN3_DS, f07F9_023B_rd16s(m, adm_seg, (uint16_t)(adm_off + 1u)), next) != 0) {
        uint16_t at = (uint16_t)(adm_off + 8u);

        wr8(m, LIN(adm_seg, at), (uint8_t)(rd8(m, LIN(adm_seg, at)) & 0xF0u));
        wr8(m, LIN(adm_seg, at), (uint8_t)(rd8(m, LIN(adm_seg, at)) | next));
        return 1;
    }
    return 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_07F9_023B(m, a[0], a[1]); }
GIN3_REGISTER("07F9_023B", 2, run)
