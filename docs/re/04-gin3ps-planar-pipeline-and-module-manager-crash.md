# RE 04 — GIN3PS.EXE：planar 管線打通、第一張 framebuffer 傾倒；停點根因定位到模組管理器（未知項收斂）

日期：2026-09-06
執行器：dosgolem-logh3 分支 `logh3-com-support`（本批 commit 見 §7）
輸入：原版未修改 `org_game/extracted/GIN3.COM`（唯讀，SHA-256 見 re/01）
規格：dosgolem-logh3 `docs/spec/009-planar-vga.md`
前篇：`03-gin3ps-ems-ok-timer-vector-crash.md`

推論標籤：confirmed／強證據／假說／未知。

---

## 1. 本體推進到哪（confirmed）

鏈路與 re/03 相同（`GIN3.COM → OPEN.EXE → GIN3PS.EXE`，EMS 全過、
19 檔載入、mode 12h）。本批新證據：

- **崩潰前遊戲正在畫畫**：trace 顯示 6C96:030C 起一段寫 ES=A000 的
  像素繪製迴圈（step ~25.754M），planar 色號陣列有 103,452 個非零像素
  （色號 15，滿版填充＋右緣一個小矩形）。
- **滑鼠輪詢在 step ~25.748M 開始**（26 次，位置正確）——
  判定已到達互動畫面的邊緣（強證據）。
- 模式切換記錄（新增觀測）：`#511038→10h`（OPEN.EXE）、
  `#5000267→03h`（OPEN 結束回文字）、`#23289863→12h`（GIN3PS）。

**第一張 framebuffer 傾倒**（planar 解碼＋DAC 調色盤）：
`workplace/re04/screen-25759000.png`（640×480）＋ `.bin` 色號陣列。
畫面全黑是**遊戲自己的調色盤還停在暗值**（DAC[15]=(8,8,8)，淡入沒有
進展）——不是管線沒東西；色號陣列證明圖形已畫進 plane。
幾何驗證用假色盤圖 `screen-25759000-fakepal.png`（**僅供目視，
不是對拍證據**）。取得方法：§7 命令列。

## 2. planar 管線（spec 009，READY 部分已實作＋測試）

- machine：4 plane × 64 KB VRAM、sequencer/GC 暫存器檔、latch、
  write mode 0/1/2、read mode 0（read 時裝 latch）、map mask、
  位元遮罩、旋轉＋AND/OR/XOR。平坦 Mem 照寫（偵錯工具用）。
- `int 10h AH=10h` 調色盤：AL=00/10（屬性暫存器）、02/12（DAC 整份/一段）。
  16 色鏈：色號 → AttrPal → DAC。實測 AL=09/17 各 1 次語意未知（已記）。
- probe `-dump-ega`（依 BDA 模式選尺寸，出色號 .bin＋PNG）、
  `-adlib`（OPL2 偵測開關）、`-watch`、`-find`、`-dump-linear`、
  ModeChanges／MemOps／VecSets 報告。
- 單元測試 7 支（map mask、位元遮罩+latch、mode 2 展開、讀裝 latch、
  旋轉 XOR、decode、mode 13h 不變），`build ./...`、`test ./...` 全綠。

## 3. 停點根因調查（confirmed 機制，未知收斂到一個指標）

完整證據鏈（全部 probe 軌跡／`-watch`／`-find`／反組譯）：

1. int 08h 向量被**直接寫 IVT**（不走 AH=25h）改成 `BED4:0097`：
   `6B93:00A4`（step 25,757,331）先寫常數 `0x2020`
   （機器碼 `26 C7 04 20 20`），`6C50:02F6`（step 25,757,761）再覆寫成
   `0xBED4`（`26 89 04`＝`mov es:[si],ax`，AX ＝ 來源表 bytes
   `BE`,`D4` 組成 `t[0]<<8|t[1]`）。
2. **整個 IVT[0]–[0B] 都被同一套表格覆寫**：每筆 4 bytes
   `{off=0x0097, seg=變動}`，段位於 A650／A8FA／ABD7／B0EA／B669／
   B7F9／B9EF／BED4（另有幾個還是預設 0x2020）。
3. 執行者是**計時器 ISR**：int 08h 當時指向 `6FB1:008A`（AH=25h 設於
   step 23,298,129，位於 TINYPROG 解出的模組區）。ISR 裡的模組管理器
   （6B93/6C50，含記錄結構 +0B 個數/+0C 項目大小/+0E+10 表指標）
   用一個 **null 基底**（trace 實測 ES=0000）把 ISR 表寫到 0:0000 起——
   也就是 IVT 本體。
4. 模組應該在的區域：0xA650 一帶被填了 0xFF（像是**清過但沒載入**的
   緩衝區），其餘全 0；0xBED40 全 0。
5. 排除過的假說（都有實驗反證）：
   - ~~EMS 頁數~~：8→16 頁崩潰位址完全不變。
   - ~~AdLib 不存在~~：`-adlib` 後逐位元組同點崩。
   - ~~tick 頻率／中斷 stack drift~~：tick 20k vs 100k 同位址。
   - ~~IOCTL 隱藏子功能~~：6 次全是 AL=00。

**目前最佳假說**：模組管理器在註冊（疑似音樂/淡入）模組時用了
一個沒有初始化的表指標（caller 6A04:0071–007B 連推數個 0 進栈當
record 欄位）。該指標在真機上的來源——某個在這裡得到不同答案的
查詢，或某個沒跑到的初始化路徑——**未知**。下一輪的第一個動作：
對 descriptor record 的建立點下 watchpoint（-watch 對 record 位址），
或反組譯 6A04:0054–007B 的參數來源。

## 4. 有沒有到畫面

**到達「互動畫面的邊緣」**：滑鼠輪詢開始、正在繪製，但淡入調色盤沒走完、
~12k 道指令後就崩。所以拿到的是**過場幀**（黑畫面＋已就緒的色號陣列），
不是穩定的標題/主選單。要把畫面對齊，得先解掉 §3 的 null 指標。

## 5. 缺口表（更新）

| # | 缺口 | 等級 | 說明 |
|---|---|---|---|
| 1 | 模組管理器的 null 表指標來源（6A04 caller 的 record 建立） | 未知 | 下一輪第一優先；工具已備（-watch/-find/擴充 trace） |
| 2 | int 33h AX=000Ch 事件回呼 | confirmed（已登錄不回呼）| 目前輪詢已通；需要時做注入事件 → far call |
| 3 | int 10h AH=10h AL=09/17 | confirmed（各 ×1，語意未知）| 看起來不擋路 |
| 4 | CRTC 位移／雙倍掃描、write mode 3、read mode 1 | 假說（未見使用）| spec 009 §4 |
| 5 | int 21h AH=0Dh | **已處理**（靜默收下，spec 009 §3）| — |

## 6. dosgolem-logh3 改動（本批）

- `docs/spec/009-planar-vga.md`（新）。
- `internal/machine/vga.go`（新）、`machine.go`（planar 攔截、埠檔、
  AttrPal、ModeChanges）、`bda.go`（ModeChange 記錄）。
- `internal/dos/bios.go`（int10 AH=10h 調色盤、int33 0x000C 登錄）、
  `int21.go`（AH=0Dh 靜默、AH=48/4A 記 MemOps）、`dos.go`
  （Mouse 事件欄位、MemOps、bh()）。
- `cmd/probe/main.go`：`-dump-ega`、`-adlib`、trace 擴充 DS/ES/BX/SI、
  模式切換記錄。
- `internal/machine/vga_test.go`（新，7 支）。

## 7. 重現命令

```
cd ~/cht/dosgolem-logh3
DOSGOLEM_ORIG=~/cht/logh3/org_game/extracted tools/go.sh run ./cmd/probe \
  -exe /orig/GIN3.COM -root /orig -steps 25759000 -tick 20000 \
  -click-x 320 -click-y 200 -click-at 5000000 \
  -dump-ega /src/workplace/re04/screen-25759000.png \
  -dump-palette /src/workplace/re04/pal.bin
```

產物（一次性證據，gitignore）：`workplace/re04/` 的
`screen-25759000.{png,bin}`（過場幀）、`screen-25759000-fakepal.png`
（假色盤幾何驗證）、`screen-25750000.*`、`opening-24000000.*`
（開頭動畫 24M 步時 A0000 尚未寫入——OPEN.EXE 的繪製高峰在更後面）。
