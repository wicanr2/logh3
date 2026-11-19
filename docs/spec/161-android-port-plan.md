# 161：Android 版規劃

狀態：**DRAFT**（2026-11-19，第174輪）。只有規劃，沒有任何 Android 程式碼。發行方式已裁決，觸控對應待裁決。
相關：[160](160-on-demand-asset-loading.md)（按需載入是 Android 版的前提）。

## 目前事實

下列各項都已在本機核對，來源寫在右欄。

| 項目 | 事實 | 來源 |
|---|---|---|
| 引擎 | Ebitengine v2.8.8 | `go.mod` |
| 行動版建置工具 | `cmd/ebitenmobile bind`，`-target` 預設 `android`，`-androidapi` 最低 21，`-javapkg` 指定 Java 套件；產出給 Android Studio 用的程式庫，附 `EbitenView.java` 範本 | 模組快取 `ebiten/v2@v2.8.8/cmd/ebitenmobile/main.go` 第 97–121 行、`_files/` |
| 行動版入口 | `mobile.SetGame(game)`，只呼叫一次；呼叫之前遊戲不會開始 | `ebiten/v2@v2.8.8/mobile/mobile.go` |
| 音訊 | oto v3.3.3 有 Android 驅動（內含 Oboe） | `oto/v3@v3.3.3/driver_android.go`、`internal/oboe/` |
| 遊戲迴圈 | 規則層、畫面層、HD 呈現都是純 Go，不依賴桌面 API | `internal/` |

## 移植阻礙（已確認）

1. **音效裝置偵測會讓 Android 編不過。** `cmd/gin3/audio_device_linux.go` 標 `//go:build linux`，
   用 cgo 連 ALSA。`GOOS=android` 也符合 `linux` 標籤，會去找 `alsa.pc`。要改成 `linux && !android`，
   Android 走 `audio_device_other.go`。
2. **入口是 `package main` 加命令列旗標。** 載入流程是 `main()` 裡的一個閉包，路徑都來自旗標與
   `os.Executable()` 推出的發行目錄（`resourcepaths.go`）。行動版需要一個可被 `ebitenmobile bind`
   匯出的套件，接收 Java 傳進來的目錄，回傳 `ebiten.Game`。做法是把載入流程抽成共用函式，
   桌面 `main` 與行動版都呼叫它。
3. **存檔目錄。** `defaultSaveDir` 用 `os.UserConfigDir()`。Go 在 Android 走 Unix 分支，
   需要 `XDG_CONFIG_HOME` 或 `HOME`（`$GOROOT/src/os/file.go`），App 程序不保證有。
   改由 Java 傳入 App 私有目錄（`Context.getFilesDir()`）。
4. **素材是一般檔案路徑。** 全部用 `os.ReadFile`／`os.Open` 讀目錄。APK 裡的 assets 不是檔案系統，
   要選一種：首次啟動解到 App 私有目錄、改走 `fs.FS` 抽象由 Java 的 AssetManager 供檔，或用商店的
   大型資源分送機制。發行包目前約 356 MB（v1.6.5 AppImage），主要是 HD 圖 101 MB、配樂 127 MB、
   音色庫 142 MB。商店對 APK／AAB 的大小限制**待查證**，查到再定。
5. **沒有觸控輸入。** `cmd/gin3` 有 11 處讀滑鼠（`CursorPosition`、`IsMouseButtonJustPressed`），
   讀觸控的地方是 0。介面也依賴滑鼠特性：左下角摘要跟著游標停的星域走（hover），右鍵逐層關窗，
   左鍵按住連捲。
6. **原版遊戲檔不能隨包。** 專案規則是發行包不內嵌原版遊戲檔與素材（`AGENTS.md` 素材邊界）。
   Android 版要讓玩家匯入自己的原版檔，例如用系統的檔案選取器選資料夾，再複製到 App 私有目錄。

## 已裁決

**發行與原版檔取得方式**（使用者 2026-11-19）：比照桌面版分兩種。

| 版本 | 原版檔 | 用途與位置 |
|---|---|---|
| 完整版 | 打包進安裝包，裝好就能玩 | 使用者自己玩；只留本機或授權的私有位置，與桌面完整版同一條規則 |
| release 版 | 不含；玩家在 App 內匯入自己的原版檔 | 可散布的版本，與桌面 release 版同一條規則 |

兩種共用同一份程式，只差安裝包內容與首次啟動是否走匯入流程。完整版的原版檔同樣不得進 Git，
也不得放進可公開下載的位置。

## 待使用者裁決

依專案規則一次只裁決一個父層決策。順序：

1. **觸控操作對應**：例如點一下 ＝ 移游標後左鍵、長按 ＝ 右鍵、雙指點 ＝ 右鍵；
   hover 摘要改成「點選才更新」或保留游標。這是玩家路徑的改變，要先定。
2. **畫面方向與比例**：建議鎖橫向，沿用海報 16:9 外框，中央 4:3。
3. **背景與中斷**：切到背景時暫停音訊；原版沒有自動存檔，是否加一個暫停時的保存點。

## 階段

| 階段 | 內容 | 完成判準 |
|---|---|---|
| A0 | 規格160第二階段量測；訂 Android 記憶體預算 | 遊戲中超過 50 ms 的格列成清單並分類 |
| A1 | 載入流程抽成共用函式；音效偵測加 `!android`；存檔與素材目錄改為參數 | 桌面版行為不變，標準測試全過 |
| A2 | Docker 內建 Android 工具鏈（SDK、NDK、JDK、gomobile），`ebitenmobile bind` 出程式庫，最小 App 外殼 | 容器內產出可安裝的 APK；不在主機安裝工具鏈 |
| A3 | 兩種安裝包：完整版內含原版檔；release 版的 App 內匯入流程 | 完整版裝好直接進主選單；release 版實機匯入後能進主選單 |
| A4 | 觸控輸入（依觸控裁決） | 兩陣營正常玩家路徑可用觸控走完：月初報告、命令、清單、戰場、存讀檔 |
| A5 | 生命週期、音訊暫停、記憶體與發熱 | 實機切背景再回來不崩潰、不丟進度 |
| A6 | 簽章與發行包結構 | 比照 `dist-all/<版本>/{full,release}/`，完整版只留本機 |

## 驗證方式

- 桌面回歸：A1 之後跑標準 21 套件與既有正常玩家 GUI 路徑。
- Android：模擬器能否在本機 Docker 內跑（需要硬體虛擬化）**待查證**。實機驗收需要使用者的裝置。
- 觸控路徑的完成定義比照畫面對拍：每條路徑的截圖、輸入序列、裝置型號與 Android 版本都要記錄。
