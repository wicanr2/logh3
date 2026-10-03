# 056 — 交戰演出（打行星那一種）與戰鬥曲的出口

狀態：**DRAFT**（入口、換曲、面板、三個視窗、**兩個圖視窗的序列與計時、按鍵排隊** CONFORMED——六批 `eng-*`
與 `eng-skip-mid` 扣掉量表與戰場兩塊逐位元組相同（`docs/re/331`，歷史收據）；
094已接既有防衛標題分支量表，戰場方塊、另外三種演出及正常同狀態／色盤仍待驗）。
證據：[`docs/re/324`](../re/324-engagement-scene.md)；素材 [`docs/internals/20`](../internals/20-comani-bmg.md)
（`COMANI.BMG`，`docs/re/328`）、[`docs/re/331`](../re/331-engagement-picture-sequence.md)（序列與計時）、[`docs/internals/37`](../internals/37-ani-pnl-a98.md)（`ANI_PNL1.A98`）。
收據：[`docs/playtest/022`](../playtest/022-engagement-scene.md)。
相關：[`019`](019-planet-assault.md)（打行星的算式；演出印的就是那一輪的數）、
[`docs/re/274`](../re/274-which-screen-plays-which-tune.md)（換曲的時點）、
[`docs/re/286`](../re/286-settings-other-rows.md)（動畫效果／速度）。

## 玩家看到什麼

戰場回合處理裡打行星那一步，每一顆有人打的行星切一段演出：整張換成 `ANI_PNL1.A98`
的面板，左上行星圖、右上艦隊圖、中間兩條量表，下方戰場上方塊互打；三個文字視窗
（格 (1,6) 15×2「帝國軍／行星名／防衛力 現在/上限」、(17,6) 6×2「防衛攻擊」、(24,6) 15×2
「第？艦隊 ？？？？？？？？？／船艦總數 現在/上限」），打完中間印「擊毀 N」、艦數更新。
開場換戰鬥曲、收場換回平時曲。左上行星圖、右上艦隊圖依序畫（行星靜圖 → 艦隊靜圖 → 行星打的動畫 → 艦隊中彈的動畫，
每格間隔 8／4／0 tick 看動畫速度），「擊毀 N」之後停頓「訊息速度」tick 才收場；播放中按鍵不跳格，按下的滑鼠鍵
在停頓時讀到就提早收場。「動畫效果」關掉時兩張圖都不畫，擊毀與停頓照舊。

## 規則（remake）

| # | 狀態 | 輸入 | 下一個狀態 | 收據 |
|---|---|---|---|---|
| 1 | `StateBattle`（實行） | 回合處理打行星有結果 | `StateEngagement` tick 0：三個視窗印反擊前的數；戰鬥曲 | `eng-open` |
| 2 | `StateEngagement` | 時間（`Tick`） | 第 I、2I 個 tick 畫行星與艦隊的 D0 第 0 格，3I–12I 行星 `P?AB`，之後艦隊 `??D{0,1,2}` 整段；畫完印「擊毀 N」、艦數扣掉反擊 | `eng-text`／`-mid`／`-mid2`／`-end` |
| 3 | 播放中 | 左鍵或右鍵 | 不動，記下來 | `fcx-engrec-click`（`docs/re/331` §三） |
| 4 | 擊毀之後 | 記過按鍵，或停頓滿訊息速度，或停頓中按左／右鍵 | 下一場；沒有了就回 `StateBattle` 做完那一回合；平時曲 | `eng-after`（停頓中） |

視窗內容用原版的格式字串（`engagement-panel.json` 的 `formats`，Big5 位元組）；屬性直接寫
字元平面量到的（帝國軍 `E2`、行星名 `E4`、防衛攻擊 `02`、擊毀 `0D`、艦隊 `C4`）。

## 已知差異

| 差異 | 類別 | 到期條件 |
|---|---|---|
| 兩條量表的其他演出分支 | 094已核常式並接既有防衛標題分支；歷史fixture仍排除舊矩形 | 閉合分段生命週期及正常同狀態／色盤；見094與驗證057 |
| 戰場 (6,207) 627×226 的方塊（`321F:008A`／`0897`／`06FC`，開場之後畫、擊毀時閃一次） | 程式沒讀 | 讀 `321F` 段 |
| 停頓起點：原版在方塊畫完（約 26 萬步 CPU）之後才開始停頓，收場晚 remake 一個 tick（235 對 234）；動畫效果關掉時原版停頓在 tick 8 才開始（98 對 90） | 時序 | 讀開場到停頓之間花掉的 tick |
| 音效（`3B74:031A`，記錄 `+6`） | 沒讀 | 讀 `2376:1585`／`1644` |
| 登陸艦攻擊（`20B4:0681`：`?LAB`、`P?D0`／`D2`）那一條的圖序列只讀了程式碼 | 沒有收據 | 走到登陸攻擊的收據 |
| 中間視窗四種標題（防衛攻擊／登陸艦攻擊／占領成功／占領失敗）各在什麼情形印只量到一種 | 假設待驗 | 走到有登陸攻擊或換手的收據 |
| 艦隊看得清（第Ｎ艦隊 提督）的門檻假設是自軍或索敵等級 3；一顆行星被幾支打時一支一場 | 假設待驗 | 走到那種局面的收據 |
| 艦隊對艦隊（`1DD6:0009`，先問「戰鬥命令」）與要塞砲（`2264:02BB`）那兩種演出沒接 | 未實作 | 各自的收據 |

## 驗證

- `internal/game/engagement_test.go`：`TestEngagementMatchesOracle`（**直接進入**：把 S3 帝國第一支與同盟費雪那一支
  搬到達希利、把那一場的數排進佇列，六個 tick 比六批收據扣掉量表與戰場）、`TestEngagementClickQueuesUntilPause`、
  `TestEngagementAnimationOffMatchesOracle`、
  `TestEngagementQueuedFromAssault`（打行星有結果就切演出、看完回合數才加）。
- `internal/audio/tune_test.go`：`TestEveryTuneIsReachable` 從 `StateEngagement` 走到 `BGM_6`／`BGM_7`。
