# 361 — Linux 音樂工具完整固定來源集合

日期：2026-10-01。工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
接續[研究 360](360-linux-bundled-music-tools.md)，只補對應來源，不修改合成器或遊戲規則。

## 身份與方法

已證實（proven）：實際工具包 BUILD.json 的 198 個 Debian 二進位套件對應
157 組去重來源套件／完整版本。BUILD.json SHA-256：
`e10ba65a53a4d7050025ae4c8b5735e263e961a644ccaa707d02265c9571a56b`。
全部工具包檔案在封裝前重驗既有雜湊，沒有以其他版本補件。

依 [Debian Policy 的 .dsc 契約](https://www.debian.org/doc/debian-policy/ch-controlfields.html#debian-source-control-files-dsc)
解析 Source、Version 與 Checksums-Sha256。157 份同版本描述均可從 Debian 正常／安全
來源庫取得，無須 Snapshot 或新版替代。PGP 包裝原樣保存；本輪未驗 Debian 簽章信任鏈，
只依實讀 HTTPS 內容建立固定 SHA-256 鎖。epoch 只在檔名省略，身份比較保留完整版本。

來源身份、描述及所有檔案 URL／SHA-256／大小保存
[`linux-music-source-lock.json`](../internals/data/linux-music-source-lock.json)，雜湊：
`f102cbe2028118a0b3e560a2666bf0d733913dcc95f41bd0d86f40c9d7dd7bac`。
集合包含 FluidSynth、FFmpeg、AdPlug 及所有實際共享依賴；
合計 157 描述＋356 去重來源檔＝513 檔，748,421,917 bytes。
來源 archive 沒有解包或執行；同名不同雜湊、錯身份、錯 SHA-256、連結與越界名稱拒絕。

## 實作與結果

依審查 READY 的[規格 073](../spec/073-linux-music-source-closure.md) 實作固定鎖準備器。
預設離線；只有 --download 才可下載。四個並行下載工作、單檔 512 MiB／總量 2 GiB
上限，串流寫暫存檔，固定大小與 SHA-256 通過後才提交。已存在損壞快取在連網前拒絕。
完整下載後於 --network none 容器重驗，所有描述再解析並與鎖定來源集合逐項核對。
五組邊界測試與 157 份實物描述通過，損壞／缺檔情境不會呼叫網路。

本機待交付來源 tar 包含 513 個來源檔、完整套件著作權／授權、工具 BUILD.json、
固定鎖與驗證 manifest，以及本專案 opl2wav 原始碼／audio Dockerfile。
734 個普通 entry、752,414,720 bytes；直接重讀 tar 逐項比對 SHA-256／大小，無缺項或連結。
SHA-256：`c0bf4070d492d41dcff9cf883c56c3aba90e5117abb84feea3bb2241ba1e3f74`。
原版遊戲檔、生成曲庫、SF2、工具二進位與共享庫沒有放入來源封包。
首次包的中文說明用字修正後只移除已驗雜湊的該份 tar，從原快取乾淨重建；
原工具包與前輪驗收收據保留不動。這是文件產物修正，沒有放寬來源比較。

## 邊界

證據工具為 Python 3.12 容器 image
`sha256:229a2c5bfa27522db7815ea81f9bed70af17ccb9de9fc7ad142b1877b5830d36`。
輸入程式父提交為 `aa7149a01064963d2aa5a74ff6782796a71c29e6`，本輪工作樹含新增建置器；
程式雜湊、來源與輸出 manifest 見[驗證 033](../playtest/033-linux-music-sources.md)。
這項結果是來源集合完整與實物身份對照，不是逐套重建、簽章信任驗證、正式公開版交付
或全部 GPL／LGPL 發行義務完成的結論。其他平台與人耳／循環接點仍按 #69 接續。
