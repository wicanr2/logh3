# 第85輪：戰場升級的重算與接線

沿遠端#79，父提交78ee450c34485a09109879de0ae548c1f4f66c4e。
開工遠端89張worklist／9張OPEN，269條verify零失效，工作樹乾淨。
第84輪的60案原版雙跑與成長模型已保存，108仍DRAFT。

本輪先補升級後297B:05B2／0E8B的能力、編制容量、速度、完整度與費用。
以既有原版完整三表及受控亂數驗算，不重做已證實的通用公式。
證據審查足夠後將108轉READY，再接8-3之後的正式升級與逐則通知，
核資料、報告矩形、保存與正常玩家抽樣。8-5及完整062保持原門檻。

工具與新證據掛回[事實索引](../internals/00-README.md)。
既有使用者裁決與#89 pending保持，不另定忠實度選項。

重算觀測器以[產生器](../../tools/scripts/battle_promotion_recalc_prepare.py)
從固定第84輪來源產生，再走[雙跑入口](../../tools/battle_promotion_recalc_oracle.sh)
與[容器內入口](../../tools/battle_promotion_recalc_native_run.sh)。
新增WORK、STELLAR與每次重算後完整表，另加六個受損／零艘數診斷。
重算指令由[固定RAM切段](../../tools/scripts/battle_promotion_recalc_ida_prepare.py)
與[IDA查詢入口](../../tools/battle_promotion_recalc_ida.sh)重生。
由[獨立重算模型](../../tools/scripts/battle_promotion_recalc_receipt.py)
驗算並凍結[重算收據](../internals/data/battle-promotion-recalc.json)。
[研究419](../re/419-battle-promotion-recalc.md)與
[對拍074](../playtest/074-battle-promotion-recalc.md)保存結果及剩餘閘門。
現行模型為[第二版](../../tools/scripts/battle_promotion_recalc_receipt_v2.py)，
[現行來源](../internals/data/battle-promotion-recalc-v2.json)補核全局移動費。
第一版未追callee的語意由第二版取代，舊收據保留，108仍DRAFT。
