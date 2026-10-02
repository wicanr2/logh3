# 044：交錯移動資料與正常玩家路徑

工作項：[#79](https://github.com/wicanr2/logh3/issues/79)。
規格080維持READY，完整同狀態畫面收據仍缺；這份驗證不宣稱整戰場或八步完成。
本輪父提交：`704f72ff733feaca73e9dda673af73c502e1093a`，測試時工作樹有本輪修改。

| 範圍 | 結果 | 限制 |
|---|---|---|
| 原版完整移動資料 | 初始化及33圈終態，7筆122-byte艦隊與55筆78-byte分隊零差異 | 同狀態資料比較，不是畫面比較 |
| 正式Game轉接 | 原版55筆清單、隱藏索敵隊、全紀錄回寫零差異 | 正常入口後明示注入原版初態 |
| 常式案例 | 38筆艦隊、55筆分隊、9筆碰撞全部通過 | 102筆包含既有7筆重複Action，沒有增加獨立案例數 |
| 正常滑鼠命令 | 長程、短程、隊形、索敵與一般分隊實行、逐幀顯示及存讀通過 | remake正常路徑抽樣，不是原版同狀態畫面 |
| 自動戰術與被動電腦 | 可選兩種戰術結果不同；被動電腦移動0、玩家損失0 | 自動戰術是remake既有功能，不宣稱原版AI同態 |
| 標準測試 | `./internal/... ./cmd/...`全部18套件通過 | 綠色測試不代替未驗原版局面 |

## 原版與工具身份

- 原版`GIN3.COM`：`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
- 原版移動入口RAM：`9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d`。
- 完整局面JSON：`b920eb42ccc70c8c447a9a2bd53a5e40f9efae5e807929fe4972fa3cc5fbe516`。
- 102案JSON：`054f1aef81a139da6ed85d8979fa72e647574c3984c6a73a3649a48c4eeba00c`。
- dosgolem專用副本：`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，本輪回讀為乾淨，未修改執行器。
- 原版雙跑完整命令及三邊界見[研究371](../re/371-battle-movement-phase.md)；沒有換名或重包舊圖。
- 兩側測試seed：`0x12345678`。移動規則沒有RNG輸入，既有原版收據的seed未變；
  正常玩家測試於實行前設定測試seed，不宣稱後續其他階段骰序一致。
- 工具沿用`golang:1.24-bookworm`及`logh3-ebiten:2`，產品測試Go 1.24.13。
  image身份依序為`sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`、
  `sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8`。

## 可重跑命令

```sh
bash tools/gui.sh test -count=1 -v -run 'TestBattleMovementPhaseMatchesOriginal|TestBattleUnitStepsMatchOriginal|TestFleetSingleStepMatchesOriginal|TestBattleMovementCollisionMatchesOriginal' ./internal/rules
bash tools/gui.sh test -count=1 -v -run TestGameMovementPhaseMatchesOriginal ./internal/game
bash tools/gui.sh xrun sh -c 'LOGH3_MOVEMENT_OUTPUT=battle79-round31-normal-fixed go test -count=1 -v -run "TestMovementCommandsExecuteAndSave|TestBattleComputerDealsNoDamage" ./internal/game'
bash tools/gui.sh test -count=1 ./internal/... ./cmd/...
```

需要原版RAM的整階段測試缺檔會明示skip。本輪原版RAM存在，沒有skip。
正式Game測試先由主選單進S3戰場，再明示載入原版第6回合初態；
只把entry餵入引擎，end僅用於最後比較，沒有排除byte。

正常命令測試走主選單、防拷、S3載入與戰略實行後的達希利戰場。
選點、朝向、隊形及「實行」全部走滑鼠命中測試。移動中重按實行及左右鍵不能重入。
保存先檢查31筆一般池或含索敵隊的32筆池之移動欄位未被SyncFleets覆蓋，
再用正式SaveSlot／LoadSlot驗完整SHIP 27456 bytes與FLEET 3904 bytes逐位元組相同。
這不宣稱中途戰鬥重新載入後可繼續，也不宣稱未抽樣資源欄位符合原版。

| 正常命令 | 起點到終點 | 移動中間幀 |
|---|---|---:|
| 長程 | (416,280)到(584,288) | 35 |
| 隊形 | (419,258)到(506,291) | 4 |
| 索敵 | (416,280)到(584,288) | 21 |
| 分隊 | (419,258)到(587,266) | 22 |

每種模式保存第2個移動幀及第2回合終態，原生640×480色號、48-byte色盤與PNG，
檔名前綴為`workplace/battle79-round31-normal-fixed-`。
色盤以`(v<<2)|(v>>4)`展開6-bit DAC，與正式Palette.RGB8相同，沒有裁切或縮放。
長程中間PNG：`240029b5a8e43ab7ca366c4dffafc0e80bc80a78e2d4c5f346f72c962f703e53`；
長程終態PNG：`c5d1e590bd14fec9648392ac453a1edee437cd8c29de722b009efd4c78b40ed1`。
全部輸出雜湊與來源身份另存本機`workplace/battle79-round31-meta.json`。
metadata SHA-256：`5fcf67462baf5319e268cc75adb19005fc717cfcc1bd5902e4d2dd2d95f92edd`。

## 測試收據與限制

| 本機收據 | SHA-256 |
|---|---|
| `battle79-round31-rules-fixed.log` | `50117d096478acb8fc9535f8365659f1ff260ddfab0fd4935568d5021d8f4006` |
| `battle79-round31-game-phase.log` | `93ba67c890fc7acf8c9ba9c2d77d9d037e130c750f734a28bd6a034864a197c0` |
| `battle79-round31-commands-fixed.log` | `55a48d437041e4af0d28d0135f0982d331f3a0c8929c7b48c63a70a06d8886d6` |
| `battle79-round31-standard-fixed.log` | `d21aff858377fe72bd5f996ee987596ae2e0ed1f4840030f65fda1f94b408cbb` |
| `battle79-round31-gui-main.log`及`battle79-round31-gui-fcmd.log` | `0beb31c5452f87f69e4bacdbe96d50cf12cd90713ea6db51c2b5f0e96fbbc34e` |

GUI檢查先後用`bash tools/gui_check.sh main1e`及`bash tools/gui_check.sh fcmd-s3-found`；
此包裝器目前不依參數篩選，兩次實際各跑60份既有收據，全部通過、零跳過，log相同。
這些是既有GUI回歸，不代替新增移動階段的同狀態畫面。

正式整階段畫面尚未保存原版／remake／差分與逐區統計；不能將上述正常PNG升格。
索敵回收資料與隱藏行為已接線，完整正常放出、回收、再放出及其畫面仍須抽驗。
第二戰場、全部存檔與第③／④等其他回合階段仍由#79接續。
