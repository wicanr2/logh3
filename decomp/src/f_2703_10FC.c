/*
 * GIN3PS.EXE 2703:10FC  叛亂度（治安）公式的「提督」項 f5：數同陣營的提督、取兩個能力欄的最大值，算 n × p × q ÷ 1000，再交給 3310:0130（開根號）。
 *   實際上交給 3310:0130 的引數永遠是 0（見下），所以回傳值恆為 3310:0130(0) 的 AX。
 *
 * 原版（238 bytes，far 呼叫，一個 far 指標參數「星域記錄」；序言 push bp／si／di、sub sp,60h、mov bp,sp，所以星域指標的位移在 [bp+6Ah]、段在 [bp+6Ch]）。
 *   回傳 AX ＝ 3310:0130 的 AX（描述檔 ret ＝ ax）。區域變數：[bp+0..51h] ＝ 提督清單緩衝區（82 bytes）、[bp+52h] ＝ 星域 [+0x11]、[bp+53h] ＝ 清單項數 n、
 *   [bp+54h] ＝ 迴圈 i、[bp+55h] ＝ 計數 count、[bp+56h] ＝ p 的最大值、[bp+57h] ＝ q 的最大值、[bp+58h..5Bh] ＝ 除法商的暫存。
 * 過程（已證實：反組譯逐指令；28E0:0103 由其差異測試證實）：
 *   a ← byte 星域[+0x11]
 *   n ← 28E0:0103(清單 ＝ SS:bp, a, flags ＝ 0x24, limit ＝ 0)        ; 清單裡留下 提督[+0]==0xFF 且 提督[+9] 的 bit7 有設者（旗標 bit2、bit5），a != 0xFF 時再留 提督[+0x0A]==a 者
 *   count ← 0；pmax ← 0；qmax ← 0
 *   對 i ＝ 0..n−1（byte，jnb 無號）：adm ＝ ds:(0xB77C ＋ 清單[i] × 0x1C)（16 位元）
 *     byte adm[+1] ＝＝ byte 星域[+0x12]（每圈重讀星域指標）→ count++；adm[+0x0B] >= pmax 則 pmax ← adm[+0x0B]；adm[+0x13] >= qmax 則 qmax ← adm[+0x13]（都是 byte 無號 jb 跳過）
 *   v ← 3CC9:0009(count, pmax)（32 位元乘，AX:BX × CX:DX，結果 AX:BX）；v ← 3CC9:0009(v, qmax)
 *   quo ← 3CC9:0020(v, 1000)（32 位元無號除，商在 AX:BX）
 *   [bp+5Ah] ← quo 的低字；[bp+58h] ← 0；push [bp+58h]            ; 推的是清成 0 的那一格；商的低字存進 [bp+5Ah] 之後再沒被讀（死儲存）
 *   回 3310:0130(0) 的 AX
 * 呼叫的方式（逐一對過 asm）：
 *   28E0:0103 的五個堆疊字（由低位址到高位址）：清單位移 ＝ bp、清單段 ＝ SS、a（`push ax`，AL ＝ a、AH 是呼叫者的殘值，0103 只取低位元組）、
 *     flags（`push dx`，DL ＝ 0x24、DH 殘值，0103 只取低位元組）、limit ＝ 0（`xor cx,cx; push cx`）。
 *   3CC9:0009 與 3CC9:0020 用暫存器傳參：AX:BX 與 CX:DX。第一次乘：AX ＝ 0、BX ＝ count、CX ＝ 0、DX ＝ pmax；第二次乘：AX:BX ＝ 第一次的結果、CX ＝ 0、DX ＝ qmax；
 *     除：AX:BX ＝ 第二次的結果、CX ＝ 0、DX ＝ 1000。
 * 語意（強推論）：docs/re/299-region-unrest-writers-and-recompute.md 第六節（第 229 到 248 行，f5 恆等於 0，推給開根號的是清成 0 的那一格）、
 *   docs/re/43-unrest-full-disassembly.md 第 118 到 126 行、docs/spec/006-unrest-at-load.md 第 76 行、docs/spec/042 第 112 行，與本檔一致。
 *   欄位名：星域 [+0x11] ＝ 星域編號（傳給清單的 a）、[+0x12] ＝ 陣營；提督 [+1] ＝ 陣營、[+0x0B] ＝ p、[+0x13] ＝ q（來自 docs/re/43，未再驗證）。
 *
 * 差異測試的設計（這一支特別說明）：
 *   真實的 3CC9:0009、3CC9:0020、3310:0130(0)（014B 見 n＝0 直接回 0:0，不寫記憶體）都沒有全域副作用，而且商的低字是死儲存，所以不用替身時，
 *   本函式從外面唯一看得到的輸出是「回傳 0」與堆疊帶內的寫入（比對時排除）：迴圈算出來的 count、pmax、qmax 完全不影響可觀察輸出，等於沒有驗證。
 *   所以描述檔把這三個被呼叫者用 stubs 取代，並以 argregs 記錄暫存器引數（AX、BX、CX、DX）：軌跡因此能比對 count、pmax、qmax、乘法的串接與除數，
 *   3310:0130 的引數（恆為 0）與它的回傳值（成為本函式的回傳值）。替身的回傳值是隨機的，**真實的 3310:0130(0) 回 0；替身下的回傳值不是 0 屬於測試設計**。
 *   乘、除與開根號本身的語意由各自的差異測試負責（3CC9:0009、3CC9:0020、3310:0130），不在本檔驗證範圍。
 *   28E0:0103 照常執行（它的篩選結果從軌跡裡的 count、pmax、qmax 間接觀察）。
 *
 * 機器模型：C 版沒有堆疊。清單緩衝區放在原版的實際位置：基底 fcmd-base.state 的 SS ＝ 0x63EB、進入點 SP ＝ 0x1F8A（探針再推 1 個 far 指標 4 bytes、返回位址 4、
 *   push bp／si／di 6、sub sp,60h）→ bp ＝ 0x1F1C，清單在 SS:1F1C..1F6D，落在探針的堆疊帶（SS:1B8A..1F89）內，寫入集合比對時雙方都排除。
 *   換基底狀態時這個位址要重算。
 * 輸入域：星域指標取星域表（基底 fcmd-base.state）；提督表 82 筆與星域欄位由突變涵蓋。星域指標指到堆疊帶內（蓋到清單）的情形不涵蓋。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint8_t f_28E0_0103(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t a_w, uint16_t flags_w, uint16_t limit_w);

#define F10FC_SS 0x63EBu   /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define F10FC_LIST 0x1F1Cu /* 進入點 SP 1F8A − 4 − 4 − 6 − 0x60 ＝ bp(1F1C)，清單在 bp+0 */

uint16_t f_2703_10FC(Mem *m, uint16_t reg_off, uint16_t reg_seg)
{
    uint8_t a = rd8(m, LIN(reg_seg, (uint16_t)(reg_off + 0x11u)));
    uint8_t n = f_28E0_0103(m, F10FC_LIST, F10FC_SS, a, 0x24, 0);
    uint8_t count = 0, pmax = 0, qmax = 0, i;
    uint16_t r[4];
    uint32_t v;

    for (i = 0; i < n; i++) {
        uint8_t idx = rd8(m, LIN(F10FC_SS, (uint16_t)(F10FC_LIST + i)));
        uint16_t adm = (uint16_t)(0xB77Cu + (uint16_t)(idx * 0x1Cu));

        if (rd8(m, DS((uint16_t)(adm + 1u))) != rd8(m, LIN(reg_seg, (uint16_t)(reg_off + 0x12u))))
            continue;
        count++;
        if (rd8(m, DS((uint16_t)(adm + 0x0Bu))) >= pmax)
            pmax = rd8(m, DS((uint16_t)(adm + 0x0Bu)));
        if (rd8(m, DS((uint16_t)(adm + 0x13u))) >= qmax)
            qmax = rd8(m, DS((uint16_t)(adm + 0x13u)));
    }
    r[0] = 0; r[1] = count; r[2] = 0; r[3] = pmax;
    v = gin3_stub(m, "3CC9_0009", 4, r);
    r[0] = (uint16_t)(v >> 16); r[1] = (uint16_t)v; r[2] = 0; r[3] = qmax;
    v = gin3_stub(m, "3CC9_0009", 4, r);
    r[0] = (uint16_t)(v >> 16); r[1] = (uint16_t)v; r[2] = 0; r[3] = 1000;
    (void)gin3_stub(m, "3CC9_0020", 4, r); /* 商的低字存進 [bp+5Ah] 之後沒人讀；推給 0130 的是清成 0 的 [bp+58h] */
    r[0] = 0;
    return (uint16_t)gin3_stub(m, "3310_0130", 1, r);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2703_10FC(m, a[0], a[1]); }
GIN3_REGISTER("2703_10FC", 2, run)
