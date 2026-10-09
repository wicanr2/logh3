/*
 * GIN3PS.EXE 25DB:0221  把在場艦隊（清單 ds:9898 的艦隊序號）的顯示物件全部從清單移除並清掉指標：艦隊圖示、各分隊的艦艇、旗艦的艦艇
 *
 * 原版（306 bytes，far 呼叫，沒有參數；序言 push bp／si、sub sp,28h、mov bp,sp），無有意義的回傳值（描述檔 ret ＝ none）。
 *   區域變數：[bp+0] ＝ 迴圈 i（byte）、[bp+1] ＝ 分隊序號 j（byte）、[bp+2]／[bp+4] ＝ 艦隊序號清單指標（位移／段，段 ＝ DS）、
 *   [bp+6]／[bp+8] ＝ 艦隊記錄指標 f（位移／段）、[bp+0Ah]／[bp+0Ch] ＝ 分隊槽指標、[bp+0Eh]／[bp+10h] ＝ 艦艇指標（只存不讀）、
 *   [bp+16h]／[bp+1Eh]／[bp+26h] ＝ 呼叫 35D1:0102 前存起來的 ES（清指標時用）。
 * 過程（已證實：反組譯逐指令；35D1:0102 由其差異測試證實）：
 *   i ← 0；清單指標 ← DS:0x9898
 *   重複：
 *     i >= word ds:9894（無號 jb 的反面，筆數每圈重讀）→ 結束
 *     idx ← byte 清單指標[0]；f ← 0xC074 ＋ low16(idx × 0x7A)（DS 內，位移 16 位元回繞）
 *     釋放(DS, f ＋ 0x6C)                                  ; 艦隊圖示物件
 *     j ← 0；槽 ← f ＋ 0x10
 *     j < f[+0x0F]（無號 byte，f 每圈重讀）時：
 *       s ← word 槽[0]；艦艇 ← (word ds:CFB4 ＋ low16(s × 0x4E)) : word ds:CFB6   ; 池基底每圈重讀（先讀段、後讀位移）
 *       釋放(艦艇的段, 艦艇 ＋ 0x38)
 *       j++；槽 ＋= 4
 *     s ← word f[+0x38]（迴圈之後才讀）；艦艇 ← 同上算法；釋放(艦艇的段, 艦艇 ＋ 0x38)  ; 旗艦的艦艇
 *     i++；清單指標 ＋= 1
 *   釋放(seg, q)：ax ← word seg:q[+2]、bx ← word seg:q[0]；(ax | bx) ≠ 0 → 35D1:0102(bx, ax)（push 段、push 位移）；之後 word seg:q[0] ← 0、seg:q[+2] ← 0
 *     （q 與 ES 在呼叫前存起來；si 被 0102 保存）
 * 語意（強推論）：呼叫者 12AA:01E3、2C5D:06E2（台帳）；35D1:0102 只從清單移除、不釋放記憶體，這裡也不釋放，所以「釋放」是「從畫面清單移除」的意思。
 *   docs/internals/62-strategic-attrition.md 第 31 行（2AD2:01EF 對每組與 +0x38 那筆作廢）與 docs/re/329-fleet-disband-is-garrison.md 第 71、93 行
 *   （艦隊解散時 +0x6C 遠指標不為 0 → 35D1:0102）是同一類操作；docs 沒有 25DB:0221 本身的記載，艦隊 +0x38 是旗艦艦艇序號是強推論。
 * 副作用：35D1:0102 的全部寫入（物件 +16..+27、+1E..+22，格表 ds:D81E 起，清單前移與 ds:EADE 減一）、艦隊 +0x6C 起 4 bytes、各艦艇 +0x38 起 4 bytes 清 0。
 *
 * 機器模型：C 版沒有堆疊；原版區域變數都存在堆疊（探針堆疊帶內，比對時排除）。word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 的
 *   read16／write16），用本檔的 w0221_rd16s／w0221_wr16s。
 * 輸入域：艦隊序號清單 ds:9898 起 256 bytes、池基底 ds:CFB4..CFB7、seeds 用到的艦隊記錄（+0x0F 到 +0x3C）與 +0x6C 指標、seeds 用到的艦艇記錄 +0x38 指標
 *   列入 freeze：這些位置的值會被當指標去改寫物件，隨機突變會指向程式碼、堆疊帶或視訊記憶體區，那些案例原版與 C 版不可比。物件本身與清單 ds:EAE0 取突變。
 *   筆數 ds:9894 超過 255 時 i（byte）回繞，原版不終止，由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_35D1_0102(Mem *m, uint16_t obj_off, uint16_t obj_seg);

static uint16_t w0221_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w0221_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

/* 釋放(seg, q)：seg:q 起的 far 指標非 0 就交給 35D1:0102，之後清 0 */
static void w0221_release(Mem *m, uint16_t seg, uint16_t q)
{
    uint16_t ax = w0221_rd16s(m, seg, (uint16_t)(q + 2u));
    uint16_t bx = w0221_rd16s(m, seg, q);

    if ((uint16_t)(ax | bx) != 0) {
        (void)f_35D1_0102(m, bx, ax);
        w0221_wr16s(m, seg, q, 0);
        w0221_wr16s(m, seg, (uint16_t)(q + 2u), 0);
    }
}

/* 艦艇池第 s 筆的 +0x38：回傳 q（位移），*seg 是池的段；段先讀、位移後讀，與原版相同 */
static uint16_t w0221_ship_slot(const Mem *m, uint16_t s, uint16_t *seg)
{
    uint16_t pseg = w0221_rd16s(m, GIN3_DS, 0xCFB6u);
    uint16_t poff = w0221_rd16s(m, GIN3_DS, 0xCFB4u);

    *seg = pseg;
    return (uint16_t)(poff + (uint16_t)(s * 0x4Eu) + 0x38u);
}

void f_25DB_0221(Mem *m)
{
    uint8_t i = 0;
    uint16_t lp = 0x9898u;

    while (i < w0221_rd16s(m, GIN3_DS, 0x9894u)) {
        uint8_t idx = rd8(m, LIN(GIN3_DS, lp));
        uint16_t f = (uint16_t)(0xC074u + (uint16_t)(idx * 0x7Au));
        uint8_t j = 0;
        uint16_t slot = (uint16_t)(f + 0x10u);
        uint16_t seg, q;

        w0221_release(m, GIN3_DS, (uint16_t)(f + 0x6Cu));
        while (j < rd8(m, DS((uint16_t)(f + 0x0Fu)))) {
            q = w0221_ship_slot(m, w0221_rd16s(m, GIN3_DS, slot), &seg);
            w0221_release(m, seg, q);
            j = (uint8_t)(j + 1u);
            slot = (uint16_t)(slot + 4u);
        }
        q = w0221_ship_slot(m, w0221_rd16s(m, GIN3_DS, (uint16_t)(f + 0x38u)), &seg);
        w0221_release(m, seg, q);
        i = (uint8_t)(i + 1u);
        lp = (uint16_t)(lp + 1u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_25DB_0221(m);
    return 0;
}
GIN3_REGISTER("25DB_0221", 0, run)
