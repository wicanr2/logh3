# 下一輪 goal — 戰場回合結算的其餘七步（#79）＋ 六組並行

日期：2026-09-18。第八輪 goal（`docs/goal/2026-09-18-battle-settlement-round-8.md`）的結果：**五組全部關掉**。

- **A #77**：`docs/re/278` 用 `f7523f9` 重錄重驗，規格 001 規則 4–7 對齊。
- **B #57**：戰場那一條全滅與戰鬥中的士氣報告接上（`181B:000A` 八步的第 ⑧ 步子步 8-1），規格 062 開出來。
- **C #55 ＋ #71**：艦艇池的欄位表進 `docs/internals/64`；**那兩格存量解出來是飛彈與艦載機**
  （`2703:14CF` ＝ 艦種表 `+0x10`／`+0x14` × 艘數，`docs/re/345`），新編艦隊照原版填。
  自家星域補充 `1AAB:0FAF` 接上（規格 063、`docs/re/346`），「士氣值已經恢復」走得到了。
- **D #74 ＋ #75**：兩位數艦隊編號的五個地方、開戰前的「發現敵方艦隊」報告。
- **E #63／#64／#65**：情報收集三個分頁、徵稅／生產／預算計劃、艦隊編號清單與結束確認。
  **#63 的門檻表其實是距離表**（`docs/re/347`）——`docs/re/185`／`241`／`298` 早就解完，卡住的是沒回頭查。

收尾狀態：`main` 已推到 `4831038`，dosgolem 副本仍是 `f7523f9`（本輪沒動執行器，收據不必重跑）。
`go test ./internal/...` 全過、`tools/gui_check.sh` 40 例逐位元組相同、`tools/feature_check.sh` 沒有 crash
且 `knownDead` 是空的、worklist verify 0 條斷言不成立。open 的 worklist issue 剩 13 條。

**本輪新開的未解項**（都寫在 `docs/re/345` 末節與規格 063）：

- `1930:1572` 沒讀。它排在 `1AAB:0FAF` 之前、屬於同一格的處理序列，多半就是**把星域層的飛彈／艦載機
  整批補給進艦隊**那一支——收據上兩格存量各多出來的 108600 正好等於星域 0 的 `+0x68`／`+0x70`。

**使用者裁決（2026-09-18）**：第九輪放七項，分六組（下面 A–F）。

下面這段可以直接當下一個 session 的開場提示詞。

---

## 提示詞

你在 `~/cht/logh3`（銀河英雄傳說III SP 的 DOS remake）。先照 `CLAUDE.md` 的開工順序：
`tools/issues_snapshot.sh` → `tools/go.sh run ./cmd/worklist verify` → 讀 `CONTEXT.md` 現況一覽 → `git status --short`
（logh3 與 `~/cht/dosgolem-logh3` 兩邊，`git config user.email` 都要是 `wicanr2@gmail.com`）。
再讀直接相關的：`docs/spec/062`（八步，DRAFT）、`docs/re/339` §六（十二條落差）、
`workplace/re-notes/round7-57-static.md` §三、`docs/spec/063` ＋ `docs/re/346`（補充，第 ① 步的鄰居）、
`docs/re/345`（兩格存量與那塊 108600）、`internal/game/battle.go`、`internal/game/battlefinal.go`、
`internal/rules/regionresupply.go`。

**本輪目標：把戰場一個回合的結算做成原版的八步模型，並清掉五個玩家撞得到或會過期的缺口。**
先前的裁決照舊：每張新畫面維持 same-state 契約（RE → spec → 收據 → READY → 實作 → 整張逐位元組），
關 issue ＋ commit ＋ push 不必再問。**#16 不要動**。subagent 不直接改 GitHub issue。

**主機上唯一允許的算術是 `printf '%x' $((...))`**——分析、轉檔、反組譯、跑測試一律走
`tools/py.sh`／`tools/go.sh`／`tools/ida.sh`／`tools/oracle.sh`。連 `python3 -c` 做一次十六進位換算都不行
（第八輪四個 RE subagent 裡有三個踩到這一條）。

**要寫下「這個還沒解／卡住了」之前，先 grep 自己的 `docs/`。** 第八輪的 #63 就是握著三個足以搜到
`docs/re/185` 的特徵、仍然寫下了 unknown（`docs/re/347` 的教訓那一節）。

### 交付順序

**A. #78 存檔會復活的那個 bug：先做，範圍最小、影響最久**

戰略階段全滅的艦隊沒有寫回存檔表，存檔再讀回來那一支會復活。第八輪替戰場那一條加了
`rules.(*Slot).AnnihilateFleet`，戰略這一條要走同一個出口。判準：全滅 → 存檔 → 讀檔，那一支不在表上，
艦艇池那幾筆也作廢（`+0x01` ＝ FF）；有 go_test 釘住。

**B. #79 八步（主軸）**

規格 062 目前只有子步 8-1 是 CONFORMED。其餘七步與 8-2–8-5 照 `docs/re/339` §六的十二條落差逐條處理。
第 ① 步（補給 `1930:019D`、修理 `1AAB:006D`、行星防衛力回復 `181B:0083`、`1261:000B(1)` 印士氣）現在
**沒有東西擋著了**——艦艇池模型（#55）已經接上。

順手把 `1930:1572` 讀掉：它與第 ① 步的補給是同一族，而且解掉之後規格 063 的「已知偏離」第一條就能清掉
（那塊 108600）。`internal/rules.TestResupplyNumbersMatchOracle` 已經把「多出來的量 ＝ 星域的飛彈存量」
這個假設釘住，讀出來對不上會立刻壞掉。

**C. #62 參謀任命**

`090A:0248`。情報收集與徵稅計劃兩個畫面的「參謀」鈕都是它的入口，第八輪剛把那兩個畫面補完，接著做順手。
`docs/re/342` 已經記到「選中且通過 `0D92:0422`／`0D92:04E3` 的合格檢查後改 `ds:8F4D[座位]`」，
合格條件本身還沒反查。

**D. #80 ＋ #53 兩個 parity 缺口**

- **#80**：徵稅／預算計劃四顆箭頭鈕剛點過時差 17–21 個像素（原版寫色 8、remake 寫色 9）。
  目前用 `planArrowSkip` 排除、在 `docs/spec/003` 與 `docs/re/343` 明標**未解、不是核准的差異**。
  起點是畫箭頭的那一支（熱區清單 `ds:14A8` 的 idx 0／1／4／5，kind 6）與逐幀收據（`-watch-screen`）。
- **#53**：戰場層打行星與位元組三處不符（`103D` 尾段、`0593` 艦種 5、Atan256 查表）。

**E. #60 ▲▼ 按住連捲**

一覽表、右下角清單、提督清單、艦隊組成四處共用。第八輪查過箭頭鈕**按住會重複**
（`-click-hold` 拉長就連動，`docs/re/343`），但重複的節奏還沒量。

**F. #66 設定畫面的「ＭＩＤＩ Ｉ／Ｆ」與 BUG!；#59 ＋ #56 補收據**

- **#66**：設定畫面裡兩個點了沒事的項目。
- **#59**：一覽表分頁沒有收據的幾種情形（同盟座位、戰鬥中、占領中、士氣低落）。
- **#56**：戰略打行星沒有收據的三種局面（多支攻方、同回合打過仗、電腦座位）。
  這兩條是**純補證據、不改規則**——如果補出來與現行實作不符，那才是新發現。

### 收尾

`tools/gui_check.sh`（目前 40 例）、`tools/feature_check.sh`（`knownDead` 要維持空的）、
`go test ./internal/...`、worklist verify 0 條斷言不成立；
`CONTEXT.md` 走 `render -w`、`VERIFICATION-MATRIX.md` 走 `render-matrix -w`（產生前先抓 issue 快照）；
`WORKLOG.md` 記這一輪的流水帳與踩到的坑。動到 dosgolem 就整組收據重跑。
