# CONTEXT — 銀河英雄傳說III SP 專案目前真相

本檔是目前真相與唯一剩餘工作。流水帳在 `WORKLOG.md`；協作規範在 `AGENTS.md`。

## 現況一覽

- 遊戲：《銀河英雄傳說III SP》（Micronet，1994），PC DOS 版，啟動程式 `GIN3.COM`。
- 原版 bundle：`org_game/银河英雄传说3SP.zip`，
  SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`，
  來自 <https://dos.zczc.cz/games/银河英雄传说3SP/>。
- `GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
- bundle 基準機器設定（`.jsdos/dosbox.conf`）：`machine=svga_s3`、`memsize=16`、
  SB16、autoexec 為 `mount c .` + `GIN3.COM`。
- 執行器是 dosgolem，工作副本 `~/cht/dosgolem-logh3`（branch `logh3-com-support`，
  已推 origin）。原版未修改的 `GIN3.COM` 在裡面可以跑進遊戲本體的**主選單**
  並停在互動等待迴圈。
- 本作是 **Big5 中文版**（載入 16 點字型 `BU16IBM.RGB`，選單文字為 Big5）。
  這一點影響所有文字相關的逆向與 remake 決策。
- **還沒有任何對拍**：沒有 same-state 收據、沒有已逆向的資料格式、沒有已驗證的
  遊戲規則、沒有 remake 程式碼。DOSBox 擷取鏈（`tools/dosbox.sh`）尚未建立。
- 已取得的畫面（`workplace/re05/`）調色盤淡入沒走完，實際很暗；能看的圖都是
  假色盤，**僅供目視，不能當對拍基準**。

## 執行器走到哪（confirmed，證據見 `docs/re/`）

- 啟動鏈：`GIN3.COM`（622-byte launcher）→ `OPEN.EXE`（開頭動畫，點一下可跳過）
  → `GIN3PS.EXE`（189,536 bytes 未打包 MZ，遊戲本體）→ `ENDING.EXE`。
- 本體在 dosgolem 內：EMS 全過（AH=41/43 各 1、AH=44 映射 ×152）、載入 19 個資產檔
  （`MUSIC.INI`、`GIN3SP.VOC`、`GRAPH.IMG`、`COMANI.BMG`、`FRAME.BLK`、`SPRITE.SPR`、
  `BLOCK_3.BLK`、`3DBLOCK.SPR`、`COM_FLAG.SPR`、`PLADAT.DAT`、`ADMIRAL.DTX`、
  `STELLAR.DTX`、`ATTR.DAT` 等）、切 VGA mode 12h（640×480×16）、開始畫像素
  （103,452 個非零色號）、滑鼠輪詢啟動。
- 主選單畫面為 **帝國／同盟／對戰／取進度**（3 欄 × 12 列的 Big5 文字視窗），
  滑鼠輪詢 205,509 次，跑滿 6000 萬道指令仍在互動迴圈。
- dosgolem 已補的服務：`LoadCOM`、EXEC（int 21h AX=4B00h／AH=4Dh）與子程式
  結束時的記憶體回收、`EMMXXXX0` 裝置、int 67h EMS 頁配置與映射、VGA planar
  讀寫（write mode 0/1/2/3、set/reset）與 16 色調色盤鏈、int 10h AH=10h、
  int 33h AX=000Ch 登錄、int 21h AH=0Dh。
- 一次完整跑到主選單約 9 秒（6000 萬道指令），回饋迴圈很快。

## 已被推翻的斷言

| 斷言 | 出處 | 推翻依據 |
|---|---|---|
| launcher 停在 EMS 是因為配置失敗早退 | re/02 假說 | re/03：EMS 全過後停點移到別處，該假說 confirmed 為真停點並修掉 |
| 停點根因是「高段載入位址算錯，依賴某記憶體查詢回答」 | re/03 §2 假說 | re/04 §3：實測 ES=0000，是模組管理器用 null 基底把表寫進 IVT，不是位址算錯 |
| 停點與 EMS 頁數有關 | re/03 待驗 | re/04：8→16 頁崩潰位址完全不變 |
| 停點與 AdLib 不存在有關 | re/04 待驗 | re/04：`-adlib` 後逐位元組同點崩 |
| 停點與計時器 tick 頻率／中斷 stack drift 有關 | re/03 待驗 | re/04：tick 20k vs 100k 同位址 |
| 停點與 IOCTL 隱藏子功能有關 | re/04 待驗 | re/04：6 次全是 AL=00 |
| 停點是「模組管理器把 ISR 表寫進 IVT」 | re/03 §2、re/04 §3 | re/05 §2：那段程式是 Big5 文字視窗的儲存格緩衝區；寫進去的「段值」逐位元組是中文字（同取度國進盟對戰） |
| null 基底來自 `6A04:0071–007B` 推 0 當 record 欄位 | re/04 §3 最佳假說 | re/05 §3：`6A04:057F` 是雙向串列插入，那兩個 0 是「插到最前面」的哨兵 |
| write mode 3 未見使用 | dosgolem spec 009 §4 | re/05 §4：鏈內用了 1160 次；當 mode 0 處理時中文字形只剩零星像素 |

## 唯一 worklist

1. ~~計算並記錄原版 bundle 與 `GIN3.COM` 的 SHA-256~~（完成，見上）。
2. ~~dosgolem 第一次接觸~~（完成，`docs/re/01`）。
3. ~~dosgolem 實作 EXEC（int 21h AX=4B00h + AH=4Dh）~~（完成，spec 007／`docs/re/02`）。
4. ~~EMS（int 67h）最小實作~~（完成，spec 008／`docs/re/03`）。
5. ~~VGA planar 管線與調色盤鏈，能把本體畫的內容解出來~~（完成，spec 009／`docs/re/04`）。
6. ~~追出停點的根因~~（完成，spec 010／`docs/re/05`）：EXEC 之後配置游標沒有
   退回，本體被載到段 `0x3922`，heap 要不到空間，malloc 回了假指標。
7. **查調色盤淡入為什麼沒走完**（目前唯一阻塞對拍的項目）：DAC 的 6 位元值停在
   2–15，畫面實際很暗。先找出淡入是誰驅動的（計時器 ISR？垂直回歸？），
   再判斷缺的是哪一個服務。做完才有真色彩的畫面可拍。
8. 淡入走完後，用 dosgolem `Save`／framebuffer 拍第一批 same-state oracle；
   DOSBox 擷取鏈（`tools/dosbox.sh`）作為交叉驗證再建。
9. `int 10h AH=10h` 的 `AL=09/17` 語意未知（各 1 次，目前不擋路）。
