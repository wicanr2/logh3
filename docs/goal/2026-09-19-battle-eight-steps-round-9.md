# 第九輪 goal — 戰場回合結算的其餘七步（#79）＋ 六組並行

日期：2026-09-18 擬，2026-09-19 重寫（第八輪收尾與 issue 盤點之後整併；修訂經過在 `WORKLOG.md`）。

## 上一輪的結果

第八輪（`docs/goal/2026-09-18-battle-settlement-round-8.md`）**五組全部關掉**：

- **#77**：`docs/re/278` 用 `f7523f9` 重錄重驗，規格 001 規則 4–7 對齊。
- **#57**：戰場那一條全滅與戰鬥中的士氣報告接上（八步的第 ⑧ 步子步 8-1），規格 062 開出來。
- **#55 ＋ #71**：艦艇池欄位表進 `docs/internals/64`；**那兩格存量解出來是飛彈與艦載機**
  （`2703:14CF` ＝ 艦種表 `+0x10`／`+0x14` × 艘數，`docs/re/345`）。自家星域補充
  `1AAB:0FAF` 接上（規格 063、`docs/re/346`），「士氣值已經恢復」走得到了。
- **#74 ＋ #75**：兩位數艦隊編號的五個地方、開戰前的「發現敵方艦隊」報告。
- **#63／#64／#65**：情報收集三個分頁、三張計劃畫面、艦隊編號清單與結束確認。
  **#63 的門檻表其實是距離表**（`docs/re/347`）——`docs/re/185`／`241`／`298` 早就解完。

## 收尾與盤點（2026-09-19）之後的現況

`main` 在 `fc13263`，dosgolem 副本仍是 `f7523f9`（沒動執行器，收據不必重跑）。
`go test ./internal/...` 全過、`tools/gui_check.sh` 40 例逐位元組相同、
`tools/feature_check.sh` 沒有 crash 且 `knownDead` 是空的、worklist verify 0 條斷言不成立。
**open 的 worklist issue 剩 15 條。**

收尾檢查與盤點又處理掉三件，**這三項不在本輪範圍**：

- **#78 關**：戰略全滅寫回存檔表。盤點同步邊界時抓到同類的第二處——`mutableAdmirals()`
  換可寫副本沒有同步 `g.slot.Admiral`，提督失蹤／升級／任命從第一次改提督起就不進存檔。
  兩支測試都驗過會抓到回歸，盤點寫進 `docs/internals/30` 最後一節。
- **#58 關**：`297B:08D8` 的位元組 `docs/re/230` ＋ `235` 早就讀完，而且**它不碰 `+0x28`**；
  真正的寫入者是 `2703:0853`（32 支走完整批重數），正是 remake 的做法。
- **#67 撞號修掉**：它與已關閉的 #17 共用 `worklist:id=battle-data-scroll`，
  `cmd/worklist` 以 id 為鍵，於是 `CONTEXT.md` 一處說完成、一處說未完成。
  改成 `battle-data-scroll-receipt`。

**新開的 #81**（見 A 組）：九支戰場測試從來沒跑過。
**使用者裁決（2026-09-19）**：#69 的 `_M` 那 11 首**內建 GM soundfont** 隨 remake 發行。

下面這段可以直接當下一個 session 的開場提示詞。

---

## 提示詞

你在 `~/cht/logh3`（銀河英雄傳說III SP 的 DOS remake）。先照 `CLAUDE.md` 的開工順序：
`tools/issues_snapshot.sh` → `tools/go.sh run ./cmd/worklist verify` → 讀 `CONTEXT.md` 現況一覽 →
`git status --short`（logh3 與 `~/cht/dosgolem-logh3` 兩邊，`git config user.email` 都要是
`wicanr2@gmail.com`）。再讀直接相關的：`docs/spec/062`（八步，其餘七步 DRAFT）、
`docs/re/339` §六（十二條落差）、`workplace/re-notes/round7-57-static.md` §三、
`docs/spec/063` ＋ `docs/re/346`（補充，第 ① 步的鄰居）、`docs/re/345`（兩格存量與那塊 108600）、
`internal/game/battle.go`、`internal/game/battlefinal.go`、`internal/rules/regionresupply.go`。

**本輪目標：把戰場一個回合的結算做成原版的八步模型，並清掉五個玩家撞得到或會過期的缺口。**
先前的裁決照舊：每張新畫面維持 same-state 契約（RE → spec → 收據 → READY → 實作 → 整張逐位元組），
關 issue ＋ commit ＋ push 不必再問。**#16 不要動**。subagent 不直接改 GitHub issue。

三條這幾輪反覆踩到的：

- **主機上唯一允許的算術是 `printf '%x' $((...))`**——分析、轉檔、反組譯、跑測試一律走
  `tools/py.sh`／`tools/go.sh`／`tools/ida.sh`／`tools/oracle.sh`。連 `python3 -c` 做一次
  十六進位換算都不行（第八輪四個 RE subagent 裡有三個踩到）。
- **要寫下「這個還沒解／卡住了」之前，先 grep 自己的 `docs/`。** 一天之內撞了三次：
  #63 的距離表（`docs/re/347`）、#58 的 `297B:08D8`、`docs/re/42` 的兩格存量。
- **綠的測試不等於驗到了那件事。** `star1eAlly` 整張逐位元組 PASS，但畫面全是「0：0」，
  對它要驗的差異是盲的（#59）；九支戰場測試一直 skip 在「沒走到戰場」（#81）。

### 交付順序

**A. #81 九支戰場測試從來沒跑過：排在 B 之前，它是 B 的護欄**

`go test ./internal/... -v | grep SKIP` 有九支的理由是 `沒走到戰場`——它們靠「跑 40 個回合
等 AI 把兩邊帶到同一格」，而預設的 `ai.Original` 一道命令都不下。涵蓋的規則是戰場決定性、
包圍、損失量級、九回合結束、結果寫回、電腦不造成傷害（規格 061）。

**B 要動的正是戰場結算的步序**，這幾支是唯一會在規則層抓到回歸的東西。判準與卡在寫在
issue 裡（`TestBattleStartsOnContact` 測的就是「會不會開戰」，不能用 `StartBattle` 繞過去）。

**B. #79 八步（主軸）**

規格 062 目前只有子步 8-1 是 CONFORMED。其餘七步與 8-2–8-5 照 `docs/re/339` §六的十二條
落差逐條處理。第 ① 步（補給 `1930:019D`、修理 `1AAB:006D`、行星防衛力回復 `181B:0083`、
`1261:000B(1)` 印士氣）現在**沒有東西擋著了**——艦艇池模型（#55）已經接上。

順手把 `1930:1572` 讀掉：它與第 ① 步的補給是同一族，排在 `1AAB:0FAF` 之前、同一格的處理
序列，多半就是「把星域層的飛彈／艦載機整批補給進艦隊」那一支——收據上兩格存量各多出來的
108600 正好等於星域 0 的 `+0x68`／`+0x70`。解掉之後規格 063 的「已知偏離」第一條就能清掉。
`internal/rules.TestResupplyNumbersMatchOracle` 已經把「多出來的量 ＝ 星域的飛彈存量」
這個假設釘住，讀出來對不上會立刻壞掉。

**C. #62 參謀任命：先只接情報收集那一顆**

**那一顆的 RE 已經 proven 到位元組**（`docs/re/342` §5，2026-09-19 盤點時翻出來的）：
熱區 `(525,378)-(562,399)`、點擊配色 `0x27`→`0x87`、呼叫 `0A02:0AC4()`、挑到合格提督時寫
`ds:8F4D[ds:8F1D] ← 新提督編號`、共用出口 `0A02:0B3A` 重畫姓名與頭像。
拿那份反組譯記的點擊序列補一份收據就能接。

徵稅計劃那一顆（`090A:0248`，換的是**內政**參謀 `ds:8F4F`）沒有對應的 RE 文件，
`docs/re/268` 只到 `090A:021B`；合格檢查 `0D92:0422`／`0D92:04E3` 的內部條件也還沒反查。
**兩顆可以分開交付**，先接完的那一顆就先算數。

**D. #80 ＋ #53 兩個 parity 缺口**

- **#80**：徵稅／預算計劃四顆箭頭鈕剛點過時差 17–21 個像素（原版寫色 8、remake 寫色 9）。
  目前用 `planArrowSkip` 排除、在 `docs/spec/003` 與 `docs/re/343` 明標**未解、不是核准的差異**。
  起點是畫箭頭的那一支（熱區清單 `ds:14A8` 的 idx 0／1／4／5，kind 6）與逐幀收據（`-watch-screen`）。
- **#53**：三處位元組不符**一處都沒修**。拆開做，難度差很多：
  1. **艦種 5 那條最便宜也最獨立**——`internal/rules/fortressgun.go` 的條件寫反了：remake 把
     艦種 5 整段跳過，原版是艦種 5 無條件縮、其他艦種才看 `0xFF`。`-poke` 造一個艦種 5
     分隊被打到的局面即可；
  2. `spreadOverFleet` 尾段那個「留一艘」夾值（原版是死碼，實際不夾）；
  3. `Atan256` → `AtanTable256` **排最後**，它會動到戰場移動與射程，既有戰場收據要整批重跑，
     動之前先做一次「哪些收據會變」的乾跑盤點。

**E. #60 ▲▼ 按住連捲**

產品程式碼裡目前**只有一條按住路徑**：`internal/game/battle.go` 的 `RightHold`，由
`cmd/gin3/main.go` 每幀送 `IsMouseButtonPressed(MouseButtonRight)`。**左鍵完全沒有按住分支**，
所以 ▲▼ 不可能連捲。照規格 057（戰場右鍵捲動）的形狀補一條左鍵的每幀路徑，先只接一覽表那一張。

範圍含徵稅／預算計劃的兩組箭頭連發（`internal/game/planscreens.go` 的註解），是同一件事。

⚠ 既有的 11 份 `s3data-scroll-*` 綁在 2026-09-18 重產過的檢查點上，取樣相位變了
（`docs/re/337` §三），**拿它們下結論之前先用 `-watch-screen` 重量步數**。

**F. #66 ＋ #59 ＋ #56：補收據與兩個沒接的項目**

- **#66**：判準四項**已達成兩項**（`BUG!` 反組譯完整、`MUSIC.INI` 格式已在 `docs/internals/17`）。
  `BUG!` 那張除錯選單的 RE 完整到**可以直接實作**（七個動作碼都只改兩三個 `WORK` 欄位）；
  剩 `midi-if`（22）要讀——`docs/re/286` 只到分派那一行。
- **#59**：四種情形零份能區分的收據。**先錄同盟座位那一份**，`-poke` 把某格的 `+0x2A`／`+0x2B`
  寫成非 0 與 `0xFF` 各一份、同時讓那格有帝國艦隊——**這一份同時回答 `+0x2A`／`+0x2B` 的語意**
  （目前還是 hypothesis），投資報酬率最高。其餘三種也都是 `-poke` 造。
- **#56**：四項裡兩項已經做完畫掉，剩「同回合打過仗守軍殘留」與「擠開壓到擋路格」。
  前者比較有價值，會改規則輸出。

**G. #69 soundfont：裁決已下，但順序不能換**

使用者裁決 `_M` 那 11 首**內建 GM soundfont** 隨 remake 發行。接下來：

1. **先確認授權能不能隨包發行**——目前轉檔用的 FluidR3_GM 要查清楚條款與姓名標示要求；
   不合用就換一份（發行包變大數十 MB 是已接受的代價）。
2. 定它放哪：**不屬於 `assets/original/`**（那是原版素材，這是第三方資源），
   位置與 `.gitignore` 規則要另外定。
3. **授權沒確認就不要先接管線**。確認之後才做判準第一項（不需要 Docker 的轉檔入口），
   `_F` 的雜湊要與 `SOURCES.tsv` 對得上。

### 不在這一輪

#16（等使用者人工查紙本）；#48 交戰演出的量表、戰場方塊與另外三種演出；
#67 戰場「資料」捲動的收據（要 `-poke` 造 >14 列的局面）；
#68 `SPTIT.350` 標題畫面；#70 其餘素材匯出。

後兩張是同一個前提：**`SPTIT.350` 的編碼**。要動的話起點寫在 #68 本文
（檔頭形狀、沒有長度表、調色盤 48 bytes、壓縮比四條線索），解掉它同時清掉 #70 的最後一項
——#70 其餘四項裡，`BLOCK_0`–`2` 用現成的 `tools/scripts/spr_export.py 20` 就能做完。

dosgolem 不動；真的要動就先開 issue，並且整條收據重跑。

### 收尾

- `tools/gui_check.sh`（目前 40 例）、`tools/feature_check.sh`（`knownDead` 要維持空的）、
  `go test ./internal/...`、worklist verify 0 條斷言不成立；
- **`go test ./internal/... -v | grep SKIP` 要真的跑一遍**，每一支 skip 的理由都說得出來——
  第八輪把這一條寫進「已知的坑」卻跳過了，補跑才發現九支戰場測試從來沒執行過（#81）；
- `find . -path ./org_game -prune -o -user root -print` 與 `docker ps -a | grep logh3`；
  **禁止任何 prune／rmi**；
- `CONTEXT.md` 走 `render -w`、`VERIFICATION-MATRIX.md` 走 `render-matrix -w`（產生前先抓 issue 快照）；
- `WORKLOG.md` 記這一輪的流水帳與踩到的坑；關 issue 前 verify 換成跑得起來的訊號。
