# 408：艦隊交戰的文字與正式色盤

2026-10-04，沿[#48](https://github.com/wicanr2/logh3/issues/48)、
[#79](https://github.com/wicanr2/logh3/issues/79)。父提交969d68b8460e983a368aa514aca25e4c912f3a43。
本輪補1DD6:05D7場景內的082F、0B90與0F6D，並修正正式Game色盤。
完整投影、圖序列及等待輸入不由文字測試升格。

## 固定來源

| 輸入／工具 | 版本或SHA-256 |
|---|---|
| 原版bundle | de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673 |
| GIN3.COM | 770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918 |
| blocks48-r70-a/scene-01.state | 07cebb1872b6b4c2af91158ee932299a34aed59679afab2dfe43784318f7d8ef |
| dosgolem | f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c，乾淨工作樹 |
| IDA／Go | IDA9.4／Go1.24.13，既有locked-v1與專案Go容器 |
| 原版字型GRAPH.IMG | 0006671aca0e216da01d817be7f6a9e2fe8f18d5d058674bacf7f32c19ec833d |
| 格式資料 | 566b1f014a16974d153e9bd135db2cb7f6cb9efb072df5171c5b58f3c28ef93e |
| 凍結收據 | 2dbff381163ce7d512af633e496bc5349dc2597210c0131f7f0d0874b0943851 |

IDA輸入由固定原版RAM擷取原段，前256bytes為COM載入頭。
691條原指令的原名、bytes、xref及IDA EA／原版CS:IP／dosgolem線性位址分欄保存。
[凍結器](../../tools/scripts/fleet_presentation_receipt.py)逐條核對原RAM及691條bytes，
571份來源指紋與三份素材雜湊在[收據](../internals/data/fleet-presentation.json)。
工具內原始列仍標unknown，以下經呼叫端與動態輸出審查的契約標proven；不改原名。

## 原版完整資料鏈

SHIP由DS:CFB4遠指標取表、stride78。082F由SHIP+0取雙側文字視窗，
由SHIP+1查DS:C074艦隊表本體、stride122；FLEET+4查DS:6428提督名稱。
SHIP+3為0印提督，其他印「第squadron+1部隊」。SHIP+4查DS:613A艦種名稱，
SHIP+8與+0A印現在／上限艘數。這條文字路徑用分隊數量，不能套行星演出的船艦總數。

3982:022D先處理side2，回FF；其他側與DS:8F25相同時回DS:8F21，否則回DS:8F22。
082F在mode2或返回值非0時，只在detail恰為3顯示已知身分；detail4仍顯示未知。
其他分支不受detail影響。此為proven bytes及固定60案，不把原始控制值猜改成IsOwn。
60案是unit80／82／102／104，mode0／1／2，detail0..4；playerSide0、控制值0／1。
另一觀看座位的分派由原指令確認，未宣稱額外正常玩家抽樣。

FLEET+2傳word，但3982:003B與0059只載入引數低byte到AL，隨後清AH。
原word3085／4097印13／1。格式資料審查時曾把DS:C074誤當遠指標，
原版082F:0875／0877明確取DS與C074表本體，足以否定；錯誤未進正式提交。

0B90從SHIP+2E及艦種26bytes參數的+0A／+0E／+12取武器型號，超過11或其他選擇回名稱12。
side0標題色2，side1色13；原字串清中窗、換色、印名稱、還原4。
0F6D跳中窗(0,1)，按攻擊側色印「擊毀  %6u」，還原4。
32案覆蓋選擇0..3、損失0／1／726／65535及雙側；其他型號名稱僅由RAM表確認。

## 原版重現與正式實作

執行：`bash tools/fleet_presentation_oracle.sh a`與`b`。
兩次各92案，276份RAM、640×480色號與logical→AC→DAC色盤逐byte相同，正規化report相同。
每案重載同一入口，完整CPU、SS:SP、返回哨兵、引數及mode／detail／weapon寫入記於report。
沒有新按鍵，不取RND，seed1319521704保持；這是direct-entry診斷，未替代正常玩家路徑。

[096](../spec/096-fleet-presentation.md)先READY，正式FleetPresentationSpec再實作。
92個內容格、游標、字色與全畫面各307200像素零差，合計28262400像素。
畫面其餘層借固定原版入口，不把局部文字對拍稱為整張場景重建。
原版字型缺檔時只做內容格核對，像素測試skip；錯來源雜湊失敗。
資料缺欄位、錯視窗／色盤形狀、非法側或缺名稱會回error，不修改畫布。

Game.Palette先前在交戰回主選單色盤，色0及14與原版不同。
原版本輪16色與既有eng-open/pal.bin前48bytes相同；後者只作交叉參考，不重生舊收據。
bootstrap正式載入小型格式資料，Resources持有096；StateEngagement使用該色盤，退出恢復既有色盤。
正式Game診斷核16色RGB、renderer全像素映射、退出狀態，正式bootstrap正常開局抽樣保持。
一份未含攻擊命令的正常首戰探針未到交戰，保留在workplace；它不是產品缺陷或正常交戰驗收。

## 停止線

056／092保持DRAFT，48／79保持OPEN。Game尚未呼叫093艦隊核心及096艦隊文字。
正式交戰色盤已修，但正常艦隊交戰、typed投影、完整圖序列、等待輸入、GUI及保存仍未驗。
下一個精確閘門是321F:0513投影生成、可達裁切域及05D7等待輸入契約，再審查完整Game接線。
音訊硬體driver不因此重開；本輪不改音色庫、OGG門檻、061CPU政策或FF排序待裁決。
成果、可見產物及標準測試見[對拍060](../playtest/060-fleet-presentation.md)。
