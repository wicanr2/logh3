/*
 * GIN3PS.EXE 090A:0B93  目前陣營所有星域的稅率重算（寫各星域 +0x26，無回傳值）
 *
 * 原版：far 呼叫，一個參數 level（word，序言 push bp／si／di、sub sp,12h，所以在 [bp+1Ch]）。沒有回傳值。
 *
 * 反組譯逐條對應：
 *   side ← byte ds:8F1D；統治者 ← ds:B77C ＋ byte ds:[8F4F ＋ side] × 0x1C（16 位元）
 *   對星域 i ＝ 0 到 31（word 有號比較，記錄 ds:A79C ＋ i × 0x7E，16 位元位移）：
 *     星域[+0x12] != byte ds:8F1D（每圈重讀）→ 下一個
 *     base ← dword ds:[(level << 2) ＋ 0x147E／0x1480]（位移 16 位元回繞）；g ← 有號 word 星域[+0x1E]
 *     g <= 0：v ← (base × 5) ÷ 100 → 直接寫入，不夾限
 *     g > 0 ：
 *       v ← (sext32(g) × base) ÷ 100
 *       v ← (v × 100) ÷ 統治者[+0x0E]                （byte，零延伸；除數 0 時 3CC9:0070 回商 0xFFFFFFFF 再依符號補回，不引發 #DE）
 *       v ← (v × 統治者[+0x15]) ÷ 100
 *       A ← (side == 0) ? 星域[+0x7C] : (0xC8 − 星域[+0x7C])   （16 位元，byte 零延伸；side != 0 且 byte > 200 時回繞成大數）
 *       m ← ((A × 統治者[+0x14]) 的低 16 位元) ÷ 100          （無號 16 位元；`mul` 之後 `xor dx,dx` 丟掉高位）
 *       m > 150 → v ← (v × 30) ÷ 統治者[+0x14]
 *       100 < m <= 150 → v ← (v × 50) ÷ 統治者[+0x14]
 *       50 < m <= 100 → v ← (v × 75) ÷ 統治者[+0x14]
 *       m <= 50 → v 不變
 *       v >= 200（有號 32 位元）→ v ← 200；否則 v <= 0 → v ← 1
 *     星域[+0x26] ← v 的低位元組
 *   所有乘法走 3CC9:0009（低 32 位元）、除法走 3CC9:0070（有號、向零取整），夾限的比較走 3CC9:00B9（旗標給 jl／jg）。
 *   本檔直接用 gin3_lmul、gin3_ldivmod 與 32 位元有號比較表達同一個結果；三支 runtime 函式各自 diff-tested。
 *
 * 與舊文件的差異
 *   docs/re/223 的虛擬碼把「v > 200 → 200、v <= 0 → 1」放在兩個分支之後共同套用；原版 g <= 0 的分支在
 *   存入前直接跳到寫入（loc_169B5），不夾限。夾限只用於 g > 0 的分支。
 *
 * 推論等級
 *   已證實（差異測試）：上列結構、常數、運算寬度與符號、16 位元回繞、g <= 0 不夾限。
 *   強推論（docs/re/223、docs/internals/55、docs/internals/08）：星域 +0x12 是擁有者、+0x1E 是成長率、+0x26 是稅率；
 *     ds:B77C 起每筆 28 bytes 是提督記錄、ds:8F4F 是各陣營內政專任參謀的記錄編號；統治者 +0x0E 是營運、
 *     +0x14 與 +0x15 是民心相關能力（名稱為假說）；星域 +0x7C 是民心類數值（side 1 取 200 − 值，假說）；
 *     ds:147E 起是 32 位元基價表（50 75 100 150 200），level 是難度等級（呼叫者傳 ds:8FD4[陣營]）。
 *   未知：m 的三段門檻（150／100／50）與倍率（30／50／75）的設計意圖。
 *
 * 輸入域：level 任意 word；ds:8F1D 任意 byte；記錄與表內容任意。字的存取跨越位移 0xFFFF 時，本版與探針都按線性位址
 *   連續處理，不涵蓋真機上回繞到段首的行為（本函式所有位址都是 DS 內固定位移加小偏移，只有 level 與
 *   ds:8F1D 造成的 `ds:[8F4F + side]` 與 `(level << 2) + 0x147E` 會落到高位址）。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

static uint32_t w0B93_div(uint32_t a, uint32_t b)
{
    uint32_t q, r;

    gin3_ldivmod(a, b, &q, &r);
    return q;
}

void f_090A_0B93(Mem *m, uint16_t level)
{
    uint16_t side = rd8(m, DS(0x8F1D));
    uint16_t ruler = (uint16_t)(0xB77Cu + (uint16_t)(rd8(m, DS((uint16_t)(0x8F4Fu + side))) * 0x1Cu));
    uint16_t reg = 0xA79C;
    int16_t i;

    for (i = 0; i < 0x20; i++) {
        if (rd8(m, DS(reg + 0x12)) == rd8(m, DS(0x8F1D))) {
            uint16_t at = (uint16_t)(level << 2);
            uint32_t base = ((uint32_t)rd16(m, DS((uint16_t)(at + 0x1480u))) << 16) | rd16(m, DS((uint16_t)(at + 0x147Eu)));
            int16_t g = (int16_t)rd16(m, DS(reg + 0x1E));
            uint32_t v;

            if (g <= 0) {
                v = w0B93_div(gin3_lmul(base, 5u), 100u);
            } else {
                uint16_t a16, m16;

                v = w0B93_div(gin3_lmul((uint32_t)(int32_t)g, base), 100u);
                v = w0B93_div(gin3_lmul(v, 100u), rd8(m, DS(ruler + 0x0E)));
                v = w0B93_div(gin3_lmul(v, rd8(m, DS(ruler + 0x15))), 100u);
                if (rd8(m, DS(0x8F1D)) == 0)
                    a16 = rd8(m, DS(reg + 0x7C));
                else
                    a16 = (uint16_t)(0xC8u - rd8(m, DS(reg + 0x7C)));
                m16 = (uint16_t)((uint16_t)(a16 * rd8(m, DS(ruler + 0x14))) / 100u);
                if (m16 > 0x96)
                    v = w0B93_div(gin3_lmul(v, 30u), rd8(m, DS(ruler + 0x14)));
                else if (m16 > 0x64)
                    v = w0B93_div(gin3_lmul(v, 50u), rd8(m, DS(ruler + 0x14)));
                else if (m16 > 0x32)
                    v = w0B93_div(gin3_lmul(v, 75u), rd8(m, DS(ruler + 0x14)));
                if ((int32_t)v >= 200)
                    v = 200;
                else if ((int32_t)v <= 0)
                    v = 1;
            }
            wr8(m, DS(reg + 0x26), (uint8_t)v);
        }
        reg = (uint16_t)(reg + 0x7E);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_090A_0B93(m, a[0]);
    return 0;
}
GIN3_REGISTER("090A_0B93", 1, run)
