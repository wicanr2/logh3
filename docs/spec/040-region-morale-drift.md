# SPEC 040 — 民心每回合往擁有者的目標值走一格

狀態：**CONFORMED**（純函式通過三份收據 192 格同狀態驗證，其中一份跨過月底；已接進 `advanceTurn`）
證據：[`docs/re/295`](../re/295-region-morale-writers.md)（反組譯 `056A:0AA5` ＋
收據 `mrx295`／`mrx295d`／`mrx295f`）
相關：[`docs/internals/22`](../internals/22-stellar-dt.md)（星域記錄 `+0x7C`／`+0x12`）、
[`docs/internals/55`](../internals/55-blend-and-tax-rate.md)（稅率的 `m`）、
[`docs/spec/006`](006-unrest-at-load.md)（未安定度的 `f2`）
執行器：dosgolem-logh3 `e0b6a03`

## 原版做什麼

每按一次「實行」，`056A:000C` 那一串的第五支 `056A:0AA5` 掃一次 32 個星域：

```
for 星域 k in 0..31:
    目標 = 星域[+0x12] == 0 ? 0 : 星域[+0x12] == 2 ? 100 : 200
    d = 星域[+0x7C] − 目標            # 民心零延伸成 word 再減
    d < 0 → 星域[+0x7C] += 1
    d > 0 → 星域[+0x7C] −= 1
```

- 擁有者讀**星域記錄的 `+0x12`**，不是依行星現算（收據 `mrx295d`）。
- 時序：扣艦隊維持費（`056A:09EB`）之後、艦隊移動（`056A:03E0`）之前；
  **同一次「實行」裡電腦座位的稅率重算在它之前**（收據 `mrx295` 的 17 號星域）。
  逐格開戰與回合尾巴重算擁有者都在它之後，所以一格易主之後，
  **下一次「實行」**才開始往新擁有者的目標走。
- 續接戰鬥的那一圈不跑 `056A:000C`（`docs/re/251`），所以一個回合只走一格。

## 民心的其他寫入者（這一份不實作）

| 寫入者 | 做什麼 | remake |
|---|---|---|
| `2703:0000`（新遊戲）| 照擁有者設 0／200／100 | `rules.NewGameRegions`／`Morale` |
| `2C5D:061A` → `332E:04BB`（讀存檔）| 整張星域表照檔案，民心不重算 | 讀存檔照檔案的用 |
| `1213:03F5`（費沙事件）| 費沙擁有者是人類座位時，那一方其餘星域民心 ← 100、行星經濟力扣 `ds:328E[回合 ÷ 100]`% | 費沙事件還沒有（`phezzan-event-trigger`）|

## remake

```go
rules.DriftRegionMorale(g.res.Stellar)   // internal/rules/morale.go
```

接在 `Game.advanceTurn` 扣維持費的迴圈之後、艦隊移動之前。remake 的
`runBrain`（含 AI 的稅率重算）排在整個回合的最前面，與原版「AI 先、民心後」
同向；逐座位的順序另見 `ai-seat-order`。

## 驗收

`internal/rules/morale_test.go`：

- `TestDriftRegionMoraleSteps`：三條分支、到目標就停、255 往下走。
- `TestDriftRegionMoraleMatchesReceipts`：拿 `mrx295`（回合 0→1→2→3）、
  `mrx295d`（回合 0→1）與 `mrx295f`（回合 9→10→11，跨過月底）的前一份星域表
  跑一次，下一份 32 格的 `+0x7C` 逐格相同；會動的格子要正好 40 格（正對照）。
  缺收據就 skip。

## 沒有驗到的

- **戰略層的正常玩家路徑沒有逐格收據**：三份收據都是 poke 出來的局面。
  remake 端也還沒有「remake 自己跑同一串輸入、跟原版比星域表」的整合收據——
  驗的是規則本身，接進 `advanceTurn` 的位置照反組譯的順序。
  自然局面的佐證是 `shipdeliver3` 的 24／28 號星域（回合 7–10 各多 1），
  那一組的取樣點有的落在處理中途，沒有拿來當驗收。
- 稅率算式的 `m` 在原版是 16 位元無號算（`docs/re/295` 稅率那一節）；
  `rules.TaxRatePerMille` 用有號整數。民心只要在 0–200 之內兩者逐位元組相同，
  而三個寫入者都不會寫出這個範圍以外的值，所以不改。
