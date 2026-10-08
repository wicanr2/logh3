/*
 * GIN3PS.EXE 1AAB:1286  艦隊裡每一艘修護艦（艦種 6）：存量重設成現有艘數，再拿存量把同艦隊其他艦艇的缺額補回來
 *
 * 原版：far 呼叫，一個參數＝艦隊記錄的 far 指標（位移在前、段在後），無回傳值（描述檔 ret = none）。
 * 艦艇池基底在 ds:CFB4（位移）與 ds:CFB6（段），每次迴圈都重讀。
 *
 * 逐條對應反組譯（IDA sub_288A6）：
 *   for i ← 0；i < 艦隊[+0x0F](byte，每圈重讀)；i++：
 *     src ← 池[艦隊[+0x10 + 4i](word)]            ; 位移 low16(idx × 0x4E)，段 ds:CFB6 在這一圈開頭讀好
 *     若 src[+0x04] != 6：下一圈
 *     階段一（重設）：src[+0x10](word) ← 0；src[+0x0E](word) ← src[+0x08]
 *     階段二（補缺額）：for j ← 0；j < 艦隊[+0x0F](byte，每圈重讀)；j++：
 *       若 src[+0x0E](word) | src[+0x10](word) == 0：結束階段二（整個存量用完）
 *       dst ← 池[艦隊[+0x10 + 4j](word)]          ; 段 ds:CFB6 在這一圈重讀
 *       若 dst[+0x04] == 6：下一圈
 *       若 dst[+0x0C] <= dst[+0x08]（無號）：下一圈
 *       need ← dst[+0x0C] − dst[+0x08]（16 位元）
 *       stock ← src[+0x0E..+0x11]（32 位元，hi:lo）
 *       give ← (hi == 0 且 lo < need) ? lo : need   ; hi 非 0 時不比低字，直接取 need（原版 cmp dx,0 後 jnz 跳過低字比較）
 *       dst[+0x08](word) += give（16 位元加，不進位到別處）
 *       src[+0x0E] −= give（32 位元，sub 低字、sbb 高字，各自從記憶體現讀）
 *
 * 推論等級
 *   已證實（差異測試的位元組、寫入集合）：兩個階段的結構、欄位位移、give 的取法、只對艦種 != 6 的艦艇補。
 *   強推論（docs/internals/64-ship-pool.md、docs/re/346 §四、docs/re/131）：+0x04 艦種（6 ＝ 修護艦）、+0x08 現有艘數、
 *     +0x0C 修理補得回來的上限、+0x0E/+0x10 飛彈存量（對修護艦而言就是它能「修」的艘數預算）。階段二就是修護艦修理
 *     同艦隊其他艦艇的動作；欄位名稱不是證據，只有位移與運算是證據。
 *   未知：呼叫者 1AAB:1235 的時機（本函式只有一個呼叫者）。
 *
 * 階段二裡「hi 非 0」的分支：正常輸入不可達，因為階段一剛把 src[+0x10] 清成 0，且階段二只往 dst[+0x08] 加；
 *   只有當 dst[+0x08] 與 src[+0x10] 重疊（池索引別名，idx 差使 dst ≡ src + 8 mod 64K）時，dst 的加法才會寫到 src 的高字。
 *   描述檔用一個別名種子（池 #10 與池 #(10 + k)，78k ≡ 8 mod 65536）走過這條路徑。
 *
 * 輸入域：艦隊指標任意（描述檔從艦隊表取）；艦隊 +0x0F 到 255、池索引任意。
 *   池索引超出 352 筆時讀寫池之後的記憶體，本版同樣照做（平面視訊記憶體的讀取除外，見 docs/re/445 限制）。
 */
#include "gin3_reg.h"

void f_1AAB_1286(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint8_t i = 0;
    uint16_t slot = (uint16_t)(fleet_off + 0x10);

    while (i < rd8(m, LIN(fleet_seg, fleet_off + 0x0F))) {
        uint16_t src_seg = rd16(m, DS(0xCFB6));
        uint16_t src = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(rd16(m, LIN(fleet_seg, slot)) * 0x4Eu));

        if (rd8(m, LIN(src_seg, src + 0x04)) == 6) {
            uint16_t count = rd16(m, LIN(src_seg, src + 0x08));
            uint8_t j = 0;
            uint16_t slot2 = (uint16_t)(fleet_off + 0x10);

            wr16(m, LIN(src_seg, src + 0x10), 0);
            wr16(m, LIN(src_seg, src + 0x0E), count);
            while (j < rd8(m, LIN(fleet_seg, fleet_off + 0x0F))) {
                uint16_t dst_seg, dst, need, lo, hi, give, borrow, cur;

                if ((rd16(m, LIN(src_seg, src + 0x0E)) | rd16(m, LIN(src_seg, src + 0x10))) == 0)
                    break;
                dst_seg = rd16(m, DS(0xCFB6));
                dst = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(rd16(m, LIN(fleet_seg, slot2)) * 0x4Eu));
                if (rd8(m, LIN(dst_seg, dst + 0x04)) != 6 &&
                    rd16(m, LIN(dst_seg, dst + 0x0C)) > rd16(m, LIN(dst_seg, dst + 0x08))) {
                    need = (uint16_t)(rd16(m, LIN(dst_seg, dst + 0x0C)) - rd16(m, LIN(dst_seg, dst + 0x08)));
                    lo = rd16(m, LIN(src_seg, src + 0x0E));
                    hi = rd16(m, LIN(src_seg, src + 0x10));
                    give = (hi == 0 && lo < need) ? lo : need;
                    wr16(m, LIN(dst_seg, dst + 0x08), (uint16_t)(rd16(m, LIN(dst_seg, dst + 0x08)) + give));
                    cur = rd16(m, LIN(src_seg, src + 0x0E));
                    borrow = cur < give;
                    wr16(m, LIN(src_seg, src + 0x0E), (uint16_t)(cur - give));
                    wr16(m, LIN(src_seg, src + 0x10), (uint16_t)(rd16(m, LIN(src_seg, src + 0x10)) - borrow));
                }
                j++;
                slot2 = (uint16_t)(slot2 + 4);
            }
        }
        i++;
        slot = (uint16_t)(slot + 4);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1AAB_1286(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("1AAB_1286", 2, run)
