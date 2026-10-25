# 135：v1.6.1三平台完整版與Linux AppImage

使用者確認本次v1.6.0試玩平台為Linux，要求重打三平台，Linux提供AppImage。152輪較早依Windows平台回覆準備的更新包保留，不能據此將這次試玩記為Windows。唯一狀態表為[153](../goal/2026-10-25-full-appimage-round-153.md)，封裝契約為[155](../spec/155-appimage-full-package.md)，來源、封包及驗證雜湊見[v161-full-delivery.json](../internals/data/v161-full-delivery.json)。

## 交付

三個完整版均為`v.1.6.1-20261025`，放在`dist-all/v.1.6.1-20261025/full/`。

| 平台 | 檔案 | bytes |
|---|---|---:|
| Linux x86_64 | `LOGH3-v.1.6.1-20261025-linux-x86_64.AppImage` | 451459576 |
| Windows amd64 | `LOGH3-v.1.6.1-20261025-windows-amd64.zip` | 451581535 |
| macOS Intel／Apple Silicon | `LOGH3-v.1.6.1-20261025-macos-universal.zip` | 461418440 |

已包含三組海報、五語、四字型、完整原版資料、音樂、音效與音色庫。含原版及衍生素材，僅保留本機。既有更新包與v1.6.0三平台、推廣影片保持，沒有公開Release或tag變更。

Windows及macOS完整解壓後，使用程式或原版／強化啟動器。Linux將AppImage設為可執行後啟動，不需另放素材。若下載工具未保留權限，可使用：

```sh
chmod +x LOGH3-v.1.6.1-20261025-linux-x86_64.AppImage
./LOGH3-v.1.6.1-20261025-linux-x86_64.AppImage
```

沒有FUSE時，官方支援從同一檔案解出並執行：[AppImage官方說明](https://docs.appimage.org/user-guide/troubleshooting/fuse.html)。

```sh
./LOGH3-v.1.6.1-20261025-linux-x86_64.AppImage --appimage-extract-and-run
```

Linux存檔仍在`$XDG_CONFIG_HOME/logh3/save`，未設定時為`~/.config/logh3/save`；AppImage不改寫這個格式或位置。工具列可直接選思源黑體及戰略AI，與系統設定共用當前值。字型與AI沿既有本次執行契約，其他已保存偏好沿原檔案。

## 來源與封裝

三平台使用同一1616項凍結來源，其中638項非vendor，與152輪已驗來源逐檔SHA-256全同；Windows及Linux內部執行檔也與152輪25日程式逐位元組相同。沿最近16 PASS／4無測試，不因未變更產品的重新封裝重跑相同單元測試。

615項共用素材與資料取自已驗v1.6.0完整包，三平台逐項全同。沒有附帶玩家偏好或測試存檔。ZIP核對精確檔案集合、CRC、manifest、SHA-256、UTF-8旗標、BOM／CRLF及權限；Windows實際ZIP解出的安裝副本在GUI後再次逐檔核對。

Linux沿`hr-appimage:runtime-recovery-r1`，固定Type-2 runtime前綴944632 bytes及SHA-256`1cc49bcf1e2ccd593c379adb17c9f85a36d619088296504de95b1d06215aebbf`。以該映像的SquashFS 4.0、zstd及固定時間建立，沒有下載浮動runtime。AppDir共741個可讀檔案，包含15個ALSA／X11函式庫、相依授權與官方runtime授權快照。由實際AppImage本身解出後，全部內容、權限與原AppDir精確相同；內部程式為ELF64 x86_64，實際ELF需求最高GLIBC_2.34。

AppRun依自身位置找素材，預設960×720、原版AI；命令列參數可覆寫。主機仍提供X11／XWayland、glibc、OpenGL／GPU驅動與音效服務，沒有附帶主機驅動。AppImage原理與runtime／檔案系統分工見[官方架構](https://docs.appimage.org/reference/architecture.html)。

macOS沿既有Go1.24.13及SDK15.5工具鏈，同源建amd64與arm64再合為通用Mach-O。兩側皆最低macOS11，依賴均為系統函式庫；ARM SHA-256臨時簽章4083頁逐頁通過。沒有Developer ID或Apple公證，也未作macOS真機驗收。

## 實際封包操作

正式AppImage與實際Windows ZIP各31張正常主選單鍵鼠操作。沒有direct-entry、script、局面或seed注入；可選遙測僅唯讀。抽測主系統按鈕及下拉列hover、思源黑體、AI3與設定同步、五語字型／AI選單、Esc及右鍵取消、F1／F7、原版反光、海報工具列與側欄防誤觸。

AppImage正常操作沒有指定`-orig`、`-data`、`-art`或`-save`，由封裝自行找到完整素材；測試只將`XDG_CONFIG_HOME`隔離至工作區。六表實際寫入`config/logh3/save`，再從正常讀檔介面選陣營讀回。Linux與Windows六表SHA-256全同，也與v1.6.0基準相同。兩平台普通畫面與海報中央各691200像素零差，沒有遮罩。

另各三條診斷啟動，驗開場、原版及強化模式、自動素材、HD、GM／MT-32與預設關閉。AppImage每條均執行正式.AppImage，從`/tmp`啟動，不改跑解出的裸程式。這些有script的診斷與上述正常鍵鼠分開記錄。

AppImage正常移動的Draw中位9.06ms，24次移動前後HD背景上傳18至18。這是Xvfb／軟體OpenGL環境抽樣，不宣稱硬體輸入延遲或每個發行版相容。`latency_ms`包含xdotool程序啟動與排程，不能用來推論端到端輸入延遲。

## 重現與限制

入口為[full_appimage_release.py](../../tools/scripts/full_appimage_release.py)、[AppImage包裝器](../../tools/appimage.sh)、[正常操作器](../../tools/scripts/pointer_toolbar_gui.py)、[AppImage診斷](../../tools/scripts/appimage_smoke.py)與[獨立驗證器](../../tools/scripts/full_appimage_receipt.py)。打包器拒絕覆寫現有輸出，原凍結輸入在`workplace/package92-v.1.6.1-20261025-release153/`，AppDir與驗證器快照在`workplace/appimage153/`。產生器與收據的SHA-256列於交付索引。

```sh
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh \
tools/gui.sh xrun python3 tools/scripts/full_appimage_receipt.py --check
```

容器沒有FUSE裝置，因此驗證採實際AppImage的extract-and-run，未驗桌面FUSE掛載；Windows採Wine。沒有新增真機、人耳或DOS parity聲明；#48正式戰術讀回重建保持OPEN。
