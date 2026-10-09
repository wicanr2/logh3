/*
 * GIN3PS.EXE 2B99:0180  兩個陣營視角各做一次 Floyd–Warshall（呼叫 2B99:01B1 兩次，矩陣指標都從 ds:D062 起的表讀）
 *
 * 原版：far 呼叫，沒有參數，沒有 C 序言，無回傳值（AX 是殘留值）。指令：
 *   push ds:D06C; push ds:D06A; push ds:D064; push ds:D062; call 2B99:01B1; add sp,8
 *   push ds:D070; push ds:D06E; push ds:D068; push ds:D066; call 2B99:01B1; add sp,8
 * 2B99:01B1 的四個堆疊字由低位址到高位址是：距離矩陣位移、距離矩陣段、中點矩陣位移、中點矩陣段，所以
 *   第一次：距離矩陣 ＝ (ds:D064):(ds:D062)、中點矩陣 ＝ (ds:D06C):(ds:D06A)    （陣營 0 的視角，2B99:0068 的第一個表）
 *   第二次：距離矩陣 ＝ (ds:D068):(ds:D066)、中點矩陣 ＝ (ds:D070):(ds:D06E)    （陣營 1 的視角）
 * 基底狀態：D000:0000、D000:0800 兩個 2048 bytes 的距離矩陣，D000:1000、D000:1400 兩個 1024 bytes 的中點矩陣。
 * 第二次呼叫的四個參數是在第一次呼叫「之後」才從 ds:D066.. 讀出（push 發生在 call 之後），所以第一次呼叫的寫入
 *   （memset 中點矩陣、對角線、鬆弛）若蓋到指標格，第二次用的是被改過的指標。
 * 語意（強推論，docs/re/243）：兩張星域圖的最短路（陣營 0、陣營 1 視角）與中繼點。名稱不是證據，呼叫順序與參數位置才是。
 * 副作用：2B99:01B1 的寫入（各 memset 1024 bytes 中點矩陣、距離矩陣對角線與鬆弛、中點矩陣記錄）。
 * 輸入域：指標格 ds:D062..D071 用描述檔 freeze 排除隨機突變（亂指標會讓矩陣蓋到程式碼或堆疊，原版與 C 版都無法重現），由 seeds 給定
 *   （基底、兩組矩陣相同、暫存區、ROM 區、中點矩陣蓋到指標格自己）。矩陣內容仍由探針突變。
 *   2B99:01B1 的輸入域限制（讀 A0000 到 BFFFF 的案例由探針丟棄；word 存取不落在位移 0xFFFF）同樣適用。每案約 300 萬步，描述檔 step_budget 調高。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern void f_2B99_01B1(Mem *m, uint16_t d_off, uint16_t d_seg, uint16_t m_off, uint16_t m_seg);

void f_2B99_0180(Mem *m)
{
    f_2B99_01B1(m, rd16(m, DS(0xD062)), rd16(m, DS(0xD064)), rd16(m, DS(0xD06A)), rd16(m, DS(0xD06C)));
    f_2B99_01B1(m, rd16(m, DS(0xD066)), rd16(m, DS(0xD068)), rd16(m, DS(0xD06E)), rd16(m, DS(0xD070)));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2B99_0180(m);
    return 0;
}
GIN3_REGISTER("2B99_0180", 0, run)
