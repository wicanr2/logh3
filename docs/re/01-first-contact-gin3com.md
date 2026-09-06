# RE 01 — GIN3.COM 第一次接觸（dosgolem 無頭執行器）

日期：2026-09-06
執行器：dosgolem（工作樹含本次最小改動，未 commit；見 §6）
探針：`cmd/probe`（通用探針，吃執行檔路徑）
輸入：`org_game/extracted/GIN3.COM`，SHA-256
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`（唯讀，未修改）

推論標籤：confirmed／強證據／假說／未知。

---

## 1. 檔案基本資訊（confirmed）

| 項目 | 值 |
|---|---|
| 大小 | 622 bytes |
| 格式 | **.COM**（無檔頭；開頭 bytes `8C 06 21 03` 不是 `MZ`）|
| 載入方式 | PSP + 映像載到 PSP:0100，無重定位；CS=DS=ES=SS=PSP 段，IP=0x100 |
| 角色 | **launcher／loader**，不是遊戲本體（622 bytes 只做環境檢查再 EXEC 別人）|

靜態可見字串（檔案偏移 0x152 起）：
`This program can not be executed in US mode.`、`Memory Not Enough!`、
`EMS Driver not found ,or Ems pages not enough !`、`Mouse Drive not Found!`、
`open.exe`、`gin3ps.exe`、`ending.exe`、`9 1\r`、`9 2\r`、`EMMXXXX0`、`-MSDOS`。

## 2. dosgolem 端的第一個發現：loader 只吃 MZ（confirmed）

`internal/machine/loader.go` 只有 `LoadEXE`（MZ 檔頭 + 重定位），
.COM 會在 `parseMZ` 被「不是 MZ 執行檔」擋掉。
本次加了最小 `LoadCOM`（§6），probe 依檔頭 magic 自動分派，CPU／DOS 層未動。

## 3. 原版 GIN3.COM 實跑結果（confirmed）

命令：

```
cd ~/cht/dosgolem
DOSGOLEM_ORIG=~/cht/logh3/org_game/extracted \
  tools/go.sh run ./cmd/probe -exe /orig/GIN3.COM -root /orig \
  -steps 5000000 -trace 40 -log-calls
```

- **只跑了 35 道指令**就主動結束，離開碼 3。
- 停點：CS:IP = 0100:014C（`int 21h AH=4Ch`），主控台輸出
  `EMS Driver not found ,or Ems pages not enough !`。
- 路徑（軌跡 + 反組譯對照，全部 confirmed）：
  1. 0100:0100 進入；設堆疊 SP=0x236E；
  2. `int 21h AH=4Ah`（BX=0x237 段）縮自己的記憶體區塊——dosgolem 回成功，
     AX 保持 0x4A00 ≠ 0 → 走 `jnz 0126`；
  3. `cmp cs,0x1600; jbe`——dosgolem 的 PSPSeg=0x0100 ≤ 0x1600，成立，
     進 EMS 檢查（**假說**：這一格是「US mode」判別的一部分，語意未定，
     見 §5 未知項）；
  4. EMS 檢查第一步不是 int 67h，而是 **`int 21h AH=3Dh` 開字元裝置
     `EMMXXXX0`**（EMM 驅動的 device header 名字）。dosgolem 的
     `resolve()` 只對映真實檔案，找不到 → AX=2、CF=1 → 回傳 FFFF →
     判定無 EMS → 印錯誤訊息 → exit(3)。

服務呼叫統計：`AH=4A ×1、AH=3D ×1、AH=09 ×1、AH=4C ×1`。
找不到的檔：`EMMXXXX0`。沒有未支援的 CPU 指令，沒有碰視訊記憶體
（A0000／B8000 全 0，顯示模式維持 03h 文字模式——launcher 不切模式）。

## 4. What-if：跳過 EMS 檢查後的完整流程（confirmed，patch 版）

為了對映 launcher 的完整行為，把 0x39 的 `call 00B0`（EMS 檢查）換成
`mov ax,1`（假裝有 EMS），另存為
`workplace/re01/GIN3_skip_ems.COM`（原檔未動，hash 已驗）。
結果：**121 道指令，離開碼 0，主控台無輸出**。

流程（軌跡 confirmed，檔名來自 DS:DX 與參數區）：

1. EMS 檢查（被 patch 跳過）。
2. 滑鼠檢查：`int 21h AH=35h AL=33h` 取向量 → `int 33h AX=0000` reset ×2。
   dosgolem 有 int 33h，**通過**（confirmed：int 33h AH=00 ×2）。
3. **連續 EXEC 三支程式**（`int 21h AX=4B00`，依序）：
   - `OPEN.EXE`（開頭動畫？——內容未知）
   - `GIN3PS.EXE`（189,536 bytes，MZ，**遊戲本體**；命令列尾巴是
     `9 1\r` 或 `9 2\r` 二選一，靜態可見於 0x317／0x31C，
     選擇條件是前面某個檢查的回傳值——**未知**）
   - `ENDING.EXE`（64,677 bytes，MZ）
4. 每次 EXEC 後呼叫 `AH=4Dh` 取回傳碼。
5. `AH=4Ch` exit(0)。

**注意：這 3 次 EXEC 全是假的**——dosgolem 沒實作 `AX=4B00`，
未實作服務表如實記了 `int 21h AH=4B AL=00 ×3`、`AH=4D ×3`。
CF 被清掉讓 launcher 以為成功，實際上 OPEN/GIN3PS/ENDING
一個 byte 都沒被載入執行。這是「宣告成功會說謊」的又一例，
probe 的 UnimplementedReport 有抓到，不算安靜地錯。

## 5. 缺失清單（coverage gap）

| # | 缺口 | 位置 | 等級 | 影響 |
|---|---|---|---|---|
| 1 | **字元裝置開啟**（`AH=3D` 開 `EMMXXXX0`）| GIN3.COM mem 0x1DF（檔案 0xDF `CD 21`）| confirmed | EMS 偵測第一關就失敗，真 DOS + EMM386 會成功 |
| 2 | **int 67h EMS**（AH=40h status、AH=42h 取可用頁數，需 ≥8 頁）| mem 0x1BA／0x1C2 `CD 67` | confirmed（靜態；本次未走到）| 就算裝置開啟成功，int 67h 未實作會回 AH 不變 → 仍判定失敗 |
| 3 | **int 21h AX=4B00h EXEC** | mem 0x246 `CD 21` | confirmed | launcher 的核心功能；不實作就永遠跑不到 GIN3PS.EXE 本體 |
| 4 | int 21h AH=4Dh 取子程式回傳碼 | mem 0x24D | confirmed | 與 EXEC 配套 |
| 5 | （潛在）`int 21h AH=63h` DBCS 前導位元組表、`int 2Fh AX=4F01h` | mem 0x188 `CD 21`／0x19B `CD 2F`（常式 0x184 起）| 假說 | 本次 patch 路徑沒走到；疑似日文 DOS 模式判別，與 `9 1`/`9 2` 尾巴或「US mode」訊息有關 |
| CPU 指令 | **無缺口**：35／121 道全程 8086 基本指令，無未支援 opcode | — | confirmed | — |
| 顯示模式 | launcher 不切換（維持 03h 文字模式）；本體用什麼模式待 GIN3PS.EXE 進來才知道 | — | confirmed（對 launcher）| — |

## 6. 對 dosgolem 的改動（未 commit，由主人決定）

兩處，都是通用層的最小新增，不影響既有 MZ 路徑
（`tools/go.sh build ./...` 與 `test ./internal/machine ./internal/dos` 全綠）：

1. `internal/machine/loader.go`：新增 `LoadCOM(data)`——映像放 PSP+100h
   （與 LoadSeg 同位址，MCB／FreeSeg 計算沿用），CS=DS=ES=SS=PSPSeg、
   IP=0x100、SP=0xFFFE 並壓 0 word。`LoadEXE` 一行未改。
2. `cmd/probe/main.go`：`-exe` 依檔頭 `MZ` magic 自動分派
   `LoadEXE`／`LoadCOM`（不看副檔名），並更新 flag 說明文字。

測試語料（CPU SingleStepTests 761 MB）不在本機，未跑 CPU 語料測試；
CPU 核心本次未動。

## 7. 下一步建議（依優先序）

1. **實作 `int 21h AX=4B00h` EXEC（load-and-execute）**：這是通往
   GIN3PS.EXE 本體的唯一門。需要先定案子行程的 PSP／MCB／記憶體配置
   語意（`docs/spec/004` 加一節，先 DRAFT）。`AH=4Dh` 一併。
2. **EMS 最小實作**：`AH=3D` 認得 `EMMXXXX0` 字元裝置 + `int 67h`
   AH=40/42（回「在、頁夠」即可，page frame 映射等 GIN3PS.EXE 真的
   用到再說——先記 unimplemented）。
3. 之後對 `GIN3PS.EXE`（MZ, 189,536 bytes，未打包——檔頭 `4D 5A 60 00`，
   無 LZEXE/PKLITE 特徵）單獨跑一次 probe，看它切什麼顯示模式、
   開哪些 `.DT8/.DT9/.CMP/.SPR/.MDS`。
4. 釐清常式 0x184（AH=63h／int 2Fh）的語意與 `9 1`/`9 2` 參數的意義
   （懷疑是 DOS/V 日文模式 vs US 模式選擇——**假說**）。
