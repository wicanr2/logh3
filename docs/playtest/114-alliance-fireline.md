# 114：同盟改選赤焰戰線與陣營星圖

使用者於2026-10-11改選「赤燄戰線」，對應既有HTML的同盟B「赤焰戰線」。帝國維持B亮藍艦橋。本收據取代[113](113-faction-themes.md)中的現行同盟A外觀，不重寫113歷史結論。

## 正式呈現

依[135的READY修訂](../spec/135-original-style-hd-ui.md)，同盟改用深褐面板、土金框體與鮮明紅色主要操作。文字仍由當幀原字模重建，原立體框、矩形、按鈕底緣、命中與流程保持。戰略星圖同步改用深褐底與土金航路，16種情報狀態各自映射，徽章、星系位置、敵我標記及選取框保持。

- 正式資料：[interface-themes.json](../../assets/hd/interface-themes.json)。同盟B核准PNG SHA-256為`ffc9e15baa744af061f14066dadd4a5f98320764552b6d471fe36ad711047ac3`，直接取自既有概念頁的B，不重繪原圖。
- 帝國B原PNG SHA-256仍為`71c5de38e33f0e2b26d14ff37460f2bbe5c464a9ab90c67c55d23f0fa3cab8e0`。
- 核准概念HTML SHA-256仍為`7e6cb3610d94a9bc8ad4b3a3370ef6b6ded3f4e50c8f81e3387abc1baf83a5a7`。

## 驗證

以正式`v.1.1.2-20261011` Linux封包程式，沿片頭、主選單、防拷及正常鍵鼠路徑驗證。沒有`-script`、direct-entry、局面或seed注入。

| 範圍 | 結果 |
|---|---|
| 兩陣營各34個完整X11視窗 | 68張、47001600像素，最終全部零差 |
| 財政、訂造、命令、設定，兩主題對核准概念 | 8組、13302720個UI像素，差異零 |
| HTML於1440及390寬 | 各34項通過，切換／圖片／配色標記／水平溢出均通過 |
| 標準測試 | 19套件，15PASS、4無測試 |
| 原索引重播、16情報色、快取隔離、載入／交棒及六表不變 | 測試通過 |

正常X11曾取到尚未更新的畫面或游標。失敗原圖及數字保留，只等待同一輸入完成繪製，沒有遮罩、改預期或重送操作。每張最終完整視窗皆零差。單元測試的第一次失敗來自測試把A的零漸層當成所有主題的條件；覆蓋測試改為隔離漸層，逐列顏色另由八組概念及正常視窗完整核對。未修改正式呈現來迎合測試。

## 重現與保存

模型由`tools/gui.sh xrun`執行`tools/scripts/hd_art_model.go -faction <0|1> -out workplace/hd126/theme-<empire|alliance>-r7`。正式X11由`tools/audio.sh`選既有`logh3-hd-preview:20261011-r1`，執行`tools/s1_record_display.sh`與`tools/scripts/hd_art_gui.py`，程式取自新版正式Linux包、縮放1.5，輸出`theme-x11-<陣營>-r7`。核准概念由`faction_hd_check.py --revision r7`核對；瀏覽器由`faction_hd_game_browser_check.py`核對。

完整命令、來源、失敗嘗試及輸出雜湊見[來源索引](../internals/data/alliance-fireline-theme.json)。離線頁位於`workplace/hd126/FACTION-HD-GAME.html`。113的舊A頁依原雜湊重生並保存在`workplace/hd126/faction113-source/workplace/hd126/FACTION-HD-GAME.html`；113的版控來源副本同目錄保存，原索引不覆寫。

本收據只宣稱remake呈現與已列正常玩家路徑。不宣稱DOS畫面parity。新版影片與封包交付另行驗證，不能由本收據推定通過。原版衍生圖片及實錄仍限本機及目前private儲存庫。
