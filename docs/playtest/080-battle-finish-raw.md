# 080：勝方占領完整原池診斷

2026-10-05，第91輪，沿#79。固定原版入口與暫存模型的診斷，尚非正式Game或正常玩家收據。
來源、位址空間及勘誤見[研究425](../re/425-battle-finish-raw.md)，
固定來源見[原池索引](../internals/data/battle-finish-raw.json)。

| 核對 | 結果 |
|---|---|
| 兩案逐指令原版雙跑 | 完整JSON相同，七次目標欄位值變動僅定位清FF；六表／清單與既有收據相同 |
| 12案原版占領與七案守軍 | 暫存模型完整六表與清單共850972 bytes相同，seed相同 |
| 同19案正式地形 | 2226次通行判定，全表再比850972 bytes零差；沒有地形拒絕案例 |
| 規則與原池範圍 | 座標／目標、索敵池、揭露、庫存、修理上限、士氣、兩項費用、全陣營維持費、計分及功績 |
| 剩餘正式驗收 | Game續接、報告renderer、typed同步、存讀及正常玩家勝負 |

## 重現

輸出不存在時，依序執行：

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_finish_raw_prepare.py
bash tools/battle_finish_raw_oracle.sh a
bash tools/battle_finish_raw_oracle.sh b
bash tools/gui.sh xrun python3 tools/scripts/battle_finish_raw_model_prepare.py
bash tools/gui.sh xrun gofmt -w workplace/finish79-r91-model-test.go
bash tools/gui.sh xrun bash tools/battle_finish_raw_probe.sh
bash tools/gui.sh xrun bash -c 'export WINNER_PROBE_GUARD=1; bash tools/battle_finish_raw_probe.sh'
bash tools/gui.sh xrun bash -c 'export WINNER_PROBE_TERRAIN=1; bash tools/battle_finish_raw_probe.sh'
bash tools/gui.sh xrun bash -c 'export WINNER_PROBE_TERRAIN=1 WINNER_PROBE_GUARD=1; bash tools/battle_finish_raw_probe.sh'
bash tools/gui.sh xrun python3 tools/scripts/battle_finish_raw_review.py
```

沿固定父保存與dosgolem HEAD，原版輸入唯讀，注入、完整步數、輸入與寫入定位均保留。
暫存模型以父提交的strategicassault.go為固定來源，退出trap移除/tmp副本，拒絕覆寫輸出。
來源變動後需依索引父提交取回相同輸入；不能用新正式程式冒充本輪暫存模型。
第89／90輪凍結收據保持；本輪原版JSON雜湊190a1aa4a1c633be6fc4bee035dd696bbb80ddea021355ffa534df4fe5bc8c52。
Go PASS與完整bytes核對分開。標準回歸與Docker收尾見WORKLOG。
110的實作驗收仍待續，完整062保持DRAFT、#79 OPEN。

另補兩案callback前後雙跑，一個實際占領報告的完整六表與清單及seed不變；
返回後提督／副官才加3300／1650，見[續接來源](../internals/data/battle-finish-cursor.json)。
重生入口為battle_finish_cursor_prepare.py、battle_finish_cursor_oracle.sh a／b及battle_finish_cursor_review.py。

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_finish_cursor_prepare.py
bash tools/battle_finish_cursor_oracle.sh a
bash tools/battle_finish_cursor_oracle.sh b
bash tools/gui.sh xrun python3 tools/scripts/battle_finish_cursor_review.py
```

110已升READY，正式規則／Game及CONFORMED仍待實作驗收。
