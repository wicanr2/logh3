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


/*
 * 硬體葉函式的替身（描述檔 stubs）。C 版呼叫它來取代「原版呼叫該硬體函式」：
 * id 是 SEG_OFF（大寫 hex），args 依序是堆疊參數字與 argregs 暫存器值。
 * 檢查程式依記錄下來的軌跡比對呼叫序列與參數，並回傳記錄下來的回傳值；
 * 呼叫序列或參數與原版不同就計為軌跡不符。替身沒有記憶體副作用。
 */
uint32_t gin3_stub(Mem *m, const char *id, int nargs, const uint16_t *args);

#endif
