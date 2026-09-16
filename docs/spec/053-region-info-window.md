# 053 — 星系情報視窗（戰略地圖介面左鍵點星圖）

狀態：**CONFORMED**：十批收據（九批 `fcmd-rinfo-*`、`fcmd-map-rclick`）整張 640×480 逐位元組相同（一塊事先列出
的排除區：小圖最後一列 128 個像素，見已知差異）；顯示層（Ebiten）`fcmd-rinfo` 同樣相同
（`tools/gui_check.sh`）。
證據：[`docs/re/320`](../re/320-region-info-window.md)、[`docs/re/314`](../re/314-fleet-standby-planet.md)
§四（小圖）、[`docs/re/89`](../re/89-minimap-unit-markers.md)（行星位置）
資料：[`docs/internals/data/region-info-panel.json`](../internals/data/region-info-panel.json)、
[`docs/internals/data/capital-panel.json`](../internals/data/capital-panel.json)（`unknown`）
執行器：dosgolem-logh3 `b5485eb`（`tools/fleetcmd_receipts.sh` 第 18 段）
收據表：[`docs/playtest/021`](../playtest/021-strategic-fleet-command.md)

## 怎麼到得了

戰略地圖介面上左鍵點星圖的一格。

## 規則（typed input → 狀態）

| # | 輸入 | 規則 | 收據 |
|---|---|---|---|
| R1 | 左鍵點右側面板看的那一格（`ds:9588`：開局是首都，換選過就是換選的那格） | 開星系情報視窗：格 (0,0) 起 20 × 19 格蓋住整個左半，色 0x97 | `fcmd-rinfo`、`fcmd-rinfo-fezzan2`、`fcmd-rinfo-ballat2` |
| R2 | 左鍵點別的格 | 換選星域（`073F:02EE`）：地圖反白換到那一格、右側面板摘要換成那一格（不是自己的：成長率／徵稅額／稅率印 `------`／`----------`／`-----`，防衛力與經濟力照印）、四個分頁重建（位移歸零）、清單改看那一格；點同一格不動 | `fcmd-rinfo-fezzan`、`fcmd-rinfo-ballat` |
| R2′ | 右鍵點星圖 | 什麼都不換：`ds:9588`、地圖反白、右側面板與清單都不動；左下角那塊只是跟著游標（規格 001） | `fcmd-map-rclick` |
| R3 | 視窗內容 | 陣營名（陣營色）＋「星系」＋ 星域名；「星系內行星」與每顆行星名（擁有者色，列 2 起）、小圖 128×128 貼在 (24,85)、外框 (23,84)–(152,213)、每顆行星從 (160, 110＋20i) 拉引線到 (24＋X>>4, 85＋Y>>4)（先畫 (−1,−1) 色 3 影子，再畫擁有者色）；「待命船艦 艦數 部隊 補給物資」七列：自己的星域印七種艦艇的待命艘數（星域 `+0x30`）、艘數 ÷ 滿編、飛彈與戰鬥機（星域層 `+0x68`／`+0x70`）；不是自己的印七行寫死的破折號（第六行「Ｓ」是原版的錯字，照印）；右下角「結束」 | 同 R1 |
| R4 | 行星位置 | 載入星域地圖時算：先收圖塊 6（最多五格），不夠補要塞的圖塊 22；每格 (欄×16＋8, 列×20＋12)，先依 y 再依 x 排序，依序對到行星（`rules.PlanetSites`，指定待命行星與戰場小地圖同一支） | `fcmd-rinfo-ballat2`（魯德米拉在要塞格） |
| R5 | 左鍵「結束」／右鍵 | 關窗，左半重畫；左鍵點別處不動。關窗那一刻左下角那塊照游標底下那一格重寫（不在星圖上就空） | `fcmd-rinfo-close`／`-rclick`／`-outside`、`fcmd-rinfo-ballat2-close` |

## 已知差異

| 差異 | 類別 | 到期條件 |
|---|---|---|
| 小圖最後一列（畫面 y 212、x 24–151）：原版那一列的來源列超出地圖，讀到的是緩衝區後面的東西；remake 畫底色。對拍時跳過那 128 個像素（與規格 051 指定待命行星那條同一件事） | 原版讀越界 | 不到期 |
| ~~換看同盟或中立星域時右側面板的「經濟力」（x 432–479、y 224–233）~~ | 已到期（2026-09-16，#45）：月底收稅改成逐座位、中立不扣（`docs/re/321`）之後整張都比，巴拉特 68282／費沙 103100 相同 | — |

## 驗證

- `TestRegionInfoMatchesOracle`：`fcmd-rinfo`／`-outside`（排除小圖最後一列）、`-rclick`／`-close` 整張。
- `TestRegionInfoOtherRegionsMatchOracle`：`fcmd-rinfo-fezzan`／`-ballat`／`-ballat2-close` 整張、`-fezzan2`／`-ballat2`（排除小圖最後一列）；`selectedRegion` 14／31。
- `TestMapRightClickKeepsPanel`：`fcmd-map-rclick` 整張；`TestRightClickOnMapKeepsSelection`：右鍵四個星域之後 `SelectedRegion()` 仍是 −1、左下角是游標底下那一格。
- GUI：`tools/gui_check.sh` 的 `fcmd-rinfo`（排除那 128 個像素）。

以上全綠（2026-09-16；`go test ./...` 與 `tools/gui_check.sh`）。
