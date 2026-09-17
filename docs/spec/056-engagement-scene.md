# 056 — 交戰演出（打行星那一種）與戰鬥曲的出口

狀態：**DRAFT**（入口、換曲、面板與三個視窗 CONFORMED——`eng-open`／`eng-end` 扣掉四塊事先
列出的排除區逐位元組相同；行星圖、艦隊圖、量表、方塊動畫與演出的計時沒做）。
證據：[`docs/re/324`](../re/324-engagement-scene.md)；素材 [`docs/internals/20`](../internals/20-comani-bmg.md)
（`COMANI.BMG`，編碼未解）、[`docs/internals/37`](../internals/37-ani-pnl-a98.md)（`ANI_PNL1.A98`）。
收據：[`docs/playtest/022`](../playtest/022-engagement-scene.md)。
相關：[`019`](019-planet-assault.md)（打行星的算式；演出印的就是那一輪的數）、
[`docs/re/274`](../re/274-which-screen-plays-which-tune.md)（換曲的時點）、
[`docs/re/286`](../re/286-settings-other-rows.md)（動畫效果／速度）。

## 玩家看到什麼

戰場回合處理裡打行星那一步，每一顆有人打的行星切一段演出：整張換成 `ANI_PNL1.A98`
的面板，左上行星圖、右上艦隊圖、中間兩條量表，下方戰場上方塊互打；三個文字視窗
（格 (1,6) 15×2「帝國軍／行星名／防衛力 現在/上限」、(17,6) 6×2「防衛攻擊」、(24,6) 15×2
「第？艦隊 ？？？？？？？？？／船艦總數 現在/上限」），打完中間印「擊毀 N」、艦數更新。
開場換戰鬥曲、收場換回平時曲。按鍵只把動畫跳到結果，演出長度不變（約 29M 步）；
「動畫效果」關掉時沒有方塊動畫、短九百萬步，面板與換曲照舊。

## 規則（remake）

| # | 狀態 | 輸入 | 下一個狀態 | 收據 |
|---|---|---|---|---|
| 1 | `StateBattle`（實行） | 回合處理打行星有結果 | `StateEngagement` 開場幀：三個視窗印反擊前的數；戰鬥曲 | `eng-open` |
| 2 | `StateEngagement` | 左鍵或右鍵 | 結果幀：「擊毀 N」、艦數扣掉反擊 | `eng-end`（`eng-click`／`-rclick`） |
| 3 | 結果幀 | 左鍵或右鍵 | 下一場；沒有了就回 `StateBattle` 做完那一回合（要塞砲倒數、索敵、回合數）；平時曲 | `eng-after` |

視窗內容用原版的格式字串（`engagement-panel.json` 的 `formats`，Big5 位元組）；屬性直接寫
字元平面量到的（帝國軍 `E2`、行星名 `E4`、防衛攻擊 `02`、擊毀 `0D`、艦隊 `C4`）。

## 已知差異

| 差異 | 類別 | 到期條件 |
|---|---|---|
| 行星圖 (8,56) 257×102、艦隊圖 (376,56) 256×102：`COMANI.BMG` 的 `PPD0`／`F?D0`／`PPAB` 段。**編碼已解**（`docs/re/328`，`internal/asset.BMG` 逐格對上 `eng-text`／`-mid`／`-mid2`／`-end`），remake 的演出還沒接上它、仍留底圖的黑 | 未接 | 把名字組法（`3958:0041`）與逐格計時（`348A:0281`，`ds:8F10[ds:8F37]`）接進 `StateEngagement` |
| 兩條量表 (297,62) 48×68 | 沒讀 | 讀 `20B4:0486` 畫量表那段 |
| 戰場 (6,207) 627×226 的透視格線與方塊動畫（`PPAB` 段、`3931`／`3B74`） | 素材與程式都沒讀 | 同上 |
| 演出用計時（約 29M 步）自己收場，**按鍵沒有可見效果**（`docs/re/328` §六推翻了「按鍵跳到結果」）；remake 靠按鍵換幀 | 未模型化的計時 | 接上逐格計時時拿掉按鍵換幀 |
| 動畫效果（`ds:8F36`）與速度（`ds:8F37`）沒有人讀：remake 本來就沒有動畫 | 隨動畫一起 | 同上 |
| 中間視窗四種標題（防衛攻擊／登陸艦攻擊／占領成功／占領失敗）各在什麼情形印只量到一種 | 假設待驗 | 走到有登陸攻擊或換手的收據 |
| 艦隊看得清（第Ｎ艦隊 提督）的門檻假設是自軍或索敵等級 3；一顆行星被幾支打時一支一場 | 假設待驗 | 走到那種局面的收據 |
| 艦隊對艦隊（`1DD6:0009`，先問「戰鬥命令」）與要塞砲（`2264:02BB`）那兩種演出沒接 | 未實作 | 各自的收據 |

## 驗證

- `internal/game/engagement_test.go`：`TestEngagementMatchesOracle`（**直接進入**：把 S3 兩支艦隊
  搬到達希利、把那一場的數排進佇列，比 `eng-open`／`eng-end` 扣掉排除區）、
  `TestEngagementQueuedFromAssault`（打行星有結果就切演出、看完回合數才加）。
- `internal/audio/tune_test.go`：`TestEveryTuneIsReachable` 從 `StateEngagement` 走到 `BGM_6`／`BGM_7`。
