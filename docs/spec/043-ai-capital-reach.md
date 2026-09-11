# SPEC 043 — 戰略 AI 的「從首都走得到嗎」

狀態：**CONFORMED**（新遊戲第一個 AI 回合同狀態；範圍見「驗收」與「沒有驗到的」）
證據：[`docs/re/298`](../re/298-ai-capital-reach-is-not-all-ones.md)（`t260_post.bin`
分頁框裡的距離表逐格、`s3aicomp` 正對照）、
[`docs/re/185`](../re/185-reachability-and-distance-table.md)（`2B99:0392` 的算式）、
[`docs/re/243`](../re/243-who-fills-the-distance-table.md)（距離表怎麼建）、
[`docs/re/244`](../re/244-region-posture-recompute.md)（什麼時候重建）
相關：[`docs/spec/029`](029-strategic-ai-turn.md)（戰略 AI 回合）、
[`docs/spec/041`](041-rebellion.md)（叛亂與 AI 共用亂數）、
[`docs/internals/26`](../internals/26-star-region-graph.md)（路徑表）

## 原版做什麼

```
2B99:0392(首都, 陣營, &旗標[32]):
    對 i = 0 … 31：旗標[i] = (2B99:028F(首都, i, 陣營) < 0800h) ? 1 : 0

2B99:028F(a, b, 陣營) = 距離表[陣營][a × 32 + b]      ; EMS 分頁框，word
距離表 = 2B99:007D(陣營) ＋ 2B99:01B1                   ; 見 internals/26
    邊權 0xFF → 0x7FFF；兩端不全屬於該陣營 → 邊權 ＋ 0x400；Floyd–Warshall
```

距離 ＝ `1024 × k ＋ 沿途邊權`，沿途邊權永遠小於 1024，所以

```
旗標[i] = 1  ⟺  從首都到 i 的路上，最少只要踏過 k ≤ 1 條「兩端不全是自己的」邊
```

**不是全 1**。星圖全連通，但從自己的地盤出發只走得到自己的地盤與緊貼著它的
那一圈。新遊戲同盟（首都 31）走不到 0–12、25，帝國（首都 0）走不到 12、17–31
（`t260_post.bin` 分頁框逐格）。

### 誰用它

| 呼叫者 | 用法 | remake |
|---|---|---|
| `2D54:0000`（階段四、五的攻擊目標分）| 旗標 0 的格子不計分——連敵方首都也一樣 | `ai.AttackBoard` |
| `2E54:019F` → `2E54:0222`（守勢優先度）| 只算「連通 ＆ 自己的」格子 | `ai.DefenceBoard` → `DefencePriority` |
| `2E54:0F09`（階段二的候選）| `2B99:028F(i, 首都) < 0x800`，方向相反，結果相同（下一節）| `StrategyInput.phaseNeutral` |

階段四對每個 `ds:D50E != 0` 的格子抽一次亂數（`2E54:08E3`），所以這張表**直接決定
一回合抽幾次**；叛亂判定與它共用 `ds:94DC`（`docs/spec/041`）。

### 方向

距離表不對稱（`ds:533A[2][6]` 那一個位元組），但罰則的條件兩端對稱、相鄰關係
對稱，最少要罰幾條邊因此對稱；跨不跨 `0x800` 只看它。所以
`2B99:028F(首都, i) < 0x800` 與 `2B99:028F(i, 首都) < 0x800` 永遠相同。

### 用哪一份擁有者

距離表在載入、新遊戲與回合尾巴（`056A:078E` 重算擁有者、費沙事件之後的
`2703:08DE`）重建。已知會寫 `+0x12` 的步驟（`2703:159B`、叛亂 `2703:1908`、
戰術戰鬥結束 `185C:02DA`、費沙事件）都在重建之前；戰略 AI 在下一個回合的
`056A:000C` 之前跑。所以 AI 查表那一刻，表 ＝ 目前擁有者的函數（**strong
inference**，新遊戲那一刻逐格證實）。

## remake 怎麼做

`internal/ai/original_reach.go`：

```
CapitalReach(圖, 擁有者, 陣營) [32]bool
    距離, _ = 圖.BuildTables(擁有者[0..31], 陣營)       ; rules：2B99:007D ＋ 2B99:01B1
    回 距離[Capitals[陣營]][i] < ReachLimit（0x800）
```

- `TurnInput.Reachable`／`StrategyInput.Reachable` 為 `nil` 時照它算，一回合一次
  （`Turn` 進門、`DefenceBoard`／`AttackBoard` 進門）；呼叫端要注入別的表才填。
- 用**呼叫當下**的擁有者，不把距離表存成遊戲狀態（上一節）。
- 圖沒有邊權表就 panic：`rules.LoadStarGraph` 保證有，沒有是組錯了圖。
  **不回「全部走得到」頂替**——那正是這份規格要修掉的錯。

## 驗收

- `ai.TestCapitalReachNewGame`：新遊戲擁有者，兩個陣營的可達表與 `t260_post.bin`
  分頁框相同，同盟那一列 32 格距離逐格相同。
- `ai.TestAttackBoardNewGameDrawsOnce`：同一局面同盟的 `ds:D50E` 只有 `[13] = 4096`、
  階段四抽 1 次。
- `ai.TestCapitalReachMatchesS3AIComp`：`s3aicomp` 的擁有者算出的候選是 13／15／22／28，
  分數 682／2048／2048／2048，與收據相同（原本是手填候選）。
- `ai.TestCapitalReachIsDirectionFree`：兩份局面、兩個陣營，跨 `0x800` 逐對對稱。
- `game.TestRebellionSameStateRandom`：remake 正式入口走 `rbx292e` 那條玩家路徑，
  第一個 AI 回合抽 **7** 次、叛亂那一圈進場種子 **AF5AAD71**、整圈之後 **936A823B**
  （`rbx297w`）、帝國 9 格 `r1 ＋ r2` 與叛亂的 7 格與收據相同。

## 沒有驗到的

- 讀檔路徑：重建點（線性 `0x04FD0`）與讀檔時擁有者重算的先後沒有逐位元組核對。
  存檔裡的 `+0x12` 與重算結果一致時沒有差別。
- 中局有易主之後的局面沒有執行期收據（新遊戲那一局沒有易主；`s3aicomp` 的分頁框
  當時不是這一頁，只有結果對得上）。
- 候選評分 `2E54:1574` 的「走不到 → 0」與距離、階段五 `2E54:1243` 的
  `2B99:028F(艦隊所在, 目標) < 0x800`，remake 仍用相鄰圖步數（`StarGraph.Hops`）。
  那兩處查的是「兩格之間」而不是「從首都」，不在這份規格裡。
- 階段二的排序鍵：原版依 `+0x78`（`2B99:028F(i, 首都)` 的距離）升冪，remake 用
  `StarGraph.Hops(首都, i)`。這一份只把「走不走得到」那道過濾換成原版的；
  兩者挑出不同的格子時會不同（新遊戲那一局兩者都是 24）。
