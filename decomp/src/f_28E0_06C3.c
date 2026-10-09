/*
 * GIN3PS.EXE 28E0:06C3  提督升級時的能力值成長：抽亂數決定哪幾項能力成長，每項成長量由 28E0:095C 算出，回傳成長了哪幾項的位元遮罩
 *
 * 原版：far 呼叫，一個參數＝提督記錄（28 bytes）的 far 指標（位移、段）。序言 push bp／si、sub sp,2Ah、mov bp,sp，所以指標在
 *   [bp+32h]（位移）與 [bp+34h]（段）。回傳 AX ＝ si（位元遮罩）。唯一呼叫者 28E0:05F9。
 *   區域變數：[bp+0] ＝ 最近一次 095C 的結果 AL、[bp+1] ＝ 成長項數 count（byte）；SI ＝ 位元遮罩；其餘是指標暫存（寫了馬上讀）。
 *
 * 語意（已證實：與反組譯逐條對過；欄位名稱為強推論，見下）：
 *   定義 stat(k, a, b)：g ← 28E0:095C(提督[+k] 的低位元組, a, b)；g != 0 才 提督[+k] ← 提督[+k] + g（byte 回繞），並回報「成長了」。
 *   定義 odd(r) ＝ 原版的 `cwd；xor ax,dx；and ax,1；add dx,ax；or dx,dx` 慣用法（對 rand() 取 % 2，結果非 0 為奇數路徑）。
 *   1. stat(0x0B, 0x12C, 0x32)：成長了 → si |= 0x004、count++。
 *   2. r ← rand()：odd(r) 為真：stat(0x13, 0xC8, 0x64) → si |= 0x008、count++；否則 stat(0x12, 0xC8, 0x64) → si |= 0x010、count++。
 *   3. r ← rand()；r % 3（有號）：0 → stat(0x0F, 0x64, 0x32) → si |= 0x020、count++；1 → stat(0x10, 0x64, 0x32) → 0x040、count++；
 *      2 → stat(0x11, 0x64, 0x32) → 0x080、count++；其餘（r 為負時才可能）什麼都不做。
 *   4. count >= 3（byte 無號）：回 si。
 *   5. r ← rand()：odd(r) 為真：stat(0x0C, 0xC8, 0x32) → si |= 0x100、count++；否則 stat(0x0D, 0xC8, 0x32) → si |= 0x200、count++。
 *   6. count >= 3：回 si。
 *   7. r ← rand()：odd(r) 為真：stat(0x0E, 0x64, 0x32) → si |= 0x400（不加 count）；否則 stat(0x17, 0xC8, 0x32) → si |= 0x800（不加 count）。
 *   8. 回 si。
 *   095C 的等級 >= 0xC8 時回 0 且不取亂數；其餘各抽兩次亂數，所以亂數序列的消耗量隨路徑改變（rand 是 3CD5:000C，寫 ds:94DC 種子）。
 *
 * 呼叫 28E0:095C 的方式：far 呼叫，參數從右往左推：先推 b（`mov ax,32h／64h；push ax`）、再推 a、最後推 `word ptr es:[提督+k]`
 *   （被呼叫者只取低位元組當等級）。stat 的 k 是十進位位移，a、b 是原版載入的常數，逐個呼叫點核對過。
 *
 * 與文件的對照（常數與順序已證實相符；欄位名稱是依順序對應，強推論）：docs/re/418-battle-promotion.md 第 45 到 50 行的升級成長規則——「成長先試統率力 300／50，再以 RND 選占領或
 *   防禦 200／100，再選光線砲、飛彈或戰鬥機 100／50。若已有三項非零成長便停止。否則選移動或回轉 200／50；仍不足三項再選營運 100／50 或
 *   運用 200／50。byte 加法與旗標位元按原版保留」——逐項對上本檔的 1 到 8 步：+0x0B 是統率力（300／50）；+0x13／+0x12 二選一是占領／防禦
 *   （200／100）；+0x0F／+0x10／+0x11 三選一是光線砲／飛彈／戰鬥機（100／50）；+0x0C／+0x0D 二選一是移動／回轉（200／50）；+0x0E 是營運
 *   （100／50）、+0x17 是運用（200／50）。文件沒有寫 +0x13 與 +0x12 各是占領還是防禦、+0x0F 到 +0x11 的順序、移動與回轉的順序，這些對應
 *   本檔不下結論（未知）。
 * 欄位語意（強推論）：提督記錄 +0x0B 到 +0x17 是能力值（docs/internals/08 第 217 行附近）；位元遮罩的 0x004 到 0x800 對應成長了哪一項
 *   （呼叫者 28E0:05F9 使用，細節未查證）。
 *
 * 輸入域：提督指標取提督表任一筆（82 筆）；各項能力值與 ds:94DC 亂數種子由突變與種子涵蓋（種子造出等級 >= 0xC8 不取亂數、
 *   成長量加到 0xFF 回繞、各路徑）。
 * 推論等級：已證實（反組譯逐指令＋差異測試，095C 與 rand 各有差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_28E0_095C(Mem *m, uint16_t level_word, uint16_t span_a, uint16_t span_b);
extern uint16_t f_3CD5_000C(Mem *m);

/* 原版的 `rand() % 2` 慣用法：cwd；xor ax,dx；and ax,1；add dx,ax；結果非 0 為奇數路徑 */
static uint16_t w06C3_odd(uint16_t r)
{
    uint16_t dx = ((int16_t)r < 0) ? 0xFFFFu : 0u;
    uint16_t ax = (uint16_t)((r ^ dx) & 1u);

    return (uint16_t)(dx + ax);
}

/* 呼叫 095C 並把成長量加到提督[+k]；回成長量（0 表示沒有成長） */
static uint8_t w06C3_stat(Mem *m, uint16_t seg, uint16_t off, uint16_t span_a, uint16_t span_b)
{
    uint8_t g = (uint8_t)f_28E0_095C(m, rd8(m, LIN(seg, off)), span_a, span_b);

    if (g != 0)
        wr8(m, LIN(seg, off), (uint8_t)(rd8(m, LIN(seg, off)) + g));
    return g;
}

uint16_t f_28E0_06C3(Mem *m, uint16_t adm_off, uint16_t adm_seg)
{
    uint16_t si = 0;
    uint8_t count = 0;
    uint16_t r;

    if (w06C3_stat(m, adm_seg, (uint16_t)(adm_off + 0x0Bu), 0x12Cu, 0x32u) != 0) {
        si |= 0x004u;
        count++;
    }

    r = f_3CD5_000C(m); /* 亂數2 */
    if (w06C3_odd(r) != 0) {
        if (w06C3_stat(m, adm_seg, (uint16_t)(adm_off + 0x13u), 0xC8u, 0x64u) != 0) {
            si |= 0x008u;
            count++;
        }
    } else {
        if (w06C3_stat(m, adm_seg, (uint16_t)(adm_off + 0x12u), 0xC8u, 0x64u) != 0) {
            si |= 0x010u;
            count++;
        }
    }

    r = f_3CD5_000C(m); /* 亂數3 */
    switch ((int16_t)r % 3) {
    case 0:
        if (w06C3_stat(m, adm_seg, (uint16_t)(adm_off + 0x0Fu), 0x64u, 0x32u) != 0) {
            si |= 0x020u;
            count++;
        }
        break;
    case 1:
        if (w06C3_stat(m, adm_seg, (uint16_t)(adm_off + 0x10u), 0x64u, 0x32u) != 0) {
            si |= 0x040u;
            count++;
        }
        break;
    case 2:
        if (w06C3_stat(m, adm_seg, (uint16_t)(adm_off + 0x11u), 0x64u, 0x32u) != 0) {
            si |= 0x080u;
            count++;
        }
        break;
    default:
        break;
    }

    if (count >= 3u) /* loc_361B4 */
        return si;

    r = f_3CD5_000C(m); /* 亂數4 */
    if (w06C3_odd(r) != 0) {
        if (w06C3_stat(m, adm_seg, (uint16_t)(adm_off + 0x0Cu), 0xC8u, 0x32u) != 0) {
            si |= 0x100u;
            count++;
        }
    } else {
        if (w06C3_stat(m, adm_seg, (uint16_t)(adm_off + 0x0Du), 0xC8u, 0x32u) != 0) {
            si |= 0x200u;
            count++;
        }
    }

    if (count >= 3u) /* loc_3623F */
        return si;

    r = f_3CD5_000C(m); /* 亂數5 */
    if (w06C3_odd(r) != 0) {
        if (w06C3_stat(m, adm_seg, (uint16_t)(adm_off + 0x0Eu), 0x64u, 0x32u) != 0)
            si |= 0x400u;
    } else {
        if (w06C3_stat(m, adm_seg, (uint16_t)(adm_off + 0x17u), 0xC8u, 0x32u) != 0)
            si |= 0x800u;
    }
    return si;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_28E0_06C3(m, a[0], a[1]); }
GIN3_REGISTER("28E0_06C3", 2, run)
