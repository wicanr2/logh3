# 下一輪 goal — 戰略艦隊命令做完整（#41 → #42），讓「下命令 → 實行 → 走 → 到」閉合

日期：2026-09-16。上一輪（2026-09-15）把第一片「移動」接上並關 #40：十五批
`fcmd-*` 收據整張逐位元組相同，規格 051 CONFORMED（第一片），詳見
`docs/playtest/021`、`docs/re/312` §七、`WORKLOG.md` 最後一條。

下面這段可以直接當下一個 session 的開場提示詞。

---

## 提示詞

你在 `~/cht/logh3`（銀河英雄傳說III SP 的 DOS remake）。先照 `CLAUDE.md`
的開工順序：`tools/issues_snapshot.sh` → `tools/go.sh run ./cmd/worklist verify`
（現況：0 條斷言不成立，open 的是 #41、#42、#43、#44、#16）→ 讀 `CONTEXT.md`
現況一覽與「玩家還碰不到的功能」→ `git status --short`。再讀直接相關的三份：
`docs/re/312`（§三、§四、§7.5–7.9）、`docs/spec/051`、`docs/playtest/021`。

**本輪目標：把戰略艦隊命令做完整。** 使用者 2026-09-15 已裁決：先做艦隊命令、
每張新畫面維持 same-state 契約（RE → spec → dosgolem 收據 → READY → 實作 →
整張逐位元組）、關 issue ＋ commit ＋ push 不必再問。不要重問這三件事。

### 交付順序（便宜的先做，每一項做完就對拍、進版控）

**A. 補規格 051 留下的兩個小洞（同一條收據鏈各多一步，半天內）**

1. 清單那一列的「移動圖示」其實是一個字：`fcmd-done/plane.bin` 列 13 欄 22 是
   Big5 `A1B8`（☆）、屬性 `0x80`（底 8、字 0）、樣式 `0x08`。錄 `fcmd-self-done`
   （從 `fcmd-pick` 點瓦爾哈拉再按決定 ＝ 取消命令）看 ☆ 有沒有，定出規則
   （假設：`+0x07 != 0xFF` 就印）。做完把 `TestFleetMoveOrderWritesFleet` 改成
   整張比、`tools/gui_check.sh` 的 `fcmd-done` 拿掉排除矩形。
2. 下了命令之後再開視窗：從 `fcmd-done` 再點清單第 0 列（滑鼠 450:268）錄
   `fcmd-reopen`，看第 5–7 列（`0DEA:03F9`／`04C0` 對 `+0x08 != 5` 的分支；
   格式字串 `move_target`／`standby_planet` 已在 `fleet-window.json`）。

**B. #41 駐留（2）與輸送（3）** — 與移動同一個目的地迴圈，只差動詞與 `+0x08`。
從 `fcmd-menu.state` 各展開：點第 2／3 項 → 目的地 → 亞爾提那 → 決定，各錄
pick／confirm／done 三張，對 `fleet0.bin` 的 `+0x08` ＝ 2／3。remake 只要把
`fleetMenuClick` 的閘門打開（`fleetTargetsAllowed` 與 `applyFleetOrder` 已照動作碼寫）。

**C. 回合處理：命令之後按「實行」會發生什麼** — 這一項決定「能玩」的迴圈有沒有
閉合，排在 B 之後、C′ 之前。從 `fcmd-done` 按 系統指令 → 實行（畫面 (544,125) →
(560,125)，滑鼠 y 減 40）錄一張：原版有沒有 `113A:0B46` 那類模態報告
（「出發前往／已經到達／已經駐留」、`113A:0797`「無法前往敵方的星系」、
`113A:0A4E`「由於移動經費不足無法移動」）。有 → 新畫面，走完整契約；沒有 →
把「走一格之後」的地圖、清單、左下角對上（remake 的 `Advance` 已會走）。

**C′. #41 待命（4）** — `0DEA:1320`「指定待命行星」是一張新畫面（行星清單）。
反組譯 `0DEA:1320`／`0DEA:1433`（`workplace/ida/fcx0654.com`，線性 `0x7960 + 0x1320`
起；`.i64` 連續開關會壞，一次多目標），寫 DRAFT → READY 段落，錄 開／選一顆／
右鍵／決定，實作。寫 `+0x08 ← 5`、`+0x07 ← 行星`。

**D. #41 占領（1）** — 允許表換成 (敵方, 中立)，從首都 k ≤ 1 走不到任何敵方／中立
星域，所以要造局面：`-poke` 把艦隊 `+0x06`（`lin:` 前綴、長度十進位）改到
一個鄰接費沙或同盟星域的格子，poke 內容寫進收據旗標與 `docs/playtest`；
右鍵選那一格讓清單列出它。費沙有兩層訊息框（`ds:2F06` 警告、`ds:2FF0`
「你確定要占領費沙嗎」）。

**E. #41 選單的停用旗標**（`docs/re/312` §7.6）— 不是自己的星域或該格有戰鬥只剩
「移動」，艦隊 `+0x00` bit0 關占領與輸送，停用色 `0x98`。局面同樣用 poke 造
（星域 `+0x12`／`+0x13`），錄停用的樣子與「點了沒反應」。

**F. #42「變更」→ 副官／編組** — 視窗迴圈鈕 1 → `0DEA:17A9(艦隊)`；從
`fcmd-win.state` 點 (490,365)（滑鼠 490:325）就是 `fcx-henkou` 那張。副官那條
多半接既有的副官清單畫面（規格 013／014），編組接艦隊組成（規格 015）——先看
分派再決定是接線還是新畫面。

做到哪算哪：**A、B、C 這一輪一定要做完**；C′、D、E、F 依序推進，做不完的把
收據與反組譯進版控、issue 本文補「卡在」。

### 每一項的完成定義

- 收據腳本化進 `tools/fleetcmd_receipts.sh`（同一條檢查點鏈；要 poke 的局面把
  poke 寫進那一行）；批次名 `fcmd-*`，探索用 `fcx-*`。
- 規格 051 各項一段 typed input → 艦隊記錄；狀態改 CONFORMED 才關 issue；
  新畫面各自一份 spec。
- 測試整張 640×480 比（`matches`），GUI 每個新畫面至少一例進 `tools/gui_check.sh`。
- `docs/re/313` 逐次紀錄（不改 312，只加指標）；`docs/internals/58` 更新事實；
  `docs/playtest/021` 追加表列（雜湊全寫）；`CONTEXT.md`／`WORKLOG.md`。
- 關 issue 前把 verify 換成 `go_test`，重抓快照、`verify` 0 條斷言不成立、
  `render -w`／`render-matrix -w`。commit 訊息尾巴 `Co-Authored-By: Claude …`，
  不放 session 連結；`git config user.email` 要是 `wicanr2@gmail.com`。

### 已知的坑（上一輪踩過的，別再踩）

- 收據旗標是滑鼠座標，畫面 y ＝ 滑鼠 y ＋ 40；`-save-state` 要放在 `-steps` 的前一步。
- `-load-state` 之後點擊沒反應是 dosgolem `b5485eb` 之前的坑，已修；用舊狀態檔要重存。
- 左下角與右側面板是「寫進去那一刻的文字」＋「游標底下那一格變了才重寫」
  （規格 001 規則 4–7、`Game.hoverProbe`／`fleetCmd.probe`）：新畫面對拍時
  列 16–17 有差先想這件事，不要去改摘要的算法。
- 底線裝飾畫在格子最後一列（色 15）；字元裝飾 `0x80` 那個位元還沒解。
- 對拍收斂三步：`tools/scripts/celldiff.py`（差異按格列）→ `planerows.py`
  （印原版字元平面那幾格）→ `crop_png.py`（兩張並排看）。
- `workplace/ida/*.i64` 連續 `idat` 會壞（`error code 4`）：一次多目標；壞了刪掉
  用 `tools/ida.sh raw idat -A -B /work/<檔>.com` 重建。
- Docker：只走 `tools/*.sh` 包裝器；一次性容器 `--rm`；收尾 `docker ps -a | grep logh3`
  與 `find . -user root`；**禁止任何 prune／rmi**。

### 不在這一輪

#43（分頁與捲動）、#44（星系情報視窗）、#16、音樂三項（交戰演出＋戰鬥曲、
自產曲庫管線、MIDI 音色對 MPU-401）——排在 #41／#42 之後，除非 C 發現回合
處理需要先動到它們。不改 dosgolem，除非收據需要；改了就整條 `fcmd-*` 重跑。
