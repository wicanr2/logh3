/*
 * 函式自我登記：每個 decomp/src/f_*.c 用 GIN3_REGISTER 宣告自己，檢查程式不必維護一張總表。
 * 參數 args 是原版堆疊上的 16 位元字，由低位址到高位址（第一個參數在 args[0]）。
 * args 依描述檔的參數順序展開：堆疊字（由低位址到高位址）與 kind=reg 的暫存器值，各佔一格。
 * run 回傳打包後的值（ax、dxax，或 pack:… 由高位到低位串接的暫存器），由向量檔的 ret 欄位決定比較範圍。
 */
#ifndef GIN3_REG_H
#define GIN3_REG_H

#include <stdint.h>

#include "gin3_mem.h"

typedef uint64_t (*Gin3Run)(Mem *m, const uint16_t *args);

void gin3_register(const char *id, int nargs, Gin3Run run);

#define GIN3_REGISTER(id, nargs, runfn) \
    static void __attribute__((constructor)) gin3_reg_##runfn(void) { gin3_register(id, nargs, runfn); }

#endif
