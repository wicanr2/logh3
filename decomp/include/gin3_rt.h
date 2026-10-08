/*
 * GIN3PS 的 runtime 輔助函式語意（Datalight C 的 long 運算區，3CC9 與 3CD5）。
 * 這裡只放「數學結果」，不含 Datalight 的程式碼。每個函式都有對應的 f_*.c 與差異測試，
 * 其他 C 參考實作要呼叫的時候用這裡的 inline 版本。
 */
#ifndef GIN3_RT_H
#define GIN3_RT_H

#include <stdint.h>

/* 3CC9:0020 的數學結果：無號除法，除數為 0 時商 0xFFFFFFFF、餘數 0。 */
static inline void gin3_uldivmod(uint32_t a, uint32_t b, uint32_t *q, uint32_t *r)
{
    if (b == 0) {
        *q = 0xFFFFFFFFu;
        *r = 0;
        return;
    }
    *q = a / b;
    *r = a % b;
}

/* 3CC9:0070 的數學結果：有號除法，商向零取整，餘數的正負號跟被除數；
 * 用無號運算表達，INT_MIN 與除以 0 都與原版的取補數步驟一致。 */
static inline void gin3_ldivmod(uint32_t a, uint32_t b, uint32_t *q, uint32_t *r)
{
    uint32_t na = a >> 31, nb = b >> 31;
    uint32_t ua = na ? 0u - a : a, ub = nb ? 0u - b : b;

    gin3_uldivmod(ua, ub, q, r);
    if (na)
        *r = 0u - *r;
    if (na != nb)
        *q = 0u - *q;
}

#endif
