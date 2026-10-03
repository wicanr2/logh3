# 第62輪：艦隊戰鬥入口與玩家命令

工作依據：[#79](https://github.com/wicanr2/logh3/issues/79)，演出相關部分沿[#48](https://github.com/wicanr2/logh3/issues/48)。
父提交`5662aed0892fc2e413befec535a40a580d80451d`；沿[062](../spec/062-battle-turn-eight-steps.md)，
完整第④步尚未READY。

目前正式ResolveBattleTurn走行星對艦隊近似，艦隊對艦隊常式1DD6:0009尚未接入。
本輪先核第四階段的參戰門檻、排序、敵方射程查詢與人類「戰鬥命令」入口，
取得固定seed、完整注入記錄與獨立原版收據；不以舊傷害單元測試代替此流程。

原版只用專用dosgolem f7523f9。固定第61輪自然索敵邊界作定位，
需要接敵時在唯一入口明示注入完整局面，不靠重擲取得。
IDA9.4沿固定RAM輸出原始名稱、位址、bytes與分級語意；未知項不猜補。

證據足以審查的契約才升READY，再接正式玩家路徑、UI與保存。
061單人CPU被動政策保持；不追整段精確骰序或硬體wall-clock。
完整#79／#48與Goal的完成門檻保持，不由入口常式或診斷收據宣稱完成。
