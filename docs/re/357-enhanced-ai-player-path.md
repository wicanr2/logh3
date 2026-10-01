# 357 — 增強 AI 的正常入口接線缺口

日期：2026-10-01。工作項目：[Issue #87](https://github.com/wicanr2/logh3/issues/87)。
輸入為 remake `2fbd0d13f4380861b9d945fd7a7377ca17ae723c`；這是產品接線調查，
不使用反組譯位址、不將增強決策稱為原版同狀態行為。

## 已證實

正式 `cmd/gin3 -ai 1` 只安裝 `Focused`／`Encircle`。`Focused` 僅實作 `Orders`，
`Game.runBrain` 只對 `ai.Builder` 建隊、對 `ai.Refitter` 整補，因此新局敵方不建隊。
同一正常 GUI 序列從主選單、防拷、月初報告、玩家編隊、實行至系統存進度，
原版模式保存玩家 1／敵方 3 支，強化模式保存玩家 1／敵方 0 支。
輸出 `workplace/playtest/ai87-before/receipt.json` SHA-256
`6ae5be352d1370cb92fe1c94efe07b7736c3b2d9c61163422b6f10366695f282`；
兩次均由 `logh3-ebiten:2` 正式 renderer、未注入的正常入口取得。
原版模式 FLEET SHA-256 `919c4daf11fcd67dd9a4fa6a600b7a1d6dc4eddba23c106a90aed6b8e70cee0c`；
強化模式 `865e605a9e572b9457725b6a75660e640ea900668ce71408777acdeda35780fb`。

`Game.LoadSlot` 末尾無條件 `SetBrain(g.OriginalBrain())`，會丟失 CLI 已選的強化策略。
`Focused` 攻勢的 `Order.Move` 使用零值，但正式 `StarGraph.PlanStep` 在下一格非己方時
要求 `MoveOccupy`，因此不能用舊測試只寫 Target／Parked 再直接 Advance 的結果
證明正式出擊能生效。戰力輸入讀 `Fleet.Strength` 快取，未序列化的新編制常為 0；
必須按[既有規格 017](../spec/017-computer-ai.md)以現有編制套原版權重。

這些是現行產品程式的已證實關係；原版規則來源仍以規格 029、031、051 及
`rules/fleetstep.go` 的既有原版定位為準，不重開無關 driver 逆向。

## 設計邊界

沿用規格 017「強化版只改決策、不改規則」、CLI `-1`／`0–3` 且偏好不進原版存檔。
將原版建隊／整補接到增強策略，改攻勢為正式占領命令；載入新底稿時重接建隊來源，
重設策略暫態，但保留模式／難度。新局同樣清除上一局的策略暫態。
這些修復既有玩家選項，不改難度常數、經濟、戰損或路徑表。

2026-09-17 的被動戰場裁決仍有效（`internal/game/passivebattle.go`）：電腦不下戰術命令、
不造成傷害。不可因換 `Encircle` 型別便稱正常戰場已包圍攻擊；本輪不取消該裁決。
來源與驗收入口：[規格 068](../spec/068-enhanced-ai-player-path.md)。

## 修正後驗收

兩陣營正常新局均實際建隊且與原版策略的首回合命令不同；正常保存／讀取保留增強偏好。
正式 CLI 兩模式各重現兩次：首回合均 1／3 支，載入後續回合均 1／4 支，命令差異保持。
固定來源雜湊、輸出雜湊與完整重生入口見[驗證 029](../playtest/029-enhanced-ai-player-path.md)、
[`enhanced-ai-player-path.json`](../internals/data/enhanced-ai-player-path.json)。
全產品測試通過。最初兩次測試 panic 為不可比較策略值及舊助手漏載權重，
同一命令補正後通過；未改 AI 參數、原版規則或被動戰場裁決。
