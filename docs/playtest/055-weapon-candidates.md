# 055：武器候選與合法戰鬥機原版診斷

日期2026-10-04。事前分類為受控原版診斷，未宣稱remake same-state或正常玩家驗收。
固定來源、工具版本及位址語意見[研究403](../re/403-weapon-candidates.md)。
[收據JSON](../internals/data/weapon-candidates.json)的SHA-256為
59a5ad2bd68a54c888179be3d85a6b99c15f16a1892b7d2b05c4de0667ad601f。

## 執行入口

```sh
bash tools/weapon_candidates_oracle.sh weapon79-r65-near-a
bash tools/weapon_candidates_oracle.sh weapon79-r65-command-a
bash tools/weapon_candidates_oracle.sh weapon79-r65-selected-a
bash tools/weapon_candidates_oracle.sh weapon79-r65-fighter-submit-a
bash tools/weapon_candidates_oracle.sh weapon79-r65-fighter-phase-end-a
bash tools/weapon_candidates_oracle.sh weapon79-r65-fighter-end-a
bash tools/go.sh run tools/scripts/weapon_candidates_probe.go > workplace/round65-probe-final.json
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 bash tools/py.sh /work/tools/scripts/weapon_candidates_receipt.py /work/workplace/new-receipt.json
```

工具拒絕覆寫已存在run、flags與receipt。已執行的a/b不重跑、不改名移植舊收據。
新環境須先依052重生固定父快照並核SHA；目前腳本有意綁這份固定入口。

48個入口案例均各雙跑，命令續接另有五組；以JSON的case及完整flags為準。
上列是合法戰鬥機操作的a側入口範例，完整凍結另須b側及其餘案例。
每組保存兩份RAM各1048576 bytes、640×480原生PNG、色號、色盤、完整report／meta及state指紋。
53組265個檔案逐byte相同。沒有裁切、遮罩或整機state一致聲明。

## 原版操作

| 步數 | 操作／狀態 |
|---|---|
| 484357728 | 唯一入口注入：seed12345678、艦隊12戰鬥、只讓自身80可命令、目標102右方20、Detail2、Out7F、朝向0 |
| 485999999 | 標頭父快照；原版候選mask7 |
| 486400000 | 虛擬滑鼠240,230，左鍵保持100000步，關標頭 |
| 490300000 | 虛擬滑鼠268,208，左鍵保持100000步，選目標102 |
| 495300000 | 虛擬滑鼠270,250，左鍵保持100000步，選戰鬥機第三列 |
| 495599999 | 原版已提交武器2／目標102；seed與庫存仍保持 |
| 508573441 | 單擊0C54入口，原版已完成8筆排序16次RND |
| 535946392 | 返回181B:0385，606艘損失已完成 |
| 699999999 | 戰場Input第2回合，分數各增30 |

後三個點擊皆有200000步移動前置。滑鼠採640×400虛擬空間，本局面PNG的Y另加40。
鏡頭只適用本局面：世界X為滑鼠X+176，世界Y為滑鼠Y+80；不作全域轉換常數。
完整初始注入清單在每個plan，包括一般分隊與偵察池對照，不能只重建表中摘要。

已目視selected-a選單三項可用，以及fighter-end-a的Input與分數。
沒有remake原圖／差分，因此本文件不能當畫面對拍完成。
標準產品測試18套件正常結束；完整命令、Game及保存仍依092 DRAFT接續。
