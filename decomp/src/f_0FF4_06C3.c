/*
 * GIN3PS.EXE 0FF4:06C3  艦隊組成畫面開場：把艦隊 +0x10 起的編制，依隊形槽抄進 ds:977A 工作表，並換格盤圖塊
 *
 * 原版：far 呼叫，一個 far 指標參數＝艦隊記錄（序言 push bp／si／di、sub sp,10h，所以在 [bp+1Ah]／[bp+1Ch]），無回傳值。
 * 逐條對應反組譯：
 *   byte ds:984B ← 0
 *   i ← 0（byte）、組 ← 艦隊 + 0x10（16 位元位移，段＝艦隊段）
 *   迴圈：
 *     i >= byte 艦隊[+0x0F]（無號，每圈重讀）→ 結束
 *     byte 組[2] == 0xFF（空組）→ 略過下面這一整段
 *     否則：
 *       inc byte ds:984B
 *       di ← 0x977A ＋ (byte 組[3]) × 4                         ; 工作表第 組[3] 格（隊形槽）
 *       word ds:[di] ← word 組[0]                                ; 4 bytes 逐段抄：先 word
 *       byte ds:[di+2] ← byte 組[2]                              ; （每一段都在前一段寫入「之後」才讀來源）
 *       byte ds:[di+3] ← byte 組[3]
 *       0FF4:0EF1(idx ＝ word 組[3]、tile ＝ word 組[2])          ; 兩個都是 word 讀（含下一個位元組），0EF1 只取低位元組
 *     i++；組 += 4
 *   byte ds:984C ← byte 艦隊[+0x0E] − byte ds:984B                 ; 應有組數減現有（byte 相減）
 * 艦隊記錄若與工作表（ds:977A 起）重疊，後面的讀取會看到前面的寫入；本版逐步照原版順序做。
 *
 * 呼叫者：0FF4:0591（docs/internals/56-fleet-composition-panel.md「開畫面時怎麼填（0FF4:06C3）」，虛擬碼與本檔一致）。
 *   回傳 AX 是殘值（0EF1 沒有回傳值），呼叫者 0FF4:0591 結尾 `call 06C3; mov sp,bp; pop bp; retf`，描述檔 ret = none。
 *
 * 推論等級
 *   已證實（差異測試）：欄位位移與寬度、迴圈條件（無號、每圈重讀組數）、空組判定、寫入順序、0EF1 的參數、984B／984C 的算式。
 *   強推論（docs/internals/56、33）：組[0..1] 艦艇池索引、組[2] 艦種、組[3] 隊形槽、艦隊 +0x0E 應有組數、+0x0F 現有組數；
 *     ds:977A 是 41 × 4 的工作表、ds:984B 已編成組數、ds:984C 差額。
 * 輸入域：艦隊指標任意（含別名到工作表本身與位移接近 0xFFFF）；組數 0..255（超過 10 時讀到艦隊記錄之後的記憶體，照原版）。
 *   寫入全在 DS 段（工作表、984B、984C）加 0EF1 的物件寫入（0EF1 自己處理 ROM 區）。
 */
#include "gin3_reg.h"

extern void f_0FF4_0EF1(Mem *m, uint16_t idx_word, uint16_t tile_word);

/* word 存取：位移各自 16 位元回繞（seg:FFFF 的高位元組在同一段的位移 0），與 dosgolem 一致 */
static uint16_t f06C3_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void f06C3_ds_wr16s(Mem *m, uint16_t off, uint16_t v)
{
    wr8(m, DS(off), (uint8_t)v);
    wr8(m, DS((uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_0FF4_06C3(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint8_t i = 0;
    uint16_t slot = (uint16_t)(fleet_off + 0x10u);

    wr8(m, DS(0x984B), 0);
    while (i < rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x0Fu)))) {
        if (rd8(m, LIN(fleet_seg, (uint16_t)(slot + 2u))) != 0xFFu) {
            uint16_t at3 = (uint16_t)(slot + 3u);
            uint16_t di;

            wr8(m, DS(0x984B), (uint8_t)(rd8(m, DS(0x984B)) + 1u));
            di = (uint16_t)(0x977Au + (uint16_t)(rd8(m, LIN(fleet_seg, at3)) << 2));
            f06C3_ds_wr16s(m, di, f06C3_rd16s(m, fleet_seg, slot));
            wr8(m, DS((uint16_t)(di + 2u)), rd8(m, LIN(fleet_seg, (uint16_t)(slot + 2u))));
            wr8(m, DS((uint16_t)(di + 3u)), rd8(m, LIN(fleet_seg, at3)));
            f_0FF4_0EF1(m, f06C3_rd16s(m, fleet_seg, at3), f06C3_rd16s(m, fleet_seg, (uint16_t)(slot + 2u)));
        }
        i = (uint8_t)(i + 1u);
        slot = (uint16_t)(slot + 4u);
    }
    wr8(m, DS(0x984C), (uint8_t)(rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x0Eu))) - rd8(m, DS(0x984B))));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_0FF4_06C3(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("0FF4_06C3", 2, run)
