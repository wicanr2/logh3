# 058 — 戰略打行星（占據那一格的打行星、反擊、士氣、換手）

狀態：**規則層 CONFORMED**：兩份收據（`fcmd-assault-ast` 亞斯提打巴溫、`fcmd-assault-phz` 費沙）以原版進入
`20B4:1497` 那一刻的記憶體驅動 `rules.StrategicAssault`，與離開時的記憶體逐欄相同（亂數種子、行星擁有者／
防衛力／存量／飛彈／艦載機、艦隊士氣與 `+0x00`、每組艘數、提督功績、兩個計分器）。攻占訊息框與士氣低落報告見規格 059。
證據：[`docs/re/330`](../re/330-strategic-planet-assault.md)；事實整理 [`docs/internals/53`](../internals/53-strategic-planet-assault.md)。
收據摘要：[`docs/playtest/021`](../playtest/021-strategic-fleet-command.md) 第 22 段。
相關：[`054`](054-fleet-arrival.md)（走到占據那一格、bit4）、[`019`](019-planet-assault.md)（戰場上的打行星，**另一條算式**）、
[`047`](047-occupation-report.md)（占據報告）。

## 規則

時機：回合的逐格掃描（`0374:03E8`，星域 0..31）碰到 `+0x13` bit4 的格子（`054` 規則 4 設的），跑一次。

1. 攻方 ＝ 這一格有帝國艦隊（`+0x00` bit0 不算）就是帝國，否則同盟。攻方在這一格的艦隊依艦隊表順序排。
2. 迴圈、一回合的開火與反擊、分攤、士氣、換手、功績、計分：照 `docs/internals/53` 的算式，逐項以 `docs/re/330`
   為準。每一支先打（1 個亂數）、沒拿下再挨反擊（1 個亂數）；士氣低落（bit0）的不開火不耗亂數。
3. 艦隊的朝向由行星座標算（`atan(0x3F8 − x, 0x3F4 − y)`，`ds:593A` 查表），行星座標 ＝ `rules.PlanetSites`
   掃這個星域的戰場地圖（與原版 `2AF5:00CB` 算的 `+0x46`／`+0x48` 相同，`TestOccupyAstarteRunsStrategicAssault`）。
4. 回合結束的星域重算（`2703:08DE`）**不重算 bit4**；bit4 只在命令階段走完艦隊之後（`056A:0469`）重算。
   所以打到士氣低落的那一格，這一回合結束時 bit4 還在、下一回合才清（收據 `fcmd-occ-phz-t3` 18、`-t4` 之後停打）。

**remake：`rules.StrategicAssault`、`Game.strategicAssault`（`scanRegionsFrom`）。**

## 已知差異

| 差異 | 類別 | 到期條件 |
|---|---|---|
| ~~拿下行星時的訊息框與士氣低落報告~~：已接，見規格 [`059`](059-turn-transition-windows.md)「報告框」（2026-09-17） | — | — |
| 同一格兩支以上攻方：原版 `297B:0BD6` 把後定位的擠開 32 像素，距離 ≠ 0 → 打掉量有衰減、方向不同；remake 一律距離 0、同朝向 | 沒有收據 | 錄一份兩支攻方同格的收據 |
| 分隊的 `+0x06`（組士氣）、`+0x0C`（戰力）、`+0x0E`／`+0x12`（兩格存量）與 `20B4:18C2` 的索敵揭露：remake 的編制沒有這幾格 | 資料模型 | 艦艇池模型化時一起 |
| 同一回合先打過仗時戰場單位表可能殘留，守軍不一定是 0 | 沒驗 | 錄一份「同回合有戰鬥又有戰略打行星」的收據 |
| ~~電腦座位的攻占訊息框會不會顯示、會不會等~~：會（`docs/re/333` §二，收據 `fcmd-turn-ai-capture`） | — | — |

## 驗證

- `internal/rules/strategicassault_oracle_test.go`：`TestStrategicAssaultMatchesOracle`（兩批逐欄）。
- `internal/game/occupation_test.go`：`TestOccupyAstarteRunsStrategicAssault`（行星座標、回合流程）、
  `TestOccupyPhezzanArrivalMatchesOracle`（打完之後 bit4 仍在）。
- `internal/game/regionturn_test.go`：`TestRecomputeRegionsMarksOccupation`（回合結束不重算 bit4）。
