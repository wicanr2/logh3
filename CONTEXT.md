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
- 本作是 **Big5 中文版**，1994 年由亨和企業．微波軟體發行，中文版程式
  Andrew Liao、翻譯邵程衍；日文原版為 BOTHTEC／QUEST ＋ Micro Vision（1993）。
  文字用 Big5，日文假名走 **ETen／HKSCS 擴充區**（標準 Big5 表會解錯）。
- 逆向出來的結構化資料整理在 **`docs/internals/`**（素材清單、執行檔與製作
  名單、啟動與記憶體、文字視窗與編碼、調色盤）。`docs/re/` 是逐次調查紀錄。
- **還沒有任何對拍**：沒有 same-state 收據、沒有已逆向的資料格式、沒有已驗證的
  遊戲規則、沒有 remake 程式碼。
- 已取得第一張真色彩畫面（主選單，`workplace/re06/`），同命令連跑兩次逐位元組
  一致。dosgolem 跑的是原版執行檔，所以那就是**原版側**的畫面；還缺的是
  remake 側——對拍要等 remake 有東西可比才成立。

## 執行器走到哪（confirmed，證據見 `docs/re/`）

- 啟動鏈：`GIN3.COM`（622-byte launcher）→ `OPEN.EXE`（開頭動畫，點一下可跳過）
  → `GIN3PS.EXE`（189,536 bytes，**GLUE V1.70 封裝**，遊戲本體）→ `ENDING.EXE`。
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
- 主選單以遊戲自己的顏色畫出：黑底、青綠選單框、白色中文字。
  `沒實作的服務（0 種）`。
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
| 畫面全黑是遊戲自己的淡入沒走完 | re/04 §1、re/05 §5 | re/06 §2：是 dosgolem 的兩個缺口——`int 10h AH=10h` 的 AL 分支接錯，以及「取向量後直接跳」的 BIOS 呼叫從未執行 |
| `GIN3PS.EXE` 是未打包的 MZ | re/01 §3 | 檔頭 `0x24` 起是明文 `GLUE V1.70 -- Program Coder (By:C.C.F.)`；整支檔案撈不到可讀字串 |
| 模組是「TINYPROG 解出的」 | re/03 §2、re/04 §3 | 整包檔案裡沒有 `TINYPROG` 字串；`TINYPROG says` 只在單獨執行 `GIN3PS.EXE` 時出現，與 GLUE 的關係未知 |
| `BU16IBM.RGB` 是 16 點字型 | 早前的現況一覽 | 48 bytes ＝ 16 色 × 3 bytes 的調色盤，分量是 0–15 的 4 位元值；字型的候選是 1.76 MB 的 `IBM16.350` |

## 唯一 worklist

1. ~~計算並記錄原版 bundle 與 `GIN3.COM` 的 SHA-256~~（完成，見上）。
2. ~~dosgolem 第一次接觸~~（完成，`docs/re/01`）。
3. ~~dosgolem 實作 EXEC（int 21h AX=4B00h + AH=4Dh）~~（完成，spec 007／`docs/re/02`）。
4. ~~EMS（int 67h）最小實作~~（完成，spec 008／`docs/re/03`）。
5. ~~VGA planar 管線與調色盤鏈，能把本體畫的內容解出來~~（完成，spec 009／`docs/re/04`）。
6. ~~追出停點的根因~~（完成，spec 010／`docs/re/05`）：EXEC 之後配置游標沒有
   退回，本體被載到段 `0x3922`，heap 要不到空間，malloc 回了假指標。
7. ~~查調色盤為什麼是黑的~~（完成，spec 011／`docs/re/06`）：不是淡入沒走完，
   是 `int 10h AH=10h` 的 AL 分支接錯，加上「取向量後直接跳」的 BIOS 呼叫
   從未執行。
8. **把主選單存成正式的 oracle 收據**：獨立 run 目錄 ＋ metadata（bundle 與
   `GIN3.COM` 雜湊、dosgolem Git HEAD、完整輸入序列與步數、顯示模式、
   framebuffer 與 DAC 雜湊），寫進 `docs/playtest/`。這是之後每一次
   remake 差分的基準。
9. **開始資料格式逆向**：`GRAPH.IMG`、`SPRITE.SPR`、`FRAME.BLK`、
   `PLADAT.DAT`、`ADMIRAL.DTX`、`STELLAR.DTX` 等已知會被載入的檔。
   dosgolem 可以在載入點下攔截，拿「解碼前的檔案 bytes」對「解碼後的
   記憶體內容」，比純靜態猜格式可靠。**每解出一個格式就寫進
   `docs/internals/`**，不要只留在 `docs/re/` 的敘事裡。
10. **字型**：跑不跳開頭動畫的路徑，確認 `IBM16.350`（1.76 MB）是不是
    中文點陣字型並解出格式。
