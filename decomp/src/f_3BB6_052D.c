/*
 * 3BB6:052D  由聲道記錄的兩個 word 與一個 byte 算出 13 位元的編碼值（near，SI ＝ 記錄指標，回 AX）
 *
 * 原版（DS ＝ seg130，dosgolem 下 0x4AFE）：
 *   ax ← word [si+0C]；dx:ax ← ax × word [si+0E]（無號 mul）；dx:ax ← dx:ax × 2（add ax,ax；adc dx,dx，進位出 bit 31 丟掉）；
 *   ax ← dx（即雙倍乘積的高 16 位元）；dl ← byte [si+04]；
 *   若 ax >= 0x03FF（無號，`cmp ax,3FFh；jb` 不成立）：dl ← dl ＋ 1（8 位元）、ax ← ax >> 1（邏輯）；
 *   dl ← dl × 4（連續兩次 add dl,dl，8 位元回繞）；ah ← ah or dl；ax ← ax and 0x1FFF。
 *   回傳 AX；殘留的 DX ＝ (雙倍乘積高字的高位元組) << 8 ｜ 最後的 dl。描述檔 ret ＝ pack:ax,dx 兩個都比對；
 *   呼叫者 3BB6:054F 只用 AX（AH、AL 分別當後續兩次輸出的值），DX 在它 call 3BB6:023F 時被重設，不使用殘留。
 *   不寫記憶體。DS：呼叫者 3BB6:02BB 載入 seg130（載入位置相依），描述檔用 ds_seg 4AFE；C 版把 ds 當參數。
 *
 * 語意（假說）：SI 指向 3BB6:0003 建立的聲道記錄（ds:0006 ＋ 0x14 × 聲道，8 個）。+0x0C 是基礎值（呼叫者 02BB 由
 *   ds:03BE 起的 word 表以「鍵 mod 12」取得）、+0x0E 初始 0x8000、+0x04 是「鍵 / 12 － 2」。算式像是：
 *   位於 [0, 0x3FF) 的 10 位元尾數 ＝ 基礎值 × (+0x0E / 0x8000)，滿 0x3FF 就尾數減半、指數加一，
 *   指數放在 ah 的 bit 2..4（dl × 4 後的低 3 位元有效），整個編碼取 13 位元。
 *   這個形狀與 FM 音源的 F-number（10 位元）加 block（3 位元）暫存器對相符，但未驗證。名稱與用途：未知。
 *   dl 的高位元（bit 3 以上）在 and 0x1FFF 後消失，只有 DX 殘值看得到。
 * 推論等級：算術已證實（逐指令對過）；欄位與用途假說。
 * 輸入域：SI 任意 16 位元值（word 存取落在位移 0xFFFF 時，高位元組取同一段的位移 0，用 f052d_rd16s 表達，
 *   不是 gin3_mem.h 的線性位址加 1）；DS 固定 seg130（0x4AFE，線性 0x4AFE0..0x5AFDF，不碰 A0000 以上）。
 */
#include "gin3_reg.h"

/* 段內回繞的 word 讀取：seg:FFFF 的高位元組在 seg:0000（dosgolem internal/cpu/cpu.go，8086 語意） */
static uint16_t f052d_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint64_t f_3BB6_052D(Mem *m, uint16_t ds, uint16_t si)
{
    uint16_t a = f052d_rd16s(m, ds, (uint16_t)(si + 0x0Cu));
    uint16_t b = f052d_rd16s(m, ds, (uint16_t)(si + 0x0Eu));
    uint32_t prod = (uint32_t)a * b; /* mul：dx:ax */
    uint16_t lo = (uint16_t)prod;
    uint16_t hi = (uint16_t)(prod >> 16);
    uint16_t carry = (uint16_t)(lo >> 15); /* add ax,ax 的進位 */
    uint16_t ax, dx;
    uint8_t dl, dh;

    hi = (uint16_t)((hi << 1) | carry); /* adc dx,dx */
    ax = hi;                            /* mov ax,dx */
    dh = (uint8_t)(hi >> 8);
    dl = rd8(m, LIN(ds, (uint16_t)(si + 4u)));
    if (ax >= 0x03FFu) {
        dl = (uint8_t)(dl + 1u);
        ax = (uint16_t)(ax >> 1);
    }
    dl = (uint8_t)(dl << 1); /* add dl,dl 第一次 */
    dl = (uint8_t)(dl << 1); /* add dl,dl 第二次 */
    ax = (uint16_t)(ax | ((uint16_t)dl << 8)); /* or ah,dl */
    ax = (uint16_t)(ax & 0x1FFFu);
    dx = (uint16_t)(((uint16_t)dh << 8) | dl);
    return ((uint64_t)ax << 16) | dx;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3BB6_052D(m, 0x4AFEu, a[0]); /* 描述檔 ds_seg：呼叫者 3BB6:02BB 載入的 seg130；參數 si */
}
GIN3_REGISTER("3BB6_052D", 1, run)
