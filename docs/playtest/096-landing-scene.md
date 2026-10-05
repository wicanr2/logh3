# 096：原版登陸戰逐格收據

2026-10-05，第108輪，沿 [#48](https://github.com/wicanr2/logh3/issues/48)。
分類：原版固定完整保存續行與明示分支注入。沒有新增remake same-state。
方法與原始位址見[研究436](../re/436-landing-scene.md)，原始身份及各檔雜湊見[來源索引](../internals/data/landing-scene.json)。

## 重生與核對

在專案根目錄執行`bash tools/landing_scene_oracle.sh <案例> <a|b>`。
案例為baseline、failure、capture、off、click。每次均使用專用dosgolem完整父state，容器/tmp編譯探針。
預設無網路、UID1000、4 GiB／4 CPU／256 PID與5分鐘上限，原版資料唯讀。
既有輸出拒絕覆寫。要重建既有批次，先另存收據；不能將探索輸出覆蓋正式收據。
完整父state在`/home/anr2/cht/dosgolem-logh3/workplace/fcmd/eng-t5.state`，SHA-256見436。
缺原版或父state時包裝器失敗，不用自製替身。

核對入口：`bash tools/gui.sh xrun python3 /src/tools/scripts/landing_scene_receipt.py --check`。
索引只以`--freeze`新建一次，後續核對不改寫。

## 原版結果

| 案例 | 每跑原始檔數 | 保存畫面 | 逐格繪製返回 | 事件 | 最終步數 |
|---|---:|---:|---:|---:|---:|
| 防衛基準 | 120 | 27 | 20 | 59 | 1187924209 |
| 登陸失敗 | 248 | 59 | 39 | 122 | 1252923503 |
| 登陸成功 | 148 | 34 | 18 | 76 | 1202723021 |
| 關動畫 | 100 | 22 | 2 | 73 | 1197123834 |
| 動畫中排入滑鼠 | 248 | 59 | 39 | 126 | 1199269708 |

每案a／b全部原始檔逐byte相同，另有各一份report，僅命令列的批次字元不同。
兩跑各201張觀測畫面、118次逐格繪製返回。畫布640×480，沒有裁切、縮放或排除區。
輸出位於`workplace/oracle/landing48-r108-r5-<案例>-<批次>`，click使用r6。
各張包含完整1048576-byte RAM、307200-byte色號、色盤與原生PNG，parent／phase6-input／end另外保留。
此處201是觀測畫面總數，不等於201個獨立動畫格。

failure傷害518，防衛力6000→5482、owner仍0；capture傷害1、防衛力1→0、owner變1。
成功不反擊，失敗有登陸、反擊及占領失敗共三次90-tick等待。
成功有登陸及占領成功兩次等待，音效22／2；失敗音效22／3／17／3。
關動畫沒有整段COMANI或其音效，兩次逐格繪製是共用底圖／返回，三次結果等待仍在。

click與failure按畫面名稱逐一核對，59張整圖、色盤與PNG相同，共18124800色號像素零差。
按下只縮短結果等待，不略過39次逐格繪製；四個回呼恢復原入口另標callback_resume，不重算為新COMANI呼叫。
click／off最終FLEET3904、SHIP27456、PLANET246 bytes及seed皆與failure相同。
沒有由等待步數推算DAC／PIT實機時序，也沒有宣稱實體音效或人耳驗收。

## 產品與完成範圍

標準18套件13PASS／5無測試，Game124.523秒、gin3 60.009秒。
第107輪3870來源與164原始輸入重核保持；正式Go、AI與存檔格式沒有改動。
118保持DRAFT，完整056／062不升格，#48保持OPEN。
下一閘門是先反擊／帝國／多人名冊及完整單幀快照，再接正常Game與GUI／保存。
本批原版探索失敗與探針修正全部保留，不能以這份收據聲明正常玩家登陸演出已完成。
