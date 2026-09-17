# 059 — 回合裡的過場視窗與報告框（出發前往／已經到達／錯誤報告／處理中面板／攻占訊息框／士氣報告）

狀態：**CONFORMED**（移動那一族：`113A:0B46`、`0797`、`0A4E`、`0899` 的畫面，與右側「戰略模式處理中」；
報告框：攻占訊息框 `20B4:1922`、士氣值低落、已全軍覆沒）。「已經恢復」**DRAFT**（remake 沒有自家星域補充，碰不到）。
證據：[`docs/re/332`](../re/332-turn-transition-windows.md)（整合 `docs/re/322` §一、§六 並更正其中 `+060F` 的條件）、
[`docs/re/333`](../re/333-turn-reports.md)（報告框）。
收據：`tools/fleetcmd_receipts.sh` 第 23、24、25 段；摘要 [`docs/playtest/021`](../playtest/021-strategic-fleet-command.md)。
相關：[`054`](054-fleet-arrival.md)（走一格的四條分支本身）、[`058`](058-strategic-planet-assault.md)（戰略打行星）、
[`047`](047-occupation-report.md)（占據報告，接在移動之後）。

## 狀態圖

原版的回合是一支同步程式，中間停在 `3765:006C(n)`（等 n tick，或滑鼠左／右鍵按下就結束）與 `113A:0899`（等點鈕）。
remake 把每個停頓點做成一個狀態，停下來的時候保存「接回哪一支」，關掉之後從那裡往下跑。

```
實行
 └ StateTurnThinking（單人模式；標題窗「電腦陣營 宇宙曆…／戰略模式思考中」、清單清空）─18 tick→
 └ advanceTurn：beginTurn → 電腦下命令 → 提督結局報告 → 維持費、民心
     └ moveFleetsFrom(i)               ← 32 支逐支（113A:0350）
          ├ 電腦座位的艦隊：直接套結果，不停
          ├ StepAskOccupy ─→ StateOccupyAsk ──點「占領」──→ 走一格那一路（見下）
          │                                   └點「待命」──→ 不動，接 i+1
          ├ StepBlockedEnemy ─→ StateTurnWindow{被擋, 90} ─結束→ 套結果、接 i+1
          ├ StepNoFunds      ─→ StateTurnWindow{經費不足, 90} ─結束→ 接 i+1
          └ StepMoved        ─→ StateTurnWindow{出發前往, 22} ─結束→ {到達句, 90} ─結束→ 走過去（可能解散）、接 i 或 i+1
          └ 每一支走完 047D（非停駐）：113A:03C5 重算士氣 → StateTurnReport{士氣／已全軍覆沒} ─看完→ 接下一支
     └ afterFleetMoves：重數艦隊 → 占據報告（047）→ 叛亂 → 逐格掃描（戰鬥／戰略打行星 → StateTurnReport{拿下行星}）
     └ 回合尾巴：易主報告 → 1261:00C8 → StateTurnReport{士氣} → 費沙事件 → 勝負 → 月底
```

「結束」＝ 計時到，或一次滑鼠左／右鍵按下。一次按下只結束一段。鍵盤不算。

「思考中」不吃點擊：`04F8:0355` 回來到 `056A:000C` 之間的呼叫樹沒有讀事件（`3765:0003` 在 AI 之前清一次佇列），
期間的點擊留在佇列，由處理中各步之間的 `056A:0983` 吃掉，不會結束後面的過場（`fcx-turnwin-clkai`，`docs/re/336` §七）。

## 規則

1. **誰會停**：艦隊的陣營是人類座位（單人模式 ＝ 玩家陣營；對戰模式兩邊都是）。原版條件是
   `(ds:8F20 ≠ 2 且 3982:022D(陣營) == 0) 或 ds:8F1C bit0`；bit0 是隱藏除錯選單的「顯示」，remake 沒有那個選單。
   電腦座位的艦隊不印、不等，照舊套結果。
2. **套結果的時點**：`047D` 先扣錢、改位置與提督 `+0x0A`（`+0585`–`+05C4`），再叫 `0B46`（`+060F`），等完才到達收尾
   （`+065B`）。駐留到了目的地的解散（`297B:025D`）因此在「已經駐留在星系上。」那一張之後。被擋的命令清除與經費不足的
   不動在印之前或之後都看不出差別，remake 放在等完之後。
3. **一張過場視窗的內容**（`3603:0006(0x97, 20, 6, 18, 6, 1)`：外框 (320,160)–(639,319)、內容 (336,180) 起 18×6 格、
   列距 20；頭像 `3958:0215(提督, 0x44, 0x23)` ＝ (544,180)，整張 80×100）：
   - 第 0 列：`\x0c` 清空、帶色陣營名、兩個半形空白、「第」、全形編號、「艦隊」；第 1 列：提督名＋軍階。
   - 出發前往（`0B46`）：第 2 列現在格名（擁有者色、左靠 16 半形寬）＋「出發前往」；第 3 列下一格名（擁有者色）。
     第二段在第 4 列加「已經到達目標星系。」（下一格 ＝ 目的地、命令不是駐留）或「已經駐留在星系上。」（命令是駐留）；
     沒到目的地不加。
   - 被擋（`0797`）：第 2 列「無法前往敵方的星系」、第 3 列「。」（原版字串的換行，照印）。
   - 經費不足（`0A4E`）：第 2 列「由於移動經費不足無法移動。」。
4. **地圖**：整回合處理中地圖沒有反白格（`ds:F182 = FF`）。出發前往第一段把現在格塗色 5，第二段現在格還原、
   下一格塗色 5，視窗關掉之後下一格的色 5 留到 32 支走完（`056A:044B`）；錯誤報告不動地圖。占據、叛亂、易主報告先整片還原、
   再把報告那一格塗色 5，關掉之後留著（`docs/re/335` §五）；回合結束回到命令階段整片還原。
5. **等多久**：出發前往第一段 `ds:8F38 >> 2`（預設 22 tick），第二段與錯誤報告 `ds:8F38`（預設 90 tick）。
   `ds:8F38` 是設定畫面的訊息速度。
6. **右側標題窗**：實行之後、移動與之後的報告期間只有「戰略模式處理中」一列，清單與摘要清空；分頁鈕、命令鈕、
   清單框照畫，清單兩端的捲動箭頭換成結算時那一組。處理中 `ds:8F1D = FF`，地圖 32 格都塗擁有者色（不看有沒有自軍艦隊，
   `3705:037F`）、沒有反白格。問話視窗（`0899`）也在這個底板上。回到下一回合命令階段才還原。電腦下命令那一段的「<陣營> 宇宙曆…／戰略模式思考中」
   （`fcmd-turn-think`）remake 不停在那裡，沒有畫面（見已知差異）。
7. **視窗的生命週期**：`056A:03E0` 開窗時不顯示，第一張過場才顯示，32 支跑完才關。多支人類艦隊共用同一個視窗，
   每張以 `\x0c` 清空重印。
8. **「自己的」讀星域記錄 `+0x12`**（`047D +04DC`／`+0514`），不是由行星數出來的擁有者（兩者平常一致，
   `fcmd-turn-block-pre` 只改記錄才分得出來）。

**remake：`Game.moveFleetsFrom`／`stepFleet`、`StateTurnWindow`（`internal/game/turnwindow.go`）、
`screen.TurnWindowSpec`（`docs/internals/data/turn-window-panel.json`）。**

## 報告框（`3826:0307`）

9. 同一個矩形、同一種等法（`ds:8F38` tick 或一次左／右鍵按下）；每一則 `\x0c` 清空重印，頭像 `3958:0215(…, 0x44, 0x23)` 整張 80×100。
   座位參數 FFFF → 頭像是那一支的提督；否則 `ds:8F4D[座位]`（情報參謀）。
10. **拿下行星**（`20B4:1922`，行星換手的當下，人類與電腦座位都印、都等）：
    人類「我軍第Ｎ艦隊／提督名／正在攻占<星域名>／星系內行星<行星名>／已經歸屬<陣營名>掌控之中」＋提督頭像；
    電腦「<陣營名>艦隊／企圖占領<星域名>／…」＋對手陣營的情報參謀頭像。星域名用星域記錄 `+0x12` 的顏色（這時還沒重算），
    行星名用換手之後的顏色。編號 `3982:0073`（< 10 全形）；名字 `3982:0296` 不補寬。
11. **士氣報告**（`1261:0104`）：bit2（士氣跨過 50）且人類座位 →「<陣營>第Ｎ艦隊  提督名／士氣值低落的關係，／已經無法戰鬥。」或
    「士氣值已經恢復，／可以進行戰鬥。」＋「艦隊士氣值　%4u」；bit3（全隊艘數 0，不看座位）→「<陣營>第Ｎ艦隊／提督名／已全軍覆沒。」；
    印完清 bit2／bit3。跑的時點：每支非停駐的艦隊走完 `047D` 之後先照編制重算士氣（`113A:03C5` → `1261:0067(f, 1)`），
    以及回合尾巴易主報告之後（`1261:00C8`）。戰略打行星裡只記 bit2、不印。
12. 艦隊留著：「已全軍覆沒」只是報告；全滅（整支拿掉）是 `0374:0608` 另一條（規格 060）。

**remake：`StateTurnReport`（`internal/game/turnreport.go`）、`Game.fleetStepMorale`、`rules.RecomputeMorale`、
`rules.StrategicAssaultResult.CapturedBy`；字串在 `turn-window-panel.json` 的 `report_formats`。**

## 已知差異

| 差異 | 類別 | 到期條件 |
|---|---|---|
| 「戰略模式思考中」的停留：原版停到戰略 AI 算完（`fcmd-done` 那條 18.6 tick、S3 第一回合 7.7 tick），remake 固定 18 tick | 使用者裁決（2026-09-17） | 不到期（裁決） |
| 「已經恢復」：remake 沒有自家星域的艦艇補充（`1AAB:0FAF`），士氣回不到 50 以上 | 規則沒接 | 補充接上時一起對 `fcmd-turn-recover`（issue） |
| `1261:0067` 的 `ds:8F72[陣營] ← 297B:10DE`：該陣營所有艦隊維持費的總和（同時重寫各艦隊 `+0x0A`／`+0x0C` 與 `ds:8F82`，`docs/re/336` §七）。remake 在讀者（預算畫面、造船基數、AI 建艦隊門檻）現算；原版只在 `1261:00A8` 與 `090A:0EA1` 更新，兩次之間編制變了的時差沒有對拍。編制裡有艘數卻沒有滿編數的組（remake 修理拆出來的）不重算士氣 | 時差沒對拍／資料模型 | 艦艇池模型化（#55）時一起 |
| 戰鬥中（`ds:8F16 != 0`）士氣報告走 `3826:0405` 的臨時視窗（(2,7) 24×5、頭像 (352,200)）：視窗版面與文字已接（`TurnWindowSpec.DrawBattleReport`，收據 `eng-morale-low` 視窗內逐位元組相同），但 remake 的戰場沒有結算最後一步（`181B:03F0` → `185C:001D` 對沒被消滅的艦隊叫 `1261:0104`）與修理階段等當場印的時點，**觸發點沒接**，也到不了同一個戰場局面，所以只比視窗 | 戰場結算沒有階段模型 | 戰場回合結算照 `181B:000A` 八步模型化時一起接上，改整張比 |
| `ds:8F1C` bit0（除錯選單「顯示」）讓電腦座位的艦隊也印：remake 沒有除錯選單 | 未實作 | 做隱藏除錯選單時一起 |
| 等待的時間以 remake 的 `Tick` 累計，不重現原版遞減相位（第一個 tick 可能差一個相位點） | 計時近似 | 需要逐 tick 對拍時照 `docs/re/331` 的相位公式 |

## 驗證

- `internal/game/turnwindow_test.go`：出發前往、已經到達、已經駐留、占領第一回合（沒到達句）、經費不足、被擋（實行前 poke 那一份）、
  問話與問完之後兩張、點一次只結束一段（`fcmd-turn-clk22`）逐位元組整張比對；計時 22／90 tick 各自結束。
- `internal/game/garrison_test.go`：`fcmd-garrison-report`、`fcmd-garrison-self-report`（就地駐留「瓦爾哈拉 出發前往 瓦爾哈拉」）。
- `tools/gui_check.sh`：`fcmd-turn-think`、`fcmd-turn-depart`、`fcmd-turn-arrive`，以及改成點掉過場的 `fcmd-exec`、`fcmd-occ-t2`、`fcmd-garrison-t1`。
- `TestTurnThinkingMatchesOracle`：`fcmd-turn-think` 整張相同、17 tick 仍在、第 18 tick 進出發前往。
- `TestBattleMoraleReportWindowMatchesOracle`：`eng-morale-low` 視窗矩形（先塗掉再畫）逐位元組相同，排除收據游標 16×20（見已知差異）。
- `TestTurnWindowThreeHumanFleetsMatchOracle`：S3 帝國第一回合三支人類艦隊（`fcmd-s3-think`、`fcmd-s3-w1a`–`w3b`，第 25 段），
  前一支的下一格色 5、視窗不重開、兩位數編號（`3982:0038` 半形）都在整張比對裡。
- `fcmd-turn-busy`、`fcmd-turn-block`（AI 之後 poke）是原版的對照，remake 沒有停在那裡的狀態，不比。
- `turnreport_test.go`：`TestTurnReportCaptureMatchesOracle`、`TestTurnReportNoShipsMatchesOracle`（收據同一個 poke）、
  `TestTurnReportAICaptureMatchesOracle`、`TestTurnReportMoraleLowMatchesOracle`，正常玩家路徑整張逐位元組相同。
  GUI：`fcmd-turn-ai-capture`、`fcmd-turn-morale-low`（另兩張要改記憶體，GUI 腳本做不到）。
- `internal/game/fleetcmd_test.go`、`garrison_test.go`、`occupation_test.go`：既有的回合後收據（`fcmd-exec`、`fcmd-garrison-t1`、
  `fcmd-occ-*`）照舊通過（`executeTurn` 會把過場視窗點掉）。
