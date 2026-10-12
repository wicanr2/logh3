# 115：赤焰戰線、雙陣營HD與AI實錄影片

2026-10-11，第126輪。使用者將同盟改選為既有B「赤焰戰線」，帝國維持B亮藍與黃金。本輪從最新正式來源重建`v.1.1.2-20261011`，交付根為`dist-all/v.1.1.2-20261011/`。沒有移動或覆寫既有版本，沒有建立遠端tag或Release。

## 範圍與結果

| 項目 | 已驗結果 |
|---|---|
| 陣營介面與星圖 | 帝國亮藍／黃金，同盟深棕／土金／赤紅。保留原版字形、立體框、按鈕位置及底緣；兩陣營各34個正常X11完整視窗，共47001600像素零差 |
| 核准概念 | 四畫面×兩陣營，共13302720個UI像素零差；同盟B不使用歷史A的7像素例外。離線實際畫面頁在1440及390寬各34項通過 |
| 敵軍顯示 | 預設關閉，設定鈕與F7切換。正常S3戰場的關／開／關／開均核完整視窗，關閉恢復圖與初次關閉圖相同；資料隔離、摘要與同輸入戰鬥結果測試保持 |
| 戰略AI | 兩模式從相同六表的新局開始，正常實行一回合，再以存讀切換陣營檢視CPU部署。正常X11實錄各6張完整視窗零差，三支CPU艦隊命令確有差異 |
| 戰術AI | 兩模式從公開S3正常載入，模型核戰略入口與首戰六表完全相同。正式程式各完成4回合，各10張完整實錄視窗零差；成片保留實際推進及敵艦進入主視野 |
| S1實際遊玩 | 正常主選單、防拷、讀檔清單顯示三個額外劇本，選S1亞姆力扎會戰，進場下令並完成3回合。錄影含艦隊移動、交火、光束及命中結果 |
| 故事影片 | 118秒，1920×1080、30fps、立體聲。HD雙旗背景左帝國右同盟，原生HD字幕與星點動畫，故事引向「改變歷史、扭轉局勢」 |
| 影音驗收 | 成片平均−33.4dB、峰值−12.9dB，沒有黑段；43個來源影格抽樣通過。停格偵測逐項分類為介面展示、實錄輸入停留或仍有星點變動的字幕段。24個成片抽樣已目視，字幕未裁切 |
| 測試及封包 | 19套件15PASS／4無測試；八封包、六ZIP逐檔manifest與SHA-256核對。Linux完整版11入口，Windows Wine完整版6入口，兩平台引擎啟動器各4條；兩平台三劇本讀取及保存重讀通過 |
| macOS | Intel／Apple Silicon通用包、依賴、ZIP權限與ARM CodeDirectory全頁雜湊通過。未在macOS真機執行，沒有Developer ID或Apple公證 |

戰略命令的模型欄位為`[艦隊索引,Region,Target,Move]`。原版規則模式為`[[23,20,15,0],[28,30,24,1],[31,30,24,1]]`，強化模式為`[[23,31,4,5],[28,20,15,0],[31,20,15,0]]`。這是該初始局面的實際結果，不推廣為所有戰局必勝。

## 正常實錄與來源

所有影片由正式Linux封包的`gin3`執行，SHA-256為`19fea68e0ceedd2afc4c63e284ec82a48fc37458af5a805f900190e2f28e4e1e`。使用X11真實鍵鼠，沒有`-script`、記憶體注入、診斷進場或seed重擲。模型是核對工具，不能單獨代替實錄。兩種AI沿產品既有初始seed `0xD3DC167E`；正常存讀與場景初始化保持正式行為。

| 錄影批次 | 正式交付名稱 |
|---|---|
| `workplace/hd126/s1-record-r4/` | `promo/RAW-S1-HD.mp4` |
| `workplace/hd126/ai-record-original-r3/` | `promo/RAW-TACTICAL-ORIGINAL.mp4` |
| `workplace/hd126/ai-record-enhanced-r2/` | `promo/RAW-TACTICAL-ENHANCED.mp4` |
| `workplace/hd126/strategy-record-original-r2/` | `promo/RAW-STRATEGY-ORIGINAL.mp4` |
| `workplace/hd126/strategy-record-enhanced-r2/` | `promo/RAW-STRATEGY-ENHANCED.mp4` |

完整命令、相對輸入時間、錄影與程式雜湊保存在各`TRACE.json`及交付的`smoke/recording-*/`。五段實錄分別量測非靜音，平均−35.2至−27.6dB。遊戲畫面與聲音在同次執行錄製；比較段採右側強化模式音軌。片頭、片尾及兩段介面靜圖使用正式遊戲的`BGM_0F.ogg`，沒有把靜圖稱為戰鬥錄影。

成片為`promo/LOGH3-v.1.1.2-20261011-promo.mp4`；逐段時碼在`promo/PLAN.json`，來源核對及影音量測在`promo/AUDIT.json`，目視總覽為`promo/CONTACT-SHEET.png`。來源影格平均通道誤差須小於4，容許H.264再次編碼；這個檢查不宣稱逐像素無損。

HD雙旗以已檢視的原旗幟照片為參考，透過imagegen重建；輸入、完整提示與權利分類保存在`promo/FLAG-BACKGROUND-HD.json`。原輸出1672×941，合成背景以Lanczos放至1920×1080；文字與星點按1920×1080逐幀產生，不冒稱來源圖原生1920解析度。世界觀及亞姆力札引導文參考[Production I.G官方介紹](https://www.production-ig.co.jp/works/ginei/story.html)及[官方第12至15話故事](https://ginei.jp/story1.php)，未使用原動畫片段。

## 重現與交付

- 編譯來源HEAD為`523ced1346358b1d3a9982148c569a69af953a5f`，工作樹含未提交內容。`BUILD.json`綁1523項實際建置輸入，完整凍結於`workplace/package92-v.1.1.2-20261011-hd/source/`。
- 工具映像沿`logh3-ebiten:2`、`logh3-hd-preview:20261011-r1`與`psychicwar-osxcross:go1.24.13-15.5-r2`。均在Docker執行，原版輸入唯讀，UID/GID為1000:1000。
- [實錄器](../../tools/scripts/ai_gameplay_record.py)、[戰術模型](../../tools/scripts/ai_promo_model.go)、[戰略模型](../../tools/scripts/strategy_ai_promo_model.go)保存正常玩家操作。製片入口為[故事製片器](../../tools/scripts/faction_story_promo.py)的`prepare`、`assemble`、`audit`；本輪計畫為`workplace/hd126/promo-r2/SCENES.json`。
- [交付核對器](../../tools/scripts/faction_delivery_finalize.py)在`tools/gui.sh xrun`容器內以`python3 tools/scripts/faction_delivery_finalize.py`執行，核凍結來源、當前素材、六個ZIP、兩份來源包及所有本輪收據，再集中到正式交付根。
- [來源索引](../internals/data/faction-hd-ai-delivery.json)保存實際來源及輸出完整SHA-256，來源副本在`workplace/hd126/delivery115-source/`。`DELIVERY.json`與`SHA256SUMS.json`位於版本根。原有113、114及第125輪證據保留歷史身份。

本次「原版規則模式」指remake既有原版策略及戰術模式。沒有新增DOS整場同狀態證據，也不宣稱AI必勝、Windows／macOS真機或新片人耳驗收。完整登陸分段及原有音訊限制保持。第125輪地圖與動畫抽樣沿用既有驗收，不宣稱本輪重新對拍所有原版動畫。

完整版、HD美術及影片含原版衍生素材，只留本機及授權私有範圍。`release/`引擎包不含原版資料、音樂或HD衍生美術，使用RRSAL-1.0。現行樣式證據見[114](114-alliance-fireline.md)，本輪唯一狀態表見[計畫126](../goal/2026-10-11-original-ui-ai-promo-round-126.md)，入口掛[專案索引](../internals/00-README.md)。
