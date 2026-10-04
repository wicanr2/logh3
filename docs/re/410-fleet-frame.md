# 410：完整交戰單幀、結果字色與開場停止線

2026-10-04，沿#48／#79。父提交3b4c54c5219c93f2779ad141e965351c45556f7d。
GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
原bundle與權利邊界沿AGENTS.md；原檔唯讀。
dosgolem f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c，Go1.24.13，IDA9.4。

來源為第70輪固定兩個06FC完整RAM與畫面，兩側RAM逐byte相同。
第69輪實際COMANI記錄與caller決定圖段列表；第71輪已解格式與原素材決定文字。
完整來源由[凍結器](../../tools/scripts/fleet_frame_receipt.py)生成
[fleet-frame.json](../internals/data/fleet-frame.json)，SHA-256
`17d968aa06150de3f67e98f8ead12f8636c2624b78a08802ab1e8bd203a4a027`。
17項輸入雜湊、兩幀typed欄位、原生期望畫面與三份IDA原bytes同列保存。
原始CS:IP、IDA EA與dosgolem線性位址各自分欄；自動名稱保持unknown。

## proven：完整合成與目標字色

`workplace/round73-fleet-frame-probe.go`從空白畫布開始，不借上方像素。
初版靜圖名稱寫成A0，庫內不存在；修正為原指令已確認D0。
尚未審查的初版資料保存在`workplace/frame48-r73-invalid-segment.json`，
SHA-256 `267d37a5098a01f2d39fc25311a8021f941b801c1925ff83a7a6ca6656643c89`。
這是來源產生器錯誤，不是原版或產品缺陷。前輪凍結收據保持。

第二版傷害前零差，傷害後266個結果文字像素不同，首差(275,182)為2／13。
1DD6:07C8、07CB、07CE、07D1的caller載入目標，傳目標+00的word及loss給0F6D。
實際0F6D stack為return07D6／1DD6／726／1001。
1001低byte為陣營1，高byte為艦隊16；不能由攻方武器字色推結果字色。
第71輪局部探針傳同一陣營給標題與結果，各常式的局部結論保持，完整組合需分開。
第三版依目標陣營寫結果，兩圖各307200像素零差。
探針產物在`workplace/frame48-r73-probe-v2/`與`-v3/`，未覆寫失敗差分。

## proven：支援編成裁切界

滿編表16項首word直接核當前RAM，兩陣營均上限4000。
來源與取艘數公式沿研究215及當前BuildComposition／FleetEditor，
不從UI外觀推上限。合成器明定目前0..4000艘支援域，外部修改值不升格。
1..8堆疊與可見Y220..376推得方塊最低192、標籤最低172，均在168裁切上緣之內。
這是公式導出的支援域結論，不宣稱任意修改存檔或所有原版生涯不變量。

## proven：開場無額外八tick契約

原版RAM SHA-256 `9dfbb2cbec5d93849e0f100511cad65b342b41f68198378cd9cdc103ebeb9279`。
由它切出1DD6:0152..030E、3958:0005..0040、348A:01DC..0227／0281..02B2，
IDA輸入與指令逐byte相符。開場畫底圖、開文字視窗及換曲，未見固定八tick另行等待。
底圖的348A:0281沿已知契約等待前一倒數，再設8／4／0。
原版動畫關閉的八tick收據不能直接解讀為Game必須增加固定延遲。
本輪保留既有Game計時，硬體時序停止線保持；新規格099只授權單幀合成。

完整056／092仍DRAFT；正常玩家進完整艦隊交戰、正式命令與核心、時間線、音效及保存仍未接。
後續產品驗證與產物雜湊見[對拍062](../playtest/062-fleet-frame.md)。
