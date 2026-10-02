# 第三十五輪：第二戰場及正常資源結果

工作項[#79](https://github.com/wicanr2/logh3/issues/79)，保持開啟。
父提交1a7c5e5b1aef3eee08baf75aa4d5ca9c9aa71087。
第34輪固定第6回合35邊界已驗，不重開。

先核原版隨附存檔及正式玩家載入路徑，找第二戰場與能出現補給／修理結果的既有局面。
事實可由檔案／可丟棄探測查明，不以隨機重擲碰局面。
需要原版固定狀態時依014記憶體注入契約，完整記錄初態／每個注入與seed。
新的完整畫面比較在執行前明列等級；不足條件時不宣稱same-state。
正常操作證據另列，不用diagnostic入口或fixture代替。

沿[079](../spec/079-battle-preparation.md)、[080](../spec/080-battle-movement-phase.md)與
[083](../spec/083-battle-phase-panel.md)的READY範圍驗收；發現新規則先回RE／spec。
不改八步順序、公開方式與音樂決定，不新增CPU逐週期或硬體wall-clock門檻。

收尾發現受控文件的證據連結無效，另在[#88](https://github.com/wicanr2/logh3/issues/88)
登錄後修正共用Issue路徑解析。既有#67／#32／#66／#25／#81僅修證據格式或檔名，
保留敘述、章節定位、原判準及完成狀態；不改遊戲規則。
