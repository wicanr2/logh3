# 第46輪：讀檔後座位與日期

工作[#89](https://github.com/wicanr2/logh3/issues/89)，父提交04ae7d659321b0b568287852b7afc1dba30121f1。
開工89張worklist／12張open，269條核實零不成立。沿remake／spec與文件職責路由。

第45輪AI一次性跳過已驗，不重開。接續[088](../spec/088-normal-battle-player-entry.md)未完成範圍：
讀檔後原版帝國／同盟／對戰選單，WORK座位控制區的來源、玩家可見路徑與保存。
各分支固定原版輸入，擷取選單及選擇後狀態；不可用原版RAM代替正式Game。
同一份S3的正常首次戰場日期380／381仍不同，追原版回合寫入與戰鬥續接，證據充分後才READY。

優先接原版已量測的正常選單，不猜補資料欄位或月底時序。
只讀相關IDA／動態控制流，不完整反編譯；修正後補受影響選擇、輸入及正式存讀與標準回歸。
完整入口沒有同態前保持DRAFT及nearby-state，不以新子範圍代替#89或正常索敵驗收。


本輪已驗限定選單及座位子契約，詳[研究384](../re/384-load-faction-selection.md)與[完整收據](../internals/data/load-seat-selection.json)。
18標準套件、21限定PASS及五個真實X11樣本通過；日期在戰場入口提前進位及活動座位未修，留#89。
