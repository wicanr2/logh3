/*
 * GIN3PS.EXE 297B:0066  為一位提督建一支新艦隊：填艦隊記錄、配兩筆艦艇池記錄（旗艦與索敵分遣隊）、更新星域與陣營的艦隊計數，回艦隊槽號
 *
 * 原版：far 呼叫，一個 word 參數＝提督編號（只取低位元組：mov al,[bp+2Ch]；序言 push bp／si／di、sub sp,22h）。回傳 AL ＝ 艦隊槽號 s（AH 殘值，
 *   描述檔 ret ＝ pack:al）。提督記錄 A ＝ ds:(0xB77C + 編號 × 0x1C)（28 bytes）。
 * 過程（已證實：反組譯逐指令；297B:0209、2AD2:0029、3C9C:000A 由其各自的差異測試證實）。下列每一步的讀寫順序與反組譯一致
 *   （A、F 與池記錄別名時，前面的寫入影響後面的重讀，C 版不快取任何欄位）：
 *   s ← byte( (A[+1] ≠ 0 ? 0x10 : 0) + A[+9] )          ; 艦隊槽號（byte 相加，回繞）
 *   F ← ds:(0xC074 + s × 0x7A)                           ; 艦隊記錄 122 bytes（16 位元乘加）
 *   F[+0] ← 0；F[+0x3C] ← 0x64
 *   F[+1] ← A[+1]（重讀）；F[+4] ← 編號（byte）；F[+5] ← A[+7]；F[+6] ← A[+0x0A]
 *   F[+0x0E] ← AL of 297B:0209(word A[+6])               ; 應有組數
 *   F[+0x0F] ← 1；F[+8] ← 5；F[+7] ← 0xFF；F[+0x48] ← 4
 *   3C9C:000A(F+0x10, 0xFF, 0x28)                        ; F[+0x10..0x37] ← 0xFF
 *   F[+0x12] ← A[+8] & 0x0F；F[+0x13] ← 0x14
 *   word F[+0x10] ← 2AD2:0029(F, A[+8] & 0x0F（重讀）, 0)          ; 旗艦組（艦種 ＝ A[+8] 低 4 位）的池槽號
 *   word F[+0x38] ← 2AD2:0029(F, 7, 0xFF)                        ; 索敵分遣隊（艦種 7）的池槽號
 *   F[+0x3A] ← 0xFF；F[+0x3B] ← 0
 *   池記錄 P ← (ds:CFB6):(ds:CFB4 + 16 位元(word F[+0x38] × 0x4E))；P[+0x1E] |= 0x80
 *   星域計數：byte ds:(0xA79C + A[+0x0A] × 0x7E + 0x28 + byte ds:8F1D) 加一   ; 16 位元位址算式
 *   陣營計數：byte ds:(0x8F49 + F[+1]) 加一
 *   回 s
 * 呼叫 2AD2:0029 的參數：第三、四個參數的高位元組是殘值（被呼叫者只取低位元組），C 版給 0。
 * 語意（強推論）：docs/internals/64-ship-pool.md 第 48 行「建新艦隊：297B:0066 無條件配兩筆——旗艦（p3 ＝ 0）與索敵分遣隊」、docs/re/42 第 46 行、
 *   docs/internals/33 第 132 行（F[+0x13] ＝ 20 是 297B:0066 寫死的）、docs/re/212 第 15 行；A[+9] 是提督在陣營內的槽位，A[+1] 是陣營，
 *   A[+0x0A] 是所在星域，A[+8] 低 4 位是艦種（名稱未驗，位移與運算才是證據）。呼叫者 2D76:0611（hw-dependent）與 seg010:01BF（IDA 自動名稱）。
 * 副作用：F[+0..+0x13、+0x38..+0x3C、+0x48]；F[+0x10..+0x37] 先填 0xFF 再被覆寫；池內兩筆新記錄（2AD2:0029）、2703:1433 領出的星域待命存量；
 *   P[+0x1E]；星域計數 byte；陣營計數 byte。
 *
 * 輸入域：編號低位元組任意；s 任意（F 在 DGROUP 內 16 位元回繞，可能蓋到其他資料，照原版運算）；池指標 ds:CFB4..CFB7 與 2AD2:0029 的限制相同
 *   （池在 RAM、不在 ROM；2AD2:0029 的 C 版檔頭）。word 存取在位移 0xFFFF 時高位元組取同段位移 0（rd16s／wr16s）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_297B_0209(Mem *m, uint16_t ax_in, uint16_t level_word);
extern uint16_t f_2AD2_0029(Mem *m, uint16_t s_off, uint16_t s_seg, uint16_t kind_word, uint16_t b_word);
extern uint32_t f_3C9C_000A(Mem *m, uint16_t off, uint16_t seg, uint16_t val, uint16_t count);

static uint16_t w0066_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w0066_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

#define W0066_A(o) DS((uint16_t)(a + (o)))
#define W0066_F(o) DS((uint16_t)(f + (o)))

uint8_t f_297B_0066(Mem *m, uint16_t idx_word)
{
    uint16_t a = (uint16_t)(0xB77Cu + (uint16_t)((uint8_t)idx_word * 0x1Cu));
    uint16_t f, idx1, idx2, w, pool_off, pool_seg, cell;
    uint8_t s, al, kind;

    al = (rd8(m, W0066_A(1u)) != 0) ? 0x10u : 0u;
    al = (uint8_t)(al + rd8(m, W0066_A(9u)));
    s = al;
    f = (uint16_t)(0xC074u + (uint16_t)(s * 0x7Au));

    wr8(m, W0066_F(0u), 0);
    wr8(m, W0066_F(0x3Cu), 0x64u);
    wr8(m, W0066_F(1u), rd8(m, W0066_A(1u)));
    wr8(m, W0066_F(4u), (uint8_t)idx_word);
    wr8(m, W0066_F(5u), rd8(m, W0066_A(7u)));
    wr8(m, W0066_F(6u), rd8(m, W0066_A(0x0Au)));
    al = (uint8_t)f_297B_0209(m, 0, w0066_rd16s(m, GIN3_DS, (uint16_t)(a + 6u)));
    wr8(m, W0066_F(0x0Eu), al);
    wr8(m, W0066_F(0x0Fu), 1u);
    wr8(m, W0066_F(8u), 5u);
    wr8(m, W0066_F(7u), 0xFFu);
    wr8(m, W0066_F(0x48u), 4u);
    f_3C9C_000A(m, (uint16_t)(f + 0x10u), GIN3_DS, 0xFFu, 0x28u);
    wr8(m, W0066_F(0x12u), (uint8_t)(rd8(m, W0066_A(8u)) & 0x0Fu));
    wr8(m, W0066_F(0x13u), 0x14u);
    kind = (uint8_t)(rd8(m, W0066_A(8u)) & 0x0Fu);
    idx1 = f_2AD2_0029(m, f, GIN3_DS, kind, 0);
    w0066_wr16s(m, GIN3_DS, (uint16_t)(f + 0x10u), idx1);
    idx2 = f_2AD2_0029(m, f, GIN3_DS, 7u, 0xFFu);
    w0066_wr16s(m, GIN3_DS, (uint16_t)(f + 0x38u), idx2);
    wr8(m, W0066_F(0x3Au), 0xFFu);
    wr8(m, W0066_F(0x3Bu), 0u);
    w = w0066_rd16s(m, GIN3_DS, (uint16_t)(f + 0x38u));
    pool_seg = rd16(m, DS(0xCFB6u));
    pool_off = (uint16_t)(rd16(m, DS(0xCFB4u)) + (uint16_t)(w * 0x4Eu));
    wr8(m, LIN(pool_seg, (uint16_t)(pool_off + 0x1Eu)), (uint8_t)(rd8(m, LIN(pool_seg, (uint16_t)(pool_off + 0x1Eu))) | 0x80u));
    cell = (uint16_t)(0xA79Cu + (uint16_t)(rd8(m, W0066_A(0x0Au)) * 0x7Eu) + 0x28u);
    cell = (uint16_t)(cell + rd8(m, DS(0x8F1Du)));
    wr8(m, DS(cell), (uint8_t)(rd8(m, DS(cell)) + 1u));
    {
        uint8_t side = rd8(m, W0066_F(1u));
        uint16_t cnt = (uint16_t)(0x8F49u + side);

        wr8(m, DS(cnt), (uint8_t)(rd8(m, DS(cnt)) + 1u));
    }
    return s;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_297B_0066(m, a[0]); }
GIN3_REGISTER("297B_0066", 1, run)
