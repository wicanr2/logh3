# 096：艦隊交戰文字與色盤

狀態：READY。工作項目[#48](https://github.com/wicanr2/logh3/issues/48)及
[#79](https://github.com/wicanr2/logh3/issues/79)。完整056／092仍DRAFT。
本規格核准1DD6:05D7交戰內的082F雙側文字、0B90武器標題、0F6D損失數字及交戰色盤資料。
不核准艦隊交戰的Game規則接線、投影、等待輸入或存檔變更。

## 證據審查

原版GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
dosgolem乾淨版本`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，IDA9.4。
原版入口state SHA-256 `07cebb1872b6b4c2af91158ee932299a34aed59679afab2dfe43784318f7d8ef`。
[重播腳本](../../tools/fleet_presentation_oracle.sh)保存完整命令及CPU、堆疊和RAM注入。
92案各重跑兩次，276份RAM／色號／色盤輸出相同；兩份report只除命令名稱後相同。
原指令、原名與IDA EA／原版CS:IP／線性位址分欄保存於
[凍結收據](../internals/data/fleet-presentation.json)。
[凍結器](../../tools/scripts/fleet_presentation_receipt.py)逐項核bytes、來源雜湊及擁有權。

2026-10-04由原版呼叫端、原始格式與92案內容格審查，文字輸入及輸出已足以實作。
FLEET表是DS:C074本體，FLEET+2為word，不能把前四bytes誤解為遠指標。
本子規格達READY；同狀態測試通過仍不代表完整092或正常玩家驗收完成。

## 已證實契約

| 原版定位 | 契約 |
|---|---|
| 082F | 清指定側視窗，印艦隊／分隊或提督，再印分隊艦種及SHIP+8／+0A艘數 |
| 3982:022D | side2回FF；side等於DS8F25取DS8F21，否則取DS8F22 |
| 082F:0892 | mode2或上述回值非0時，僅detail恰為3顯示身分；其餘未知 |
| 082F:0952 | mode非2且回值0時，身分已知，不受detail影響 |
| 082F:08E1／0967 | FLEET+2編號交3982:0038，沿用既有fleetNumberBytes |
| 082F:0907 | squadron0印提督；其他印全形「第squadron+1部隊」 |
| 082F:09D4 | 艦種用DS613A名稱，兩個艘數是分隊艘數，不能用全艦隊總數 |
| 0B90 | weapon0／1／2選艦種參數+0A／+0E／+12；其他選名稱12；型號大於11亦選12 |
| 0B90 | 標題色side0為2、side1為13，印完文字色還原4 |
| 0F6D | 中窗跳(0,1)，按攻擊側2／13印「擊毀  %6u」，還原文字色4 |

版面格座標由[格式資料](../internals/data/fleet-presentation-spec.json)驅動：
左(1,6)15×2、右(24,6)15×2、中(17,6)6×2。格原點y40，單格16×20。
未知字串是「第？艦隊  第？部隊」，與行星交戰未知提督字串不同。
格式採原始Big5 bytes，不以Unicode重排寬度。

## Typed輸入與失敗方式

分隊輸入包含side、squadron、艦隊編號、提督與艦種Big5名稱、現在／上限艘數及detail。
可見度輸入包含mode、playerSide及原始兩格sideControl，保留3982回值語意，不猜改為IsOwn。
武器輸入包含原始選擇byte及三類武器型號；損失是16位元無號整數。
正式合成器只接受side0／1、playerSide0／1、mode0..2和16位元数量，
錯誤回error且不改畫布；不擅自生成缺失名稱或正規化detail。
文字合成不取亂數、不改SHIP／FLEET或保存格式。

## 色盤與驗收

使用原版logical色號經AC映射後的16色6位元DAC，不能直接把logical當DAC索引。
以92案各307200像素與原版內容格逐項核對；保存原版、remake、差分、區域數字及SHA-256。
圖外和未受修改的其他層來自原版固定入口，明確標為局部same-state合成診斷。
需要原版字型的測試缺檔即skip，不用替代字型宣稱逐像素相同。
執行時資料鏈須由bootstrap載入格式資料，提供專屬交戰色盤；Game選取色盤後只改變StateEngagement。
正常玩家入口須有最小抽樣，但完整艦隊交戰仍須等092解開投影與等待輸入後才接線。
原版資源及凍結像素留在私有研究範圍，不因本規格取得公開散布權。
