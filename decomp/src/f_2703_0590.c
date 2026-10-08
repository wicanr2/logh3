/*
 * GIN3PS.EXE 2703:0590  把星域層的待命存量，按行星經濟力的整數百分比分回同陣營的行星
 *
 * 原版：far 呼叫，兩個參數：星域記錄 far 指標（位移在前、段在後）、陣營（word，只讀低位元組）；
 * 無有意義的回傳。呼叫端 2703:0540 對陣營 0、1 各呼叫一次。
 * 語意（已證實：反組譯逐指令；docs/spec/039、docs/internals/22 一致，補充了下列細節）：
 *   星域的來源欄位：七種艦艇存量 S[k]＝dword(星域+0x30 + 陣營×28 + k×4)，k＝0..6；
 *     A＝dword(星域+0x68 + 陣營×4)（飛彈池）、B＝dword(星域+0x70 + 陣營×4)（艦載機池）。
 *     名稱來自 docs/spec/039、docs/internals/22（強推論：欄位本義未獨立驗證，運算本身逐指令對過）。
 *   行星切片＝ds:CFF8（位移）／ds:CFFA（段）起、ds:CFFC（byte）筆，每筆 82 bytes（與傳入的星域無關）。
 *   total ＝ 同陣營（行星[+0x11] == 陣營）各行星 dword(+0x1A)（經濟力）的 32 位元和。
 *   第二趟逐筆同陣營行星 P：
 *     econ ＝ dword(P+0x1A)；若「有號」econ >= best（best 初始 0）則 best ← P（同值後者取代，負值永遠不選）
 *     pct  ＝ 有號除法((econ × 100 的低 32 位元) ÷ total)         （3CC9:0070；total 為 0 時商為 0xFFFFFFFF）
 *     對 k＝0..6：q ＝ ((星域 S[k] 現值 × pct) 的低 32 位元) ÷ 100（無號，3CC9:0020）
 *                 dword(P+0x20+4k) ← q；局部存量 L[k] −= q
 *     dword(P+0x3C) ← (A 現值 × pct 的低 32 位元) ÷ 100；dword(P+0x40) ← (B 現值 × pct) ÷ 100
 *     A餘 −= dword(P+0x3C)；B餘 −= dword(P+0x40)
 *   最後若 best 非空：dword(best+0x20+4k) += L[k]（k＝0..6）；dword(best+0x3C) += A餘；dword(best+0x40) += B餘。
 *   即逐顆整數捨去後，餘額全給經濟力最大的行星（同值取最後一顆）。
 * 與 docs/spec/039 的差異：asm 沒有「total 為 0 就不分配」的判斷；total 為 0 時照走 3CC9:0020 的除零路徑
 *   （商 0xFFFFFFFF），在遊戲內是否可達為 unknown。docs/spec/039 的「為 0 就不分配」描述的是 remake 規則。
 * 局部存量 L：原版用 3CA9:0002 把 28 bytes 搬到堆疊（ss:bp+0x12）。堆疊不在差異測試的比對範圍，
 *   C 版用 L[7] 局部陣列取代，來源讀取位址與原版相同（星域+0x30+陣營×28 起 28 bytes）。
 * 副作用：寫行星 +0x20..+0x3B、+0x3C..+0x3F、+0x40..+0x43；星域記錄只讀。
 * 輸入域：行星切片與星域記錄互不重疊（遊戲內如此）；步數超過 200,000 的執行（切片筆數很大）由探針丟棄。
 *   指標的 dword 不跨越 64KB 段尾（記憶體模型不做段內回繞）。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

#define F_SF 0x0080u
#define F_OF 0x0800u

static uint32_t rdw32(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint32_t)rd16(m, LIN(seg, off)) | ((uint32_t)rd16(m, LIN(seg, (uint16_t)(off + 2))) << 16);
}

static void wrw32(Mem *m, uint16_t seg, uint16_t off, uint32_t v)
{
    wr16(m, LIN(seg, off), (uint16_t)v);
    wr16(m, LIN(seg, (uint16_t)(off + 2)), (uint16_t)(v >> 16));
}

/* (v × pct) 的低 32 位元，再無號除以 100（3CC9:0009 接 3CC9:0020） */
static uint32_t scale100(uint32_t v, uint32_t pct)
{
    uint32_t q, r;

    gin3_uldivmod(gin3_lmul(v, pct), 100, &q, &r);
    return q;
}

void f_2703_0590(Mem *m, uint16_t region_off, uint16_t region_seg, uint16_t side_arg)
{
    uint8_t side = (uint8_t)side_arg;
    uint16_t side_w = side;
    uint16_t side28 = (uint16_t)(side_w * 0x1C);
    uint16_t side4 = (uint16_t)(side_w << 2);
    uint16_t best_off = 0, best_seg = 0;
    uint32_t best_econ = 0;
    uint32_t stock[7];
    uint32_t a_rem, b_rem, total = 0;
    uint16_t src = (uint16_t)(region_off + 0x30);
    uint16_t p_off, p_seg;
    uint8_t i;
    int k;

    /* 3CA9:0002(ss:bp+0x12, 星域+0x30+陣營×28, 0x1C)：28 bytes 搬到堆疊局部存量 */
    src = (uint16_t)(src + side28);
    for (k = 0; k < 7; k++) {
        uint32_t v = 0;
        int b;

        for (b = 3; b >= 0; b--)
            v = (v << 8) | rd8(m, LIN(region_seg, (uint16_t)(src + 4 * k + b)));
        stock[k] = v;
    }
    a_rem = rdw32(m, region_seg, (uint16_t)((uint16_t)(region_off + 0x68) + side4));
    b_rem = rdw32(m, region_seg, (uint16_t)((uint16_t)(region_off + 0x70) + side4));

    p_off = rd16(m, DS(0xCFF8));
    p_seg = rd16(m, DS(0xCFFA));
    for (i = 0; i < rd8(m, DS(0xCFFC)); i++) {
        if (rd8(m, LIN(p_seg, (uint16_t)(p_off + 0x11))) == side)
            total += rdw32(m, p_seg, (uint16_t)(p_off + 0x1A));
        p_off = (uint16_t)(p_off + 0x52);
    }

    p_off = rd16(m, DS(0xCFF8));
    p_seg = rd16(m, DS(0xCFFA));
    for (i = 0; i < rd8(m, DS(0xCFFC)); i++) {
        if (rd8(m, LIN(p_seg, (uint16_t)(p_off + 0x11))) == side) {
            uint32_t econ = rdw32(m, p_seg, (uint16_t)(p_off + 0x1A));
            uint16_t flags = f_3CC9_00B9((uint16_t)(econ >> 16), (uint16_t)econ, (uint16_t)(best_econ >> 16),
                                         (uint16_t)best_econ);
            int less = ((flags & F_SF) != 0) != ((flags & F_OF) != 0); /* jl */
            uint32_t pct, qd, rem;
            uint16_t dst = (uint16_t)(p_off + 0x20);
            uint32_t f3c, f40;

            if (!less) {
                best_seg = p_seg;
                best_off = p_off;
                best_econ = econ;
            }
            gin3_ldivmod(gin3_lmul(econ, 100), total, &qd, &rem);
            pct = qd;
            for (k = 0; k < 7; k++) {
                uint32_t v = rdw32(m, region_seg, (uint16_t)((uint16_t)(region_off + 0x30) + side28 + (uint16_t)(4 * k)));
                uint32_t q = scale100(v, pct);

                wrw32(m, p_seg, (uint16_t)(dst + 4 * k), q);
                stock[k] = stock[k] - q;
            }
            f3c = scale100(rdw32(m, region_seg, (uint16_t)((uint16_t)(region_off + 0x68) + side4)), pct);
            wrw32(m, p_seg, (uint16_t)(p_off + 0x3C), f3c);
            f40 = scale100(rdw32(m, region_seg, (uint16_t)((uint16_t)(region_off + 0x70) + side4)), pct);
            wrw32(m, p_seg, (uint16_t)(p_off + 0x40), f40);
            a_rem = a_rem - rdw32(m, p_seg, (uint16_t)(p_off + 0x3C));
            b_rem = b_rem - f40;
        }
        p_off = (uint16_t)(p_off + 0x52);
    }

    if (best_off != 0 || best_seg != 0) {
        for (k = 0; k < 7; k++) {
            uint16_t at = (uint16_t)(best_off + 0x20 + 4 * k);

            wrw32(m, best_seg, at, rdw32(m, best_seg, at) + stock[k]);
        }
        wrw32(m, best_seg, (uint16_t)(best_off + 0x3C), rdw32(m, best_seg, (uint16_t)(best_off + 0x3C)) + a_rem);
        wrw32(m, best_seg, (uint16_t)(best_off + 0x40), rdw32(m, best_seg, (uint16_t)(best_off + 0x40)) + b_rem);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2703_0590(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("2703_0590", 3, run)
