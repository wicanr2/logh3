# 079：勝方占領核心模型與非零守軍診斷

2026-10-05，第90輪，沿#79。這是固定原版與暫存規則模型的診斷，
未修改正式Game或rules。原版來源與工具定位見[研究424](../re/424-battle-finish-attack.md)，
分級來源見[battle-finish-attack.json](../internals/data/battle-finish-attack.json)。

| 核對 | 結果 |
|---|---|
| 既有12案有開火核心 | 完整行星、艘數／位置、旗標／士氣、計分／功績及seed相同 |
| 七案守軍原版雙跑 | 完整JSON逐byte相同，16次查詢／守軍量符合獨立模型 |
| 七案占領組合 | 同上述核心欄位及seed，模型差異列表為空 |
| 實際影響 | 近距與Out控制案SHIP終態差20 bytes、WORK差2 bytes |
| 第89輪基準 | 60案收據與48案無開火完整表保持，不覆寫 |
| 正式驗收 | 完整原池、production、正常玩家、報告renderer及保存仍未驗 |

## 重現

原始JSON不存在時：

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_finish_guard_prepare.py
bash tools/battle_finish_guard_oracle.sh a
bash tools/battle_finish_guard_oracle.sh b
bash tools/gui.sh xrun bash tools/battle_finish_attack_probe.sh
bash tools/gui.sh xrun bash -c 'export WINNER_PROBE_GUARD=1; bash tools/battle_finish_attack_probe.sh'
bash tools/gui.sh xrun python3 tools/scripts/battle_finish_attack_review.py
```

兩份原版固定dosgolem HEAD、IRQ0每20000步與seed12345678，父保存與原版唯讀。
所有座標、Out及艘數注入均以lin位址及完整bytes保存，報告press／release仍由原控制器自動輸入。
模型在/tmp複製rules，退出trap移除；來源雜湊檢查拒絕別的strategicassault.go。
輸出拒絕覆寫。第一版模型與第二版基準逐byte相同，19案核心差異全部為零。

## IDA查詢

`battle_finish_attack_ida_prepare.py`從固定父RAM產生六個COM與匯出器。
每個COM先複製到容器/tmp；透過`tools/ida.sh raw`呼叫idat，外層timeout90s。
`-S`共用`/work/finish79-r90-ida.py`，輸出、segment與範圍如下：

| 輸出JSON | segment | 範圍 |
|---|---|---|
| finish79-r90-0DEA.json | 0DEA | 1759:ret |
| finish79-r90-1261.json | 1261 | 0067:ret、032B:ret |
| finish79-r90-17A0.json | 17A0 | 00AC:ret |
| finish79-r90-203A.json | 203A | 0424:ret、04D1:ret、0593:ret、0743:ret |
| finish79-r90-203A-full.json | 203A | 0593:0743 |
| finish79-r90-20B4.json | 20B4 | 0D96:ret、0E7E:ret、103D:ret、15E1:ret、1662:ret、18C2:ret |
| finish79-r90-297B.json | 297B | 08D8:ret、0ABF:ret |

203A完整範圍補回艦種7早退後的主體，保留早期匯出，不把輸出檔存在或工具exit0當語意證據。
每筆原始位址、bytes、operand、xref與分級保存於索引。

標準產品回歸18套件13 PASS、5無測試，日誌在workplace/finish79-r90-standard.log；Docker／擁有權收尾見WORKLOG。
110與完整062保持DRAFT，#79保持OPEN。
