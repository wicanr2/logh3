# 422：空分隊清理的原池與引用

2026-10-05，第88輪，遠端#79。[規格109](../spec/109-battle-empty-groups.md)。
第87輪的升級入口保存已位於8-1之後，無法證明清理輸入。
本輪從固定eng-t5重新到185C:001D前保存，沒有改CPU入口。
原始GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`，
bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。

## 可重現方法

```sh
bash tools/battle_cleanup_entry_oracle.sh a
bash tools/battle_cleanup_entry_oracle.sh b
bash tools/gui.sh xrun python3 tools/scripts/battle_cleanup_prepare.py
bash tools/battle_cleanup_oracle.sh a
bash tools/battle_cleanup_oracle.sh b
bash tools/gui.sh xrun python3 tools/scripts/battle_cleanup_review.py
```

腳本使用固定dosgolem f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c、乾淨工作樹，
Go1.24.13、tick20000。完整父命令由entry包裝器及每份meta.txt保存。
父RAM兩跑1MiB相同，SHA-256 `1575a57e438f8803a0fd0200268bc0e883a64d7210b4bcc42faab1e35f035ecf`。
父a保存SHA-256 `a62af47fb7d15a136e2e2ed51760ef0e9dd82538f68d5f7e0be4db5f2b0011b8`。
父保存的序列化含非排序映射，不以其全檔相同取代RAM相同。

控制器由第86輪固定來源產生，來源SHA-256
`895f92e7ac5f3431569f2b84cf5bf342114e249f87051f6d07345fafd6573b8b`。
生成觀測器SHA-256 `673cb53134aa548b85a19ac1942bf54add25fcdc782d802f73663e0e49d5ad39`。
兩側先恢復完整機器與DOS保存，再核RAM，重新恢復以消除VGA讀取的latch副作用。
每案前固定seed12345678、WORK+0D/+0E=1、顯示等待0，明示寫指定非旗艦SHIP+08=0。
注入逐項保存lin:地址及bytes，沒有重新選seed或跳到函式中間。

艦隊12／16／17／28／29／30／31各測無死亡、首個非旗艦、末組、全部非旗艦、交錯空組。
35案完整六表、戰場清單、原始指令入口bytes、參數與呼叫次序雙跑相同。
兩份6534142-byte JSON SHA-256
`80f108f6db41137c245dd52040616234779cf93349fbd7c0fae87d1aa2ef9949`。
沒有報告輸入，末端seed保持78563412，不消耗RND。

## 已證實事實

| 項目 | 原版證據 | 結論 |
|---|---|---|
| 清理條件 | 185C:0177與35案SHIP+08注入 | 只移除艘數0的非旗艦 |
| 逐次壓縮 | IDA 185C:021A、024F、0254／025A、0267／0275 | 減組數後搬完整四byte引用，尾槽不清 |
| 作廢 | 77個死亡池完整前後byte | 擁有者FF、+19清bit6、+38..3B清零 |
| 未重新編號 | 完整SHIP比較 | 存活組+03／+05保持，連續空組仍全部移除 |
| 其他欄位 | 完整六表與seed | 不重算能力／速度／費用，不動其餘四表、不取RND |
| 清單重建 | 12AA:03BC呼叫與DS:9894的164bytes | 七艦隊保留，55池引用扣除死亡數，偵察池仍按艦隊附後 |

以上等級proven，限明示診斷域。旗艦全滅與士氣報告不在新增35案範圍。
77個死亡池中，64個+19為3F，13個為C5，顯示指標皆非零。
清除+38..3B不能被誤寫成「只有bit6設置時才清」。

## IDA保存

主工具IDA Pro9.4，只在一次性COM建立資料庫，不改正式.i64。
COM為256-byte零header加父RAM線性185C0起4096bytes，SHA-256
`1b6b4aae3a0b652f48e8d3c24f863c39fde0b4c386ddf9afe7ab8c425d206bd7`。
沿tools/ida.sh raw執行idat -A，五段001D／00B2／0177／01ED／020C至ret，188指令。
原始operand及IDA／CS:IP／線性定位全部保留，直接語意逐項附於資料索引；未審查維持unknown。
IDAPython來源與輸出完整雜湊見[資料索引](../internals/data/battle-cleanup.json)。

## 實作根因與界線

舊battleFinalPhase只按Fleet.Composition.Ships移除結構化組，
原FLEET引用未即時壓縮，BattleUnit.Squadron也未重編。
交戰已寫回原池但Composition保留前值，故死亡組可能未被移除。
原池升級、補給與保存都會消費錯誤引用或舊資源；修正須在8-1提交，不能留待保存推導。
原版曲線、升級成長與全滅規則本輪不重開，8-5仍待原閘門。
正式驗收見[077](../playtest/077-battle-empty-groups.md)，本檔保留調查過程。

## 清理至升級的追加診斷

沿同一完整父保存另生四案，艦隊12與29各測首個非旗艦及全部非旗艦死亡。
對該艦隊提督明示一次足額功績20000、等級28；固定種子仍12345678，不挑結果。
從8-1自然執行至8-5入口，於185C:060D另保存升級輸入六表。
兩份1113851-byte JSON完全相同，SHA-256
`0e45c28c9cfe015a6db29aa8d6a37a9e51a2dc559a981cf0f3b9a317e822b3db`。
生成觀測器SHA-256 `32cf996fb4b16a61b63cc08d57cf0501b9b09ec336c520bb3c2dd9a091bf415c`。

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_cleanup_chain_prepare.py
bash tools/battle_cleanup_chain_oracle.sh a
bash tools/battle_cleanup_chain_oracle.sh b
```

每案一人升級，清理後組數分別7／1／6／1。
艦隊12兩案seed末值6AC77C90，艦隊29兩案253C349A。
正式Game分階段核升級前與後完整五表、存活池的士氣／三速度與seed相同，
proven限明示診斷鏈，不代表自然功績與正常GUI升級。

舊battlefinal.go「拿掉艦隊是唯一會讓表過期的事」被空組壓縮反例推翻。
舊歷史保留，現行程式移除此斷言。第一次標準回歸另發現兩組移動fixture只放兩個分隊，
但原FLEET仍多組；fixture現在明示同步一組上界，不放寬正式清理的資料完整性檢查。
