/*
 * GIN3PS.EXE 2E54:0C27  對陣列裡「已到達」(+0x79 == 6)的艦隊下「出擊」命令（far，三個參數，無回傳值；ret = none）
 *
 * 參數（堆疊字，由低位址到高位址；序言 push bp／si、sub sp,0Ah，所以第一個參數在 [bp+12h]）：
 *   list     far 指標（位移、段）：艦隊 far 指標的陣列，每筆 4 bytes（位移 word 在前、段 word 在後）
 *   count    一個字，只用低位元組（cmp al,[bp+16h] 是 byte 比較，jnb）：筆數
 *   target   一個字，只用低位元組（mov dl,[bp+18h]）：目標星域編號
 *
 * 反組譯逐條對應（IDA sub_3BCD7）：
 *   對 i = 0 … count−1（byte 計數）：
 *     fleet ← far 指標 list[i]（讀位置 list.off + i × 4，16 位元回繞；位移 word、段 word 各讀一次）
 *     fleet[+0x79] == 6 才處理（byte 比較，jnz 跳過）：
 *       fleet[+0x08] ← 1
 *       fleet[+0x07] ← target
 *       fleet[+0x79] ← 1
 *   三個寫入的位移都是 fleet.off + 常數（16 位元回繞），段不變。每一圈都重讀陣列項，
 *   所以陣列與艦隊記錄重疊時，前一圈的寫入會影響後一圈讀到的指標（別名）。
 *   與 2E54:0C83 的差別只有 +0x79 == 6 的條件。
 *
 * 推論等級
 *   已證實（差異測試的位元組與寫入集合）：迴圈結構、條件值 6、寫入值、欄位位移、byte 截斷。
 *   強推論（docs/re/170、docs/internals/54）：+0x79 == 6 = 「已經到了」(docs/re/170 第 91 行)，+0x08 是行動碼
 *     （1 = 出擊／增援）、+0x07 是目標星域、+0x79 = 1 是「已下命令」。名稱不是證據，只有位移與常數是證據。
 *
 * 輸入域：list 取暫存緩衝區（任意內容；項目可指到任意位址）或種子指定位置；count 低位元組 0..255，高位元組任意；
 *   target 任意。word 存取在位移 0xFFFF 時，dosgolem 的高位元組取同一段的位移 0（w0C27_rd16s）。
 *   線性 0xF0000 以上是 ROM，寫入在 dosgolem 被安靜丟棄（internal/machine/bios_rom.go），用 w0C27_put8 表達。
 *   讀到 A0000 到 BFFFF 的案例由探針丟棄；寫到那裡的平面模式案例同樣丟棄。
 */
#include "gin3_reg.h"

/* 段內 word 讀取：位移各自 16 位元回繞（seg:FFFF 的高位元組在 seg:0000）。 */
static uint16_t w0C27_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效。 */
static void w0C27_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_2E54_0C27(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w, uint16_t target_w)
{
    uint8_t count = (uint8_t)count_w;
    uint16_t i;

    for (i = 0; i < count; i++) {
        uint16_t si = (uint16_t)(list_off + (uint16_t)(i << 2));
        uint16_t f_seg = w0C27_rd16s(m, list_seg, (uint16_t)(si + 2u));
        uint16_t f_off = w0C27_rd16s(m, list_seg, si);

        if (rd8(m, LIN(f_seg, f_off + 0x79u)) == 6) {
            w0C27_put8(m, LIN(f_seg, f_off + 8u), 1);
            w0C27_put8(m, LIN(f_seg, f_off + 7u), (uint8_t)target_w);
            w0C27_put8(m, LIN(f_seg, f_off + 0x79u), 1);
        }
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2E54_0C27(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("2E54_0C27", 4, run)
