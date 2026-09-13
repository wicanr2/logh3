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
| [`001`](001-region-summary-panel.md) | 星域摘要面板（左下角） | CONFORMED |
| [`002`](002-capital-panel.md) | 右側面板（日期、錢、自己的首都） | CONFORMED |
| [`003`](003-budget-panel.md) | 預算計劃／徵稅計劃的數字 | CONFORMED |
| [`004`](004-production-panel.md) | 生產計劃的數字 | CONFORMED |
| [`005`](005-save-slot-loading.md) | 讀一格存檔 | CONFORMED |
| [`006`](006-unrest-at-load.md) | 載入時重算未安定度 | CONFORMED |
| [`007`](007-load-game-path.md) | 「取進度」那條玩家路徑 | CONFORMED |
| [`008`](008-save-slot-list.md) | 存檔畫面那 11 列 | CONFORMED |
| [`009`](009-save-slot-writing.md) | 存檔寫出 | CONFORMED |
| [`010`](010-save-game-screen.md) | 「存進度」畫面 | CONFORMED |
| [`011`](011-opening-report.md) | 疊在戰略底板上那幾張裡跟局面走的內容 | CONFORMED |
| [`012`](012-admiral-detail.md) | 提督詳細資料 | CONFORMED |
| [`013`](013-fleet-organisation.md) | 艦隊編成 | CONFORMED |
| [`015`](015-fleet-composition.md) | 艦隊組成 | CONFORMED |
| [`016`](016-versus-mode.md) | 對戰模式 | CONFORMED |
| [`017`](017-computer-ai.md) | 電腦 AI：原版行為與強化版 | 兩層都有了 |
| [`018`](018-battle-resolution.md) | 戰鬥結算（射程、攻擊名單、損失分攤）| 部分 `READY`（逐節標示）|
| [`019`](019-planet-assault.md) | 攻打行星 | READY |
| [`020`](020-intro-animation.md) | 開頭動畫 | READY（畫面與速度）／播放順序未解 |
| [`021`](021-battle-data-screen.md) | 戰術選單「資料」 | CONFORMED |
| [`022`](022-fleet-detail.md) | 艦隊詳細資料 | CONFORMED |
| [`023`](023-battle-command-menus.md) | 戰場的命令選單 | CONFORMED |
| [`024`](024-battle-move-order.md) | 戰場的「移動」命令（分隊層） | READY／已實作 |
| [`025`](025-battle-formation-order.md) | 戰場的「編隊」命令 | CONFORMED |
| [`026`](026-battle-supply-repair.md) | 戰場的「補給」與「修理」 | READY |
| [`027`](027-fleet-move-and-battle-order.md) | 艦隊選單的「移動」與「戰鬥」 | READY |
| [`028`](028-fortress-gun-fire.md) | 要塞砲開火 | READY |
| [`029`](029-strategic-ai-turn.md) | 原版的戰略 AI 回合 | READY |
| [`030`](030-blend-and-tax-rate.md) | `Blend` 與稅率重算 | CONFORMED |
| [`031`](031-fleet-refit.md) | AI 回合的艦隊整補 | CONFORMED |
| [`032`](032-region-posture.md) | 星域態勢重算（`+0x13` 與 `+0x24`） | CONFORMED |
| [`033`](033-region-ownership-chain.md) | 回合結束的星域歸屬鏈 | CONFORMED |
| [`034`](034-region-flip-report.md) | 星域易主的戰略報告 | CONFORMED |
| [`035`](035-month-end-gate.md) | 月底的閘門與每回合的重算 | READY |
| [`036`](036-ship-delivery.md) | 月底交付的船艦放到哪裡 | READY |
| [`037`](037-unit-price-from-staff.md) | 船艦單價由軍事專任參謀決定 | CONFORMED |
| [`038`](038-monthly-growth-and-tax-rate.md) | 月底重算成長率與稅率 | CONFORMED |
| [`039`](039-region-supply-pool.md) | 星域的飛彈／艦載機存量每回合回充 | DRAFT |
| [`040`](040-region-morale-drift.md) | 民心每回合往擁有者的目標值走一格 | CONFORMED |
| [`041`](041-rebellion.md) | 叛亂：判定、效果與接進回合的位置 | CONFORMED |
| [`042`](042-region-unrest-recompute.md) | 星域未安定度重算與呼叫點 | CONFORMED |
| [`043`](043-ai-capital-reach.md) | 戰略 AI 的首都可達表 | CONFORMED |
| [`044`](044-ending-image-position.md) | 400 列結局 VO5 的垂直位置 | CONFORMED |
| [`045`](045-ship-delivery-screen-details.md) | 月底船艦接收報告的底板細節 | CONFORMED |
