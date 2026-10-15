# 128：v.1.4.0三平台發行與推廣片重拍

本輪狀態集中於[第139輪狀態表](../goal/2026-10-15-release-and-promo-round-139.md)。可選規則沿[127](127-windows-and-rule-settings.md)，設定頁修正依[147](../spec/147-compact-system-settings.md)。舊候選完整歸檔，現行封包、設定頁及影片各有新收據。

Windows／Linux／macOS完整版與引擎包已從修正後程式重建。正式交付根目錄為`dist-all/v.1.4.0-20261015/`，維持`full/`、`release/`及`promo/`。完整包與影片只留本機，GitHub Release只發布三平台引擎、來源及校驗檔。

## 封包結果

三平台使用產品提交`024a851d`的同一份1576項凍結來源，其中598項非vendor來源與目前程式逐檔相同。七封包經ZIP／來源包內容、必要資料、manifest、權限與SHA-256核對。完整清單為`SHA256SUMS-platforms.json`，公開引擎／來源另有`release/SHA256SUMS.txt`；[交付索引](../internals/data/v140-release-promo.json)保存實際雜湊與驗收收據身份。

| 項目 | 結果 |
|---|---|
| 標準測試 | 19套件：15 PASS、4無測試 |
| Linux完整包 | 11條公開CLI入口、兩AI、戰場及存讀通過；實際ZIP另解壓後逐檔、權限核對 |
| Linux引擎包 | 原版／強化兩啟動器各驗主選單、財政確認，共4案；原版僅由外部路徑匯入 |
| macOS完整／引擎 | Intel／Apple Silicon通用程式，最低macOS11；只依賴系統庫，ARM3180個程式頁SHA-256全部相同，ZIP逐檔及權限通過 |
| Windows完整／引擎 | 修正版實際ZIP解壓到中文含空格目錄，直接啟動、兩啟動器、缺檔及外部匯入共8條Wine入口通過 |
| 設定頁正常操作 | 實際Linux完整ZIP帝國、Windows完整ZIP Wine同盟，各驗敵軍預設關閉／手動開關、AI五值、四字型、F7、兩規則及音色取消 |
| 戰場設定 | 兩平台由正常片頭進S1，開／關敵軍的實際圖示有差；回合、單位及鏡頭不變。戰場內AI3→原版→1與敵軍開關後正常返回 |

macOS未在真機執行；Intel切片未簽章，Apple Silicon為ad-hoc簽章，沒有Developer ID或Apple公證。Linux封包及單元測試不代稱Windows／macOS真機或DOS整場同狀態通過。

修正前未發布候選及原凍結目錄移至`workplace/promo139/precompact/`，`MAPPING.json`保存原路徑對照與七包原雜湊。舊127及138收據保持歷史身份，不拿舊來源雜湊核對新程式。已發布1.3.0與1.2.0的tag、Release與封包皆未改動。

## 設定頁與README

音色庫標籤原本覆蓋敵軍按鈕，兩規則又超出視窗底部。147以12×12全形、6×12半形字格整合設定，繪製與命中共用控制矩形。AI選單只呼叫現有`SetAILevel`，沒有更動演算法、六表或新增AI偏好格式。敵軍以使用者最後決定預設關閉，兩項規則預設開啟。

正式Linux收據在`workplace/promo139/compact-delivery-f0/RESULT.json`，Windows收據在`workplace/music69-windows/compact139-gui-f1/RESULT.json`；平台啟動收據與完整SHA-256由交付索引綁定。[README三圖來源](../images/readme-v140-compact.json)另存完整960×720原圖身份，已目視並用`--check`獨立重核。畫面刻意手動開啟敵軍，沒有改動預設值。

## 全新遊戲錄影

全部遊戲鏡頭來自本輪v.1.4.0實際Linux完整ZIP的正常入口，沒有`-script`、局面或seed注入。戰鬥錄影取本輪147排版修正前的同規則候選；新版設定頁與AI鏡頭取147後重新打包的正式ZIP。帝國取960×720、同盟取本輪1280×960新錄影前段、要塞取960×720；字幕合成保持4:3與原速。1988動畫與官方海報沿既有已核來源，不稱本輪新拍動畫。

成片選用23場完整艦隊交火及2支分隊要塞演出，逐格與結果時點通過。成片約50.53秒艦隊交火涵蓋三種武器，要塞保留通知、首次發射及逐分隊戰損，試射包含友軍。未選用來源仍有漏格或未完整場次，已在PLAN及SOURCE-PLAYBACK另列；不宣稱整份來源錄影或全遊戲無漏格。原版／HD切換、設定頁與AI選單均由實際正常操作錄製。

正式分鏡36段、約102.13秒。單首BGM_0F由0秒連續播放，僅整片頭尾淡入淡出，各片段原聲中的配樂不混入。遊戲音效依本輪首次`played`觀測重混，不稱逐取樣同步。成片聲畫與連續取樣結果由交付索引保存，不宣稱使用者已聆聽新混音。

實際成片為102.133333秒、1920×1080、60fps、雙聲道，11816346 bytes，SHA-256 `5465b7693a75469daeec20a5be8aeafc4068cd4ffa88ab74433068c8299d2733`。35個剪接點連續來源取樣核對通過，AAC全片混音訊噪比34.23dB，無削波、長靜音或黑段。低變動區間依實際來源抽樣分類，保留靜態卡片與正常等待；已目視36格總覽及原尺寸設定、要塞、片尾，字幕安全矩形完整。合成器只重用本輪已核對同來源時碼及字幕的無聲片段，配樂和音效整片重新生成，不沿用舊影片音軌。

影片與16項證據檔位於`promo/`，另有`SHA256SUMS.json`。來源、平台收據及每項輸出由交付索引綁定；Release發布後另核對實際五個下載附件、校驗清單及tag的598項產品來源。

工具入口：[三平台封包](../../tools/scripts/v140_platform_delivery.py)、[正常錄影](../../tools/scripts/ai_gameplay_record.py)、[設定頁驗證](../../tools/scripts/compact_settings_gui.py)、[製片與連續音軌驗證](../../tools/scripts/promo_v140_movie.py)、[公開附件下載核對](../../tools/scripts/v140_release_verify.py)。#48正式戰術讀回重建與DOS整段同狀態仍保持OPEN。
