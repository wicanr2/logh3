/*
 * GIN3PS.EXE 1588:0945  戰場鏡頭：依座位找到第一支屬於該座位陣營的艦隊，把艦隊的鏡頭座標換成格座標，夾進範圍，寫進該座位的鏡頭槽
 *
 * 原版：far 呼叫，一個堆疊字 seat（序言 push bp／si、sub sp,0Eh，所以在 [bp+16h]，只取低位元組），回 AX。
 * 逐條對應反組譯：
 *   faction ← byte ds:(0x8F25 ＋ seat)                         ; 座位 → 陣營表（`[bx-70DBh]`）
 *   i ← 0（byte）、清單位移 ← 0x9898
 *   迴圈：
 *     i >= byte ds:9894（無號，注意是 byte；1C03:0006 讀同一個筆數是 word）→ 結束
 *     艦隊 ← ds:(0xC074 ＋ low16(byte ds:[清單位移] × 0x7A))
 *     byte 艦隊[+1] == faction → 結束（找到）
 *     i++；清單位移++
 *   找不到時迴圈因 i >= 筆數離開，清單位移指在清單最後一項之後；兩種離開都接著用「清單位移指到的那個 byte」當艦隊編號
 *   （原版沒有檢查，找不到時就用 ds:9898[筆數] 那一格）：
 *   艦隊 ← ds:(0xC074 ＋ low16(byte ds:[清單位移] × 0x7A))
 *   x ← 有號 16 位元 ((word 艦隊[+0x5C] ＋ 0xFF20) ÷ 0x10)       ; 即 (+0x5C − 224) ÷ 16，idiv 向零取整
 *   y ← 有號 16 位元 ((word 艦隊[+0x5E] ＋ 0xFF4C) ÷ 0x14)       ; 即 (+0x5E − 180) ÷ 20
 *   1588:04A0(&x, &y)                                          ; x 夾進 0 到 100、y 夾進 0 到 84（兩個局部變數在堆疊上）
 *   word ds:(0x8F3A ＋ seat × 4)     ← x << 4（16 位元）
 *   word ds:(0x8F3A ＋ seat × 4 ＋ 2) ← y × 20（16 位元有號乘法，取低 16 位元）
 *   回 AX ← y × 20（最後一個寫進 AX 的值）
 * 注意：夾限把負值一律夾成 0，所以 x、y 的除法是「向零取整」還是「向負無限取整」在輸出上無法區分；
 *   有號與無號除法則可區分（艦隊 +0x5C < 224 時有號為負、夾成 0，無號會是大數、夾成 100）。
 *
 * 堆疊局部變數：x、y 在原版放在堆疊（ss:bp+6、ss:bp+8），04A0 透過 far 指標改它們。差異測試排除堆疊區，所以本版
 *   借用線性 0x9D000（探針的暫存緩衝區，本函式不用）：先存原來的 4 bytes、寫 x、y、呼叫 04A0、讀回、再還原 4 bytes，
 *   對記憶體的淨影響為零，與原版（只動堆疊）一致。
 *
 * 呼叫者：1588:0930（呼叫兩次，一次一個座位，docs/re/393-battle-work-control.md「1588:0945 依座位找首個對應陣營艦隊，
 *   以 FLEET+5C／+5E 計算鏡頭」，與本檔一致；補上：除數 16／20、夾限由 04A0 做、找不到時用清單 [筆數] 那一格）。
 *
 * 推論等級
 *   已證實（差異測試）：表位址、筆數的寬度、迴圈條件與找不到時的行為、算式與常數、04A0 的參數次序、寫入位置與值、回傳值。
 *   強推論（docs/re/393、326）：艦隊 +1 是陣營、+0x5C／+0x5E 是艦隊的戰場座標（像素），ds:8F3A 起每座位 4 bytes 是鏡頭槽
 *     （x 以 16 像素為單位的像素座標、y 以 20 為單位）。名稱為假說。
 * 輸入域：seat 任意 word（取低位元組）；清單項目 0..255、筆數 0..255；艦隊記錄固定在 DS 段，位移 16 位元回繞。
 */
#include "gin3_reg.h"

extern uint16_t f_1588_04A0(Mem *m, uint16_t p_off, uint16_t p_seg, uint16_t q_off, uint16_t q_seg);

#define F0945_TMP_SEG 0x9D00u /* 線性 0x9D000：探針的暫存緩衝區，本函式不用；借來放 x、y 兩個堆疊局部變數 */

/* DS 段內的 word 存取：位移 16 位元回繞，seg:FFFF 的高位元組在位移 0（與 dosgolem 一致） */
static uint16_t f0945_rd16(const Mem *m, uint16_t off)
{
    return (uint16_t)(rd8(m, DS(off)) | (rd8(m, DS((uint16_t)(off + 1u))) << 8));
}

static void f0945_wr16(Mem *m, uint16_t off, uint16_t v)
{
    wr8(m, DS(off), (uint8_t)v);
    wr8(m, DS((uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_1588_0945(Mem *m, uint16_t seat_word)
{
    uint8_t seat = (uint8_t)seat_word;
    uint8_t faction = rd8(m, DS((uint16_t)(0x8F25u + seat)));
    uint8_t i = 0;
    uint16_t list = 0x9898u;
    uint16_t fleet, slot;
    int16_t x, y;
    uint8_t saved[4];
    int k;

    while (i < rd8(m, DS(0x9894u))) {
        uint16_t f = (uint16_t)(0xC074u + (uint16_t)(rd8(m, DS(list)) * 0x7Au));

        if (rd8(m, DS((uint16_t)(f + 1u))) == faction)
            break;
        i = (uint8_t)(i + 1u);
        list = (uint16_t)(list + 1u);
    }
    fleet = (uint16_t)(0xC074u + (uint16_t)(rd8(m, DS(list)) * 0x7Au));
    x = (int16_t)((int16_t)(uint16_t)(f0945_rd16(m, (uint16_t)(fleet + 0x5Cu)) + 0xFF20u) / 0x10);
    y = (int16_t)((int16_t)(uint16_t)(f0945_rd16(m, (uint16_t)(fleet + 0x5Eu)) + 0xFF4Cu) / 0x14);

    for (k = 0; k < 4; k++)
        saved[k] = rd8(m, LIN(F0945_TMP_SEG, (uint16_t)k));
    wr16(m, LIN(F0945_TMP_SEG, 0), (uint16_t)x);
    wr16(m, LIN(F0945_TMP_SEG, 2), (uint16_t)y);
    (void)f_1588_04A0(m, 0, F0945_TMP_SEG, 2, F0945_TMP_SEG);
    x = (int16_t)rd16(m, LIN(F0945_TMP_SEG, 0));
    y = (int16_t)rd16(m, LIN(F0945_TMP_SEG, 2));
    for (k = 0; k < 4; k++)
        wr8(m, LIN(F0945_TMP_SEG, (uint16_t)k), saved[k]);

    slot = (uint16_t)(0x8F3Au + (uint16_t)(seat << 2));
    f0945_wr16(m, slot, (uint16_t)((uint16_t)x << 4));
    f0945_wr16(m, (uint16_t)(slot + 2u), (uint16_t)((uint16_t)y * 20u));
    return (uint16_t)((uint16_t)y * 20u);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1588_0945(m, a[0]); }
GIN3_REGISTER("1588_0945", 1, run)
