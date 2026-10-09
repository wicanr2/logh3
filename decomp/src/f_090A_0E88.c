/*
 * GIN3PS.EXE 090A:0E88  某陣營（座位）的回合預算：先算維持費與移動經費（297B:10DE），再算「收入 − 支出」依撥款比率分出的預算
 *
 * 原版（253 bytes，far 呼叫，一個堆疊字 f；push bp／si／di、sub sp,6，所以 f 在 [bp+10h]）：
 *   up ← 297B:10DE(f)                                  ; 已 diff-tested：回 AX:BX ＝ 該陣營的維持費合計（32 位元），並寫
 *                                                       ; ds:8F82 ＋ 4 × (f & 0xFF) 的移動經費合計（10DE 只取 f 的低位元組）
 *   s ← f << 2（整個 word，16 位元）                   ; 之後所有表都以 s ＝ 4 × f 當索引（f 的高位元組有效）
 *   ds:(s ＋ 0x8F74) ← up 高字；ds:(s ＋ 0x8F72) ← up 低字      ; `mov [si-708Ch],ax`；`mov [si-708Eh],bx`（先高字後低字）
 *   net ← dword ds:(s ＋ 0x8F52) ＋ dword ds:(s ＋ 0x8F6A) − dword ds:(s ＋ 0x8F72) − dword ds:(s ＋ 0x8F82)   ; add／adc、sub／sbb，32 位元回繞
 *   3CC9:00B9(net, 0:0)：結果旗標 SF ＝ 1（jns 不成立）→ net 為負：dword ds:(s ＋ 0x8F5A) ← 0（兩個 word 都寫 0），結束
 *   否則 3CC9:00B9(net, 0x0147:0xAE14)：jl 成立（net < 0x0147AE14，有號 32 位元）：
 *       prod ← net × (byte ds:(f ＋ 0x8F92))（3CC9:0009，32 位元回繞）；out ← prod ÷ 100（3CC9:0070，有號）
 *     否則：q ← net ÷ 100（有號）；out ← q × (byte ds:(f ＋ 0x8F92))（回繞）
 *     dword ds:(s ＋ 0x8F5A) ← out（先寫高字 ds:(s ＋ 0x8F5C)、再寫低字 ds:(s ＋ 0x8F5A)）
 *   無回傳值（描述檔 ret ＝ none）。比率的 byte 以 f（不是 s）當索引。
 * 兩條路徑用不同的先乘後除、先除後乘，目的是避開 32 位元有號溢位：門檻 0x0147AE14 ＝ 21,474,836 ＝ floor(2^31 ÷ 100)。比率 byte 超過 100 時，
 *   net 略小於門檻的乘積仍會超過 2^31（當有號數變負）或 2^32（回繞），本版與原版一致地照 32 位元運算。
 *
 * 所有表位移都是偶數（s 是 4 的倍數、常數是偶數），word 存取不會落在位移 0xFFFF，所以不必用段內回繞的 word 存取。
 * 呼叫者（台帳）：090A:0E6D（hw-dependent）。docs/internals/27-treasury-globals.md 第 54、140、141 行「`ds:8F5A` 的算式
 *   （`090A:0E88(陣營)`）：297B:10DE 算維持費 → ds:8F72／ds:8F82；ds:8F52 + ds:8F6A − ds:8F72 − ds:8F82 → ds:8F5A」，
 *   docs/spec/042-region-unrest-recompute.md 第 51 行與本檔一致。補充：負值寫 0；net ≥ 0x0147AE14 時先除後乘（避免溢位），
 *   其餘先乘後除；乘積在比率 > 199 時仍可能回繞（docs 沒有記載）。各表的語意（國庫、收入、維持費、移動經費、撥款比率）
 *   取自 docs/internals/27，名稱不是證據，位移與運算是證據。
 * 輸入域：f 全部 16 位元值（f >= 256 時 10DE 只用低位元組，表索引用整個 word，位移 16 位元回繞）；艦隊表與艦艇池取 S3 基底。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位語意為強推論。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

extern uint32_t f_297B_10DE(Mem *m, uint16_t faction_word);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

#define F_SF 0x0080u
#define F_OF 0x0800u

static uint16_t w0E88_r16(const Mem *m, uint16_t s, uint16_t k) { return rd16(m, DS((uint16_t)(s + k))); }

static uint32_t w0E88_r32(const Mem *m, uint16_t s, uint16_t lo)
{
    return (uint32_t)w0E88_r16(m, s, lo) | ((uint32_t)w0E88_r16(m, s, (uint16_t)(lo + 2u)) << 16);
}

void f_090A_0E88(Mem *m, uint16_t f)
{
    uint32_t up = f_297B_10DE(m, f);
    uint16_t s = (uint16_t)(f << 2);
    uint32_t net, out, q, r;
    uint16_t fl;

    wr16(m, DS((uint16_t)(s + 0x8F74u)), (uint16_t)(up >> 16));
    wr16(m, DS((uint16_t)(s + 0x8F72u)), (uint16_t)up);
    net = w0E88_r32(m, s, 0x8F52u) + w0E88_r32(m, s, 0x8F6Au);
    net = net - w0E88_r32(m, s, 0x8F72u);
    net = net - w0E88_r32(m, s, 0x8F82u);

    fl = f_3CC9_00B9((uint16_t)(net >> 16), (uint16_t)net, 0, 0);
    if ((fl & F_SF) != 0) {
        wr16(m, DS((uint16_t)(s + 0x8F5Cu)), 0);
        wr16(m, DS((uint16_t)(s + 0x8F5Au)), 0);
        return;
    }
    fl = f_3CC9_00B9((uint16_t)(net >> 16), (uint16_t)net, 0x0147u, 0xAE14u);
    if (((fl & F_SF) != 0) != ((fl & F_OF) != 0)) {
        uint32_t prod = gin3_lmul(net, rd8(m, DS((uint16_t)(f + 0x8F92u))));

        gin3_ldivmod(prod, 100u, &q, &r);
        out = q;
    } else {
        gin3_ldivmod(net, 100u, &q, &r);
        out = gin3_lmul(q, rd8(m, DS((uint16_t)(f + 0x8F92u))));
    }
    wr16(m, DS((uint16_t)(s + 0x8F5Cu)), (uint16_t)(out >> 16));
    wr16(m, DS((uint16_t)(s + 0x8F5Au)), (uint16_t)out);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_090A_0E88(m, a[0]);
    return 0;
}
GIN3_REGISTER("090A_0E88", 1, run)
