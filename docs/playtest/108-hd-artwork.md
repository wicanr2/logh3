# 原畫風 HD 人物與本機測試版

使用者確認原畫風 HD 修復試樣後，首批 5 張人物已接入遊戲：萊因哈特、吉爾菲艾斯、奧貝斯坦、楊威利、卡介倫。這是第一批美術，其他 78 張人物與艦艇、地圖尚待修復。

## 試玩入口

本機完整版版本：`v.0.1.4-20261006`。交付根為 `dist-all/v.0.1.4-20261006/`，含原版及衍生素材，只留本機。正式封包、SHA-256 及抽驗結果由該目錄的 `BUILD.json`、`SHA256SUMS.json` 與 `smoke/` 保存。

| 平台 | 本機完整版資料夾 | 啟動 |
|---|---|---|
| Linux | `full-local/LOGH3-v.0.1.4-20261006-linux-amd64/` | `start-original.sh` 或 `start-enhanced.sh` |
| Windows | `full-local/LOGH3-v.0.1.4-20261006-windows-amd64/` | `start-original.cmd` 或 `start-enhanced.cmd` |
| macOS | `full-local/LOGH3-v.0.1.4-20261006-macos-universal/` | `.app` 或兩個 `.command` |

兩個啟動器表示原版或強化 AI，均預設使用 HD 圖片。HD 顯示畫布為 1920×1440，保留原有 4:3 比例；啟動視窗為 960×720，可調整大小。只有已修復的人物使用新圖，其他美術與字型沿用原圖。

可優先測帝國／同盟新局、預算變更與完成、一覽表人物資料，以及游標停在人像上時的遮擋。存檔沿既有平台使用者目錄，AI、規則、音訊與存檔格式保持。

## 素材與重建

- [規格122](../spec/122-hd-artwork.md)保存使用者裁決與顯示契約。
- [HD 素材入口](../../assets/hd/README.md)與[來源雜湊](../../assets/hd/MANIFEST.json)保存實際 5 張及完整提示詞。使用內建 image_gen；定版第 0 號圖保持原輸出雜湊。
- 原人物來自 FACE.AMG，原圖尺寸 80×100，索引沿內部參考 09、研究 271 及既有匯出器。HD 圖含重繪，不能稱為原版逐像素還原。
- [正常操作模型](../../tools/scripts/hd_art_model.go)、[真實 X11 操作](../../tools/scripts/hd_art_gui.py)及[全視窗擷取](../../tools/scripts/hd_art_capture.go)沿正式 Game 與 renderer，兩陣營各 28 張。
- 建置沿 [既有封包器](../../tools/scripts/local_test_package.py)的可選 `--art assets/hd`，macOS沿現有通用程式工具鏈；不修改或覆寫前版交付。

正式來源執行方式：

```sh
bash tools/gui.sh run ./cmd/gin3 -orig /orig -graphics hd -art assets/hd -scale 1.5
```

`-graphics original` 回到原圖顯示。PNG 截圖依所選顯示模式；色號傾印仍為 640×480。替換圖缺少單張時回退原圖，整個 HD 目錄缺少有效圖、圖片損壞或索引越界時明確報錯。

## 驗收界線

合成器已驗同色覆蓋、圖層順序、界外裁切、部分原圖裁切、缺圖回退、損壞圖拒絕與滑鼠座標。原模式與 HD 的同輸入色號傾印保持 307200 bytes 相同。

實際Linux包完成帝國1.5倍與同盟2倍各28張正常X11視窗，共53760000像素零差。標準19套件14PASS／5無測試；Windows Wine六入口、macOS雙架構與ARM程式頁簽章、三平台ZIP內全部HD圖及MANIFEST通過。[來源索引](../internals/data/hd-artwork.json)與[核對器](../../tools/scripts/hd_art_finalize.py)集中完整結果。

GUI 採正常片頭、主選單與鍵鼠輸入，沒有直接進入人物畫面或局面注入。完整視窗與同版本模型逐像素比較；HD 來源縮到實際視窗使用已固定的最近鄰取樣。這證明 remake 顯示與輸入接線，不是原版對拍。

Windows 使用 Wine 抽驗，macOS核對通用架構、依賴及程式頁簽章；兩者仍未真機驗收。人耳與原有未完成項保持。前版 72 秒推廣片保持，本輪沒有新影片。最終結果與 Docker 清理寫入 WORKLOG 及 CONTEXT。
