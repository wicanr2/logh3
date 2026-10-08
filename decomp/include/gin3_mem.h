/*
 * GIN3PS 的 C 參考實作共用的記憶體模型與 runtime 輔助函式。
 *
 * 模型：1 MB 平坦記憶體，8086 的 seg:off 以 LIN() 換算，位移是 16 位元（超過 0xFFFF 回繞），
 * 線性位址遮到 20 位元。C 函式的參數與原版 far／near 呼叫的堆疊參數一一對應，
 * 指標參數拆成 (off, seg) 兩個 uint16_t，與堆疊上的順序相同（位移在前）。
 *
 * DS 固定為 GIN3_DS。全域變數用 DS(off) 取線性位址，位移是 DGROUP 內的位移。
 */
#ifndef GIN3_MEM_H
#define GIN3_MEM_H

#include <stdint.h>

#define GIN3_MEM_SIZE 0x100000u
#define GIN3_DS 0x54B3u /* dosgolem 載入位置（映像在 0349）下的 DGROUP 段 */

typedef struct {
    uint8_t b[GIN3_MEM_SIZE];
} Mem;

#define LIN(seg, off) ((((uint32_t)(uint16_t)(seg) << 4) + (uint16_t)(off)) & 0xFFFFFu)
#define DS(off) LIN(GIN3_DS, (off))

static inline uint8_t rd8(const Mem *m, uint32_t lin) { return m->b[lin & 0xFFFFFu]; }

static inline uint16_t rd16(const Mem *m, uint32_t lin)
{
    return (uint16_t)(m->b[lin & 0xFFFFFu] | (m->b[(lin + 1) & 0xFFFFFu] << 8));
}

static inline uint32_t rd32(const Mem *m, uint32_t lin)
{
    return (uint32_t)rd16(m, lin) | ((uint32_t)rd16(m, lin + 2) << 16);
}

static inline void wr8(Mem *m, uint32_t lin, uint8_t v) { m->b[lin & 0xFFFFFu] = v; }

static inline void wr16(Mem *m, uint32_t lin, uint16_t v)
{
    m->b[lin & 0xFFFFFu] = (uint8_t)v;
    m->b[(lin + 1) & 0xFFFFFu] = (uint8_t)(v >> 8);
}

static inline void wr32(Mem *m, uint32_t lin, uint32_t v)
{
    wr16(m, lin, (uint16_t)v);
    wr16(m, lin + 2, (uint16_t)(v >> 16));
}

/*
 * Datalight 的 long 輔助函式（GIN3PS 的 3CC9 區）：運算元在 AX:BX 與 CX:DX，結果回 AX:BX。
 * 這裡直接用 C 的 32 位元運算表達同一個數學結果。語意來自 docs/re/146、internals/51
 * 與 3CC9 的反組譯，不含 Datalight 的程式碼。
 */
static inline uint32_t gin3_lmul(uint32_t a, uint32_t b) { return a * b; }
/* 除以 0 回 0xFFFFFFFF（3CC9:0020 的行為） */
static inline uint32_t gin3_uldiv(uint32_t a, uint32_t b) { return b ? a / b : 0xFFFFFFFFu; }

#endif
