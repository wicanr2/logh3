# 074：戰場升級重算原版診斷

工作：[Issue #79](https://github.com/wicanr2/logh3/issues/79)，[108仍DRAFT](../spec/108-battle-promotion.md)。
這份收據比較固定原版雙跑與獨立模型，沒有remake畫面、正常玩家或完整回合同態聲明。
來源、原始定位、公式與雜湊見[研究419](../re/419-battle-promotion-recalc.md)。

## 結果

| 核對 | 結果與範圍 |
|---|---|
| 原版決定性 | 66案，兩份20108439-byte JSON完全相同 |
| 成長與亂數 | 121次嘗試、53次成功、473個LCG輸出及末端seed均符合 |
| 逐次資料重算 | 53個返回點，完整ADMIRAL／FLEET／SHIP逐byte相同 |
| 最終資料 | 66案完整三表相同，未修改欄位也納入比較 |
| 費用與造艦預估 | 53次兩陣營維持費／移動費相同；造艦預估以原版後稅額為輸入核對相同 |
| 報告 | 原版保留11則；沒有本輪remake報告或差分圖 |
| 全局經濟／稅／未安定度 | 原版完整WORK／STELLAR與前後差分已保存，尚未獨立模型 |
| 產品與GUI | 本輪未改產品，正式升級、軍階UI、保存與正常玩家驗收待續 |

涵蓋第84輪60案，加兩陣營各半數／一成／零艘三個診斷。
各案固定seed於執行前，不重擲；損害與功績以明示lin位址注入。
能力255、等級100／255及零艘均為helper診斷，不外推正常生涯。
11則報告在兩輪觀測器中均保留；畫面樣本來源仍沿073，不升格。

## 重生命令

先確認原版資料、固定父保存與第84輪產生的Go檔存在；缺來源即受阻，沒有替代資料。
所有產生器與原版入口拒絕覆寫現有收據。新一輪重跑須另定版本及輸出名。

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_promotion_recalc_prepare.py
bash tools/battle_promotion_recalc_oracle.sh a
bash tools/battle_promotion_recalc_oracle.sh b
bash tools/gui.sh xrun python3 tools/scripts/battle_promotion_recalc_ida_prepare.py
bash tools/battle_promotion_recalc_ida.sh
bash tools/gui.sh xrun python3 tools/scripts/battle_promotion_recalc_receipt_v2.py
bash tools/gui.sh xrun python3 tools/scripts/battle_promotion_recalc_annotations_v2.py
bash tools/battle_promotion_recalc_review_ida_v2.sh
bash tools/gui.sh xrun python3 tools/scripts/battle_promotion_recalc_review_receipt_v2.py
```

[原版入口](../../tools/battle_promotion_recalc_oracle.sh)檢查專用dosgolem固定HEAD與乾淨工作樹，
使用該專案tools/go.sh，原版唯讀掛載，UID1000，有CPU／記憶體／PID與5分鐘逾時。
IDA沿既有tools/ida.sh與locked-v1，兩批各90秒；暫存資料庫有trap，未修改遊戲輸入。
查詢範圍297B為0209:025D、05B2:083F、0E8B:10DE、10DE:116C、116C:12FB、12FB:131C；
090A為0E6D:0E88及0E88到0F84 retf。

原版兩份JSON放`workplace/promotion79-r85-recalc-{a,b}.json`；
SHA-256 `5049465c845318759894b9b1d957c3a2fbc807998608eae1eb0776fb84988c97`。
[第二版來源收據](../internals/data/battle-promotion-recalc-v2.json)含22項來源，
SHA-256 `cda14b5b122e4e875a825948c61194c96b534a07c7a563c770b26570b35291e6`。
[第二版分級收據](../internals/data/battle-promotion-recalc-review-v2.json)含8項來源，
SHA-256 `844038b4feed02e0c1ecb87cfd18f06832b2ecb37639946d9e8db916eeb29a9a`。
808條原始指令bytes相同，42項proven語意已自動合併，其他仍unknown。
第一版漏追全局移動費callee；保存舊來源並由第二版取代，勘誤見研究419。

本輪沒有產品變更，標準產品測試沿第83輪最近通過結果，不重跑或擴大聲明。
本輪來源、工具語法、連結與入口、擁有權及Docker收尾另由WORKLOG記錄。
