# SPEC 030 — `Blend` 與稅率重算

狀態：**CONFORMED**（已實作；收據 `turn260mem` 的軍事費、總稅收與建艦隊支數逐一相同）
證據：[`docs/re/222`](../re/222-ai-sets-its-own-blend.md)、
[`docs/re/223`](../re/223-tax-rate-recompute-formula.md)、
[`docs/internals/55`](../internals/55-blend-and-tax-rate.md)、
[`docs/internals/data/tax-rate.json`](../internals/data/tax-rate.json)、
收據 `turn260mem`（`workplace/t260_pre.bin`／`t260_post.bin`）
相關：[`docs/spec/006`](006-unrest-at-load.md)（`Blend` 是未安定度的參數）、
[`docs/spec/029`](029-strategic-ai-turn.md)（建艦隊的預算閘門）

## 要做什麼

原版的戰略 AI 在自己的回合開頭，按那個陣營的**軍事費**把
`Blend`（`ds:8FD4[陣營]`）重設成 0–4 的一個等級，接著**整批重算該陣營
每個星域的稅率**。稅率一動，總稅收跟著動，建艦隊的閘門
（維持費 × 20 >= 總稅收）就會提早或延後收手。

remake 目前完全沒有這一段：稅率停在開局值，AI 的預算永遠是開局的稅收。
`docs/re/219` 那個「多建一支艦隊」的差異就是它。

## 證據審查

| 項目 | 等級 | 依據 |
|---|---|---|
| `2D10:0319` 的全文與兩個呼叫點 | proven | `docs/re/222`；`farcall_to.py` 掃全傾印，兩處 |
| 軍事費 → 等級的分段函式 | **對拍** | `turn260mem`：同盟 45848 → 等級 1，監看抓到 `02 → 01` |
| 回合結尾歸零（非月底）| **對拍** | 同一份監看的 `01 → 00`，回合數 0，`0 % 10 != 9` |
| 回合結尾照軍事費重算（月底）| **對拍** | 收據 `monthblend`（`docs/re/238`）：把回合計數器 poke 成 9，結尾那一次寫入消失（值沒變），`Blend[同盟]` 留在 1、總稅收 10213，對照組是 0／6696 |
| 玩家座位不會被改動 | **對拍** | 兩份傾印帝國那一格從頭到尾都是 2 |
| `090A:0B93` 的算式 | **對拍** | `docs/re/223`：23 格逐一相同（帝國 10、同盟 13）|
| 稅率算式的民心三段（`m > 50`）| ×30／×50 有收據（`mrx295`，`docs/re/295`），×75 proven（無收據）| `mrx295` 把同盟兩格的民心改成 0／50，AI 那一次重算與算式相同；×75 那一段還沒有局面走到 |
| 成長率 `<= 0` 的那一條 | proven（無收據）| 量到的 23 格成長率都是正的 |
| 每個陣營一份軍事費（`ds:8F52[陣營×4]`）| proven | `docs/internals/27`、`docs/internals/30`；存檔 `WORK.DT?` 的 `+0x3E`／`+0x42` |

## 規則

### 等級

```
BlendForFunds(軍事費):
    >= 50000 → 2
    >= 40000 → 1
    >= 10000 → 2
    >=  5000 → 3
    >=     0 → 4
    <      0 → 2
```

### AI 回合的兩次寫入

```
AI 回合開始（建艦隊之前）：
    Blend[電腦陣營] = BlendForFunds(軍事費[電腦陣營])
    重算該陣營所有星域的稅率
AI 回合結束：
    回合數 % 10 == 9（月底）→ 同上重算一次
    否則                    → Blend[電腦陣營] = 0，一樣重算
```

玩家陣營那一格不走這條路徑。

### 稅率

算式與 `rules.TaxRatePerMille` 相同（新遊戲載入時已經在用），差別只有
**下標用 `Blend` 不是 `Rate`**、而且民心取星域記錄的 `+0x7C` 現值：

```
對每個「擁有者 ＝ 該陣營」的星域：
    星域[+0x26] = TaxRatePerMille(成長率, 民心, 擁有者, Base[Blend[陣營]], 統治者記錄)
```

新遊戲載入那一次兩個下標都是 2，所以既有的收據不受影響。

## remake 要動的地方

| 檔案 | 改什麼 |
|---|---|
| `internal/rules/growth.go` | `MonthEndTotals`：一趟結算回**兩個陣營**的徵稅額 |
| `internal/rules/taxrate.go`（新）| `BlendForFunds`、`RecomputeTaxRates` |
| `internal/rules/save.go` | `Military`／`TaxTotal` 改成每陣營一格，另讀 `Blend`（工作區塊 `+0xC0`）|
| `internal/game/game.go` | 軍事費與總稅收改成每陣營一格；月底兩邊都結算；`runBrain` 前後各設一次 `Blend` |
| `internal/game/originalbrain.go` | `Support.Money` 的 `tight` 改成 `Blend[陣營] >= 3` |

## 驗收

- `TestFirstAITurnMatchesReceipt` 建出來的艦隊數從 4 變成 **3**，
  前三支的提督／類別鍵／目的地不變。
- `TestTaxRateRecomputeMatchesDump`（新）：拿 `t260_pre` 的星域表與
  統治者記錄重算，13 格與傾印相同。
- 既有的月底結算、面板與存檔測試全綠——玩家那一邊的數字不變。

## 不在這份規格裡

- ~~電腦那一方的造船支出~~：**量過了**（`docs/re/236`）。兩個陣營跑的
  是同一套開局結算——`2703:0430` 加總稅收、`0654:00D4` 扣造船費，
  各在自己的階段做；同盟 50000 → 61087 → 45848 與 remake 算出來的相同。
  切給造船的**百分比**也解了（`docs/re/256`、`docs/re/257`）：`ds:8F92`
  每個陣營一個 byte，電腦在回合收尾設成 `min(總稅收 ÷ 900, 35)`。
  那筆錢在七個艦種之間怎麼分也解了：`ds:8FD6[陣營 × 7 ＋ 列]`，電腦寫回
  一張固定表（帝國 `30 30 8 5 9 9 9`、同盟 `30 25 10 5 10 10 10`）。
- ~~艦隊的移動花費~~：**接上了**（`docs/re/237`）。`+0x0C` 就是
  `rules.FleetOtherCost` 那條算式（艦種列 `+0x05`），三支算出來
  263／95／134 與收據逐一相同；維持費那一筆同樣接上了
  （`056A:09EB`，`docs/re/236`）。第一個 AI 回合跑完同盟是
  45848 − 426 − 492 ＝ **44930**，與第二個 AI 回合的傾印相同。
- `090A:0372` 那支 `±1`（`docs/re/206`；早期記成 `0x09413`，那是 prologue 的
  第二個 byte，`docs/re/299`）：在已量到的局面裡沒有被走到。
- ~~未安定度~~：**接上了**（[`docs/spec/042`](042-region-unrest-recompute.md)）。
  `2D10:0319` 與 `2D10:0213` 寫完之後都跑 `090A:0B93` → `090A:0E6D`，後者經
  `2703:02E5` 把 32 格的 `+0x7D` 整份重算；`refreshBlend` 的尾巴跟著跑
  `recomputeUnrest`。收據 `urx299w` 量到第一個 AI 回合裡重算了七次
  （`Blend` 2→1、建艦隊與下命令四次、參謀換人、`Blend` 1→0）。
- **月底分支的回合數差一格**：`runBrain` 用 `g.turn % 10 == 9`，但 remake 的
  `g.turn` 在 `advanceTurn` 開頭就先加了（第一次「實行」AI 跑的時候原版
  `ds:8F18` 是 0、`g.turn` 是 1），原版的條件對應 `(g.turn − 1) % 10 == 9`
  （`docs/re/299` 順帶）。`TestMonthEndKeepsBlend` 設 `g.turn = 8` 對的是原版
  計數器 8，收據 `monthblend` poke 的是 9。還沒改。
