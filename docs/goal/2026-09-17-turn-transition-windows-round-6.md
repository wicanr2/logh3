# 下一輪 goal — 回合裡的過場視窗與報告（#52）→ 艦隊全滅（#57）

日期：2026-09-17。第五輪 goal（`docs/goal/2026-09-17-list-tabs-strategic-assault-round-5.md`）三項都做了：
一覽表分頁（#51 關，`docs/re/327`）、駐留解散與戰略打行星（#46 關，`docs/re/329`／`330`、規格 058）、`COMANI.BMG` 解碼與
交戰演出的圖、計時、按鍵（#48 前進，`docs/re/328`／`331`）。途中使用者裁決解出來的原版素材進版控（`assets/original/`，
1,884 檔）。收尾狀態：`main` 已推（最後一個 commit 是登記 issue 與這份 goal，見 `WORKLOG.md` 最後一條），dosgolem 副本 `b5485eb`，
沒有殘留 `logh3-*` 容器，`verify` 249 條、0 條斷言不成立。

第五輪留下的與原本沒有 issue 的都登記成 issue 了（2026-09-17，使用者指示）：#52–#70，加上既有的 #16、#48、#53。
open 清單的權威是 GitHub 上掛 `worklist` 的 issue，這份 goal 只是這一輪的計畫快照。使用者定的主軸是 **#52 回合過場視窗**。

下面這段可以直接當下一個 session 的開場提示詞。

---

## 提示詞

你在 `~/cht/logh3`（銀河英雄傳說III SP 的 DOS remake）。先照 `CLAUDE.md` 的開工順序：
`tools/issues_snapshot.sh` → `tools/go.sh run ./cmd/worklist verify`（現況：0 條斷言不成立、8 條 `manual`；open 的是
#16、#48、#52–#70）→ 讀 `CONTEXT.md` 現況一覽與「玩家還碰不到的功能」（已改成 issue 索引）→ `git status --short`。
再讀直接相關的：`docs/re/322` §一、§六（`113A:047D` 的 `+060F`、`0797`／`0A4E`／`0B46`／`0899`）、`docs/re/330` §2.6
（`20B4:1922`）與 §6（`1261:0067`／`0104`）、`docs/re/329` §二（`297B:034A`）、`docs/spec/054`／`058` 已知差異表、
`internal/game/game.go` 的 `advanceTurn`／`moveFleetsFrom`／`StateOccupyAsk`（回合中途停下來的現成寫法）。

**本輪目標：「實行」之後玩家看得到原版回合裡的每一張過場視窗。** 先前的裁決照舊：每張新畫面維持 same-state 契約
（RE → spec → dosgolem 收據 → READY → 實作 → 整張逐位元組），關 issue ＋ commit ＋ push 不必再問。
**#16 不要動**（使用者在人工查紙本）。subagent 不直接改 GitHub issue。

### 交付順序（每一項做完就對拍、進版控）

**A. #61 font_extract.py 的坑：暖身，小**

`tools/scripts/font_extract.py` 重跑會整份蓋掉人工補過的 `docs/internals/data/font-graph-img.json`（62 行舊格式 vs 77 行）。
先比兩份差在哪幾個欄位，選一個做法（腳本改輸出新格式，或人工欄位拆到另一個檔），改完重跑一次確認 JSON 不變、
`assets/original/font/` 雜湊與 `MANIFEST.tsv` 相同；拿掉腳本開頭那句警告、verify 換 `absent`。
兩種做法都不動資料格式的語意，如果發現必須改欄位意思，停下來問使用者。

**B. #52 回合裡的過場視窗與報告：這一輪的主力**

1. **先讓回合處理可以中途停下**。`advanceTurn` 現在同步跑完；`StateOccupyAsk` 已經示範了「停在第 i 支、回答之後從
   i＋1 接回」。把每一支艦隊走一格（`stepFleet`）、戰略打行星（`strategicAssault`）、士氣報告做成可以停在任一張視窗、
   關掉（或逾時）再接回的狀態。**先寫規格 059 的狀態圖再動程式**；既有收據（`fcmd-occ-*`、`fcmd-garrison-*`、
   `fcmd-assault-*`、月底與叛亂那幾條）一條都不能壞。
2. **出發前往／已經到達／已經駐留**（`113A:0B46`）：人類座位（`3982:022D(陣營) == 0` 或 `ds:8F1C` bit0）每一支走一格印一張：
   `\x0C%s  第Ｎ艦隊\n提督名＋軍階\n`＋頭像 `3958:0215(提督, 0x44, 0x23)`＋「現在格名 出發前往 下一格名」；地圖現在格塗色 5 →
   等 `ds:8F38 >> 2` → 塗回、下一格塗色 5；到了目的地加「\n已經到達目標星系。」（駐留「\n已經駐留在星系上。」），等 `ds:8F38`。
   視窗 `056A:03E0` 開的 `3603:0006(0x97, 20, 6, 18, 6, 1)`，32 支跑完才關。收據：從 `fcmd-exec` 那條（第１艦隊往亞爾提那）與
   `fcmd-garrison-done` 的實行，用 `-watch-screen` 找每張的時點，**錄在停頓中間**（不要錄在塗色或換行的暫態上——`eng-end`
   那種教訓）；`fcmd-garrison-report`（#570M）已經是「已經駐留」那一張。
3. **兩張錯誤報告**（`113A:0797`「無法前往敵方的星系\n。」——句號在下一行是原版字串；`113A:0A4E`「由於移動經費不足無法移動。」）：
   同一個視窗、同一種等法。局面用 `-poke` 造（規格 014）：經費不足把 `ds:8F52[陣營]` 寫到小於艦隊 `+0x0C`；被擋把路上那一格的
   擁有者改掉。
4. **右側面板「戰略模式處理中」**（電腦座位「戰略模式思考中」）與清單暫時清掉：量它什麼時候出現、什麼時候還原。
5. **拿下行星的訊息框**（`20B4:1922`，`docs/re/330` §2.6）：「我軍第Ｎ艦隊\n提督\n正在攻占＋星域名＋\n星系內行星＋行星名＋\n
   已經歸屬%s掌控之中」、`3826:0307` 訊息框＋提督頭像，人類座位約 18M 步。收據：`fcx-assault-cap` 那種 `-poke` 巴溫防衛力
   100 的局面，腳本化成 `fcmd-assault-cap-*`（進 `tools/fleetcmd_receipts.sh`）。
6. **士氣報告家族**（`ds:32B4`：「士氣值低落的關係，\n已經無法戰鬥。」「已經恢復，\n可以進行戰鬥。」「艦隊士氣值　%4u」
   「已全軍覆沒。」）：`1261:0104`（`+0x00` bit2 → `1261:0138`、bit3 → `1261:0267`，然後清 bit2／bit3）。先找誰用旗標 1 叫
   `1261:0067` 或 `1261:00C8` 逐支迴圈、在回合裡的哪一步；收據 `fcmd-occ-phz-t3` 已經停在「士氣值低落」那一張。
7. 問話視窗 `113A:0899`（`StateOccupyAsk`，版面照反組譯、沒有畫面收據）：正常路徑要路上那一格中途易主才碰得到，
   用 `-poke` 把占領路徑上的格子改成敵方造一份收據，把現在的版面驗掉。
8. remake：拿掉 `rebellion.go` 的「回合過場的訊息框還沒畫」；規格 054／058 已知差異表對應的列改成規則；#52 的 verify 換 `go_test`。

**C. #57 艦隊全滅：B 做完士氣報告之後接**

`297B:034A`（`docs/re/329` §二）：艦艇不還、提督與副官 `28E0:057F(…, 0)` 記成失蹤、「…第Ｎ艦隊…已經全滅。」報告
（`0374:073F`）、戰場那條 `185C:016A`。先讀 `0374:0608` 的呼叫者與 `0374:06BC`／`0896`，局面用 `-poke` 把第 0 組艦艇
`+0x08` 改成 0 造（艦艇池 far 指標 `ds:CFB4` 每批先讀再算位址）。remake 的新遊戲艦隊沒有艦艇池（#55），全滅判定照
`rules.FleetGroup.Ships` 做，差異列進規格。拿掉 `fleetstep.go` 的自承註解、verify 換 `go_test`。

### 每一項的完成定義

- 收據腳本化：B、C 進 `tools/fleetcmd_receipts.sh`（新段號接在第 22 段之後；要 poke 的局面把 poke 寫進那一行；要倒記憶體的
  用 `recmem`），批次 `fcmd-*`，探索用 `fcx-*`；單批重跑 `tools/fleetcmd_rerun.sh <批次>`（它認得 `rec` 與 `recmem`；
  用到的變數要放在腳本標頭）。
- 規格：B 開 059（回合過場視窗與報告），054／058 已知差異改指標；C 改 054 或開 060。
- 測試整張 640×480 比（`matches`；事先列出的排除區才用 `matchesExcept`），GUI 每個新畫面至少一例進 `tools/gui_check.sh`
  （出發前往、錯誤報告、攻占訊息框、士氣報告各一）。
- 文件：`docs/re/332` 起逐次紀錄（不改舊的，只加指標）；`docs/internals/` 更新事實（33 艦隊表 `+0x00` 各位元、53 戰略打行星、
  視窗與報告的版面）；`docs/playtest/021` 追加表列（雜湊全寫）；`CONTEXT.md`／`WORKLOG.md`。
- 關 issue 前把 verify 換成 `go_test`，重抓快照（剛開／剛改的 issue 要等 GitHub 列表更新，快照張數對了再跑）、`verify` 0 條
  斷言不成立、`render -w`／`render-matrix -w`。commit 訊息尾巴 `Co-Authored-By: Claude …`，不放 session 連結；
  `git config user.email` 要是 `wicanr2@gmail.com`。
- 做的過程中發現的新缺口：照既有 issue 格式開新條目（主 session 開），不要只寫在文件裡。

### 已知的坑（第五輪踩過的，別再踩）

- **subagent 的結論要自己重現再寫進文件**：第五輪三個 subagent 的結論都重跑過（驗證腳本、位元組、收據雜湊），其中一個
  推翻了既有文件三句話。subagent 也有一次在主機直接跑 Python（唯讀），派工時把 Docker-only 寫得更明白。
- **收據的時點不要落在暫態**：`eng-end`（#1169M）截在 `321F:0897` 畫到一半；`eng-after` 被舊文件誤寫成「回戰場」。
  錄報告類畫面先用 `-watch-screen` 找穩定區，錄在中間。
- **步數換計時器 tick 要照遞減相位**：第 n 次遞減在「設定計時器那一步之後的第一個相位點」＋ (n−1) × 200,000 步
  （`-tick 20000`），不是從設定那一步起算（`docs/re/331`，`TestEngagementMatchesOracle` 開頭註解）。訊息速度 `ds:8F38`
  預設 90 tick、`3765:006C` 停頓期間讀到滑鼠左／右鍵按下就結束，鍵盤無效。
- **回合結束的 `refreshRegionPosture` 不重算 bit4**：`056A:0469` 只在命令階段走完艦隊之後跑（`prepareOccupationReports`）。
  改回合流程時不要把它加回去（`TestRecomputeRegionsMarksOccupation`）。
- **`297B:0332`／`03D2` 不是函式入口**（是呼叫 `090A:0E6D` 的位址）；判定進入點先看前面是不是 epilogue、掃描器會不會漏定址模式。
- 戰略打行星是直接驅動 `rules.StrategicAssault` 對 `mem-in`／`mem-out`；改回合流程時它的呼叫時點（`scanRegionsFrom`
  占據那一格）與亂數消耗不能變（`TestStrategicAssaultMatchesOracle`、`TestOccupyAstarteRunsStrategicAssault`）。
- 移動與按鍵在同一步，原版會把那一下丟掉：收據一律先 `-mouse-at` 搬、隔 10M 步再按（`mvc`／`mvr`）；
  系統指令、實行那種不換反白的鈕可以直接 `-clicks`。
- remake 的叛亂亂數第三回合起與原版分岔：過場視窗的收據盡量取前兩回合，或用 `recmem` 在那一步倒記憶體直接驅動。
- `-poke` 要落在畫面重畫之前；位址一律 `lin:` 前綴、長度十進位；`report.txt` 核對括號裡的實際位址。
- 這一版 dosgolem 的 `-watch A-B,C-D` 只看第一段、`-regs-at` 偶爾漏抓（#54）：分開寫多個 `-watch`，不要依賴「沒命中」。
- `go fmt` 會動到不相關的檔：commit 前 `git checkout --` 那幾個（`internal/ai/original_score.go`、`internal/audio/library.go`、
  `internal/game/admiralfatereport_test.go`、`internal/rules/planetassault_oracle_test.go`）。
- `tools/py.sh` 掛 `docs/`、`workplace/`、`tools/`（唯讀）、`assets/`；Docker 只走 `tools/*.sh` 包裝器；收尾
  `docker ps -a | grep logh3` 與 `find . -path ./org_game -prune -o -user root -print`；**禁止任何 prune／rmi**。

### 不在這一輪

#16（等使用者人工查紙本）；#48 交戰演出的量表、戰場方塊與另外三種演出；#53 戰場層打行星的三處位元組出入；#54 dosgolem
的兩個旗標（修了要整條收據重跑，另排一輪）；#55 新遊戲艦艇池；#56 戰略打行星的三種邊角局面；#58 就地駐留不扣移動費與
`297B:08D8`（B 讀 `113A` 時順手讀掉就關）；#59 一覽表分頁沒有收據的情形；#60 ▲▼ 按住連捲；#62 參謀任命；#63 情報收集分頁；
#64 徵稅／生產／預算的鈕；#65 艦隊編號清單與結束確認；#66 ＭＩＤＩ Ｉ／Ｆ 與 BUG!；#67 戰場「資料」捲動；#68 SPTIT.350
標題畫面；#69 自產曲庫管線（`_M` 音色要使用者裁決）；#70 其餘素材匯出。不改 dosgolem，除非收據需要；改了就整條收據重跑。
