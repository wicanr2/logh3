# v.1.7.0-20261119 四平台完整版與推廣片

2026-11-19，第177輪。使用者要求打包四平台完整版並重拍推廣片；裁決版號 v.1.7.0、Android 先補 HD 圖、建本機發行金鑰、推廣片沿用 v1.6.4 故事片加入 Android 段。機器收據見 [v170-full-delivery.json](../internals/data/v170-full-delivery.json)。

本版內容是 Android 版（[規格161](../spec/161-android-port-plan.md) M1）、素材按需載入第一階段（[規格160](../spec/160-on-demand-asset-loading.md)）與參謀更換修正（[規格064](../spec/064-staff-appointment.md)、[驗收145](145-staff-pick-plan-screens.md)）。

## 封包

位置在 `dist-all/v.1.7.0-20261119/`，只限本機，未公開上傳。

| 平台 | 檔案 | bytes | SHA-256 |
|---|---|---:|---|
| Windows amd64 | `full/LOGH3-v.1.7.0-20261119-windows-amd64.zip` | 372498347 | `1d0893ca199372712ffadd5bf25ee0d59bb7f06773a211dcd2de0b0ba00ec7f1` |
| Linux x86_64 | `full/LOGH3-v.1.7.0-20261119-linux-x86_64.AppImage` | 372488696 | `aa8d9270fedb06096beadf17d05a15da403632289e6635166633c67d8246ddef` |
| macOS 通用 | `full/LOGH3-v.1.7.0-20261119-macos-universal.zip` | 382408901 | `d1fc4257ceb28e1aab193aadcf3cdee547176362636ade2207032145b2731885` |
| Android（arm64、x86_64） | `full/LOGH3-v.1.7.0-20261119-android.apk` | 158250692 | `ac1d63ca7d7ecfd7c2c1adef8053104d789a97b77c6a20207f32121f5bd668f9` |
| 推廣片 | `promo/LOGH3-v.1.7.0-20261119-promo.mp4` | 22798102 | `5b41152325afa053ae08f5fc6df82bfe68beea736477dbeb06c705bd259c205b` |

桌面三平台從 main `56fbbc9b` 凍結 1679 個來源檔建置，652 個共用素材與 v1.6.5 逐檔相同，只換程式與說明檔。發行資料與目前 `docs/internals/data` 逐檔比對相同，所以素材整包沿用。

Android APK 從 `e51caa7d` 建置，素材取自同一份共用素材：原版檔、99 個資料檔、根層 50 個原版配樂與音效 OGG、去掉海報的 HD 圖。海報畫在畫面兩側，行動版兩側放側欄，所以不收。凍結之後的產品來源變更只在 `internal/mobileshell`，桌面執行檔不引用它。APK 以本機發行金鑰簽章（v1、v2），憑證 `CN=LOGH3 remake (local)`，SHA-256 `b90d5052…6f574ef9`。金鑰與密碼在 `workplace/android-state/`，不進 Git。之後的版本都用這把金鑰簽，才能覆蓋安裝、保留 App 內存檔。

## 驗收

| 路徑 | 平台 | 結果 |
|---|---|---|
| 新遊戲帝國：星域15開戰、九回合、存讀檔、再開戰 | AppImage、Wine | 通過 |
| 新遊戲同盟：星域13開戰、九回合、存讀檔、再開戰 | AppImage、Wine | 通過 |
| 前線駐留、換提督、當地領船、E／T拒絕、C旗艦、出擊、存讀 | AppImage | 通過，存檔與 v1.6.5 相同 |
| 參謀更換：預算、建造、生產三入口，兩陣營 | AppImage 解出的執行檔 | 通過，前後截圖逐張目視 |
| 新遊戲兩陣營整場（同上路徑），每一下走行動版觸控 | 行動版殼（桌面驅動） | 通過，另含 16:9 寬度 |
| S3 讀入、戰略版面、選取、星系情報、一覽表、系統指令、實行、戰場捲動、點分隊 | Android 模擬器，release APK | 通過（推廣片錄影同一次） |
| 同盟新遊戲、防拷、月初報告、存進度、強制關閉重開、取進度 | Android 模擬器，release APK | 通過，讀回日期與資金相同 |

新遊戲兩陣營存出的七個檔案，在 AppImage、Wine、行動版殼三者之間逐位元組相同，也和 v1.6.5 相同。行動版殼的整場測試是 `internal/mobileshell` 的 `TestMobileNewGameFullBattle`，戰略版面裁掉的命令列改按側欄、右鍵改按「退回」、防拷改按側欄鍵盤，手機畫面上點不到的位置直接判失敗。

截圖逐張目視：新遊戲兩陣營、兩平台各九張；參謀更換兩陣營各二十一張；模擬器存讀檔七張。同盟參謀更換第一次執行時，照固定秒數操作的腳本在主機負載高時跟不上，預算與建造兩個入口沒有走到；負載降下後重跑通過，收據取第二次。

標準測試 `go test ./cmd/... ./internal/...`：17 個套件通過、6 個沒有測試、0 失敗。

## 推廣片

沿用 [v1.6.4 已驗故事片](142-v164-full-scout-march-promo.md)。舊片的片段來源已在 11 月 4 日清理時刪除，不能整片重剪，所以直接沿用成片：

- 開頭兩張與結尾一張字卡，用與舊片相同的算法重繪新版號。先以舊版號重繪，和舊片同一格比對，三張都在 39 dB 左右，差異只來自重新編碼；剪接格位置也逐一量出（第120、427、547、7493格）。
- 舊片其餘影格原樣解碼再編碼，不增刪、不變速；抽樣比對最低 37.0 dB。
- 結尾字卡前插入五段 Android 實錄，共 25.3 秒：戰略地圖選取、星系情報、側欄開一覽表、戰場方向鈕捲動、點分隊開作戰選單。來源是交付的 release APK 在模擬器上的正常玩家路徑，原速，觸控點以系統的「顯示觸控」標出。
- 配樂：舊片的 BGM_0F 到插入點只剩 2 秒，不循環。Android 段改用這段畫面在遊戲中實際的平時曲 BGM_3F（S3 存檔 `WORK+1F` 音源 1 為 FM、`+20` 平時曲 3），音量沿舊片混音同一倍率（量得 5.98，擬合 37.5 dB），交界 0.8 秒淡出淡入。BGM_3F 開頭 1.25 秒與 3.0 秒附近的短暫無聲是原曲休止。

全片 9131 格、152.18 秒、1920×1080、60 fps，AAC 立體聲 44100 Hz，峰值 0.846，沒有黑幀。舊段音軌對舊片 37.5 dB。目視過全片抽格表、Android 段與結尾字卡原尺寸。沒有人耳驗收。產生器是 `tools/scripts/v170_promo.py`（analyze、render、audit、publish），分析與核對結果和影片放在同一個 `promo/` 目錄。

## 可重現入口

所有命令在儲存庫根目錄，經 Docker 包裝器執行。

```sh
bash tools/release.sh python3 tools/scripts/v170_full_release.py prepare   # 依序 windows、linux
LOGH3_APPIMAGE_BATCH=release177 LOGH3_APPIMAGE_VERSION=v.1.7.0-20261119 bash tools/appimage.sh
bash tools/release.sh python3 tools/scripts/v170_full_release.py appimage
bash tools/macos.sh macos-release177 bash /src/tools/scripts/local_test_macos_build.sh v.1.7.0-20261119 workplace/package92-v.1.7.0-20261119-release177/source
bash tools/release.sh python3 tools/scripts/v170_full_release.py macos      # 之後 extract-appimage、extract-windows
LOGH3_ANDROID_BASE=$PWD/workplace/package92-v.1.7.0-20261119-release177/common tools/android.sh stage workplace/android-v170/stage
tools/android.sh aar workplace/android-v170/aar
tools/android.sh release-apk workplace/android-v170/aar workplace/android-v170/apk workplace/android-v170/stage v.1.7.0-20261119 10700
bash tools/release.sh python3 tools/scripts/v170_full_release.py android
bash tools/release.sh python3 tools/scripts/v170_promo.py analyze          # 之後 render、audit，目視後 publish
bash tools/release.sh python3 tools/scripts/v170_full_receipt.py
```

`android` 之前要先用 `apksigner verify --print-certs` 把簽章寫進 `workplace/android-v170/APKSIGNER.txt`。正常玩家路徑的批次是 `workplace/release177/verify-linux.sh`，Wine 走 `tools/player_wine.sh`，命令列記在各自的 `RESULT.json`。行動版整場測試：`LOGH3_MOBILE_E2E=1 LOGH3_ORIG=/orig go test -run TestMobileNewGameFullBattle ./internal/mobileshell/`。

## 限制

- Windows 只在 Wine 驗證，沒有實機。Wine 同盟戰後那一張截圖仍是戰場畫面；遊戲記錄的狀態已是戰略畫面，v1.6.5 同一張也一樣，判斷是 Wine 截圖時視窗還沒重繪。
- macOS 核對架構、簽章與素材，沒有真機。
- Linux 用 extract-and-run，沒有測 FUSE。
- Android 只在模擬器驗證，沒有實機收據；切背景回來、聲音、觸控手感都要實機。release 版（App 內匯入原版檔）、觸控板模式與 M2 未做。
- 不新增 DOS 同態聲明。未打 tag。
