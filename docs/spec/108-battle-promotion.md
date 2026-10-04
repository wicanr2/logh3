# 108：戰場提督與副官升級

狀態：DRAFT。遠端#79，完整062仍DRAFT。

玩家在第8步索敵回收後取得提督／副官升級通知。
第84輪已補原版60案雙跑與成長模型，完整提督表相同；
艦隊重算、typed回寫、正式報告與正常玩家存讀仍缺，不得據此實作。
原版會改SHIP速度，不能只更新ADMIRAL。

輸入為唯讀GIN3.COM，SHA-256
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
主工具為固定dosgolem f7523f9與IDA Pro 9.4，
一次性COM位址、原版CS:IP與線性位址分列。

[本輪計畫](../goal/2026-10-04-battle-promotion-round-84.md)、
[IDA輸入產生器](../../tools/scripts/battle_promotion_prepare.py)、
[完整062](062-battle-turn-eight-steps.md)。

已證實輸入為ADMIRAL 28-byte記錄、FLEET 122-byte記錄、SHIP 78-byte記錄，
參戰清單順序與固定受控亂數。成長門檻、12級軍階、旗標及通知見
[研究418](../re/418-battle-promotion.md)、[原版診斷073](../playtest/073-battle-promotion-oracle.md)。
原版來源索引SHA-256 `ab96e1d8d0213da5dfc8b34acc1f46bf1b83440eb3d552ed7f37efd8c2b0c661`；
分級指令索引SHA-256 `3d52c61e0e93433297f6240078614c11da38cf27e24347a4cfe926cfc4f451b6`。

原版提督通知固定旗標2。每次成功依序重算、報告等待、再試同一人。
副官使用實際成長旗標OR1。正式實作需保持此順序。
現有RankOf的28門檻與10級上限由新原指令推翻；修正納入本契約，未READY前保留原程式。

READY前須補297B:05B2／0E8B重算、所有受影響typed欄位與窄寫入契約，
並指定原版資料／正式回合／報告／保存的驗收方法。
正常玩家與GUI缺口不由局部診斷代替。研究索引皆保留未證實語意與停止線。
