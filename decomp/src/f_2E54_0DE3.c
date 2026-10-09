/*
 * GIN3PS.EXE 2E54:0DE3  從候選陣列尾端挑 1 或 2 支艦隊派去支援目標星域（far，三個參數，無回傳值；ret = none）
 *
 * 參數（堆疊字，由低位址到高位址；序言 push bp／si、sub sp,6，所以第一個參數在 [bp+0Eh]）：
 *   list     far 指標（位移、段）：艦隊 far 指標的陣列，每筆 4 bytes（位移 word 在前、段 word 在後）
 *   count    一個字，只用低位元組（cmp byte [bp+12h],2；mov bl,[bp+12h]）：筆數
 *   target   一個字，只用低位元組（mov bl,[bp+14h]；mov dl,[bp+14h]）：目標星域編號
 *
 * 反組譯逐條對應（IDA sub_3BE93）：
 *   word ds:D54E[target] >= 0x0C（無號，jb 反向）→ 返回，不呼叫 rand
 *   r ← 3CD5:000C()                                 ; rand()，0..0x7FFF
 *   n ← (r >= 0x2000（有號 jge，r 非負故同無號）) ? 1 : (count >= 2 ? 2 : 1)
 *   對 j = 0 … n−1（byte 計數）：
 *     fleet ← list 的第 (count − 1 − j) 筆（位置 list.off + count × 4 − 4 − j × 4，16 位元回繞；
 *             位移 word、段 word 各讀一次）          ; 從陣列尾端往前取
 *     kind ← fleet[+0x09]（byte）
 *     ds:5F52[kind]（byte）== 3 → 跳過（不派）
 *     否則：fleet[+0x08] ← 1；fleet[+0x07] ← target；fleet[+0x79] ← 1
 *   每一圈都重讀陣列項，陣列與艦隊記錄重疊時，前一圈的寫入會影響後一圈讀到的指標（別名）。
 *   count 小於 2 時 n 恆為 1；count 為 0 時讀的是 list.off − 4 的那一筆（16 位元回繞）。
 *
 * 推論等級
 *   已證實（差異測試的位元組、寫入集合、亂數種子 ds:94DC..94DF 的寫入）：上列分支結構、常數、尾端取法、
 *     rand 只在閘門通過後才呼叫、byte 截斷。
 *   強推論（docs/internals/54「派支援」、docs/re/168）：D54E = 需求（敵方下一步會在這一格的槽數）、0x0C = 12 槽的上限；
 *     fleet[+0x09] 是艦隊類別，ds:5F52 是類別表（類別 3 不派）；+0x08 = 1 是出擊／增援、+0x07 是目標星域、
 *     +0x79 = 1 是「已下命令」。名稱不是證據，只有位移與常數是證據。
 *
 * 輸入域：list 取暫存緩衝區（任意內容；項目可指到任意位址）或種子指定位置；count 低位元組 0..255，高位元組任意；
 *   target 低位元組 0..255（D54E 讀到 D54E + 2 × 255 為止，表只有 32 筆，後面是鄰近記憶體，原版同樣如此）。
 *   word 存取在位移 0xFFFF 時，dosgolem 的高位元組取同一段的位移 0（w0DE3_rd16s）。
 *   線性 0xF0000 以上是 ROM，寫入在 dosgolem 被安靜丟棄（internal/machine/bios_rom.go），用 w0DE3_put8 表達。
 *   讀到 A0000 到 BFFFF 的案例由探針丟棄。
 */
#include "gin3_reg.h"

extern uint16_t f_3CD5_000C(Mem *m);

/* 段內 word 讀取：位移各自 16 位元回繞（seg:FFFF 的高位元組在 seg:0000）。 */
static uint16_t w0DE3_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效。 */
static void w0DE3_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_2E54_0DE3(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w, uint16_t target_w)
{
    uint8_t count = (uint8_t)count_w;
    uint8_t target = (uint8_t)target_w;
    uint16_t r;
    uint8_t n, j;

    if (rd16(m, DS((uint16_t)(0xD54Eu + 2u * target))) >= 0x0Cu)
        return;
    r = f_3CD5_000C(m);
    if (r >= 0x2000u || count < 2)
        n = 1;
    else
        n = 2;
    for (j = 0; j < n; j++) {
        uint16_t si = (uint16_t)(list_off + (uint16_t)(count << 2) + 0xFFFCu - (uint16_t)(j << 2));
        uint16_t f_seg = w0DE3_rd16s(m, list_seg, (uint16_t)(si + 2u));
        uint16_t f_off = w0DE3_rd16s(m, list_seg, si);
        uint8_t kind = rd8(m, LIN(f_seg, f_off + 9u));

        if (rd8(m, DS((uint16_t)(0x5F52u + kind))) != 3) {
            w0DE3_put8(m, LIN(f_seg, f_off + 8u), 1);
            w0DE3_put8(m, LIN(f_seg, f_off + 7u), target);
            w0DE3_put8(m, LIN(f_seg, f_off + 0x79u), 1);
        }
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2E54_0DE3(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("2E54_0DE3", 4, run)
