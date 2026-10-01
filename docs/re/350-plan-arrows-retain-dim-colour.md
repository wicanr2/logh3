# 350 — 點過的預算箭頭保留暗色，月初入口漏了動態繪圖

日期：2026-10-01。工作項：[#80](https://github.com/wicanr2/logh3/issues/80)、
[#82](https://github.com/wicanr2/logh3/issues/82)。調查採 dosgolem 動態實驗，沒有反組譯或修改執行器。
正式契約在 [規格 003](../spec/003-budget-panel.md)，穩定事實回填
[熱區參考](../internals/23-hotspot-lists.md)。重生入口：`bash tools/plan_arrow_receipts.sh`。

## 輸入、工具與位址基準

- `org_game/银河英雄传说3SP.zip` SHA-256：
  `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
- `org_game/extracted/GIN3.COM` SHA-256：
  `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
- dosgolem-logh3：`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，工作樹乾淨，
  容器 `golang:1.24-bookworm`；分析為 `tools/py.sh` 的 `python:3.12-slim`。
- 原版 VGA mode 12h，640×480 色號陣列；remake 同尺寸，沒有裁切、縮放或遮罩。
- 這次所有 `lin:` 均為 dosgolem 執行期實體線性位址。DS=54B3 時，
  `lin:5DB04` 對應 `ds:8FD4` 稅率檔位、`lin:5DAC2` 對應 `ds:8F92` 建造費百分比。
  熱區的穩定定位仍用 `ds:14A8`，第 0／1／4／5 筆，沿用 `docs/re/343` 的位元組證據。
- 原版初始狀態來自正常玩家路徑的完整檢查點；箭頭不使用亂數。
  正式批次的 `flags.txt` 保留完整參數，`inputs.json` 保存父檢查點 SHA-256 與 image。
  稅率上界探針的注入完整列在旗標中，不當成正常玩家路徑。

## 已證實的呈現規則

1. **單純游標懸停不改箭頭色號。** `fcx-20261001-arrow-hover` 的箭頭與未操作底稿
   除游標覆蓋之外相同；screen.bin SHA-256：
   `e67311e83f45448c1553e95f66712a6be763afe3d65dd4c2df45d0b1489f9f4b`。
2. **點擊後該箭頭全部色號 9 改成 8，其餘色號、幾何與游標不變。**
   向上箭頭有 26 個色號 9；游標遮住 5 個，因此既有收據只看見 21 個差分。
   原版點擊後移開游標仍有 26 個差分，皆為原版 8／remake 9：
   `fcx-20261001-arrow-move` SHA-256
   `e19004cdd2c55d3e770ceac1a933bc740bf0635f57042f1c208203103cca1bad`。
3. **各箭頭獨立保留暗色，不只保留最後一顆。** 先加後減時，上箭頭的 26 個暗色仍在，
   下箭頭也變暗。`fcx-20261001-arrow-updown` SHA-256：
   `06e7f2c6c9b24f99cfc73978e0f8360b440150fe4e936d624add455a532e5b3c`。
4. **關窗再開重設四顆箭頭。** `fcx-20261001-arrow-reopen` 與初始底稿的四個
   箭頭矩形逐位元組相同；SHA-256
   `f707bf502f1607b7fa1ebf532f0dfc0c42c4131f9f16ae5a0f447b5e69faca64`。
5. **數值夾在上界也會變暗。** 注入 `lin:5DB04@451000000=04` 後按稅率＋，
   `-peek` 仍為 04，箭頭已轉色；`fcx-20261001-arrow-clamp` SHA-256
   `5022a517d85f68acf0881a0d74fb172ba521a69e2576a467ef3133cd24a6b602`。
6. **第四顆建造費－同規則。** `fcx-20261001-arrow-pct-down`，游標未遮住的
   色號 9→8 有 17 點；SHA-256
   `940c54e9920caeb399c87302472ed322127cf132ccc3755b578e8c0905b276f0`。

這些是執行期已證實的有限行為；造成轉色的遊戲內部呼叫與用途仍未知，
不把「按下樣式」「停用」等名稱當成已證實語意，也不為呈現修正另挖無關 helper。

## 月初預算入口的獨立缺口（#82）

原版 `budget1` 正常路徑後按建造費＋，百分比 25→26，建造費 15185→15792，
剩餘 45556→44949；兩條等價輸入（冷啟動與從完整檢查點展開）得到同一個雜湊：
`07e28842e9f8d05d2d0f4c3ac190c76c4c36767dc74271fc88308871af4e983f`。
修正前正式 Ebiten 顯示層差 531 像素，範圍 (368,324)–(483,355)。

稅率＋時原版標籤變重稅、稅額 16070、預估軍事費 66070、建造費 16517、剩餘 49553。
原版仍以報告國庫 50000 加新的預估稅額，百分比仍為 25；不是改用命令階段國庫 45592。
`fcx-20261001-budget-tax-up` SHA-256：
`42b002f5852f1949e068755457b1c5f5525f56899758330a9904bf7a4cc1ab77`。
減稅與減建造費另有 `fcx-20261001-budget-tax-down`、`fcx-20261001-budget-pct-down`，
正式批次四顆全收。remake 原先 `StateBudgetPlan` 沒有接即時游標、稅率標籤與
動態報告預算，這才是較大的差異來源。

## 月初試作被新證據否定

試作把 `monthReport.Tax` 換成現行 `taxTotal`，兩個建造費箭頭可對上，
但稅率＋仍差 995、稅率－差 846 像素。原版 `fcx-20261001-budget-memory`
傾印證明兩側不在同一結算階段：原版還沒有扣稅，remake 已完成 `CollectTax`。
兩側稅率與成長率相同，但十個帝國星域經濟力均不同；星域 0 為 85020／82302。
因此加稅後原版總稅收 16070，remake 15472，不能靠改繪圖達成完整機制對拍。

工具、位址基準同上。原版完整記憶體：

| 檔名（dosgolem 工作副本 workplace 下） | SHA-256 |
|---|---|
| `arrow-budget-base-memory.bin`，1,048,576 bytes | `990c271c2193be1b99c4681ca996b569acebf1ae2586c9b89326d0b9d9ee442e` |
| `arrow-budget-tax-memory.bin`，1,048,576 bytes | `5ce61722f9fa40d3a251911995ca63e6b855a9ee16a75e84901cbcf07b5e1db0` |

比較範圍：dosgolem 執行期 `lin:5F2CC` 起 4032 bytes，32 筆 × 126；
原版重算前後只改星域 `+0x26` 稅率與 `+0x7D` 未安定度，經濟力不變。
行星比較範圍為 `lin:53132` 起 6642 bytes，81 筆 × 82；remake 對應
`Planets.Data` 跳過 64-byte 檔頭。沒有混用 IDA 位址或檔案偏移。

月初試作與臨時診斷測試已撤回。規格 003 的這個擴充退回 DRAFT，#82 保持 open。
下一步必須完成真實收稅與造船的前後狀態鏈，不能用預算畫面一致取代結算一致。
開窗原版游標與標籤也要另外處理：月初標籤是 `report` 第 9 行，徵稅則是 `overlay` 第 8 行。

## 對舊結論的勘誤

`docs/re/343` §七第 3 點將差異描述成「游標周圍一小塊」，並提出按下圖塊假說。
游標移開仍有 26 點且先加後減兩顆同時保留，足以排除游標形狀／位置是根因；
差異是整顆箭頭的固定色號轉換。舊收據與歷史保留，新規格取消 `planArrowSkip`。
按住連發仍屬 #60，這次沒有量測其時程或宣稱完成。

## 正式收尾

14 份正式原版收據兩次重錄的色號陣列雜湊全部相同，執行器未改。
八份徵稅正常玩家入口的完整畫面、RGB 與逐區差分均為零，
另有上界注入診斷及既有三份整張測試。正式收據與差分腳本入口在
[對拍 025](../playtest/025-plan-arrows.md)。月初結算仍留 #82，不升格。

月初記憶體探針的完整可重播參數（探索，未當成正式驗收）：

```sh
tools/oracle.sh fcx-20261001-budget-memory \
  -load-state /src/workplace/planarrow-budget.state -steps 210000000 \
  -mouse-x 471 -mouse-y 249 -mouse-at 170000000 \
  -clicks 185000000:471:249 -click-hold 100000 \
  -dump-mem-at '165000050:lin:0:1048576:/src/workplace/arrow-budget-base-memory.bin;209999999:lin:0:1048576:/src/workplace/arrow-budget-tax-memory.bin' \
  -peek lin:5DB04:2,lin:5DAC2:2
```

往後關窗探針 `fcx-20261001-budget-tax-command` 的命名不代表已到命令階段。
實際截圖仍停在船艦訂造報告；不能用它宣稱結算鏈已核對完整。
