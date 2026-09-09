# SPEC 031 — AI 回合的艦隊整補

狀態：**CONFORMED**（已實作；收據 `refit2` 的七個可觀測量逐一相同）
證據：[`docs/re/227`](../re/227-ai-turn-refit-step.md)、
[`228`](../re/228-refit-callees.md)、[`229`](../re/229-refit-chain-complete.md)、
[`230`](../re/230-composition-change-refresh.md)、
[`231`](../re/231-refit-receipt.md)、
[`docs/internals/33`](../internals/33-fleet-table.md)、
[`docs/internals/54`](../internals/54-strategic-ai-decision.md)
相關：[`docs/spec/029`](029-strategic-ai-turn.md)（AI 回合的順序）、
[`docs/re/215`](../re/215-ship-allocation-from-region-stock.md)（星域待命部隊）

## 要做什麼

原版 AI 每個回合在建艦隊之後、五個階段之前，對**自己陣營、停在自己
星域、而且那一格沒有戰鬥**的每一支艦隊做兩件事：

1. **併殘部**：同一個艦種裡不到七成五的那幾組拆掉，艘數還回星域待命
   部隊，再從待命部隊補到剛好湊滿整數組。
2. **補編制**：照類別鍵的樣板把組數補回「應有組數」（`+0x0E`）。

remake 完全沒有這一段——AI 的艦隊打完仗就一直缺著，而原版會在自己
地盤上補回來。

## 這是 AI 專屬的

`2D76:0732`／`07BA`／`0A44` 各只有一個呼叫者，全在 AI 的段裡
（`docs/re/229`）。玩家補編制走「艦隊編成」的另一條路。
**所以 remake 要把它放在 AI 那一層，不是 `internal/rules` 的共用編制。**

## 規則

完整虛擬碼在 [`docs/re/229`](../re/229-refit-chain-complete.md)。要點：

| 項目 | 值 |
|---|---|
| 觸發 | AI 回合，`2D10:0005` 的第三步 |
| 篩艦隊 | `+0x04 != 0xFF`、`+0x01 == 目前陣營`、星域 `+0x12 == 陣營`、星域 `+0x13 & 4 == 0` |
| 「缺編」判準 | 現有 × 4 < 滿編 × 3 |
| 補多少 | `滿編 − (該艦種缺編組的現有總和 mod 滿編)`，星域待命部隊夠才補 |
| 併完剩幾組 | `總和 ÷ 滿編` |
| 還艘數 | `星域[+0x30 + 陣營×28 + 艦種×4] += 艦艇[+0x08]`（艦種 7 不還）|
| 補編制跳過 | `+0x09 == 0xFF` 或 `+0x08 == 4` |
| 樣板給 0 | 用**旗艦的艦種**（無條件——與建艦隊的 `0FF4:11F4` 不同）|
| 滿了要犧牲誰 | 隊形槽已經不在樣板前 `+0x0F` 組裡的那一組 |
| 排序 | 每次改動後 `[2]` 升冪、相同再比 `[3]` 升冪 |

## remake

| 檔案 | 什麼 |
|---|---|
| `internal/rules/fleetrefit.go` | `MergeUnderStrength`（`07BA`）、`TopUpComposition`（`0A44`）、`Refit` |
| `internal/rules/region.go` | `ReturnRegionShips`（`2703:155B`，`TakeRegionShips` 的反向）|
| `internal/ai/ai.go` | `Refitter` 標記介面，比照既有的 `Builder`；只有 `OriginalStrategy` 有 |
| `internal/game/originalbrain.go` | `refitFleets`；`runBrain` 在 `Build` 之後、`Orders` 之前呼叫 |

## 驗收

收據 `refit2`（`docs/re/231`）：用 `-poke` 把提督 52 那一支的兩組艦艇
`+0x08` 改成 1000（滿編 3000）造出缺編局面，整補前後各倒一份 1 MB。
**七個可觀測量全中**——三次 `+0x0F`（4→2→3→4）與寫入它的常式、
兩組的最終艦種與艘數（艦種 0 補滿 3000、艦種 2 換上 4000）、
兩種待命部隊的餘額（艦種 0 剩 2000、艦種 2 剩 0）。
釘在 `TestRefitMatchesReceipt`；`TestRefitLeavesFullFleetAlone` 釘
「全滿編就什麼都不做」，那正是 `turn260mem` 第一個 AI 回合的局面。

收據 `refit3`（`docs/re/232`）另外釘下**編制已滿要犧牲一組**那一條：
把一組的隊形槽 `-poke` 成 99（樣板裡沒有），整補把它挑出來重建成樣板
要的槽 10，待命部隊前後相同（拆掉還回去與重建拿走是同一個數）。
釘在 `TestRefitSacrificesStaleSlot`。

收據 `refit4`（`docs/re/233`）釘下**替代清單裡的 `0xFE`**：把那一格的
艦種 0–4 存量清成 0、5 與 6 留著，補編制走替代清單到 `0xFE`，兩邊都是
0 組時先試 6。釘在 `TestRefitBalancesFiveSixFromReceipt`。

收據 `slot2who`（`docs/re/234`）釘下最後一條：**覆蓋同槽**。把一組的
`[2]`（艦種）`-poke` 成樣板要的以外的值，監看抓到補編制自己把它寫回去
（`#242632789 616B4: 02 → 03 ← 2D76:0BE5`，也就是
`mov es:[si+2], al`），時間點在整補進入之後 1757 道指令。
釘在 `TestRefitOverwritesWrongKindInSlot`。

**四條路徑都有收據了**：不動（全滿編）、併殘部、犧牲一組、覆蓋同槽，
加上替代清單的 `0xFE`。

## 一處刻意的簡化

原版 `2D76:0A44` 的迴圈上界 `艦隊[+0x0E]` 是**每一圈重讀**的，而且
迴圈自己會改它——每補一組就走 `297B:03EB` → `297B:0546` →
`297B:05B2`，那一支把 `+0x0E` 重算成 `297B:0209(提督等級)`
（`docs/re/233`）。remake 的 `want` 是固定參數；正常局面下兩者相同
（`rules.FleetSlots` 算的就是同一個數），所以不模擬。

## 不在這份規格裡

- **改完編制之後的那三支重算**（`297B:03EB` 的尾段、`297B:08D8`、
  `297B:0546`，`docs/re/230`）。remake 的 `rules.Fleet` 沒有獨立的艦艇池，
  組序同步不用做；能力重算對應的是 `Operations` 與 `Upkeep`，
  兩者在整補前後都沒變（提督與副官沒換），所以這一輪不動。
- 玩家「艦隊編成」補編制的那一條路。
