# WORKLOG — 銀河英雄傳說III SP

唯一流水帳。新条目加在最上面。

## 2026-09-06

- 主選單互動對拍到五個狀態逐位元組一致：反白第 0–3 列各一份收據、游標在選單外
  一份。解出的模型：屬性是「底色<<4|字色」（正常 0x97、反白 0x22，反白時兩者
  相同所以字看不見）、游標是 13×16 的白箭頭加色號 3 描邊、畫在
  `(滑鼠x+1, 滑鼠y+41)`、命中判定用的是**游標熱點**不是滑鼠座標本身，
  繪製順序是「一般列 → 游標 → 反白列」。遊戲把滑鼠垂直範圍設成 0–399
  （螢幕是 480 列），那是座標偏移的來源。
- 輸入打通：dosgolem 補 IRQ1 鍵盤（埠 0x60）與滑鼠虛擬座標倍率依視訊模式決定
  （spec 012，`4c94d1a`）。倍率本來寫死 2（mode 13h 用的），本作是 mode 12h，
  所以點在 x=300 遊戲收到 600、落在選單外——畫面完全正常，症狀只是
  「點了沒反應」。修好之後點選單立刻出現反白。
- **收據綁執行器版本**：滑鼠倍率一改，跳過開頭用的點擊落點整個變了，主選單
  收據與剛完成的對拍一起失效。改用 `tools/oracle.sh` 產生收據並記錄執行器
  HEAD，規則寫進 AGENTS。新收據換成「主選單 ＋ 第 3 列反白」的狀態
  （乾淨、無游標、120M 步仍不變），remake 加上反白後再次**逐位元組一致**。
- remake 開工，主選單達成 **逐位元組一致**的同狀態對拍：色號陣列與調色盤的
  SHA-256 與 oracle 收據相同。Go 專案骨架（`gfx` 色號畫布／`font` 原版字型／
  `ui` 斜角視窗與文字格／`screen` 畫面組裝／`cmd/logh3` 出圖／`cmd/oradiff`
  差分），版面數字全部從 `docs/internals/data/screen-main-menu.json` 讀，
  不寫死在程式裡。收據 `docs/playtest/001-main-menu.md`，回歸測試
  `TestMainMenuMatchesOracle`（缺字型或缺收據就 skip，不用代用品）。
  差分過程：539 → 31 → 7 → 0。三個修正各自對應一個發現——文字有立體邊
  （字形每列左移一位、往上一列、色號 3）、左移是對位元做的不是整字平移、
  框角規則對不上（七個像素，先明列修正並標成技術債）。
  建立 `VERIFICATION-MATRIX.md` 作為完成度的唯一數字來源。
- re07：中文字型解出來了。`GRAPH.IMG` 是完整的中文點陣字型（半形 8×15、
  全形 16×15、每列 2 bytes 高位在前），**內容與倚天中文系統的 `stdfont.15`
  與 `ascfont.15` 逐位元組相同**。方法是拿主選單九個已知字當答案去反查，
  三條獨立證據都對上：點陣逐位元吻合、九次 seek 偏移與公式完全相同、
  與倚天字型 100% 相同。三次「找不到」各有成因——假設 16 列（實際 15 列）、
  賭檔名 `IBM16.350`（實際是 `GRAPH.IMG`）、搜記憶體與 EMS（字型按需
  seek+read，從沒整塊存在過）。dosgolem 加了檔案存取記錄與 EMS 內搜尋
  （`30fbbb0`）。建立 `tools/py.sh` 分析包裝器、`docs/internals/data/*.json`
  機器可讀輸出與 `workplace/assets/` 的 PNG 匯出。
- 方針：dosgolem 成為唯一 oracle，DOSBox 降為原始碼參考（缺功能或行為存疑時
  讀它的實作弄清楚真機語意，不照抄、不當證據等級）。bundle 的 `dosbox.conf`
  保留為原版機器規格來源。
- 新增 `docs/internals/`：逆向出來的事實整理成保存用的參考，與 `docs/re/` 的
  調查紀錄分工。段值不穩定（修 EXEC 回收後本體降了一萬多個段），一律用結構
  位移與位元組簽章記錄，九組簽章實測可重新定位。內容含 162 個素材的清單與
  雜湊、`GIN3PS.EXE` 的 GLUE V1.70 封裝、`ENDING.EXE` 裡的完整製作名單
  （中文版：亨和企業．微波軟體／邵程衍／Andrew Liao）、堆積配置器與視窗
  管理器的結構、文字視窗與 Big5（假名走 ETen 擴充區）、調色盤與淡入公式。
  更正三個錯誤斷言：`GIN3PS.EXE` 不是未打包 MZ、模組不是 TINYPROG 解出的、
  `BU16IBM.RGB` 是調色盤不是字型。
- re06：拿到第一張真色彩畫面（主選單，黑底青綠框白字），同命令連跑兩次逐位元組
  一致。黑畫面不是遊戲的淡入沒走完，是 dosgolem 的兩個缺口疊在一起：
  `int 10h AH=10h` 的 AL 分支接錯（AL=02h 是 17 bytes 的屬性調色盤卻被接成
  768 bytes 的 DAC，AL=10h 是設單筆 DAC 卻被接成設屬性暫存器，讀回的
  07/08/09/15/17 全缺），以及「`AH=35h` 取向量 → 自己疊框架 iret 跳進去」的
  BIOS 呼叫從未執行——服務層掛在 CPU 的 INT 指令 hook 上，而向量指向裸 iret，
  C 的 `int86x` 那條路安靜地什麼都不做。改成每個向量一段 `int n; iret` 的
  stub。dosgolem-logh3 `769b030`（spec 011），已推 origin。
  注意：該分支上另有其他 session 的提交（rich2 選單修正、`cmd/scenetime`），
  本批以 rebase 接上。詳 `docs/re/06-palette-alive-first-real-screen.md`。
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
