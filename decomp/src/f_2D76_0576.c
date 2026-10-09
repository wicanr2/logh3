/*
 * GIN3PS.EXE 2D76:0576  為提督抽一個「艦隊類別鍵」（0 到 15）寫進提督記錄 +8 的高四位
 *
 * 原版：far 呼叫，兩個堆疊字（push bp、push si、sub sp,0Eh，所以參數從 [bp+16h] 起），無回傳值
 *   （AX 是殘留值，呼叫者 2D76:02FA 不用；描述檔 ret=none）：
 *   args[0] ＝ idx（word，只取低位元組：mov al,[bp+16h]）＝ 提督編號；記錄在 ds:(B77C + idx × 28)（idx × 28 是 16 位元 mul）
 *   args[1] ＝ kind（word，只取低位元組：mov bl,[bp+18h]）＝ 艦隊類別 0 到 3
 *
 * 過程：
 *   rec ← ds:(0xB77C + idx × 28)
 *   kind 的低位元組 <= 3：查跳躍表 ds:[602A + kind × 2]（本檔讀記憶體裡的位址，對應到四個 case 的程式位移），選一張鍵清單
 *     0x05A4 → ds:5F62（類別 0）　0x05AE → ds:5F70（類別 1）　0x05B8 → ds:5F6C（類別 2）　0x05C2 → ds:5F74（類別 3）
 *   清單格式：第 1 個 byte 是筆數 n，後面接 n 個 byte 的鍵。
 *   n ← 清單[0]；rec[+8] ← rec[+8] & 0x0F（先清高四位）；
 *   r ← rand()（3CD5:000C，AX，15 位元）；k ← 清單[1 + (r % n)]（n 是 byte，除法 DX:AX ÷ CX，DX 先清 0）；
 *   rec[+8] ← rec[+8] | (k << 4)（byte 左移，只留低位元組）
 * 副作用：寫 rec[+8] 兩次（先 and 後 or）、呼叫 rand 一次（種子 ds:94DC 被寫）。
 *
 * 與原版的偏離：kind 的低位元組 > 3 時本版直接返回，不做 and、rand 與 or；原版照常執行（見上）。該輸入不在比較範圍。
 * 輸入域：kind 的低位元組必須 <= 3（> 3 時走 default，指標是堆疊上未初始化的區域變數，不可重現）；
 *   n == 0 時原版 #DE（探針丟棄，本版回傳以避免 SIGFPE）。跳躍位址被突變成其他值時，原版跳到不相干的程式位置，
 *   本版視同不做事；這類案例若被探針收錄就是輸入域外（絕大多數會因 INT、步數上限或讀視訊記憶體被丟棄）。
 * 與舊文件的出入：docs/re/204 第 15 行寫「類別 > 3 → 跳過設定」，與反組譯不符：cmp bx,3；ja 跳到的 def_3A870 是 switch 之後的共用尾段
 *   （不是函式結尾），and、rand、or 照常執行，只是清單指標是未初始化的堆疊區域變數。呼叫端只給 0 到 3，所以實際不會發生。
 * 欄位語意（強推論）：rec[+8] 高四位 ＝ 艦隊類別鍵（docs/internals/08 第 38 行、docs/re/204、287 第 19 行）；清單內容與 n 取自記憶體（基底狀態與突變）。
 * 證據：asm 逐指令；跳躍表位址與 case 本體由函式位元組核對（insns.tsv 的指令位移 +0x2E、+0x38、+0x42、+0x4C、default +0x54）；
 *   docs/internals/54 第 574 行「2D76:0576(提督, 類別) 抽類別鍵（docs/re/204）：依類別查跳躍表」與本檔一致。
 * 推論等級：已證實（反組譯逐指令＋差異測試；rand 由 3CD5:000C 的差異測試證實）。
 */
#include "gin3_reg.h"

extern uint16_t f_3CD5_000C(Mem *m);

/* 跳躍表的目標位址（程式位移）→ 鍵清單的 DGROUP 位移；0 表示不是已知的 case */
static uint16_t w0576_list(uint16_t target)
{
    switch (target) {
    case 0x05A4: return 0x5F62;
    case 0x05AE: return 0x5F70;
    case 0x05B8: return 0x5F6C;
    case 0x05C2: return 0x5F74;
    default: return 0;
    }
}

void f_2D76_0576(Mem *m, uint16_t idx_word, uint16_t kind_word)
{
    uint8_t idx = (uint8_t)idx_word;
    uint8_t kind = (uint8_t)kind_word;
    uint16_t rec = (uint16_t)(0xB77Cu + (uint16_t)(idx * 28u));
    uint16_t list;
    uint8_t n;
    uint16_t r;
    uint8_t key;

    if (kind > 3)
        return;
    list = w0576_list(rd16(m, DS(0x602A + kind * 2)));
    if (list == 0)
        return;
    n = rd8(m, DS(list));
    list = (uint16_t)(list + 1);
    wr8(m, DS(rec + 8), (uint8_t)(rd8(m, DS(rec + 8)) & 0x0Fu));
    r = f_3CD5_000C(m);
    if (n == 0)
        return;
    key = rd8(m, DS(list + (uint16_t)(r % n)));
    wr8(m, DS(rec + 8), (uint8_t)(rd8(m, DS(rec + 8)) | (uint8_t)(key << 4)));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2D76_0576(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2D76_0576", 2, run)
