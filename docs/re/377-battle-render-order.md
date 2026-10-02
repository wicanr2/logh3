# 377：第二戰場分隊繪製鍵與殘留列

工作項[#79](https://github.com/wicanr2/logh3/issues/79)，[第37輪](../goal/2026-10-02-battle-render-order-round-37.md)。
比較器證據使用第36輪43邊界原版雙跑；dosgolem `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`、clean。
GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
入口RAM SHA-256 `9095dfa08ab36201784e6d4555b5e42312ac76bb20049614cf53bf93bc09e7ec`。

## 比較器

既有IDA Pro 9.4 image、UID1000，從入口RAM切35D1段，前置256零byte，offset0對應IDA0x10200。
切段SHA-256 `3001278d30a67b0eb7dc5f0bd03a1597d501d8450515855329290a226b88218b`。
一次性匯出`workplace/ida/re79-render-r37-35D1.json`，SHA-256
`49d0d57a04d127bd6957aefb1fd2b2a9c5de471b74bce5b731339cd350302559`。
原始CS:IP、dosgolem線性位址與IDA位址在[分級索引](../internals/data/battle-draw-order.json)分欄。

| 原版定位及bytes | 語意 | 等級 |
|---|---|---|
| 35D1:02B4 `26 8a 47 03`、02BE `26 8a 4f 03` | 讀兩物件+3的byte | proven |
| 02C2 `3a c1`、02C4 `73 05`、02C6 `b8 ff ff` | unsigned較小回傳-1 | proven |
| 02DF `3a c1`、02E1 `76 05`、02E3 `b8 01 00` | unsigned較大回傳1 | proven |
| 02EE `26 8b 47 04`、02F8 `26 8b 4f 04` | 優先值相等時讀+4的word | proven |
| 02FC `3b c1`、02FE `76 05`、0300 `b8 ff ff` | unsigned較大回傳-1 | proven |
| 0319 `3b c1`、031B `73 05`、031D `b8 01 00` | unsigned較小回傳1 | proven |
| 0322 `31 c0` | 兩鍵相等回傳0 | proven；通用排序器的相等鍵置換unknown |

全部43邊界DS:EADE為153，DS:EAE0完整清單均按+3昇序、+4降序。
其中135個分隊物件優先值均128，word+4低byte與SHIP+1艦隊索引相同，高byte與SHIP+3序號相同。
每幀135鍵全部唯一，包括0xFF偵察分隊。[核對工具](../../tools/scripts/battle_order_evidence.py)保存9筆分級指令與全部分隊鍵。
可丟棄測試只重畫排序後的繪製副本，規則次序未改。43幀從434..664降到各38點，Field全相同。
方向440點及整份FLEET／SHIP維持相同；探測已移除，log `workplace/battle79-round37-order-probe.log`另存。
第一次資料核對混淆圖塊索引，第二次錯用SHIP+0；回讀既有欄位契約後改用SHIP+1，全部相符。
兩次失敗是驗證腳本錯誤，不分類成產品缺陷。

## y203

S1該列位元模式`CE CE CF 56 CF 56 CF 56 CF 56 CF 56 CF 56 CF 56`，
在入口RAM唯一匹配dosgolem線性0x70EE6。匹配為proven，來源關係仍hypothesis。
S3舊遮罩在本RAM沒有匹配，不加星域特製遮罩。
2376切段SHA-256 `6f8f3986efe5936572c28a40dd26496d5fca3b8ed2303a69bb5f1e61a2106a3e`；
IDA匯出`workplace/ida/re79-render-r37-2376.json`，SHA-256
`df68ba2f4a0d3815fb26ed453805d51aaab0198eb8e3c528e72caf6820e1afc0`。
10B7仍有19列／20列合成，不能據此推論本S1列的寫入者。
從固定menu.state跑至1239379061。初次缺-vram-sites，讀CLI後以同輸入補-vram-sites -vram-at 3FAC，
完整報告仍未列寫入者。僅代表本次續跑沒觀測到，不能宣稱不存在寫入端。
探索輸出`workplace/oracle/battle79-round37-s1-strip-{explore,sites}/`保留flags與身份，不升格正式對拍。
下一步需在戰場初畫的更早父狀態觀測，不能以RAM匹配猜補。

本輪隨後從固定750M戰略過場續跑至900M，捕捉該位址三次2376:10EA／110B／1132寫入，
另有一次0B74與0B90，完整flags在`battle79-round37-s1-initial-strip`。
移除過晚的regs-from後，`battle79-round37-s1-strip-early`記錄664次10B7輸入；
但預估的格起點DI=39BC沒有匹配，需先校正實際DI轉換，仍未解出最後來源。
所有探索的輸入與輸出雜湊均綁進第37輪完整收據，不把零匹配當成沒有呼叫端。

## 規格審查

[085](../spec/085-battle-unit-draw-order.md)先DRAFT；原始比較器、43幀唯一鍵與可丟棄探測相符後審查，
限定分隊繪製排序，轉READY才接正式程式。不深入通用排序helper，不外推其他物件的相等鍵效果。

正式修正後43幀Field均零差異，新增斷言的修正前43次失敗log保留。
全畫面仍每幀38點，y203未知，不改046遮罩或golden。085只在分隊繪製範圍CONFORMED。
