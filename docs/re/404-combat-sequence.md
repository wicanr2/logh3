# 第66輪：距離曲線、多分隊排序與累積戰損

2026-10-04，沿[#79](https://github.com/wicanr2/logh3/issues/79)及[#48](https://github.com/wicanr2/logh3/issues/48)。
父提交b61764b4e524e9885ffb799b0be870a726e405b9，正式玩法未改，092仍DRAFT。
入口為[本輪計畫](../goal/2026-10-04-combat-sequence-round-66.md)、
[原版診斷056](../playtest/056-combat-sequence.md)及[完整收據](../internals/data/combat-sequence.json)。
收據SHA-256：199f328b4463d49a817e54c2ade49dd148a24d45e94e94d135488d699c3205a8。

## 來源與位址

- 原版bundle：org_game/银河英雄传说3SP.zip，SHA-256 de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673。
- GIN3.COM：SHA-256 770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918，原始檔唯讀，未patch。
- dosgolem專用副本：f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c，工作樹乾淨。
- Go、GUI、Python及IDA image ID完整保存在收據images，不以tag代替身份。
- 本文函式位址為原版CS:IP。IDA9.4 EA以0x10200為COM程式基址；dosgolem線性位址另列runtime_linear_hex。
- 固定RAM來源normal79-r61-a/end-mem.bin，SHA-256 6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4。
  新297B COM及三份沿用IDA匯出的全部原指令bytes均與此RAM核對。

## 零距離讀未初始化的前項指標

證據等級proven，只限固定原指令及受控診斷。
203A:00EC在currentX小於q時才寫local前項far pointer。
0161的jge讓q=0在第一筆X=0直接跳至017E，跳過0163起的初始化。
017E讀BP+14／+16的殘留指標；不能把它解釋為curve[-1]。

實際艦種表映射的11種武器類型為0、1、2、3、5、6、7、8、9、10、11。
每型各有距離0及20雙跑。型別4曲線存在但未由本次艦種表映射，不拿它冒充可用武器。
固定零距離樣本的前項為00D7:0001，線性0xD71，11案攻擊力均0。

另在型別0的203A:012A之後，只把前項指標改為54B3:45CC。
注入為lin:65A7A@506950550=CC 45 B3 54，完整記在plan與flags。
與未注入案例相比，兩筆SHIP78、兩筆FLEET122、兩筆ADML28及seed4 bytes均相同。
攻擊力卻由0變18，損失由25變145。此案是局部stack反例，不能稱正常玩家結果。

使用者已裁決：**remake零距離取曲線首點，記錄為原版差異**。
這修正原版未初始化stack造成的不穩定結果，不改存檔格式。
本輪模型仍重建原版實際stack結果，正式穩定規則尚未接線。

## 排序在跳過未下令分隊之前

證據等級proven，限本輪同艦隊八筆一般分隊及兩筆戰鬥機命令。
FLEET編制原序為80、82、83、84、85、87、86、88。每筆都先取兩個RND，之後才跳過武器FF／3。
排序鍵為：

```text
RND()%base + base
+ RND()%ADML[提督].byte(+0B) + ADML[提督].byte(+0B)
+ SHIP.byte(+06) + FleetMobility
base = DS4426[kind.byte(+09)]
```

297B:083F的加值沿[既有機動力公式](../internals/56-fleet-composition-panel.md)。
新IDA原指令確認取提督／副官+17較大值；副官FF時只取提督。
中間乘積取低word，signed idiv100，最後100減中間項，以AL byte返回。
本次能力106、編制8的DS4E2E係數49，得到72，八次0522回值全部相同。

| 池 | 排序鍵 |
|---|---|
| 80 | 325 |
| 82 | 332 |
| 83 | 294 |
| 84 | 387 |
| 85 | 339 |
| 86 | 354 |
| 87 | 365 |
| 88 | 352 |

梳狀排序以unsigned word降序，完整結果84、87、86、88、85、82、80、83。
跳過FF後只剩82、80。不能由UI先下令80推斷先打80。
末端SHIP+42已被其他用途覆寫成1，必須在0539／033E邊界讀鍵與排序結果。

## 多目標與同目標

原版UI提交80→102，再提交82→104或82→102，武器都為2。
兩次傷害入口的BP+08／+0A是攻方far pointer，+0C／+0E是目標；不能用殘留SI／DI代替。

| 案例 | 第一擊 | 第二擊 | 末端 |
|---|---|---|---|
| 兩目標 | 82→104，損636 | 80→102，損527 | 104剩2364，102剩2473 |
| 同目標 | 82→102，損726 | 80→102，損374 | 102由3000→2274→1900 |

模型從原始編制順序、預設seed12345678獨立生成鍵與排序，未拿原版命中順序當模型輸入。
兩案逐擊攻擊／防禦／射擊量／百分比／requested／actual及中間艘數、修理上限、彈藥符合。
分數逐擊除20後相加為57／54；提督功績581／550、副官290／274。
末端seed均95190066，來自16次排序與2次傷害RND，正式遊戲不鎖此seed。
計分表的陣營為目標的敵方：8FE4+(1-targetFaction)*4及8FEC+(1-targetFaction)*4。

## 取消及驗證工具勘誤

cancel-outer在選目標等待時按右鍵，觀測到沒有提交、演出、傷害或返回181B:0385，seed保持。
本次只能宣稱仍在選擇等待，不能稱整回合取消。武器子選單右鍵FF返回分隊等待的既有第63輪契約保持。

- 單擊probe初稿只驗帝國計分格，改成目標陣營後仍錯；沿203A:06EB／3982:0250改成目標敵方陣營後23案通過。
- 多擊probe漏讀SS:SP中的SS，排序陣列讀到線性0；補實際SS欄位後修正。
- 同目標排序比兩目標晚13264步。same-target-ranked保存錯相位RAM，不用它驗排序。
  same-target-measured改用0539=508262599、033E=508264696；兩個傷害入口仍為523550689／561173442。
- 收據凍結器最初把第63輪舊schema當成新schema；現從實際flags核parent_state_sha256並保存祖先計畫。

這些是診斷工具與擷取設定問題，不寫成產品缺陷。舊輸出保留，不改原版或已凍結收據。

## 本輪邊界

35組70次原版執行，189組完整RAM／色號／色盤／PNG逐byte相同；830份來源、1532條IDA原指令、21項proven保存。
18產品套件正常結束，13有測試套件通過、5無測試檔。返回戰場兩圖已目視。
沒有新的remake same-state結果。死亡、其他取消分支、完整演出owner、正式Game／UI／保存仍缺。
092保持DRAFT，#79／#48保持OPEN；061單人CPU被動政策不變。

## 2026-10-04第67輪實作續記

本節追加實作現況，不改寫第66輪原版收據或stack模型。沿
[093子規格](../spec/093-fleet-combat-core.md)先審查READY後實作
[純規則核心](../../internal/rules/fleetcombat.go)。零距離已取曲線首點；
凍結第66輪JSON的production_implemented=false仍是當輪事實。

[正式測試](../../internal/rules/fleetcombat_test.go)直接讀原版AX／BP及RAM作期望，
覆蓋31戰損、11近距離、合法戰鬥機與兩組連打。另12個零距離／stack反例只驗核准首點，
不稱原版未初始化stack戰損parity。12組完整六點及排序／機動表由既有RAM再讀確認。

首輪測試要求合法戰鬥機收據不存在的0C6D等係數觀測，因而失敗。
這是驗證腳本問題。該收據改核實際保存的呼叫數與末端資料，其他案例仍逐係數核原版。
同一容器、同一命令乾淨重跑通過，沒有更改原版、正式算式或golden資料。

索引與來源雜湊保存於[第67輪純核心收據](../internals/data/fleet-combat-core.json)，
可重建方法見[056追加驗證](../playtest/056-combat-sequence.md)。
正式Game／UI／保存及完整演出待接，093 READY、092 DRAFT；#79／#48保持OPEN。
