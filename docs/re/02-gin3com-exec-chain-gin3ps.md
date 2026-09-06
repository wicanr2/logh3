# RE 02 — GIN3.COM 原生路徑跑通 EXEC 鏈進 GIN3PS.EXE（dosgolem）

日期：2026-09-06
執行器：dosgolem 分支 `logh3-com-support`
（`916bb21` spec 007 ＋ `632e0d4` 實作；前一棒 `c868cd6` LoadCOM）
輸入：原版未修改 `org_game/extracted/GIN3.COM`（SHA-256
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`，唯讀）
規格：dosgolem `docs/spec/007-exec-ems.md`（READY 部分已實作）
前篇：`01-first-contact-gin3com.md`

推論標籤：confirmed／強證據／假說／未知。

---

## 1. 結論（confirmed）

**launcher 的原生路徑（無 patch）在 dosgolem 裡一路 EXEC 進了本體
`GIN3PS.EXE`。** 原版 GIN3.COM 直接跑：

```
DOSGOLEM_ORIG=~/cht/logh3/org_game/extracted \
  tools/go.sh run ./cmd/probe -exe /orig/GIN3.COM -root /orig \
  -steps 20000000 -tick 20000 -log-calls -trace 30
```

- EMS 偵測**通過**：`EMMXXXX0` 裝置開啟成功、`int 67h AH=40h/42h` 各 1 次
  （回 8/8 頁，恰好過 launcher 的 `cmp bx,8`）。
- 滑鼠偵測通過，進 EXEC 鏈：`int 21h AX=4B00h` → `open.exe`。

不點滑鼠的話，OPEN.EXE 的開頭動畫**會一直跑**（確實在畫，不是死迴圈，
見 §3），所以「進本體」的收據用了一次點擊跳過開頭：

```
  -steps 200000000 -tick 20000 -click-x 320 -click-y 200 -click-at 5000000
```

結果（23,289,631 道指令，整鏈離開碼 0）：

```
開過的檔：EMMXXXX0 open.exe BU16IBM.RGB gin3ps.exe EMMXXXX0
int 21h AH=4B ×2、AH=4C ×3、AH=4D ×2
int 67h AH=40 ×2、41 ×1、42 ×2、43 ×1
```

- EXEC 鏈 confirmed：`GIN3.COM → OPEN.EXE → GIN3PS.EXE`。
  OPEN.EXE 結束（點擊跳過）→ launcher 取回碼（AH=4D）→ EXEC
  `gin3ps.exe` → GIN3PS.EXE 跑了一段 → exit 0 → launcher 見回碼 0，
  跳過 `ending.exe` → 自己 exit 0。
- **ENDING.EXE 這條路沒跑到**（只在 GIN3PS 回碼 ≠0 時才 EXEC，
  尾巴 `9 1`／`9 2` 二選一；confirmed by RE 01 反組譯）。

## 2. 本體 GIN3PS.EXE 的停點（confirmed 停點／假說 原因）

GIN3PS.EXE 進場後自己再做一次 EMS 偵測（又開了一次 `EMMXXXX0`），
然後呼叫了 launcher 沒用到的兩個 EMS 功能：

| 呼叫 | 語意 | dosgolem 現況 |
|---|---|---|
| `int 67h AH=41h` | 取頁框段位址 | 未實作（記錄，AH=84h）|
| `int 67h AH=43h` | 配置 EMS 頁 | 未實作（記錄，AH=84h）|

之後很快就 exit 0（整鏈 23.3M 道指令內結束；OPEN.EXE 點擊後的清理
也佔一部分）。**假說**：EMS 配置失敗 → 判定記憶體環境不足 → 安靜離開
（沒有任何主控台輸出）。支持證據：呼叫時序（41/43 是最後的 EMS 呼叫）、
launcher 的檢查表把 EMS 列為硬性需求（RE 01 §3）。

補充：**單獨跑 GIN3PS.EXE**（不經 launcher）走的是另一條路——
印 `TINYPROG says, "Bad program file!"`、exit 255。
TINYPROG 是 overlay/壓縮載入器殼（日系 DOS 常見）；
「Bad program file」在沒有 launcher 命令列尾巴（`9 1`／`-MSDOS` 參數區）
時出現，所以**單跑不是代表性路徑**，對拍一律從 GIN3.COM 進。
（confirmed：直接 probe 收據；殼的細節是假說。）

## 3. OPEN.EXE 開頭動畫：會跑、會畫，只是慢（強證據）

- 視訊模式切到 **10h（EGA 640×350）**；寫過的埠：
  3C4/3C5（sequencer）、3CE/3CF（graphics controller）、3D4/3D5（CRTC）、
  388/389（**AdLib FM 音源**）、040/043（PIT）、020（PIC）、
  216–266（每隔 0x10 一組，**未知**，疑似第二片 PIC 或音效相關）。
- 開過的檔：`BU16IBM.RGB`（色盤）、`IBMFLS16.350`、`IBM16.350`
  （16 點陣字型）、`MUSIC.INI`、`TITSPT.RGB`、`OPEN_M.MDS`（音樂）。
- A0000 非零像素隨指令數成長（50M 步 3,534 → 400M 步 35,133），
  幀解碼熱點在 OPEN.EXE 檔案偏移 0x13E4 的 XOR 繪製迴圈
  （`add di,0x50 / cmp di,0x46f0 / loop`，80 bytes/列）。
- 等待迴圈（檔案偏移 0x8EF–0x911）：等三個條件之一——
  旗標 byte 清 0（計時器 ISR 清）、計數到 0x176（374 tick）、
  或滑鼠按下（`int 33h AX=5` 的 press 計數）。**點擊能跳過是
  confirmed**（§1 的收據）；前兩個計時器路徑有送達（machine 的 IRQ0
  會跑程式自裝的 int 08h 或轉 int 1Ch），但預設 165k 指令/tick 下
  374 tick ≈ 62M 指令才逾時，純等非常慢。
- ⚠ 模式 10h 是 EGA planar：probe 的 `A0000 非零像素`（mode 13h 線性
  語意）只是粗略指標，**planar 內容的逐點驗證還沒有工具**。

## 4. 本批缺口清單（下一個工作項的輸入）

| # | 缺口 | 證據 | 等級 | 影響 |
|---|---|---|---|---|
| 1 | **int 67h AH=41h**（頁框段）/ **AH=43h**（配置頁）——可能還要 AH=44h 映射、45h 釋放 | 鏈內軌跡各 ×1 | confirmed（被呼叫） | GIN3PS.EXE 疑似因此早退（§2） |
| 2 | int 21h AH=43h AL=00（取檔案屬性）| 鏈內 ×5、×1 | confirmed | 目前 CF 清 0 但 CX 未填＝垃圾；暫未觀察到阻斷 |
| 3 | int 21h AH=0Dh（flush disk buffers）| 鏈內 ×3 | confirmed | 無害（收下即可），尚未實作 |
| 4 | EGA mode 10h planar 的 framebuffer 觀測（probe 只有 mode 13h 與 CGA dump） | §3 | confirmed | 開頭動畫無法逐點對拍 |
| 5 | AdLib（388/389）與 216–266 埠語意 | 埠寫入清單 | 未知 | 音樂／音效；可先只記不實作 |
| 6 | TINYPROG 殼的格式與「Bad program file」判定條件 | §2 | 假說 | 若要單獨對拍 GIN3PS 才需要 |

## 5. dosgolem 側的改動（分支 `logh3-com-support`，已 push）

- `916bb21`：`docs/spec/007-exec-ems.md`（SDD：先有 spec；
  每條行為附 GIN3.COM 反組譯 bytes 或 probe 軌跡出處）。
- `632e0d4`：實作 READY 部分＋單元測試。
  - `internal/machine/loader.go`：`LoadEXEAt`／`MZImageParags`，
    `initPSP` 參數化；`LoadEXE`／`LoadCOM` 行為不變。
  - `internal/dos/exec.go`（新）：EXEC、childExit、`AH=4Dh`、
    `int 67h`（40h/42h）、EMMXXXX0 判別。
  - `internal/dos/{dos.go,int21.go,files.go}`：exec 堆疊與 curPSP 欄位、
    `exit` 深度感知、`setBlock` 以 curPSP 為準、裝置讀寫防禦。
- 分層：全部在通用層，**沒有任何 logh3 專屬位址**進 dosgolem
  （spec 006 判準：換一支 binary 這些都還成立）。
- 測試：`tools/go.sh test ./...` 全綠（cpu 套件 cached ok；
  SingleStepTests 語料 761 MB 不在本機，CPU 核心本次未動）。
- 未 commit 的既有 untracked：`cmd/scenetime/`、`scenetime`（非本批產物，
  留給主人）。

## 6. 下一步建議

1. **EMS 補完**：AH=41h（頁框段，選 0xE000 一類慣例值）、AH=43h
   （配置：回 handle＋成功）、以及大概率立刻用到的 AH=44h（映射）
   與 45h（釋放）。先量測 GIN3PS 叫哪些再寫 spec（SDD），
   目標是讓本體過記憶體檢查進到第一個畫面。
2. probe 加 EGA planar（mode 10h）dump：照 3CE/3CF 與 3C4/3C5 的
   目前設定把四個 plane 拼回 640×350×16 色，否則開頭動畫無法對拍。
3. `AH=43h` 屬性查詢補上（便宜，防「安靜地錯」）。
