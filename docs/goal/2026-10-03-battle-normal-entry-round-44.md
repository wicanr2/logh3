# 第四十四輪：正常玩家戰場入口配對

工作項 [#79](https://github.com/wicanr2/logh3/issues/79) 保持開啟。
父提交5751b76bce200b3b229fafcda51fd39a2bbd3907；遠端88張worklist、11張open，268條核對零不成立。

先核對主選單取進度、防拷138、S3、帝國、首次戰略實行至達希利的正常路徑。
兩側戰略實行前固定D3DC167E，不重擲；原版只注入seed，不改艦隊、情報、鏡頭或原始檔。
原版從冷啟動獨立雙跑。remake使用與cmd/gin3共用的bootstrap.Load及Game.New，
只送正式玩家輸入，不用原版RAM還原或EnterBattle。
亂數量測與固定由[測試期Go覆蓋](../../tools/scripts/battle_normal_probe_seed.py)提供；正式產品檔案不改。

先比較入口艦隊／分隊身分、日期、全紀錄及完整畫面。入口不同就記錄差異，
不能繼續把第三階段151個受控幀升格為正常玩家同態驗收。
資料、畫面與seed差異要分類；沒有原版證據不改產品或golden。
087保持READY，其他八步及音樂已裁決範圍保持。

重播入口：[原版擷取器](../../tools/battle_normal_entry_oracle.sh)、
[正式資源玩家探測](../../tools/scripts/battle_normal_probe.go)。
結果追加至[046](../playtest/046-battle-spotting-phase.md)與[研究381](../re/381-battle-spotting-display.md)。

## 本輪結果

載入後三張全表相同，正常首戰參戰7對3艦隊、回合380對381、WORK控制區三bytes不同。
原版5份與remake17份輸出獨立雙跑相同；省略實行前診斷存檔的六表／畫面仍同。
完整307200像素差5863，明標nearby-state，保留原圖／並排／差分。
新上游工作[#89](https://github.com/wicanr2/logh3/issues/89)與[088 DRAFT](../spec/088-normal-battle-player-entry.md)
先追最早分歧；[研究382](../re/382-normal-battle-player-entry.md)保存證據，
[完整收據](../internals/data/battle-normal-entry.json)SHA-256
`d9e719ca5ffca52447c7476154deec1023c87d057987af9ae68d5f163036181e`。
既有索敵25個PASS事件零跳過，151受控UI幀保持；087 READY、#79 OPEN，正式產品與dosgolem未改。
