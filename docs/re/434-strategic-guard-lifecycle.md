# 434：戰後清單跨入下一星域

日期：2026-10-05。工作項 [#56](https://github.com/wicanr2/logh3/issues/56)。

## 固定輸入與方法

沿用[研究433](433-strategic-assault-edges.md)的原版身份。
GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
dosgolem `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，乾淨工作樹，Go 1.24.13。
完整父保存 `workplace/oracle/cleanup79-r88-entry-a/entry.state` 的 SHA-256
`a62af47fb7d15a136e2e2ed51760ef0e9dd82538f68d5f7e0be4db5f2b0011b8`。
從正常8-1邊界續行，IRQ0每20000步；沒有CS/IP注入或人工返回。

可重建入口為[產生器](../../tools/scripts/strategic_guard_lifecycle_prepare.py)、
[原版包裝器](../../tools/strategic_guard_lifecycle_oracle.sh)與
[容器控制器](../../tools/strategic_guard_lifecycle_native_run.sh)。先在gui.sh容器執行產生器，
再跑原版包裝器a、b；輸出使用排他建立，不能覆蓋先前收據。
v6控制器 SHA-256 `c4930096819aefe76b7039d2fca9b0d255ccdad6a3c3786cbb0018d65fe3771b`。

兩案均在8-5前固定seed12345678，令帝國艦隊12失去戰意、六支同盟艦隊保持可戰。
下一星域案另將既有帝國艦隊0放到星域27，明設該星域bit4。
每次注入的線性位址與bytes完整保存在JSON；報告輸入亦保存步數與落點。
這是明示注入的原版控制流證據，不宣稱自然遊玩的全程同狀態。

## 已證實

| 比較點 | 結果 |
|---|---|
| 12AA:03BC重建 | 先55→0，再逐筆回55；共56次筆數寫入 |
| 勝方20B4:1512 | 參數低byte為星域26，callback185C:0B18 |
| 185C:001B戰後返回 | 原有164 bytes清單逐byte保留 |
| 下一星域20B4:1497 | 原始堆疊參數為星域27，正常呼叫者0374:05AB |
| 下一星域20B4:1512及1588載圖 | 清單仍逐byte相同，沒有重建或清零 |
| 20B4:1497返回 | 清單仍55筆，完整原SHIP池在每個觀測點保存 |

原始指令位址空間均為dosgolem執行時段:位移。筆數位址DS:9896，線性lin:5E3C6。
三個寫入定位為12AA:03DC、0441、045D；原始12 bytes與每次前後值在收據內。
原版沒有按星域重篩清單，也沒有用所有艦隊替代清單。守軍查詢仍依433已證實的
原池陣營、種類、Out與現在位置篩選，讀取時才計算行星半徑。

`workplace/assault56-r100-lifecycle-v6-a.json`與b完整相同，兩案分別續跑15840144與7077921步。
兩檔 SHA-256 均為`d5689d017f511288e74080ca91f25e22d20112a54d47d048b1c88f38d9e1afd0`。

## 探索修正與限制

v3基準案把返回地圖入口設成3705:0000，未命中而耗完固定預算。
v4補堆疊後證實第一次1512是星域26勝方入口；不能稱下一星域。
v5才取得真正星域27入口，v6再跑至其返回並完成雙跑。
這些是探針停止點修正，沒有改執行器或產品，也沒有增加預算或重擲seed。

本輪清單來源足以支持[115](../spec/115-strategic-assault-guard.md)正式接線。
這份原版控制流證據不取代正式Game、保存與正常玩家驗收，也不證明六表完整副作用。

## 正式原池接線補驗

typed接線五案指定欄位先全同，但保存前後完整比較仍各有FLEET7、SHIP88 bytes差異。
差異為朝向、目的座標與分隊工作欄位；既有110已核的共用1512原池常式包含這些寫端。
[原池探針](../../tools/strategic_guard_raw_probe.sh)只在容器副本放寬池必屬目前星域的驗證，
以回合0跑同五案，完整六表／清單與seed均零差。兩個原版批次獨立重建相同，
JSON SHA-256 `cebc4dfd0ac7e09708f25aced9fa1866bbf5cb63d82f7a1ef99116e90aefe856`。

115增補READY契約後，正式NewStrategicOccupation共用原池常式並固定回合0；
戰場勝方入口保留同星域及正式引用檢查。Game沿現況Slot讀寫，重讀艦隊，功績／分數只加一次。
正式規則與Game五案逐byte核完整六表／池序／seed，六檔讀回、壞索引停止、戰後清單保留亦通過。
驗證及聲明範圍見[088](../playtest/088-strategic-guard.md)。

收尾發現058正文頂部已更新，表格仍留三條過期限制，已同步修正。
首次377來源快照保持於`workplace/assault56-r100-guard-first-index.json`，SHA-256
`7b8445120660c45fcd0b59678fba23a2e799c88cd7a67e3227d74d04d67dcb88`；它是本輪中間輸入。
現行索引重建時包含該快照身份，沒有覆寫原版、舊輪或本輪首次凍結資料。
