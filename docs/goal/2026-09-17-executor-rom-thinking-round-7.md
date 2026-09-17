# 下一輪 goal — 執行器 ROM 尾巴（#73）→ 回合過場邊角（#72）→ 艦隊全滅（#57）

日期：2026-09-17。第六輪 goal（`docs/goal/2026-09-17-turn-transition-windows-round-6.md`）的結果：

- A #61 關（font_extract.py 改成讀版控的 JSON 核對，不再寫回）。
- B #52 關：回合過場視窗、處理中底板、四種報告框都在正常路徑整張相同（規格 059 CONFORMED）。途中把電腦 AI 對齊到第四回合
  （命令種類 `+0x08`、艦隊配槽 `07F9:03AC`、電腦座位問話＝待命），戰略打行星多支攻方逐欄相同（#56 的一種）。
- C #57 暫停：全滅轉跑 `ENDING.EXE` 的成因是 dosgolem 的 BIOS ROM 尾巴是 0（`docs/re/334`），另開 #73。

收尾狀態：`main` 已推（`eb1c0f4`），dosgolem 副本 `b5485eb` 工作樹乾淨，沒有殘留 `logh3-*` 容器，`verify` 252 條、0 條斷言不成立，
GUI 32 例全同。

**使用者裁決（2026-09-17）**：

1. #73 排進這一輪。
2. #72 要實作。「戰略模式思考中」的停留**固定 18 tick**（dosgolem 量到的長度：座位 1 印出 #551.02M → 處理中 #554.75M，
   約 3.7M 步，1 tick ＝ 200,000 步），畫面收據 `fcmd-turn-think` 整張比。規格寫明這是在執行器上量到的 AI 運算時間，
   不是遊戲的計時器。

open 清單的權威是 GitHub 上掛 `worklist` 的 issue，這份 goal 只是這一輪的計畫快照。

下面這段可以直接當下一個 session 的開場提示詞。

---

## 提示詞

你在 `~/cht/logh3`（銀河英雄傳說III SP 的 DOS remake）。先照 `CLAUDE.md` 的開工順序：
`tools/issues_snapshot.sh` → `tools/go.sh run ./cmd/worklist verify`（現況：0 條斷言不成立；open 22 張：#16、#48、#53–#60、
#62–#73）→ 讀 `CONTEXT.md` 現況一覽 → `git status --short`（logh3 與 `~/cht/dosgolem-logh3` 兩邊都看，兩邊的
`git config user.email` 都要是 `wicanr2@gmail.com`）。
再讀直接相關的：`docs/re/334`（全段）、`docs/re/332` §一與 §二、§五、§七、`docs/re/333` §三、§四、`docs/spec/059` 已知差異表、
`docs/re/329` §二（`297B:034A`）、`internal/game/turnwindow.go`、`internal/game/turnreport.go`。
dosgolem 那邊先讀 `~/cht/dosgolem/CLAUDE.md`、副本的 `docs/spec/000-index.md` 與 `189-a20-gate-and-hma-addressing.md`。

**本輪目標：修好執行器的 ROM 尾巴、整條收據換到新版執行器，再把回合過場剩下的邊角與艦隊全滅接上。**
先前的裁決照舊：每張新畫面維持 same-state 契約（RE → spec → dosgolem 收據 → READY → 實作 → 整張逐位元組），
關 issue ＋ commit ＋ push 不必再問（dosgolem 那邊只 push `logh3-oracle` 分支，合併回 master 由使用者主導）。
**#16 不要動**（使用者在人工查紙本）。subagent 不直接改 GitHub issue。

### 交付順序（每一項做完就對拍、進版控）

**A. #73 dosgolem 的 BIOS ROM 尾巴（＋#54 兩個旗標）：先做，後面全部依賴它**

A 做完之前不要錄新收據——錄了也要重錄。

1. **普查**：改唯讀之前，先查既有收據有沒有程式寫 `F0000–FFFFF`。用現版執行器跑有代表性的幾條長收據（開頭動畫、主選單、
   `fcmd-occ-phz-t4`、`eng-*` 一條、月底一條），`-watch` 每段分開寫（#54：逗號後面那段會被忽略）。有寫入就停下來看是誰、
   為什麼，再決定唯讀的範圍。
2. **#54 一起修**：`-watch A-B,C-D` 要看到每一段（或解析時直接報錯）；`-regs-at` 漏抓先找出條件（指令邊界、REP 前綴、
   中斷中），寫成可以重現的單元測試再修。它和 #73 都要整條收據重跑，合在同一次重跑，不要分兩輪。
3. **ROM 尾巴**：副本開規格（編號接 `189` 之後），寫硬體語意（PC/AT 重置向量 `FFFF0`、BIOS 日期 `FFFF5`、機型位元組 `FFFFE`＝`FC`）
   與參考來源（DOSBox-X `bios.cpp` 只讀原始碼、不照抄）；`FFFFF` 那一格照原版執行環境（js-dos 的 DOSBox 後端）取捨並寫明理由；
   `F0000–FFFFF` 寫入忽略照普查結果。單元測試：開機後讀到的 16 bytes、寫入不改內容。
4. **`.state` 補存 `a20`／`hma`／`hmaOwned`／`a20Local`**：改狀態檔版本，舊狀態檔的讀法（拒讀或預設值）寫進規格。
5. **收據換版本**：副本 commit 之後，logh3 這邊先跑 `tools/check_receipts.sh`（只比對、不覆蓋，會跑好幾個小時，可以接續），
   拿到「哪幾份變了」的表；再用 `tools/rerun_receipts.sh` 重跑。**雜湊有變的逐條核對原因**（ROM 尾巴、`-watch`／`-regs-at`
   的輸出、還是別的），寫進 `docs/playtest/021` 與 `docs/re/336`；對不上就重做收據，不要改 remake 去遷就舊收據。
   `tools/fleetcmd_receipts.sh` 等腳本裡寫死的步數如果因此失效，照新版重量並改腳本。
6. 重跑之後：`go test ./internal/...` 全過、`tools/gui_check.sh` 全同；`fcmd-occ-phz-t4` 不再轉跑 `ENDING.EXE`
   （`report.txt` 沒有 `ending.exe`、主控台沒有「NULL function pointer」）。
7. #73、#54 的 verify 換成可以跑的訊號（dosgolem 單元測試名或重錄後的收據），關 issue；`docs/re/333` 附節與 #57「卡在」加指標。

**B. #72 回合過場的邊角**

1. **「戰略模式思考中」**（使用者裁決：固定 18 tick）：實行之後兩個座位各印一次標題窗「<陣營> 宇宙曆…／戰略模式思考中」
   （`ds:06E9`）＋清單窗整窗空白；人類座位那一次立刻被蓋掉，電腦座位那一次停 18 tick 再進「戰略模式處理中」。
   先用新版執行器重量一次（`-watch-screen`，A 可能改變步數）：量出來不是 18 tick 就停下來回報，不要自己改常數。
   remake 做成一個狀態（例如 `StateTurnThinking`），規格 059 已知差異那一列改成規則；`fcmd-turn-think` 從「不比」改成整張比，
   GUI 加一例。停留期間滑鼠點擊要不要提早結束：先查 AI 那一段有沒有讀事件佇列（`3765:0003` 已在 #549.06M 清過），查不到就照原版不理。
2. **兩支以上人類艦隊**：前一支「下一格」的色 5 留到下一支第一段上畫面才消失、視窗只開一次。用 `-poke` 或正常路徑下兩道命令造局面，
   錄收據（進 `tools/fleetcmd_receipts.sh` 第 23 段），remake 對上。
3. **戰鬥中的士氣報告**（`ds:8F16 != 0` → `3826:0405`，(2,7) 24×5、頭像 (0x2C,0x28)）：造局面錄收據、接上。
4. **`1261:0067` 的 `ds:8F72[陣營] ← 297B:10DE`**、**`056A:011C` 印完標題窗只剩一列的原因**（與 `docs/internals/49` 不一致）：讀位元組，
   寫進 `docs/re/` 與 `docs/internals/`；影響畫面就補收據。
5. **`ds:8F1C` bit0（隱藏除錯選單「顯示」）**：remake 沒有除錯選單。只記錄成規格 059 的已知差異，不實作；要實作的話先問使用者。
6. 每一件做完就在 #72 本文劃掉；全做完 verify 換 `go_test` 關 issue。

**C. #57 艦隊全滅：A 做完才能錄收據**

使用者前一輪的裁決照舊：「已經全滅」報告**修成正常文字**（「<陣營>第Ｎ艦隊／提督名軍階／已經全滅。」＋頭像），列成規格的刻意差異
（原版格式字串 `ds:01CD` 開頭多一個 `\x03`）。

1. 用新版執行器重錄 `fcmd-occ-phz-t4`（費沙那條第四回合自然全滅），確認不再崩潰，並錄全滅報告那一張、回合結束後的記憶體（`recmem`）。
2. 戰略階段（`0374:0556` → `0374:0608(星域)`）：陣營 ≠ 星域擁有者、`+0x00` bit0（士氣 < 50）→ `0374:06BC` 每組扣 `+0x0A ÷ 6` 艘
   → 第 0 組 0 艘 → `0374:073F` 報告 ＋ `297B:034A`；否則 `0374:0896`（`08D9`–`08E8` 那段本意 unknown，先讀）。
   `297B:034A`：艦艇不還、提督與副官 `28E0:057F(…, 0)` 失蹤、功績歸零、星域 `+0x28` 與 `ds:8F49` 減一。
3. remake：新遊戲沒有艦艇池（#55），耗損與全滅判定照 `rules.FleetGroup.Ships` 做，差異列進規格（改 054 或開 060）。
   艦隊表、提督、星域逐欄與 `recmem` 對上；報告框正常路徑整張比（排除的只有裁決 2 那幾格文字，座標事先列進規格）。
4. 戰場那條（`185C:016A`，`185C:028E`／`0177`）：讀位元組；錄得到收據就一起接，錄不到就留在 #57 並把「卡在」寫清楚，issue 不關。
5. 拿掉 `internal/rules/fleetstep.go` 的自承註解、verify 換 `go_test`。

### 每一項的完成定義

- 執行器：副本開規格、單元測試全過（CPU 層的 SingleStepTests 不能退步）、commit 並 push `logh3-oracle`；
  logh3 這邊的收據全數在新版 HEAD 下重跑過，`meta.txt` 記的是新 HEAD。
- 收據腳本化：B、C 進 `tools/fleetcmd_receipts.sh`（第 23、24 段接著加；要 poke 的局面把 poke 寫進那一行；要倒記憶體的用 `recmem`），
  批次 `fcmd-*`，探索用 `fcx-*`；單批重跑 `tools/fleetcmd_rerun.sh <批次>`。
- 規格：B 改 059，C 改 054 或開 060。
- 測試整張 640×480 比（`matches`；事先列出的排除區才用 `matchesExcept`），每個新畫面至少一例進 `tools/gui_check.sh`
  （思考中、兩支人類艦隊、戰鬥中士氣、已經全滅各一）。
- 文件：`docs/re/336` 起逐次紀錄（不改舊的，只加指標）；`docs/internals/` 更新事實（33 艦隊表、49 標題窗、53、61）；
  `docs/playtest/021` 追加表列（雜湊全寫）；`CONTEXT.md`／`WORKLOG.md`。
- 關 issue 前把 verify 換成可以跑的訊號，重抓快照（剛開／剛改的 issue 要等 GitHub 列表更新，快照張數對了再跑）、`verify` 0 條
  斷言不成立、`render -w`／`render-matrix -w`。commit 訊息尾巴 `Co-Authored-By: Claude …`，不放 session 連結。
- 做的過程中發現的新缺口：照既有 issue 格式開新條目（主 session 開），不要只寫在文件裡。

### 已知的坑（前幾輪踩過的，別再踩）

- **執行器一改，先比對再重跑**：`rerun_receipts.sh` 會把舊畫面蓋掉（備份只留最原始的一份在 `workplace/oracle-prev/`），
  所以先跑 `check_receipts.sh` 拿到差異表。它會跑好幾個小時、可能被 OOM 砍，結果逐份附在 `workplace/receipt-check.tsv`，
  重跑會接下去；跑的時候注意這台機器有別人在用，`JOBS` 不要開大。
- **測試差異先懷疑 remake 的 fixture，不要急著注入**：第六輪「第三回合後對不上」的原因全在 remake（fixture 缺進軍路線表、
  艦隊配槽、新遊戲艘數、電腦座位問話），注入原版局面只是定位工具，修完之後正常路徑就整張相同。
- **subagent 的結論要自己重現再寫進文件**：第六輪 #57 的執行器調查由 subagent 做，主 session 重跑了決定性的 `fcx-a20-rom` 才寫進 `docs/re/334`。
  派工時把 Docker-only、只准寫 `workplace/` 與 dosgolem 副本、禁止 prune／rmi 寫進 prompt。
- **收據的時點不要落在暫態**：錄報告類畫面先用 `-watch-screen` 找穩定區，錄在中間。
- **步數換計時器 tick 要照遞減相位**：第 n 次遞減在「設定計時器那一步之後的第一個相位點」＋ (n−1) × 200,000 步（`-tick 20000`）。
  訊息速度 `ds:8F38` 預設 90 tick、`3765:006C` 停頓期間讀到滑鼠左／右鍵按下就結束，鍵盤無效。
- **地圖色 5**：報告與過場塗過的格子原版不還原（`056A:059E`／`06F3`／`089D`），下一支出發前往的「下一格」留到 `056A:044B`；
  remake 的 `mapMark` 在 `afterFleetMoves`／`beginTurn` 清掉，改流程時不要漏。
- **戰略打行星的呼叫時點與亂數消耗不能變**（`TestStrategicAssaultMatchesOracle`、`TestOccupyAstarteRunsStrategicAssault`）；
  全滅判定插進逐格處理時，確認它在原版的哪一步消耗亂數。
- **進入點先看前面是不是 epilogue**；`297B:0332`／`03D2` 不是函式入口；`docs/re/333` 引用的 `2376:03E6`／`04DE` 是 IDA 名字的位移，真正目標是 `01E6`／`02DE`。
- 移動與按鍵在同一步，原版會把那一下丟掉：收據一律先 `-mouse-at` 搬、隔 10M 步再按（`mvc`／`mvr`）；系統指令、實行可以直接 `-clicks`。
- `-poke` 要落在畫面重畫之前；位址一律 `lin:` 前綴、長度十進位；`report.txt` 核對括號裡的實際位址。
- worklist 的 `go_test` verify 的 `run` 只吃單一函式名，寫 `A|B` 會報「沒有這支測試」。
- `go fmt` 會動到不相關的檔：commit 前 `git checkout --` 那幾個（`internal/ai/original_score.go`、`internal/audio/library.go`、
  `internal/game/admiralfatereport_test.go`、`internal/rules/planetassault_oracle_test.go`）。
- `tools/py.sh` 掛 `docs/`、`workplace/`、`tools/`（唯讀）、`assets/`；Docker 只走 `tools/*.sh` 包裝器（dosgolem 副本走它自己的
  `tools/go.sh`）；收尾 `docker ps -a | grep logh3` 與 `find . -path ./org_game -prune -o -user root -print`；**禁止任何 prune／rmi**。

### 不在這一輪

#16（等使用者人工查紙本）；#48 交戰演出的量表、戰場方塊與另外三種演出；#53 戰場層打行星的三處位元組出入；#55 新遊戲艦艇池；
#56 戰略打行星剩下的局面（同回合打過仗的守軍殘留、擠開後壓到擋路格）；#58 就地駐留不扣移動費與 `297B:08D8`；
#59 一覽表分頁沒有收據的情形；#60 ▲▼ 按住連捲；#62 參謀任命；#63 情報收集分頁；#64 徵稅／生產／預算的鈕；
#65 艦隊編號清單與結束確認；#66 ＭＩＤＩ Ｉ／Ｆ 與 BUG!；#67 戰場「資料」捲動；#68 SPTIT.350 標題畫面；#69 自產曲庫管線；
#70 其餘素材匯出；#71 自家星域的艦艇補充（依賴 #55）。dosgolem 只改 A 列的三件事（ROM 尾巴、`.state` 的 A20 欄位、#54 兩個旗標），
其他執行器問題開 issue 另排。
