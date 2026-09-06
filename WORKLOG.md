# WORKLOG — 銀河英雄傳說III SP

唯一流水帳。新条目加在最上面。

## 2026-09-06

- re05：停點解掉，本體跑進**主選單**（帝國／同盟／對戰／取進度），滑鼠輪詢
  205,509 次。根因不在圖形也不在中斷——dosgolem 的 EXEC 之後配置游標沒有退回，
  `OPEN.EXE` 用掉的 230 KB 沒還，本體被載到段 `0x3922`，向 DOS 要 `0x6A07` 段
  （超過 640 KB）被拒，heap 拿不到空間，malloc 從位在 `0000:0000` 的假區塊切出
  指標，選單文字就落在 IVT 上。re03／re04 讀成「模組管理器寫 ISR 表」是誤讀：
  那八個「段值」逐位元組解出來是中文字（同取度國進盟對戰）。另補 planar
  write mode 3（鏈內 1160 次，先前當 mode 0 處理導致字形破碎）與 mode 0 的
  enable set/reset。dosgolem-logh3 `803cb06`（spec 010），已推 origin。
  詳 `docs/re/05-menu-reached-heap-starved-by-exec-leak.md`。
- re04：VGA planar 管線打通（dosgolem-logh3 `518028d`，spec 009），拿到本體的
  第一張 framebuffer 傾倒。畫面全黑是遊戲自己的淡入調色盤還停在暗值
  （DAC[15]=(8,8,8)），色號陣列證明圖形已畫進 plane（103,452 個非零像素）；
  滑鼠輪詢已開始，到達互動畫面邊緣。停點根因由「載入位址算錯」的假說收斂成
  confirmed 機制：計時器 ISR 內的模組管理器用 null 基底（ES=0000）把 ISR 表
  寫進 IVT[0]–[0B]。EMS 頁數、AdLib、tick 頻率、IOCTL 四個假說各有實驗反證。
  剩下的未知只有一個：那個表指標在真機上從哪裡得到值。
  詳 `docs/re/04-gin3ps-planar-pipeline-and-module-manager-crash.md`。
- dosgolem 第一次接觸（詳 `docs/re/01-first-contact-gin3com.md`）：
  `GIN3.COM` 是 622-byte launcher；dosgolem 新增 `LoadCOM` 後原版路徑跑 35 道
  指令停在 EMS 偵測（`EMMXXXX0` 字元裝置開啟失敗，exit 3）。patch 版證實
  EXEC 鏈：`OPEN.EXE` → `GIN3PS.EXE`（本體）→ `ENDING.EXE`。
  缺口：字元裝置開啟、int 67h EMS、EXEC（4B00/4D）。
- 建立 GitHub repo `wicanr2/logh3`（private）並推送。
- 從 dos.zczc.cz 下載 js-dos bundle（SHA-256
  `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`），
  存入 `org_game/`（gitignore），解出至 `org_game/extracted/`。
- 建立專案骨架：`AGENTS.md`（參考 sangokushi 協作規範 + dosgolem 對拍方法論）、
  `CONTEXT.md`、`.gitignore`、`docs/{playtest,spec,re}/`、`workplace/`。
- dosgolem 改動收在 fork 分支 `logh3-com-support`（wicanr2/dosgolem），已推送；
  後續對 dosgolem 的對拍驗證支援都走該分支，MR 合併由使用者主導。
- 建立 logh3 專用的 dosgolem 本地副本 `~/cht/dosgolem-logh3`（branch
  `logh3-com-support`，origin 指 GitHub wicanr2/dosgolem），與主 checkout
  `~/cht/dosgolem` 隔離；後續對拍驗證改動一律在副本內進行。
- re03：EMS 頁框/映射實作完成（dosgolem-logh3 `43e3460`，spec 008），本體
  GIN3PS.EXE 通過 EMS 配置（AH=44 映射 ×152 全過）、載入 19 個檔、切 VGA mode
  12h、A0000 有 ~38KB planar 內容。新停點：模組管理器把 int 08h 向量改到
  未載入的 BED4:0097（根因假說：高段載入位址算錯，依賴某記憶體查詢回答）。
  詳 docs/re/03-gin3ps-ems-ok-timer-vector-crash.md。
