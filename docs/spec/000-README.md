# 規格

`AGENTS.md` 的流程：

```
RE 證據 → DRAFT spec → 證據審查 → READY → implementation → 同狀態驗證 → CONFORMED
```

只有標 `READY` 的規格可以動 production path；標 `CONFORMED` 表示實作完成
而且有同狀態收據撐著。每一份規格的頂端都要寫現在是哪一個狀態。

`docs/internals/` 是逆向出來的事實，`docs/re/` 是怎麼查到的，
這裡是**要照著做什麼**——把事實翻成 remake 的行為與驗收條件。

| # | 規格 | 狀態 |
|---|---|---|
| [`001`](001-region-summary-panel.md) | 星域摘要面板（左下角）| `CONFORMED` |
| [`002`](002-capital-panel.md) | 右側面板（日期、錢、自己的首都）| `CONFORMED` |
| [`003`](003-budget-panel.md) | 預算計劃／徵稅計劃的數字 | `CONFORMED` |
| [`004`](004-production-panel.md) | 生產計劃的數字 | `CONFORMED` |
| [`005`](005-save-slot-loading.md) | 讀一格存檔（`*.DT[0-9A]` 六個檔）| `CONFORMED` |
| [`006`](006-unrest-at-load.md) | 載入時重算未安定度 | `CONFORMED` |
| [`007`](007-load-game-path.md) | 「取進度」那條玩家路徑 | `CONFORMED` |
| [`008`](008-save-slot-list.md) | 存檔畫面那 11 列（從 `SAVEDATA.IDX` 算）| `CONFORMED` |
| [`009`](009-save-slot-writing.md) | 存檔寫出 | `CONFORMED` |
| [`010`](010-save-game-screen.md) | 「存進度」畫面 | `CONFORMED` |
| [`011`](011-opening-report.md) | 疊在戰略底板上那幾張裡跟局面走的內容 | `CONFORMED` |
| [`012`](012-admiral-detail.md) | 提督詳細資料 | `CONFORMED` |
| [`013`](013-fleet-organisation.md) | 艦隊編成 | `CONFORMED` |
| [`015`](015-fleet-composition.md) | 艦隊組成 | `CONFORMED` |
| [`016`](016-versus-mode.md) | 對戰模式 | `CONFORMED` |
| [`017`](017-computer-ai.md) | 電腦 AI：原版行為與強化版 | 兩層都有了 |
| [`018`](018-battle-resolution.md) | 戰鬥結算（射程、攻擊名單、損失分攤）| 部分 `READY` |
