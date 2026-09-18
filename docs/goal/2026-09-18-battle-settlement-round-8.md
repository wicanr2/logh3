# 下一輪 goal — 戰場回合結算與戰場全滅（#57）＋ 四組並行

日期：2026-09-18。第七輪 goal（`docs/goal/2026-09-17-executor-rom-thinking-round-7.md`）的結果：

- A #73、#54 關：dosgolem `f7523f9`（規格 `191-bios-rom-tail`）。ROM 尾巴照 DOSBox 填、`F0000–FFFFF` 唯讀、狀態檔 v3 存 A20；
  `-watch` 多段、`-regs-at` 不去重。全滅不再轉跑 `ENDING.EXE`。
- B #72 關：「戰略模式思考中」固定 18 tick（裁決）、S3 同回合三支人類艦隊（收據第 25 段）、戰鬥中士氣報告視窗（交戰第 5 段）、
  靜態三項（思考中不吃點擊、`ds:8F72`、換行溢出清窗）。
- C #57 戰略那一條：規格 060（耗損、全滅、提督失蹤），費沙第四回合正常路徑整張／逐欄相同。
- 途中裁決：戰場上電腦不造成傷害（規格 061）、收據重跑改成「測試用到的先跑、`fcx-*` 不重跑」、驗證重點是功能選項有作用且不能 crash
  （`tools/feature_check.sh`）。修掉兩件：二十多支戰場測試從第六輪起被靜默 skip、遊戲本體按「存進度」不寫檔。
- 開的 issue：#74、#75、#76（已裁決關）、#77。

收尾狀態：`main` 已推，dosgolem 副本 `f7523f9`（`logh3-oracle` 已推，合併回 master 由使用者決定），收據除 `fcx-*` 外都在 `f7523f9`
重跑過，`go test ./internal/...` 全過、`tools/gui_check.sh` 34 例全同、`tools/feature_check.sh` 沒有 crash。

**使用者裁決（2026-09-18）**：

1. #57 繼續做完。
2. #76：執行器不動，**remake 只要做出一樣的效果就好**——這個專案是把遊戲機制抓出來重寫，不是複製原版的執行環境。
   位元組讀不到或執行期對照不可信時，照解包後的原碼語意做，效果對得上就算數（規格 060 已知差異表）。
3. 這一輪除了 #57，還要放 #77、#71＋#55、#74＋#75、#63／#64／#65。

下面這段可以直接當下一個 session 的開場提示詞。

---

## 提示詞

你在 `~/cht/logh3`（銀河英雄傳說III SP 的 DOS remake）。先照 `CLAUDE.md` 的開工順序：
`tools/issues_snapshot.sh` → `tools/go.sh run ./cmd/worklist verify` → 讀 `CONTEXT.md` 現況一覽 → `git status --short`
（logh3 與 `~/cht/dosgolem-logh3` 兩邊，`git config user.email` 都要是 `wicanr2@gmail.com`）。
再讀直接相關的：`docs/spec/060`（戰略那一條全滅）、`docs/re/336` §九（含戰場那一條的位元組）、
`workplace/re-notes/round7-57-static.md` §三（`185C:0001`／`001D`／`028E`／`00F0`／`0177`、`181B:000A` 的八步）、
`docs/spec/059` 已知差異（戰鬥中士氣報告的觸發點）、`docs/re/278` ＋ `docs/spec/001` 規則 4–7（#77）、
`docs/internals/33`／`62`、`internal/game/battle.go`、`internal/game/annihilation.go`、`internal/rules/annihilation.go`。

**本輪目標：把戰場回合結算做成原版的八步模型，接上戰場那一條全滅與戰鬥中的士氣報告；同時清掉四組會過期或卡著玩家的缺口。**
先前的裁決照舊：每張新畫面維持 same-state 契約（RE → spec → 收據 → READY → 實作 → 整張逐位元組），關 issue ＋ commit ＋ push 不必再問。
**#16 不要動**。subagent 不直接改 GitHub issue。

### 交付順序

**A. #77 `docs/re/278` 重驗：先做，放久了更難查**

那張表是在 dosgolem 掉了 `-click-move-lead` 的期間錄的（`6841e68`，工作樹還不乾淨）。同一份旗標在 `2e9dd3b` 是墨 2,624、
`e0b6a03` 之後是 88（`docs/re/336` §三，P）。用 `f7523f9` 重錄那一整組（含對照組 `rgnAlly180`／`rgnAllyStay`），重新判「左下角凍住的
內容」跟什麼有關（座位？開選單之前的最後一個動作？游標位置？）；結論寫 `docs/re/`，規格 001 規則 4–7 與 remake 照新結論對齊或說明不變。
`s3data-scroll-*` 那 11 份的步數也重量（#60 要用時才不會再踩一次）。

**B. #57 戰場那一條全滅 ＋ 戰場回合結算的八步：這一輪的主力**

1. **先把 `181B:000A` 的八步讀完並寫規格**（`workplace/re-notes/round7-57-static.md` §三只讀到「第 8 步 `181B:03F0` → `185C:0001`」）：
   每一步做什麼、順序、哪些步驟耗亂數。remake 現在的 `AdvanceBattleTurn` 是自己排的順序（移動 → 交火 → 要塞砲 → 打行星 → 收尾），
   要對齊成原版的步序。**規格先 READY 再動 production path。**
2. **戰場全滅**（`185C:001D`）：只看第 0 組艘數、`028E` 報告走 `3826:0405`（版面已接，規格 059）、`00F0` 擦小地圖點 ＋ `2AD2:01EF`
   作廢、`297B:034A`（與戰略那一條共用，`rules.MarkAdmiralMissing` 等已在）、最後 `12AA:03BC` 重建戰場表。
3. **戰鬥中的士氣報告觸發點**（規格 059 已知差異）：`185C:008F` 對沒被消滅的艦隊叫 `1261:0104`；修理階段 `1AAB:0182`、
   電腦戰術 AI、修理／補充指令、撤退 `185C:0D34` 也會當場印。收據 `eng-morale-low` 已經有（只驗過視窗），接上之後改成整張比。
4. 收據：戰場全滅要造局面（規格 014 `-poke`，例如把某支艦隊第 0 組艘數寫成 0 或讓它被打光），`eng-*` 段或新開一段；
   mem-in／mem-out 逐欄對（提督、艦隊表、戰場單位表 `ds:9898`、小地圖）。**注意 `297B:034A` 會先跑 `35D1:0102` 那段長迴圈**，
   mem-out 要錄在它之後（第七輪踩過：錄在 #1050M 時提督還沒改，改到 #1057M 才對）。
5. 拿掉 `internal/rules/fleetstep.go` 的自承註解、#57 的 verify 換 `go_test`。

**C. #55 新遊戲的艦艇池 ＋ #71 自家星域的補充**

戰場全滅與耗損都踩在同一個資料層上：`rules.FleetGroup` 現在沒有艦艇池索引、兩格存量、分隊士氣與戰力。先把艦艇池模型化（#55），
再接補充（`1AAB:0FAF`：擁有者的艦隊從星域存量補到編制數、補完 `1261:0067(f, 1)` 可能印「士氣值已經恢復」，收據 `fcmd-turn-recover` 已在）。
規格 059 的「已經恢復」那一列與 060 的耗損欄位（`+0x0C`／`+0x0E`／`+0x12`）到期。

**D. #74、#75 兩個小缺口**

- #74：`3982:0038` 兩位數印 `%2d` 半形，四個畫面（艦隊視窗、敵艦隊視窗、待命、交戰演出）照做；S3 開局帝國槽 9 起編號 ≥ 10，
  用 `fcmd-s3-base` 當載具各錄一份收據。
- #75：開戰前「<星域>星系內發現敵方艦隊。」（`0374:0489` 一帶還沒讀），remake 停在那張報告再進戰場；S3 帝國第一回合 #390M 那一段有畫面。

**E. #63、#64、#65：玩家點了沒反應的三處**

`tools/feature_check.sh` 目前把它們列為已知（`knownDead`）。接上一個就把對應的名字從那張表拿掉——表裡還有的項目開始有反應時測試會失敗，
提醒回頭關 issue。

- #63 情報收集的三個分頁（治安／勢力／敵艦）與「確認」。
- #64 徵稅計劃與生產計劃的「結束」、預算計劃的箭頭（`26E2`）。
- #65 艦隊編號清單的列、結束確認的中止／決定。

### 每一項的完成定義

- 規格：B 開新的一份（戰場回合結算的八步）或擴 `docs/spec/028`／`019`；C 改 059／060 的已知差異；A 改 001 與 `docs/re/278` 的指標。
- 收據腳本化（`tools/*_receipts.sh`），批次 `fcmd-*`／`eng-*`，探索用 `fcx-*`；**探索批次不重跑**（第七輪裁決）。
- 測試整張 640×480 比（`matches`；事先列出的排除區才用 `matchesExcept`），新畫面至少一例進 `tools/gui_check.sh`；
  跑 `tools/feature_check.sh` 確認沒有 crash、沒反應的項目只剩還沒接的。
- 文件：`docs/re/337` 起逐次紀錄；`docs/internals/` 更新事實；`docs/playtest/` 追加表列（雜湊全寫）；`CONTEXT.md`／`WORKLOG.md`。
- 關 issue 前 verify 換成可以跑的訊號，重抓快照、`verify` 0 條斷言不成立、`render -w`／`render-matrix -w`。

### 已知的坑（第七輪踩過的，別再踩）

- **測試 skip 是沉默的失敗**：`battleTestGame` 在第六輪加入過場視窗之後一直停在 `StateTurnWindow`，二十多支戰場測試 skip 了一整輪沒人發現。
  改動回合流程之後，**跑 `go test ./internal/... -v | grep SKIP` 看一遍**，skip 的理由要說得出來。
- **戰場測試要照原版時記得關掉「電腦不造成傷害」**（`g.originalBattleAI = true`，規格 061）；忘了的症狀是敵方不動、損失 0。
- **收據綁執行器版本**：`rgnAlly*` 與 `s3data-scroll-*` 這次變了，因為舊值錄在更早、甚至工作樹不乾淨的版本上（#77）。
  新錄收據時 `meta.txt` 的 HEAD 與「未提交檔案 0 個」都要看。
- **`-poke` 造局面要寫翻轉之後的狀態**：第七輪為了戰鬥中的士氣報告，改艘數不會觸發重算（那一回合沒有扣艘數的步驟），
  直接寫 `+0x00` bit0｜bit2 與 `+0x3C` 才會印。
- **mem-out 的時點**：`297B:034A` 先跑 `35D1:0102` 的長迴圈（`+0x6C` 是 `FFFFFFFF`），錄太早會看到「什麼都沒改」。
- **量「有沒有作用」要排除游標**：點擊會把游標搬過去，畫面雜湊一定變；先 `MoveMouse` 再取基準。疊加視窗開起來之後要重新進入，
  不然後面的格子都點在視窗上。
- **步數換 tick**：1 tick ＝ 200,000 步（`-tick 20000`）；訊息速度 `ds:8F38` 預設 90 tick，停頓時左／右鍵按下就結束、鍵盤無效。
- **Docker-only**：分析、反組譯、測試一律走 `tools/*.sh`；主機只做 docker／git／檔案編輯（第七輪主 session 犯過一次，記在 WORKLOG）。
- `go fmt` 會動到不相關的檔：commit 前 `git checkout --` 那幾個（`internal/ai/original_score.go`、`internal/audio/library.go`、
  `internal/game/admiralfatereport_test.go`、`internal/rules/planetassault_oracle_test.go`、`cmd/worklist/importmatrix.go`）。
- 收尾：`docker ps -a | grep logh3`、`find . -path ./org_game -prune -o -user root -print`；**禁止任何 prune／rmi**。

### 不在這一輪

#16（等使用者人工查紙本）；#48 交戰演出的量表、戰場方塊與另外三種演出；#53 戰場層打行星的三處位元組出入；
#56 戰略打行星剩下的局面；#58 就地駐留不扣移動費；#59 一覽表分頁沒有收據的情形；#60 ▲▼ 按住連捲（#77 會順便重量步數）；
#62 參謀任命；#66 ＭＩＤＩ Ｉ／Ｆ 與 BUG!；#67 戰場「資料」捲動；#68 SPTIT.350 標題畫面；#69 自產曲庫管線；#70 其餘素材匯出。
dosgolem 不動（#76 的裁決）；真的要動就先開 issue，並且整條收據重跑。
