# 下一輪 goal — 一覽表分頁（#51）→ 戰略打行星與艦隊解散（#46）→ 交戰演出的圖（#48）

日期：2026-09-17。第四輪 kickoff 的六片裡關了四片：月底逐座位扣稅（#45，`docs/re/321`）、
一覽表四張的列（#47，`re/323`、規格 055）、敵艦隊分頁有列的時候（#49，`re/325`）、
戰場視野的滑鼠捲動與滑鼠作用總表（#50，`re/326`、規格 057、internals 60）。另外兩片做了前半：
艦隊走到之後的四條分支、到達收尾與占據旗標（#46，`re/322`、規格 054），以及交戰演出的入口、
三個視窗與戰鬥曲出口（#48，`re/324`、規格 056）。途中新開 #51（一覽表的分頁鈕）。
收尾狀態：工作樹乾淨、`main` 已推，dosgolem 副本 `b5485eb` 已推，沒有殘留 `logh3-*` 容器，
`verify` 231 條、0 條斷言不成立。詳見 `WORKLOG.md` 最後六條。

同一天使用者另外裁決了三件事，已經落地：

- 解出來的原版素材可以進版控，放 `assets/original/`（`AGENTS.md`「Docker-only 與原版素材
  邊界」）。頭像 83 張已經放進 `assets/original/face/`。
- #16 改由使用者人工查：對照表 `docs/lookup/admiral-portraits.md`（頭像 ＋ 十個已命名欄位 ＋
  `+20`／`+21`／`+22`／`+24`，由 `tools/scripts/admiral_lookup.py` 產生）。
- 艦隊解散（`297B:0332`）併進 #46，判準已更新。

下面這段可以直接當下一個 session 的開場提示詞。

---

## 提示詞

你在 `~/cht/logh3`（銀河英雄傳說III SP 的 DOS remake）。先照 `CLAUDE.md` 的開工順序：
`tools/issues_snapshot.sh` → `tools/go.sh run ./cmd/worklist verify`（現況：0 條斷言不成立，
open 的是 #51、#48、#46、#16）→ 讀 `CONTEXT.md` 現況一覽與「玩家還碰不到的功能」→
`git status --short`。再讀直接相關的：`docs/re/323` §五、`docs/spec/055`、`docs/re/322` §五–§七、
`docs/spec/054` 已知差異表、`docs/re/310` 那張函式表（`297B:0332`）、`docs/re/324` 結論。

**本輪目標：先把一覽表收尾，再讓占領真的打得下行星、艦隊解散得掉。** 使用者 2026-09-17
定的順序是 **#51 → #46 → #48**。先前的裁決照舊：每張新畫面維持 same-state 契約（RE → spec →
dosgolem 收據 → READY → 實作 → 整張逐位元組），關 issue ＋ commit ＋ push 不必再問。
**#16 不要動**——使用者正在拿 `docs/lookup/admiral-portraits.md` 人工查紙本資料表，
結果由使用者帶回來。

### 交付順序（每一項做完就對拍、進版控）

**A. #51 一覽表的分頁鈕：先做，小**

收據已經錄好（`tools/fleetcmd_receipts.sh` 第 20 段）：`fcmd-lst-star-tab-assign`、
`fcmd-lst-star-tab-tax`、`fcmd-lst-star-reopen`、`fcmd-lst-fleet-tab-unit`、`fcmd-lst-fleet-tab-ops`。

1. 反組譯 `0D05:02FC`（依 `ds:96E6` 印三種列）與 `0C7F:0339(分頁)`（印三種列並存
   `ds:8F31[座位]`），切片在 `workplace/ida/fcx0D05.com`／`fcx0C7F.com`。要回答：配屬那一格
   `BHCADTE` 的「有無」看的是哪個存量（星域 `+0x30` 起的艦艇艘數，還是別的）、稅金的
   「徵稅額 ( 稅率%)」取哪兩格、艦隊資料的士氣／部隊（`1/ 8` 的分子分母）／作戰（命令碼 →
   字串表）各讀艦隊記錄哪一格。
2. 版面從那五批的字元平面（`plane.bin`）讀，併進 `list-screen-panel.json`（單一來源，
   測試不另抄座標）。
3. 分頁每座位記住（`ds:8F2D[座位]`／`ds:8F31[座位]`），`fcmd-lst-star-reopen` 驗關了再開仍在稅金頁。
4. remake：`listscreen.go` 接分頁熱區（碼 3／4／5），拿掉「一覽表的分頁鈕還沒接」那句自承註解；
   五批整張逐位元組；規格 055 補分頁一節；GUI 至少一例（`fcmd-lst-star-tab-tax`）。
5. ▲▼ 按住連捲（規格 055 已知差異）**不在 A 裡**，除非讀 `0C7F`／`0D05` 的事件迴圈時順手讀清楚。

**B. #46 戰略打行星＋駐留解散＋艦隊解散：這一輪的主力**

1. **先補新遊戲的艦艇池**。這是 B 的前置，也是新遊戲走進戰鬥會撞到的同一個洞。現況：
   remake 新遊戲的艦隊沒有艦艇記錄（`Slot` 是 nil），`rules.PlanetAssault`／`assaultCounter`／
   `spreadOverFleet` 都吃艦艇記錄。原版新遊戲從 `SHIP.DTX`（352 × 78，`docs/internals/30`）起步，
   艦隊組成「決定」時分配池索引（新組拿到 2、3、4…，`docs/spec/015` 第 129 行的已知差異）。
   要讀的：艦隊組成決定時寫艦艇池的那一支（從 `0FF4:000A` 的決定分支往下找）、艦隊記錄
   `+0x10` 起十組 4 bytes 怎麼填（`docs/internals/38`）。收據用既有的 `fcmd-*` 檢查點，建完
   艦隊之後倒 `ds:CFB4` 指的池（far 指標每批都要倒，不要沿用舊值）。
2. **戰略打行星**（`20B4:1497` → `1512` 九回合迴圈 → `1662`，`docs/re/322` §五已讀完）：
   接在 `0374:0556` bit4 那條（remake `rebellion.go` 的 `scanRegionsFrom`）。亂數要與原版同序
   （`ds:94DC`）。remake 的叛亂亂數在第三回合就分岔，**不要靠多跑幾回合等它剛好對上**：
   依規格 014，在 `fcmd-occ-*` 打行星那一回合之前的檢查點，把原版的亂數狀態與相關表讀出來，
   用直接進入的方式把同一份狀態餵給 remake，只比打行星那一步（行星 `+0x14`／`+0x11`、
   艦隊 `+0x3C`、`ds:8FE4`／`8FEC`、艦艇記錄）。注入的內容寫進收據。
   驗收點：亞斯提 4200 → 3768、`+0x3C` 100 → 67；費沙 8800 → 8792、`+0x3C` 47。
3. **駐留到達的解散**（`297B:025D`，已讀：釋放提督與副官、`+0x04`／`+0x05 ← FF`、艦艇還回
   星域）＋ **玩家下的艦隊解散**（`297B:0332`，沒讀）：先找玩家從哪張畫面下解散（艦隊視窗或
   命令清單，從 `297B:0332` 的呼叫者往上找），讀完兩支的差別；收據從 `fcmd-*` 建好艦隊的
   檢查點下解散、倒艦隊表／星域 `+0x28` 艦隊數／`+0x30` 起的艦艇艘數／提督 `+0x09`。
4. 士氣報告家族（`ds:32B4`，費沙那條 `+0x3C` 47 會跳）：門檻與誰印，碰到就讀；不讀的話
   在規格 054 寫明是已知差異。
5. remake：拿掉 `rebellion.go`「戰略打行星還沒接」與 `game.go` 那句「remake 還沒有解散那一支」
   的註解；規格 054 已知差異表對應三列改成規則、CONFORMED；#46 的 verify 換 `go_test`。

**C. #48 交戰演出的圖與動畫：做到哪算哪**

先解 `COMANI.BMG` 的編碼（`docs/internals/20`）。這一場讀 `PPD0`（行星圖）、`FÀD0`（同盟艦圖）、
`PPAB`（打的動畫），`eng-text` 等收據有原版畫出來的樣子可以當明文對照；`ANI_PNL1.A98` 的容器
同 `ANI_PNL0`（`docs/internals/37`）。格式解出來之後匯出成 PNG 進 `assets/original/comani/`
（新規則，匯出腳本進 `tools/scripts/`、雜湊進 `docs/internals/data/`）。之後才是兩條量表、
方塊動畫（`3931`／`3B74` 段）與計時（約 29M 步）。艦隊對艦隊與要塞砲那兩種演出這一輪不接。

### 每一項的完成定義

- 收據腳本化：A、B 進 `tools/fleetcmd_receipts.sh`（同一條檢查點鏈，要 poke 的局面把 poke
  寫進那一行），批次 `fcmd-*`，探索用 `fcx-*`；C 進 `tools/engagement_receipts.sh`。單批重跑
  `tools/fleetcmd_rerun.sh <批次>`／`-s <段號>`。
- 規格：A 改 055；B 改 054（艦艇池若規則夠多就開 058）；C 改 056。
- 測試整張 640×480 比（`matches`；事先列出的排除區才用 `matchesExcept`），GUI 每個新畫面
  至少一例進 `tools/gui_check.sh`。
- 文件：`docs/re/327` 起逐次紀錄（不改舊的，只加指標）；`docs/internals/` 更新事實
  （33 艦隊表、38 艦艇池、59 一覽表、20 COMANI）；`docs/playtest/` 追加表列（雜湊全寫）；
  `CONTEXT.md`／`WORKLOG.md`。
- 關 issue 前把 verify 換成 `go_test`，重抓快照、`verify` 0 條斷言不成立、`render -w`／
  `render-matrix -w`。commit 訊息尾巴 `Co-Authored-By: Claude …`，不放 session 連結；
  `git config user.email` 要是 `wicanr2@gmail.com`。

### 已知的坑（第四輪踩過的，別再踩）

- **移動與按鍵在同一步，原版會把那一下丟掉**（`docs/re/323` §六）：`-clicks` 把游標搬到會換
  反白的列並在同一步按下，`0D05:0141` 0 次。一律先 `-mouse-at` 搬、隔 10M 步再按
  （第 20 段的 `mvc`／`mvr`）。按分頁鈕同樣適用。
- ▼ 按住 0.3M 步是兩列（清單重印一次約 0.15M），0.1M 才是一列。
- `newTestGame` 直接 `LoadSlot` 的 S3 治安是 99%、原版 93%——比畫面的測試走
  `loadSlotFromDisk(10)`（載入當下重算未安定度）。
- 艦隊回合 120M 步一回合；60M 太密，會把下一次「實行」吃進上一回合的過場。
- remake 的叛亂亂數第三回合起分岔（`TestOccupyAstarteArrivalMatchesOracle` 第三回合起只比
  記憶體）——B-2 用注入狀態，不要等亂數。
- 看收據的時點要落在動作完成之後（`cam-long` 第一次在放開之前看，看到捲一半的畫面）。
- **remake 自己的舊假設不是證據**：動到沒有收據的既有規則前，先錄一份對照批次。
- `-poke` 要落在畫面重畫之前；訊息框有逾時（`3765:006C`），點了之後 23M 步截；位址一律
  `lin:` 前綴，`report.txt` 核對括號裡的實際位址。
- 收據旗標是滑鼠座標，畫面 y ＝ 滑鼠 y ＋ 40；`-save-state` 放在 `-steps` 的前一步。
- 對拍收斂三步：`tools/scripts/celldiff.py` → `planerows.py` → `crop_png.py`；
  `LOGH3_DUMP_REMAKE=1` 讓失敗的測試把 remake 畫面倒到 `workplace/remake-<批次>.bin`。
- `go fmt` 會動到不相關的檔：commit 前 `git checkout --` 那幾個（`planetassault*.go`、
  `cmd/worklist/importmatrix.go`、`internal/ai/original_score.go`、`internal/audio/library.go`、
  `internal/game/admiralfatereport_test.go`）。
- Docker：只走 `tools/*.sh` 包裝器；`tools/py.sh` 只掛 `docs/`、`workplace/`、`tools/`（唯讀），
  匯出到 `assets/original/` 要先寫到 `workplace/` 再複製；收尾 `docker ps -a | grep logh3` 與
  `find . -user root`；**禁止任何 prune／rmi**。

### 不在這一輪

#16（等使用者人工查紙本）；回合裡的過場視窗（處理中／思考中、出發前往／已經到達）與問話視窗
`113A:0899` 的畫面收據；艦隊對艦隊、要塞砲兩種交戰演出；參謀任命（`090A:0248`）；情報收集
三個分頁與參謀／確認、徵稅與生產計劃的參謀／結束、預算計劃的箭頭、艦隊編號清單的列、
結束確認的中止／決定、設定畫面的「ＭＩＤＩ Ｉ／Ｆ」（這些還沒有 issue）；「資料」的捲動與
`SPTIT.350` 標題畫面；其餘已解素材（`SPRITE.SPR`、`.VO5`、`.MDS` 等）搬進 `assets/original/`；
自產曲庫管線與 MIDI 音色。不改 dosgolem，除非收據需要；改了就整條 `fcmd-*` 重跑。
