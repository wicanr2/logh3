# 收據 003 — 防拷問答畫面（same-state，逐位元組一致）

比較等級：**`same-state`**，未核准差異 **0**。

| 項目 | 值 |
|---|---|
| 原版側收據 | `workplace/oracle/cb1/`（一次性證據，gitignore） |
| 執行器 | dosgolem-logh3 `6841e68`（工作樹乾淨；`2e9dd3b` 上重跑逐位元組相同） |
| 固定序列 | `tools/oracle.sh cb1 -steps 130000000 -tick 20000 -clicks 40000000:300:170,95000000:300:170 -click-hold 3000000` |
| `GIN3.COM` SHA-256 | `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918` |
| `screen.bin` SHA-256 | `10dd957db9f5a5d3c0954176b34e9c6e1f3ba885bfde0dc4a6de756f4f88ff61` |
| remake 側 | `screen.CopyProtection`（規格 `docs/internals/data/screen-copy-protection.json`） |
| 比對 | 640×480 全畫面色號陣列，**0 / 307,200 不同** |
| 回歸測試 | `internal/screen/copyprot_test.go`（缺原版或收據就 skip） |

滅相位另有一份同序列的收據：

| 項目 | 值 |
|---|---|
| 原版側收據 | `workplace/oracle/cb1dark/`（一次性證據，gitignore） |
| 執行器 | dosgolem-logh3 `453c32e`（工作樹乾淨） |
| 固定序列 | `tools/oracle.sh cb1dark -steps 110000000 -tick 20000 -clicks 40000000:300:170,95000000:300:170 -click-hold 3000000`（停在 [`docs/re/281`](../re/281-copy-protection-caret-blinks.md) 量到的滅半週期 #107.6M–#114.0M 之內） |
| `screen.bin` SHA-256 | `6b11b9bb5d054b81cdac9cf8759db9c5e6897c2583066e50e18eb31370f9b702` |
| 與 `cb1` 的差異 | 只有 (240,283) 起 16×2 那 32 個像素：15 → 9 |
| remake 側 | `screen.CopyProtectionAt(…, CopyProtCaretHalfTicks)` |
| 比對 | 640×480 全畫面色號陣列，**0 / 307,200 不同**（`TestCopyProtectionDarkPhaseMatchesOracle`） |

## 這個畫面由什麼組成

| 元素 | 來源 |
|---|---|
| 背景 | 色號 3 整片 |
| 視窗 350×117 @(129,181) | 8 層外框，**前三層左緣是 1:1 交錯（8／15）**，填色 9 |
| 頭像 80×100 @(144,192) | `FACE.AMG` 第 0 張（`internal/asset`） |
| 三行中文 | `GRAPH.IMG` 字型，色號 7、陰影 3、陰影往上一列 |
| 輸入框 36×20 @(414,260) | 2 層內凹外框（`{7,3,3,7}`），填色 8，**畫序 TRBL** |
| 輸入游標 16×2 @(240,283) | 下一個要打的那一格底下的底線，色號 15；畫／擦各 320 個 PIT 中斷（[`docs/re/281`](../re/281-copy-protection-caret-blinks.md)）|
| 四個角落 | 26 個量出來的修正像素 |

## 三件會安靜出錯的事

**交錯的邊當成純色**：框看起來完全正常，只有一半的像素不對。

**框的畫序**：角落歸最後畫的那一邊。主選單那種外凸的框用 `LBTR`
（角落歸上緣與右緣），這裡的輸入框是內凹的，要 `TRBL`。順序錯了
整片框其他部分完全相同，只有四個角不一樣。

**輸入游標會閃**：週期 640 次計時器中斷（約 1.1 秒），`cb1` 停在
亮著的那半個週期、`cb1dark` 停在滅的那半個。remake 以 `CopyProtectionAt`
的可重播 PIT 相位驅動，亮／滅兩個半週期各有一份同狀態收據逐位元組對過；
錄新收據時要先看 `-steps` 落在哪個半週期，否則紅的是相位不是版面。
