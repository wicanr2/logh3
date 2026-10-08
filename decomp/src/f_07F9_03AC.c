/*
 * GIN3PS.EXE 07F9:03AC  替提督挑艦隊槽（far，一個 far 指標參數，回 AX，不寫記憶體）
 *
 * 參數：admiral = 提督記錄的 far 指標（位移、段）。艦隊表在 ds:C074，一筆 122 bytes、32 筆（docs/internals/33）。
 * 讀取：adm[+0x01]（陣營，0 或非 0）、adm[+0x09]（完整一個 byte，沒有遮 bit7；呼叫端 2D76:0611 先把 bit7 清掉）、
 *   艦隊[+0x04]（0xFF = 空槽）。
 * 算式：
 *   base = (adm[+0x01] == 0) ? 0 : 16               ； 陣營 0 用艦隊表前 16 筆，否則後 16 筆
 *   若 艦隊[base + adm[+0x09]][+0x04] == 0xFF：回 adm[+0x09]       ； 原本那一槽是空的就沿用
 *   否則 i = 0 … 15：艦隊[base + i][+0x04] == 0xFF 就回 i
 *   16 筆都不是空的：回 0xFF
 *   槽索引超過表尾時（adm[+0x09] >= 32 - base）原版照樣讀表尾之後的記憶體，本版同樣處理（位移 16 位元回繞）。
 *
 * 回傳 AX 的高位元組（codegen 殘留，已由差異測試證實）：
 *   原版用 mul 算 base × 122 再只改 AL 讀 adm[+0x09]，三條回傳路徑都沒有清 AH，
 *   所以 AH = (base × 122) 的高位元組：陣營 0 時 0x00，陣營非 0 時 0x07（16 × 122 = 0x07A0）。
 *   即回傳是 0x00XX（陣營 0）或 0x07XX（陣營非 0）。呼叫端 2D76:0611 只存 AL（mov es:[bx],al），不受影響。
 *   本版逐位元重現，避免 AX 比對不符。
 *
 * 推論等級
 *   已證實（差異測試）：算式與回傳高位元組。
 *   強推論（docs/re/181、docs/internals/54）：艦隊[+0x04] == 0xFF 是空槽；adm[+0x09] 低七位是艦隊槽、bit7 是
 *     「沒有帶艦隊」（docs/re/182）。
 * 輸入域：admiral 取提督表 8 筆代表（兩個陣營、adm[+0x09] 起點 0／127／129 等），另有 DS+0x10 的段變體驗證段參數確實被使用。
 *   函式只讀 adm[+0x01]、adm[+0x09] 兩個 byte；這兩個 byte 與艦隊表內容由突變與 seeds 決定，0..255 全域，
 *   所以 8 筆代表在欄位值域上沒有縮小。seeds 用來打「16 槽全滿」「第 15 槽才是空的」這類突變打不到的單點。
 */
#include "gin3_reg.h"

uint16_t f_07F9_03AC(Mem *m, uint16_t adm_off, uint16_t adm_seg)
{
    uint16_t base = rd8(m, LIN(adm_seg, adm_off + 1u)) ? 16u : 0u;
    uint16_t base_bytes = (uint16_t)(base * 122u);
    uint16_t ah = (uint16_t)(base_bytes & 0xFF00u);
    uint16_t slot = rd8(m, LIN(adm_seg, adm_off + 9u));
    uint16_t cur = (uint16_t)(0xC074u + (uint16_t)(slot * 122u) + base_bytes);
    uint16_t ptr = (uint16_t)(0xC074u + base_bytes);
    uint8_t i;

    if (rd8(m, DS(cur + 4u)) == 0xFF)
        return (uint16_t)(ah | slot);
    for (i = 0; i < 16; i++) {
        if (rd8(m, DS(ptr + 4u)) == 0xFF)
            return (uint16_t)(ah | i);
        ptr = (uint16_t)(ptr + 122u);
    }
    return (uint16_t)(ah | 0xFFu);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_07F9_03AC(m, a[0], a[1]); }
GIN3_REGISTER("07F9_03AC", 2, run)
