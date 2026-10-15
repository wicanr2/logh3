# Windows啟動與兩項可選規則

第139輪交付位置勘誤：使用者要求設定頁修正後重建未發布v.1.4.0，本篇第138輪候選完整保留於`workplace/promo139/precompact/dist-all/v.1.4.0-20261015/`，原Windows凍結來源在同一`precompact/workplace/`對應目錄。舊來源索引及雜湊不變，其`dist-all/v.1.4.0-20261015/`路徑應以上述歸檔根回查；現行交付以[驗收128](128-v140-release-and-promo.md)為準。這是尚未發布候選替換，不移動已發布tag或Release。

2026-10-15，第138輪。契約為[146](../spec/146-flagship-transfer-and-rule-settings.md)，目前狀態見[唯一狀態表](../goal/2026-10-15-windows-and-rule-settings-round-138.md)。來源與交付雜湊集中於[收據索引](../internals/data/windows-rule-settings.json)。本輪為使用者指定的 remake 擴充。

## Windows路徑修正

舊程式直接啟動會使用開發目錄 `org_game/extracted`。既有完整版實際把原版放在執行檔旁的 `original`，因此直接雙擊會找不到 `GRAPH.IMG`；原啟動器另外帶旗標才避開問題。

新版在封包中自動讀取執行檔旁的 `original`、`data`、`art` 與 `music`。明確指定的旗標優先，缺原版的引擎包仍提示缺資料。沒有改以目前工作目錄猜測另一份遊戲。

## 規則

| 設定 | 預設與行為 |
|---|---|
| 轉移旗艦 | 開啟。旗艦分隊由正艘數被擊毀為零時，立即判定一次；謬拉100%，其他提督約25%。轉往同艦隊艘數最多的存活戰艦或高速戰艦分隊；同艘數選較小分隊序 |
| 名將移動加成 | 開啟。保留強化模式中萊因哈特／楊威利參戰時的陣營1.5倍新額度；關閉後恢復一般額度，名將戰術AI仍保留 |

兩個開關在系統設定下方，偏好保存於玩家存檔目錄的 `preferences.json`，與音色鍵共存。轉移只交換現有編制與旗艦槽，不移動或增加艦艇，不回補彈藥；提督、副官及池索引保持。舊旗艦保持零艘，於末端清理作廢。轉移關閉時不消耗額外亂數。

艦隊交火在傷害動畫快照固定後轉移，要塞砲在逐擊之間轉移，行星反擊在下一階段前轉移。三條正式接點共用同一判定。戰略非戰鬥耗損沿原規則；沒有把新規則稱為DOS原版行為。

## 驗證結果

| 項目 | 結果與範圍 |
|---|---|
| 標準測試 | 19套件：15 PASS、4無測試，紀錄為 `workplace/rules138-tests-final.log` |
| 機率與候選 | 0至99閾值、固定seed 1失敗／2成功、謬拉100%、副官不觸發保證、兩艦種／最多艘數／同艘數、其他艦隊／艦種／離場排除 |
| 單次與連續擊毀 | 同次失敗不重擲，新旗艦再次被擊毀可再次轉移；無候選及關閉不取亂數 |
| 原池與保存 | 保持單位索引及池引用；末端清理後原旗艦作廢，公開六表保存及LoadSlot解析仍以新旗艦為第0組 |
| 三條傷害路徑 | 使用實際S3資源的受控傷害測試，覆蓋艦隊、要塞與行星反擊；演出快照保留當次舊旗艦。這不是自然遊玩觸發或DOS對拍的聲明 |
| 實際移動引擎 | 開／關／開驗新艦隊與分隊額度及目的地範圍，基本速度不改、不累乘，原版AI模式不啟用名將加成 |
| Linux正常GUI | 帝國與同盟皆由片頭、防拷、月報、訂造進入設定；兩項開關、原圖／HD、F7、退出與重開偏好通過，共10張完整視窗 |
| Windows實際ZIP | 完整／引擎包逐檔及SHA-256通過；中文含空格目錄、`/tmp`工作目錄、直接片頭／主選單、兩啟動器、引擎缺檔及合法外部資料匯入共8案通過 |

Windows正常GUI另以實際ZIP解壓的程式驗證；結果及完整視窗由同一收據索引保存。測試在Wine容器執行，不等於Windows真機。新規則的完整戰損驗證為受控S3測試，正常GUI驗證覆蓋系統設定與偏好，不冒稱完整戰役或原版亂數對拍。

## 本機交付

版本為 `v.1.4.0-20261015`，本輪只建立Windows二進位及引擎來源。完整包包含先前已驗的音色選擇、要塞演出與戰術保存返回。Linux原始碼及正常GUI已驗；沒有重做Linux／macOS封包，也沒有建立遠端tag或Release。

| 檔案 | 路徑 |
|---|---|
| Windows完整包 | `dist-all/v.1.4.0-20261015/full/LOGH3-v.1.4.0-20261015-windows-amd64.zip` |
| Windows引擎包 | `dist-all/v.1.4.0-20261015/release/LOGH3-v.1.4.0-20261015-windows-amd64.zip` |
| 引擎來源 | `dist-all/v.1.4.0-20261015/release/LOGH3-v.1.4.0-20261015-engine-source.tar.gz` |
| 封包校驗 | `dist-all/v.1.4.0-20261015/SHA256SUMS.json` |

完整包、原曲衍生OGG、HD及實錄保留本機；引擎包不含原版、HD、OGG、SF2或ROM。既有v.1.3.0發行內容保持原身份。#48仍OPEN，正式戰術讀回重建與DOS整段同狀態不由本輪結案。

## 重現入口

- [封包產生器](../../tools/scripts/windows_hotfix_package.py)：由受驗v.1.3.0素材、現行runtime JSON與凍結程式乾淨建立完整版及引擎，拒絕覆寫既有版本。
- [ZIP驗證](../../tools/scripts/windows_delivery_smoke.py)：`bash tools/player_wine.sh python3 /tools/scripts/windows_delivery_smoke.py`，驗證實際ZIP及啟動器；輸入ZIP先放入該腳本指定的Wine工作區。
- [正常GUI](../../tools/scripts/rule_settings_gui.py)：Linux由`tools/gui.sh xrun`執行；Windows以`LOGH3_WINE_IMAGE=eob-remake-wine-verify:ubuntu-noble-20261008-r2 bash tools/player_wine.sh`執行。完整參數保存於RESULT.json。
- [路徑負向對照](../../tools/scripts/windows_resource_paths.py)：保留舊版找開發路徑失敗與修正版成功的四案。

Wine軟體呈現較慢時，正常GUI腳本等待遊戲幀確認按下／放開。載入小視窗轉大視窗後，會核對虛擬螢幕邊界並將超出螢幕的視窗移回，避免右側滑鼠座標被X server夾限。這是驗證工具修正，沒有改遊戲時鐘或產品命中座標。失敗批次與量測保留於本輪工作區。
