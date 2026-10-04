# 108：戰場提督與副官升級

狀態：DRAFT。遠端#79，完整062仍DRAFT。

玩家在第8步索敵回收後取得提督／副官升級通知。
第85輪已補原版66案雙跑、53次完整ADMIRAL／FLEET／SHIP重算與473個LCG輸出。
全局稅收／星域、typed回寫、正式報告與正常玩家存讀仍缺，不得據此實作。
原版會改SHIP速度與分隊士氣，不能只更新ADMIRAL。

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

297B:05B2／0E8B的能力、容量、三項速度、四項最小值、完整度與旗標已補證。
090A:0E6D會更新全局費用與造艦預估，不只是升級艦隊本身。
公式、原始bytes與66案邊界見[研究419](../re/419-battle-promotion-recalc.md)、
[原版診斷074](../playtest/074-battle-promotion-recalc.md)。
第二版來源SHA-256 `cda14b5b122e4e875a825948c61194c96b534a07c7a563c770b26570b35291e6`。
速度先截byte、折扣乘積取signed16低word；費用保持unsigned32乘除與低word保存。
零艘診斷先寫完整度100再設bit3，不用remake既有完整度helper的零回值猜補。

READY前須核本局面全局經濟力／預估稅／未安定度，列出所有受影響typed欄位與窄寫入契約。
正式8-3返回後按參戰順序，提督升級成功即重算／報告，再試同一人，副官最後處理。
每則報告等待用原版word的訊息速度乘二，返回後恢復設定，不改玩家永久設定。
驗收需比較每次三表、全局更新、報告文字／矩形／頭像／時點、軍階UI及存讀；
正常公開入口補抽樣，不以功績注入的診斷代替正常玩家證據。
正常玩家與GUI缺口不由局部診斷代替。研究索引皆保留未證實語意與停止線。
