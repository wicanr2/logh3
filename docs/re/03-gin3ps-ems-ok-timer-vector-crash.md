# RE 03 — GIN3PS.EXE 本體：EMS 過關、資產載入、mode 12h，停在計時器中斷向量指向空模組

日期：2026-09-06
執行器：dosgolem-logh3 分支 `logh3-com-support`（本批 commit 見 §6）
輸入：原版未修改 `org_game/extracted/GIN3.COM`（SHA-256
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`，唯讀）
規格：dosgolem-logh3 `docs/spec/008-ems-paging.md`
前篇：`02-gin3com-exec-chain-gin3ps.md`

推論標籤：confirmed／強證據／假說／未知。

> §2 把停點解讀成「模組載入器 / ISR 表」。那是誤讀——實際是中文文字視窗的
> 儲存格緩衝區，根因在 EXEC 之後配置游標沒退回。以 `05` 與 `CONTEXT.md` 為準。

---

## 1. 結果一覽（confirmed，全部來自 probe 軌跡）

**EMS 過關，本體大幅推進。** 跑法（上一輪證實點擊可跳過開頭動畫）：

```
cd ~/cht/dosgolem-logh3
DOSGOLEM_ORIG=~/cht/logh3/org_game/extracted \
  tools/go.sh run ./cmd/probe -exe /orig/GIN3.COM -root /orig \
  -steps 200000000 -tick 20000 -click-x 320 -click-y 200 -click-at 5000000 \
  -log-calls -trace 30 -watch-video
```

- `int 67h AH=41h`（頁框段 → 0xE000）×1、`AH=43h`（配置）×1 成功、
  **`AH=44h`（映射）×152 次全部成功**——上一輪的「EMS 配置失敗早退」
  假說 confirmed 為真停點，本批修掉。
- 本體載了 **19 個檔**（依序）：
  `EMMXXXX0 open.exe BU16IBM.RGB gin3ps.exe EMMXXXX0 MUSIC.INI
  GIN3SP.VOC GRAPH.IMG COMANI.BMG FRAME.BLK SPRITE.SPR BLOCK_3.BLK
  3DBLOCK.SPR COM_FLAG.SPR PLADAT.DAT PLADAT.DAT ADMIRAL.DTX
  STELLAR.DTX ATTR.DAT`
  ——音樂設定、語音、圖庫、sprite、地圖 block、勳章/提督/星域資料全開了，
  且每次 open 都有配對的 read/close（AH=3F ×26、3E ×17、42 ×10）。
- 視訊模式 **12h（VGA 640×480×16）**；`int 10h AH=10h` 調色盤 ×37；
  A0000 有約 38 KB 非零內容（planar raw，未解碼——見 §4）。
- 滑鼠輪詢回報正常（26 次，位置 (320,200) ＝ 注入的點擊）。

## 2. 新停點：IRQ0 → 空模組（confirmed 機制／假說 根因）

停在第 25,760,001 道指令，護欄攔下：
`CS:IP ＝ BED4:0099（線性 BEDD9，全 0 區域）`。

證據鏈（全部 confirmed by probe 軌跡／`-find`／`-watch`）：

1. 程式**直接寫 IVT**（不走 AH=25h）把 int 08h 向量改成 `BED4:0097`：
   - step 25,757,317：`6B93:00B2` 寫 offset；step 25,757,731：
     `6B93:00A4` 寫 segment ＝ **0x2020**（該處機器碼是
     `26 C7 04 20 20` ＝ `mov word es:[si],0x2020`，常數）；
   - step 25,757,761：`6C50:02F9` 把 segment 改寫成 **0xBED4**
     （0xBED4 ＝ 0x2020 ＋ 0x9EB4）。
2. 下一個 IRQ0 送進來 → `CPU.Interrupt(0x08)` 跳 `BED4:0097` → 全 0 →
   護欄攔在 `BED4:0099`。
3. 0x20200 與 0xBED40 兩處在崩潰時**都是全 0**——模組從來沒被載進去。
4. **不是計時器時序問題**：`-tick 20000` 與 `-tick 100000` 崩在同一個
   位址、幾乎同一個指令數（25,760,001 vs 25,800,001）——確定性、
   資料驅動。
5. 直接寫 IVT 的程式碼位在 6B93／6C50（約 430 KB 處的**已配置記憶體**，
   是 TINYPROG 解出來的模組），上下文是驅動/模組管理器
   （記錄結構欄位 +0B/+0C/+0E/+10、乘法算大小、連結串列）。

**假說**：這是音樂/音效驅動（`MUSIC.INI`、`GIN3SP.VOC`、388/389 AdLib、
216–266 埠）的模組載入器：某個模組應該被載到 0xBED4 一帶的高段
（0x9EB4 段 ≈ 636 KB 的位移，疑似由「記憶體上限/自由空間」算出），
但載入沒有發生或位址算錯，ISR 向量於是指向空區域。
是哪一個查詢/計算在我們這裡得到不同答案——**未知**，
是下一輪的第一個問題（§5 #1）。

排除過的假說（有反證）：
- ~~EMS 配置失敗早退~~（上一輪的假說）——本批 EMS 全過，程式繼續跑了
  200 多萬道指令才停在別處。
- ~~中斷鏈 stack imbalance（far call 到 iret stub）~~——崩潰位址與
  tick 頻率無關，且正規的 `pushf+call far` 鏈對 iret stub 是平衡的。

## 3. 過程中新增的觀測工具（通用，已進 dosgolem-logh3）

- probe `-watch <lo>-<hi>`：監看記憶體寫入，印步數與 CS:IP
  （本案用它定位 IVT[8] 是誰寫的）。
- probe `-find <hex>`：全記憶體搜 byte 串（本案用它定位 BED4 出處 ＝ IVT）。
- probe `-dump-linear <檔>`：A0000 64 KB raw 傾倒（planar 未解碼）。
- dos 層 `VecSets`：`AH=25h` 設向量全程記錄並進 probe 報告。
- int 21h AH=44h 的 AL≠00 子功能從「安靜成功」改為記 unimplemented
  （鏈內 6 次全是 AL=00，無新缺口）。

## 4. EGA/VGA planar 的判斷（confirmed）

機器記憶體是平坦 1 MB，寫 A0000 直接落 Mem，**沒有 map mask／latch
語意**——Mem 裡是「四個 plane 混在一起」的超集，拼不回畫面。
mode 10h/12h 的逐點對拍要先做 planar 寫路徑（渲染語意，需另開 spec），
不是工具面的活。本批證據檔：
`workplace/re03/a0000-at-crash.bin`（65,536 bytes，崩潰點的 A0000 raw，
SHA-256 `d51190a6e2a944c2edb560ac6ef444ce0f715e53a41ef092b833a56a5adc2631`，
由 §1 命令列加 `-dump-linear` 取得；**不能**當畫面對拍用）。

## 5. 下一批缺口

| # | 缺口 | 等級 | 說明 |
|---|---|---|---|
| 1 | **模組載入器的高段位址從何而來**（0x2020 → 0xBED4 的 0x9EB4 段位移）| 假說 | 先反組譯 6B93/6C50 的模組管理器，找出它讀哪個記憶體大小/配置答案；八成是某個 DOS 記憶體查詢在我們這裡回答不同 |
| 2 | int 33h AX=000Ch（滑鼠事件 handler）| confirmed（鏈內 ×1，未實作）| 主選單之後可能改用事件驅動；目前輪詢路徑已通 |
| 3 | VGA mode 12h planar 寫路徑＋dump | confirmed（缺）| 到主畫面後的對拍前提，另開 spec |
| 4 | int 21h AH=0Dh（flush）| confirmed（×3）| 無害，補上收下即可 |
| 5 | AdLib（388/389）與 216–266 埠 | 未知 | 音樂相關；先記錄 |

## 6. dosgolem-logh3 改動（分支 `logh3-com-support`）

- `docs/spec/008-ems-paging.md`：新規格（先 DRAFT 取證，44h 經軌跡
  ×152 升格 READY；45h 已實作但 0 次呼叫，保留不列驗收）。
- `internal/dos/ems.go`（新）：頁框 0xE000、4 實體頁、8 頁池、
  複製式分頁（映射複製進框、換映射/釋放寫回）。
- `internal/dos/exec.go`：int67 移到 ems.go（41h 新增）。
- `internal/dos/int21.go`：AH=43h 檔案屬性（AL=00）、AH=44h 誠實記帳、
  AH=25h 記 VecSets。
- `internal/dos/ems_test.go`（新）：41h/43h 三路、44h 映射讀寫與
  換頁寫回、45h 釋放、檔案屬性兩路。
- `cmd/probe/main.go`：`-watch`／`-find`／`-dump-linear`。
- 驗證：`tools/go.sh build ./...`、`test ./...` 全綠
  （cpu 為既有 cached；CPU 核心未動）。
- 整合驗收：§1 的原版 GIN3.COM 鏈（未修改、唯讀掛載）。
