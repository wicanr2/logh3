# 060：艦隊交戰文字與正式色盤對拍

2026-10-04。比較等級是指定文字合成器的`same-state`。
原版其他層借固定入口；Game色盤另作正式診斷，不宣稱完整正常艦隊交戰或GUI通過。
版本、來源、完整注入與推論等級見[研究408](../re/408-fleet-presentation.md)及
[096 READY](../spec/096-fleet-presentation.md)。

## 可重現方法

原版輸出：`workplace/oracle/presentation48-r71-{a,b}/`，
由`bash tools/fleet_presentation_oracle.sh a`與`b`產生，拒絕覆寫。
固定dosgolem f7523f9、scene-01.state雜湊07cebb18；完整值在408及report。
兩次92案，276份RAM／色號／色盤相同；report只除命令名稱後相同。
凍結器：`bash tools/py.sh /work/tools/scripts/fleet_presentation_receipt.py`。
凍結JSON SHA-256 `2dbff381163ce7d512af633e496bc5349dc2597210c0131f7f0d0874b0943851`。

正式合成器最終輸出：`workplace/presentation48-r71-reviewed-remake/`。
空目錄下執行：

```bash
bash tools/gui.sh xrun env LOGH3_PRESENTATION_OUTPUT=/src/workplace/presentation48-r71-reviewed-remake \
  go test ./internal/screen -run 'TestFleetPresentationNative|TestFleetPresentationRejectsInvalidInput' -count=1
```

92案內容格、游標、文字色與全畫面各307200像素零差，總計28262400像素。
552份原版／remake／差分PNG及完整色號已保存；沒有裁切、縮放或像素排除。
修改區域由正式規格的視窗座標產生，698880個文字區域像素亦零差，逐區及全圖數字記在comparison.json。
logical→AC→DAC RGB逐色相同。圖片已目視，包含detail3顯示與正式Game色盤產物。

| 產物 | SHA-256 |
|---|---|
| comparison.json | 74b52252e130cf481427cc0d452c5058434a2a243be032e5f030ca7d0f15ad35 |
| sha256.json | 971042d304a64296843a840ce4813a04c85b758ac3900a311eaf6f39179fedb1 |
| round71-presentation-reviewed.log | 8fe041b08ec77468377fce20b3357823f9610f0db9b6a37dd6ed57eca4537e42 |
| round71-palette-game.log | 708c4255282dd9c859ea58f4485db35d9d45b4aa431db311ef148977d2cb66b5 |
| presentation48-r71-game/engagement.png | 0de06a163756ca08a74de3d2429c168e942b16b1e9f31ce911aae1fcb6252900 |

Game診斷命令使用同一容器：

```bash
bash tools/gui.sh xrun env LOGH3_PRESENTATION_GAME_OUTPUT=/src/workplace/presentation48-r71-game \
  go test ./internal/game -run TestEngagementPaletteMatchesNative -count=1
```

此測試由typed交戰fixture驗色盤及退出，不代替正常入口。
正式bootstrap與正常主選單開局由TestOpeningRuntimeTablesMatchesOriginal抽樣，資源鏈已接。
等待輸入未在本輪改動；投影、艦隊圖序列、正常同狀態與保存等待092。

標準命令`bash tools/gui.sh test ./cmd/... ./internal/... -count=1`正常結束，
18套件中13PASS、5無測試檔；round71-standard.log SHA-256
`8333561372328eefd281d6e96f1fee6520e875456480877510862e5d6d3bf22c`。
逐區統計補齊後，同一screen目標命令乾淨重跑通過；產品程式沒有再改。
最新89張遠端工作項目／9張OPEN，269核對零失效、2 manual、122一致、145歷史文件證據。
專案root-owned／誤建.md目錄為零，相關容器清理完畢；原版及dosgolem未修改。
