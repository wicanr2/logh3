# 155：Linux完整版AppImage

狀態：限定CONFORMED。2026-10-25，限[135實際AppImage與封包驗收](../playtest/135-v161-full-appimage.md)；採extract-and-run與Xvfb，未驗桌面FUSE掛載。使用者已指定Linux AppImage，並確認實際試玩平台為Linux。這是既有正常遊戲入口的封裝，不改遊戲規則、資料、原版忠實度或偏好格式。

- 使用v.1.6.1-20261025與152輪逐檔相同的凍結來源，三平台共用已驗v1.6.0完整素材。Windows及macOS保留既有ZIP；Linux交付x86_64 Type-2 AppImage。
- 沿既有`hr-appimage:runtime-recovery-r1`，固定runtime SHA-256為`1cc49bcf1e2ccd593c379adb17c9f85a36d619088296504de95b1d06215aebbf`，前綴944632 bytes。由該映像的mksquashfs建唯讀檔案系統，不下載浮動runtime。可重現來源在`/home/anr2/cht/hr/tools/docker/Dockerfile.hr-appimage-r1`及其runtime恢復腳本，本專案只沿用已安裝映像。
- AppDir的`usr/bin/gin3`旁包含原版、data、art、music、soundfont及licenses。AppRun從自身位置找資料，切到該目錄，預設960×720、原版AI，命令列參數可覆寫。保留正常開場、主選單與系統設定，不使用診斷入口。
- 附帶ALSA與X11使用者空間函式庫及其相依與授權，保留主機glibc、OpenGL／GPU驅動、音效服務。記錄實際ELF需求，不冒稱所有Linux皆可用。
- 存檔與偏好沿既有`$XDG_CONFIG_HOME/logh3/save`或`~/.config/logh3/save`，不寫AppDir，不附帶測試存檔。保持先前Linux進度可用。
- 容器沒有FUSE裝置，正式抽測使用同一AppImage的官方extract-and-run入口，核對其內EXE雜湊、全部素材及正常鍵鼠、字型、AI、hover、海報與存讀檔。另核開場、原版／強化參數與自動資料路徑。不能將extract-and-run稱為桌面FUSE掛載已驗。

官方契約：[AppImage架構](https://docs.appimage.org/reference/architecture.html)、[執行與可執行權限](https://docs.appimage.org/user-guide/run-appimages.html)、[無FUSE時extract-and-run](https://docs.appimage.org/user-guide/troubleshooting/fuse.html)。

唯一狀態與交付入口：[153](../goal/2026-10-25-full-appimage-round-153.md)。
