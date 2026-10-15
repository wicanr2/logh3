# 129：系統設定小字清晰化

本輪依[148限定CONFORMED](../spec/148-clear-settings-font.md)處理設定頁字模縮小遺失筆畫及陰影重疊。唯一狀態表為[140](../goal/2026-10-15-settings-font-round-140.md)。已發布v.1.4.0封包、tag及影片保持，這是原始碼更新。

設定頁專用Noto Sans Mono CJK TC Regular，與啟動畫面同源的開放字庫。12×12全形／6×12半形供原色畫面，36×36／18×36遮罩供HD直接繪製；各自以適合字格的像素大小產生，保留內邊距及完整墨跡，不由原版16×15丟行丟列。部分ASCII超過字格時將完整墨跡縮入，禁止直接裁掉。小字不疊陰影，選取項目以深色文字配亮底。

字庫、工具版本、來源SHA-256與各產生物見[字型來源](../../internal/font/settings_font.SOURCE.json)，完整[OFL授權](../../internal/font/settings_font.LICENSE.txt)隨引擎內嵌資產保存。[產生器](../../tools/scripts/settings_font.py)沿固定`logh3-hd-preview:20261011-r1`映像執行，`--check`在記憶體獨立重生，全部檔案逐位元組一致。

標準`tools/gui.sh test ./internal/... ./cmd/...`為19套件：15 PASS、4無測試。設定專用路徑不依賴遊戲HD字型，原色重播相同；四種遊戲字型下裁切與後畫遮擋通過。控制矩形沿147，沒有更動命中區、規則或存檔格式。

最終程式`workplace/settings140-gin3-r2`沿已驗v.1.4.0完整ZIP的資料／素材，由正常片頭至S1操作，不使用腳本入口或局面／seed注入。帝國與同盟各完成原版／HD、四款遊戲字型、音色取消、兩規則、AI五值及戰場內3→原版→1；敵軍預設關閉，手動開啟後實際敵方圖示可見，關／開返回後回合、單位與鏡頭保持。完整收據位於`workplace/settings140/f0-r2/RESULT.json`及`f1-r2/RESULT.json`，身份由[實錄索引](../images/readme-settings-clear.json)綁定。

兩陣營HD與原版、AI選單原尺寸截圖已目視。HD不再顯示縮小字模及陰影黏連；原版顯示保留點陣呈現。README匯出三張完整960×720原圖並用`--check`獨立核對，不裁切或後製文字。設定頁統一專用字型，介面字型選項仍影響其他遊戲畫面。

本輪沒有重包或覆寫已發布v.1.4.0，未新增Windows／macOS真機或DOS字形parity。#48既有未完成範圍保持。

## 第141輪：保留陰影並補回可見左框

設定視窗原本左側只用近黑色陰影，帝國深藍介面看起來沒有外緣。新框在原矩形內加色號15金屬外層，原右上亮／左下暗改為內層。調色盤由`Game.Palette`及`screen-main-menu.json`確認，15才是金屬色，8為按鈕底色；未改全局調色盤或其他視窗。

標準19套件15 PASS／4無測試；兩陣營正常S1、四字型、F7、AI五值與敵軍實際圖示開／關、返回戰場保持通過。22張設定／選單實錄各檢查530行，帝國#82a4cb、同盟#b99050、原版#b2b2c3外緣連續，內側黑色陰影完整。兩陣營HD原尺寸已目視，README三張完整960×720原圖與[來源索引](../images/readme-settings-frame.json)由[核對工具](../../tools/scripts/settings_border_receipt.py)匯出及`--check`獨立重核。前輪字型收據保持歷史身份，新邊框依此索引。

重新打包進度見[狀態表141](../goal/2026-10-15-settings-repackage-round-141.md)；已發布v.1.4.0的版號例外或新修訂號仍待裁決。
