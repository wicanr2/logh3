# 431：標題等待、左鍵與重播

第97輪，GitHub #68，父提交 `481221ad222b02f8c88b2f1f1644fb3a8c6797af`。
第96輪只解標題圖；本輪補正常流程，不重查已解的壓縮格式。

## 輸入與定位

- 原版 bundle：`银河英雄传说3SP.zip`，SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
- `GIN3.COM`：SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
- dosgolem 專用副本：`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，乾淨，Go image `golang:1.24-bookworm`。
- 當前載入器：911000000 指令時擷取 `lin:3490:4096`，4096 bytes，SHA-256 `0138b89c7e5dbd9bcf708630921835806d01e712d5c8ad7c1d9bbb1b78fb3674`。
- IDA 9.4：一次性 COM 資料庫，檔案位移 0 映射到 EA `0x10100`。以下 CS:IP 是 dosgolem 執行期定位，IDA EA 另由 `0x10100 + IP` 換算，不混用。
- [匯出器](../../tools/scripts/title_loader_ida.py)保留原始函式名、bytes 與推論等級；自動命名不當作語意證據。

## 已證實的最小呼叫鏈

| 執行期定位 | 原始 bytes | 事實 |
|---|---|---|
| `0349:0538` | `B8 7E 04` | 尾段一趟完成後傳入1150 |
| `0349:05EE` | `B8 D0 07` | 三趟完成後再傳入2000 |
| `0349:06F6..0705` | `BA D4 03 B9 00 00 B0 0C 8A E5 EF B0 0D 8A E1 EF` | CRTC顯示起始位址寫0，切到完整標題頁 |
| `0349:0706` | `B8 28 23` | 標題傳入9000 |
| `0349:070E..0713` | `83 C4 02 0B C0 74 77` | 等待回0走清理離場，回1續重播 |
| `0349:0776` | `E9 DA F9` | 回到動畫迴圈 `0349:0153` |
| `0349:091B..0923` | `B8 05 00 BB 00 00 CD 33 89 5E FE` | 只讀INT33功能5、按鈕0的按下次數，沒有鍵盤或右鍵判斷 |
| `0349:0B22..0B31` | 參見原指令匯出 | 參數同時寫入倒數兩欄 |
| `0349:0B3E` | `74 1A` | 沒有左鍵就回倒數檢查；有左鍵回0 |
| `0349:0B5E` | `26 83 3E 06 01 03` | 倒數大於3才繼續等待，逾時回1 |

動態堆疊在 `0349:0B22` 已建立，`[bp+06]` 分別為 `047E`、`07D0`、`2328`。
回傳位址分別為 `0540`、`05F6`、`070E`，與靜態呼叫點相同。
入口 `0AFA` 的舊BP仍屬caller，不拿其 `[bp+]` 當callee參數。

## 決定性收據

所有批次走 `tools/oracle.sh`，完整旗標及版本保存在各批次 `flags.txt`、`meta.txt`。
來源快照沒有記憶體注入。時間不受RND控制，沒有重擲。

```sh
bash tools/oracle.sh title68-r97-waits-b -tick 20000 \
  -load-state /src/workplace/title68-r96-tail.state -steps 916000000 \
  -regs-at 0349:0B22,0349:0706 -regs-max 8
# title68-r97-waits-c 用完全相同旗標再跑一次。
bash tools/oracle.sh title68-r97-timeout-a -tick 20000 \
  -load-state /src/workplace/title68-r96-title.state -steps 1150000000 \
  -regs-at 0349:070E,0349:075E -regs-from 1050000000 -regs-max 8
# title68-r97-timeout-b 用完全相同旗標再跑一次。
```

尾段快照 SHA-256 `fb586cd18c1d528922df032a83e58462f0cc0012516fb25ea31b4ddb7108c666`；
標題快照 SHA-256 `154f8e0a6786cb996fbc7abe47ea223a5b4a799d68e80b66f870e3f991c95308`。
等待雙跑224000色號相同，SHA-256 `addde5e0929eb5f99a5430a714b19041a2cbe5530a4720a3641abebbfa3bf8ae`。
逾時雙跑224000色號相同，SHA-256 `cc161242d4f90fe2059a1d351b2d6df0d5cd154528f5a534d6473730a0d403e5`。
逾時返回位址 `070E` 在1094885089指令回AX=1，之後重新讀 `BU16IBM.RGB`、`IBM16.350`、`OPEN_M.MDS`，不是主選單。
第96輪左鍵雙跑到主選單；右鍵、普通A、BIOS A仍停標題。靜態按鈕0輪詢補足「只接受左鍵」的證據。

## 時間與停止線

PIT除數2048沿既有020；秒數採既有輸入頻率1193182 Hz。
末格後兩次等待名義長度為 `(1150-3 + 2000-3) * 2048 / 1193182`，約5.396秒。
標題為 `(9000-3) * 2048 / 1193182`，約15.442秒。
這是 hardware-spec approximation；不模擬解碼、重畫耗時與中斷相位，不宣稱同硬體wall-clock。
`-tick 20000` 的指令數只作收據定位，不能直接換成秒。

## 工具勘誤

末格雙跑 `title68-r97-last-a/b` 從同一850M快照跑到860M，完整224000色號SHA-256
`b9b8f179788ec26d9812d5acd63cdd89c294ab59ccb6b1e0a9f0d73340ba0db9`，palette亦相同。
末格16色逐byte符合 `TITSPT.RGB` 的既有DAC換算。
完整正常第96輪時間線記錄658185104指令換此色盤、914963445指令換標題色盤。
既有 `docs/internals/19` 已記尾段換色，但Game.Palette仍一律回BU16IBM。
本輪補尾段349起用TITSPT；020先前「完整動畫色盤已完成」不能涵蓋這個未接分支。

首次IDA匯出錯把COM檔案位移0當EA `0x100`，雜湊斷言失敗。
IDA本身已成功載入；改用 `ida_loader.get_fileregion_ea(0)` 得EA `0x10100` 後成功。
失敗腳本與log另存本機，不改原始bytes或正式資料庫，不建立替代image。
分級索引匯出另遇ASCII locale不能讀中文JSON，指定讀寫UTF-8後同image重跑成功。
最終匯出681條原指令、九項proven，未分級指令保留unknown警示。

正式玩家契約見[112](../spec/112-title-player.md)，實際驗收見[086](../playtest/086-title-player.md)。
