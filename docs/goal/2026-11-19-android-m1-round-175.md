# 第175輪：Android M1 起步

使用者核准規格161的版面與觸控提案，交代「接下來都你來做，不用 Gemini」，並要求 Android 版完成後重拍推廣片。
起點main `f92bc0a3`。

## 目前狀態表

| 項目 | 狀態與契約 |
|---|---|
| 工具鏈映像 | PASS，`logh3-android:20261119-r1`：Go 1.24.13、Ebitengine／ebitenmobile 2.8.8、SDK 35、NDK 27.2、Gradle 8.9 |
| 模擬器映像 | PASS，`logh3-android-emulator:api35-20261119-r1`，執行時用 KVM |
| 程式結構 | `cmd/gin3` 原地不動（被 105 份文件與多支工具引用）；新增 `internal/mobileshell`（殼）、`mobile`（只露出 `Start`、`Back` 給 Java）、`cmd/mobile-preview`（桌面預覽行動版） |
| AAR | PASS，只編 64 位元（arm64、x86_64）；規則碼假設 int 為 64 位元 |
| APK | PASS，完整版 37.8 MB（原版 162 檔、data 99 檔，尚未含 HD 圖與音樂）|
| 模擬器 | PASS，載入 0.3–0.9 秒；開頭動畫、輕觸跳過、原版主選單 |
| 側欄與自繪數字鍵盤 | PASS，原版命令列按鈕樣式與點陣字；左側「退回」＝右鍵，防拷問答時右側出數字鍵盤（０–９、退格、決定）。模擬器以觸控輸入 138 進入月初報告，兩下「退回」到戰略地圖 |
| 輕觸 | PASS，改由 Java 在手指放開時排進 Go 佇列；逐格取樣在模擬器上會漏接 |
| 繪製分層、側欄、跳出視窗、觸控板模式 | 未做 |
| 推廣片重拍 | Android 版完成後 |

## 這一輪踩過的坑（已寫進 `tools/android.sh`、`tools/android_emulator_check.sh` 註解）

- ebitenmobile v2.8.8 依賴的 x/tools v0.25.0 在 Go 1.26 編不過，Android 映像的 Go 換成 1.24.13。
- gomobile 在 `go list -m -json all` 失敗時吞掉錯誤、寫出空 go.mod；離線建置前要先補齊整張模組圖。
- gobind 把被綁定套件的每個匯出識別字都轉成 Java，回傳兩個值的函式無法轉換；殼要放在 internal。
- gobind 把中文註解複製進 Java，javac 要設 UTF-8。
- Java 端沒呼叫 `Seq.setContext` 時，Ebitengine 取不到螢幕密度，外部尺寸變 +Inf，第一格就 panic、畫面全黑。
- 模擬器 2400×1080 加軟體 OpenGL 會讓系統介面未回應，測試改用 1280×576。
- 腳本迴圈裡的 `adb shell` 會吃掉 `while read` 的標準輸入。
- 模擬器只有每秒十幾格，`adb input tap` 的按下與放開常落在同一格之間，Ebitengine 的逐格觸控取樣看不到；
  改由 Java 的 `ACTION_UP` 送進佇列。測試腳本也改成等 logcat 出現「載入完成」再送輸入。

## 下一步

- 側欄在開頭動畫期間跟著動畫的灰階調色盤變色，要固定用遊戲本體調色盤。
- 繪製分層與戰略畫面新版面（地圖靠左滿高、卡片槽）、跳出視窗。
- 觸控板模式；完整版加入 HD 圖與音樂；存讀檔路徑實測；生命週期。

## 文件

- 規格：[161](../spec/161-android-port-plan.md)
