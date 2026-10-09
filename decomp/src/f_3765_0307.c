/*
 * GIN3PS.EXE 3765:0307  從事件佇列（環狀緩衝區）移除所有「第一個 word 等於 key」的事件，後面的事件往前補
 * 呼叫：far，一個 word 參數 key（[bp+16h]；序言 push bp／si／di、sub sp,0Ch）。回傳 AX：結束時的 cur 段（無意義，描述檔以 ax 比對）。
 * 資料結構（位移是 DGROUP 內的位移，段 ＝ DS）：
 *   環狀緩衝區：ds:F184 起，16 筆 × 12 bytes ＝ 0xC0 bytes（終點 ds:F244）。一筆事件的第一個 word 是事件碼。
 *   ds:F244 word（尾端位移）、ds:F246 word（尾端段）：下一個要寫入的槽，tail。
 *   ds:F248 word（開頭位移）、ds:F24A word（開頭段）：最舊的事件，head。
 *   ds:F24C word：事件筆數。
 * 呼叫者 3765:0255（強推論，asm）：先 cli，再用 (1,4) 或 (2,8) 兩個事件碼各呼叫一次本函式，清掉佇列裡的這些事件
 *   （滑鼠／鍵盤的按下與放開事件，碼與按鍵語意未查證），再 sti。
 * 算法（asm 逐條）：
 *   cur ← head。迴圈：
 *     若 cur ＝ tail（段與位移都相等）：返回。
 *     若 word(cur) ≠ key：cur.off ← cur.off + 12；若 (cur − (DS:F184)) 當有號 32 位元 ≥ 0xC0 就 cur ← DS:F184；繼續。
 *     否則（相等）：p ← cur；迴圈直到 p ＝ tail：
 *         q ← p + 12（位移加 12，同樣超過終點就繞回起點 DS:F184）；
 *         若 q ≠ tail：memmove(p, q, 12)（3CA9:0002，已 diff-tested）；
 *         p.off ← p.off + 12（同樣繞回）。
 *       然後 tail 退一格：若 tail ＝ (DS:F184) 先把 tail.off 設成 0xF244、tail.seg 設成 DS；再 tail.off ← tail.off − 12（16 位元）；
 *       ds:F24C ← ds:F24C − 1（16 位元）。不改 head。回到迴圈開頭，**同一個 cur** 重新檢查（原位置已補上下一筆事件）。
 *   「(x − DS:F184) ≥ 0xC0」的比較是 32 位元：(seg:off 當 32 位元數) − (DS:F184)，`sub bx,si; sbb ax,cx`，再呼叫 3CC9:00B9（有號比較，
 *   已 diff-tested）與 0x000000C0 比；jl／jge 分別是旗標 SF≠OF、SF＝OF。cur／p／q 的段不是 DS 時，高 16 位元差不為 0，行為照 32 位元比較。
 * 讀寫順序與重讀：tail 在每次比較前重讀（memmove 可能改到它）；cur、p、q 是區域變數（堆疊）。
 * 輸入域：佇列結構良構、尾端可由 head 依 +12 與繞回走到（會終止）的狀態。尾端不可達時原版無窮迴圈，探針以步數上限丟棄；
 *   本版用 guard 上限（1,048,576 次迴圈）在域外輸入下返回，域外行為不比對。word 讀取在位移 0xFFFF 時取同段位移 0 的高位元組（dosgolem 的 CPU 語意）。
 *   memmove 的目的落在 F0000 以上（ROM）時 3CA9:0002 的 C 版未模型化 ROM；需要 head 段 ≥ 0xF000 且該處 word 剛好等於 key，種子與突變實際到不了。
 * 推論等級：算法已證實（asm 與差異測試）；環狀緩衝區與「事件佇列」的語意為強推論（呼叫者 3765:0255、資料形狀）；事件碼的意義未知。
 */
#include "gin3_reg.h"

extern uint32_t f_3CA9_0002(Mem *m, uint16_t doff, uint16_t dseg, uint16_t soff, uint16_t sseg, uint16_t count);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

#define F3765_0307_SF 0x0080u
#define F3765_0307_OF 0x0800u

/* 段內 word 讀取：兩個 byte 的位移各自 16 位元回繞（dosgolem cpu.go read16） */
static uint16_t f_3765_0307_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* `sub bx,si; sbb ax,cx; xor cx,cx; mov dx,0C0h; call 3CC9:00B9`：回傳 (seg:off − DS:F184) 當有號 32 位元數是否 < 0xC0（jl） */
static int f_3765_0307_before_end(uint16_t seg, uint16_t off)
{
    uint32_t d = (((uint32_t)seg << 16) | off) - (((uint32_t)GIN3_DS << 16) | 0xF184u);
    uint16_t fl = f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, 0x0000, 0x00C0);

    return ((fl & F3765_0307_SF) != 0) != ((fl & F3765_0307_OF) != 0);
}

uint16_t f_3765_0307(Mem *m, uint16_t key)
{
    uint16_t cur_off = rd16(m, DS(0xF248));
    uint16_t cur_seg = rd16(m, DS(0xF24A));
    uint32_t guard = 0;

    for (;;) {
        uint16_t t_off = rd16(m, DS(0xF244));
        uint16_t t_seg = rd16(m, DS(0xF246));
        uint16_t p_off, p_seg;

        if (++guard > 0x100000u)
            return cur_seg;
        if (cur_seg == t_seg && cur_off == t_off)
            return cur_seg;
        if (f_3765_0307_rd16s(m, cur_seg, cur_off) != key) {
            cur_off = (uint16_t)(cur_off + 12u);
            if (!f_3765_0307_before_end(cur_seg, cur_off)) {
                cur_off = 0xF184;
                cur_seg = GIN3_DS;
            }
            continue;
        }
        p_off = cur_off;
        p_seg = cur_seg;
        for (;;) {
            uint16_t q_off, q_seg;

            t_off = rd16(m, DS(0xF244));
            t_seg = rd16(m, DS(0xF246));
            if (++guard > 0x100000u)
                return cur_seg;
            if (p_seg == t_seg && p_off == t_off)
                break;
            q_off = (uint16_t)(p_off + 12u);
            q_seg = p_seg;
            if (!f_3765_0307_before_end(q_seg, q_off)) {
                q_off = 0xF184;
                q_seg = GIN3_DS;
            }
            t_off = rd16(m, DS(0xF244));
            t_seg = rd16(m, DS(0xF246));
            if (!(q_seg == t_seg && q_off == t_off))
                f_3CA9_0002(m, p_off, p_seg, q_off, q_seg, 12);
            p_off = (uint16_t)(p_off + 12u);
            if (!f_3765_0307_before_end(p_seg, p_off)) {
                p_off = 0xF184;
                p_seg = GIN3_DS;
            }
        }
        if (rd16(m, DS(0xF246)) == GIN3_DS && rd16(m, DS(0xF244)) == 0xF184) {
            wr16(m, DS(0xF244), 0xF244);
            wr16(m, DS(0xF246), GIN3_DS);
        }
        wr16(m, DS(0xF244), (uint16_t)(rd16(m, DS(0xF244)) + 0xFFF4u));
        wr16(m, DS(0xF24C), (uint16_t)(rd16(m, DS(0xF24C)) - 1u));
        /* 重新檢查同一格 */
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3765_0307(m, a[0]);
}
GIN3_REGISTER("3765_0307", 1, run)
