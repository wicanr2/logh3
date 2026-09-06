# RE 05 — 停點根因是 EXEC 沒還記憶體；主選單畫出來了

日期：2026-09-06
執行器：dosgolem-logh3 分支 `logh3-com-support`（`803cb06`）
輸入：原版未修改 `org_game/extracted/GIN3.COM`（唯讀，SHA-256 見 re/01）
規格：dosgolem-logh3 `docs/spec/010-exec-memory-reclaim-and-planar-write-modes.md`
前篇：`04-gin3ps-planar-pipeline-and-module-manager-crash.md`

推論標籤：confirmed／強證據／假說／未知。

---

## 1. 結果（confirmed）

`GIN3PS.EXE` 跑過原停點，進入互動等待迴圈：6000 萬道指令仍在跑，滑鼠
輪詢 205,509 次，畫面上是主選單 **帝國／同盟／對戰／取進度**。

畫面檔在 `workplace/re05/full-wm3.png`（本機證據，gitignore）。那張圖用
假色盤——遊戲自己的調色盤淡入還沒走完（§5），所以**僅供目視**，不是對拍
證據。驗收依據是色號陣列 `screen-60M-wm3.bin`。

## 2. 那批寫進 IVT 的值是中文字（confirmed）

re/03 與 re/04 把停點解讀成「模組管理器把 ISR 表寫進 IVT」。反組譯
`6B93:0068` 與 `6C50:02CC` 之後，那段程式是**中文文字視窗的儲存格
緩衝區**：

- `6B93:0068` 是清除迴圈：`for i in 0..w*h-1: cell[i] = {word 0x2020,
  byte attr, byte 0}`。`w = rec[+0x0B] = 3`、`h = rec[+0x0C] = 12`、
  `attr = rec[+0x01] = 0x97`。`0x2020` 是兩個空格。
- `6C50:02CC` 是寫字元：`ax = src[0]<<8 | src[1]`，寫進儲存格的 word，
  屬性取自 `rec[+0x1D]`（反白）或 `rec[+0x1F]`（正常）。

re/04 §3 列出的八個「段位址」逐位元組解成 Big5 就是
`A650 同`、`A8FA 取`、`ABD7 度`、`B0EA 國`、`B669 進`、`B7F9 盟`、
`B9EF 對`、`BED4 戰`——正好是主選單那四個詞的字。

re/04 觀察到「每筆都是 `{off=0x0097, seg=變動}`」也跟著解釋掉了：
緩衝區起點是 `0000:0002`，儲存格與 IVT 條目差 2 bytes，於是屬性
`0x97` 落在 offset 欄、字碼落在 segment 欄。

本作是 **Big5 中文版**（開頭載入的 `BU16IBM.RGB` 是 16 點字型），
不是日文版；這一點會影響之後所有文字相關的逆向。

## 3. 根因：EXEC 之後配置游標沒有退回（confirmed）

證據鏈：

1. 緩衝區來自 `6B93` 建構子：算 `w*h*4` bytes，far call `70A8:0000`
   要記憶體，把回傳的 `AX:BX` 存進 `rec[+0x10]/[+0x0E]`。實測回傳
   `0000:0002`。
2. `70A8:0000` 是 malloc 包裝：`cmp ax,0 / jnz ret; cmp bx,0 / jnz ret`
   ——只有 `0000:0000` 才算失敗，`0000:0002` 被當成配置成功。
3. 真正的 malloc 在 `72C2:0047`：循環自由串列，頭指標在 `DS:94E0`。
   找不到夠大的區塊時呼叫 `70AD:000D` 向系統要更多。
4. 本體向 DOS 要記憶體的軌跡：`#23305663 AH=4A BX=6A07 ES=3922 →
   失敗（可用 66DD）`。`0x3922 + 0x6A07 = 0xA329`，超過 640 KB。
5. `0x3922` 是 `GIN3PS.EXE` 的 PSP。`OPEN.EXE` 的 PSP 是 `0x0339`，
   它結束時 dosgolem 的 `childExit` 沒有還原配置游標——bump 配置器
   只升不降，那 230 KB 永遠不還，本體因此被載高了。

修法（spec 010 §1）：`execFrame` 記下 EXEC 之前的 `freeSeg`，
`childExit` 一併還原。回收的是所有權不是內容——視訊記憶體與 IVT 照舊
保留，只有「下一次配置從哪開始」退回去。

實測：本體 PSP `0x3922 → 0x0339`，`AH=4A BX=6A07` 由失敗轉成功，
之後還成功要到 `0x6A27`。停點消失。

**re/04 §3 的「最佳假說」（`6A04:0071–007B` 連推數個 0 當 record 欄位）
是誤讀**：`6A04:057F` 是雙向串列插入，那兩個 0 是「插到最前面」的哨兵。

## 4. 中文字形要 write mode 3（confirmed）

回收記憶體之後選單框畫得出來，但字形只剩零星像素。probe 新增的統計
顯示鏈內 write mode 使用次數 `0=42048 1=0 2=480 3=1160`——**mode 3 用了
1160 次**，而 spec 009 §4 當時把它列為「未見使用，照 mode 0 做」。

write mode 3 的 CPU 位元組是**遮罩**不是顏色，顏色一律取自 set/reset
暫存器。實作後字形完整（§1 的圖）。

同時發現 mode 0 的 enable set/reset（`gc[1]`）完全沒實作。硬體語意上
那是錯的，已補；實測影響 297,587 個像素的底色（色號 15 → 3，對照檔
`workplace/re05/ab-nosetreset.bin`）。**底色究竟該是哪一個要等原版對拍
才算定案**，本輪不宣稱畫面 parity。

## 5. 還沒到的地方

- **調色盤淡入沒有走完**：DAC 的 6 位元值多半停在 2–15，畫面實際很暗。
  §1 的圖是假色盤。淡入的驅動來源未查（未知）。
- 因此**還沒有任何對拍等級的畫面**。要拿 `same-state` 收據，得先讓
  淡入完成，並建立原版側的固定輸入序列。
- `int 10h AH=10h` 的 `AL=09/17` 各 1 次語意仍未知（不擋路）。

## 6. 重現命令

```
cd ~/cht/dosgolem-logh3
DOSGOLEM_ORIG=~/cht/logh3/org_game/extracted tools/go.sh run ./cmd/probe \
  -exe /orig/GIN3.COM -root /orig -steps 60000000 -tick 20000 \
  -click-x 320 -click-y 200 -click-at 5000000 \
  -dump-ega /src/workplace/re05/screen-60M-wm3.png
```

一次執行約 4 秒（2576 萬道指令）／約 9 秒（6000 萬道）。反組譯用
`workplace/re05/disasm.sh <起始偏移> "<hex bytes>"`（objdump 走容器），
記憶體用 `-peek "<段>:<偏移>:<長度>"`。

產物（一次性證據，gitignore）：`workplace/re05/`。
