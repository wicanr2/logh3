# 088 正常讀檔至首次戰場入口

Status: DRAFT

工作項 [#89](https://github.com/wicanr2/logh3/issues/89)；它是[#79](https://github.com/wicanr2/logh3/issues/79)
正常玩家第三階段同態配對的上游閘門。證據：[研究382](../re/382-normal-battle-player-entry.md)、
[完整收據](../internals/data/battle-normal-entry.json)、[第四十四輪](../goal/2026-10-03-battle-normal-entry-round-44.md)。
087保持READY；既有151個受控UI幀、全原池資料及有效RGB成果不重開、不升格。

## 已量測事實

原版GIN3.COM、dosgolem f7523f9、S3及兩側實行前D3DC167E固定。
原版正常冷啟動雙跑1MiB RAM、畫面、色盤相同；remake用bootstrap.Load／Game.New及正式輸入雙跑相同。
載入後FLEET3904／SHIP27456／STELLAR4032 bytes全部相同；WORK+0C／+0E／+10不同。
首次達希利戰場原版7艦隊、戰場池55筆，remake3艦隊、同口徑27筆；
WORK戰略回合380／381。這些是proven差異，原因尚未證實。
未做實行前診斷存檔的對照，六張全表與完整畫面相同；差異不是該存檔副作用。

## 未達READY

- 讀檔後陣營小選單及WORK控制區的寫入。007的早期CONFORMED明示排除此選單。
- 原版首次實行的AI命令、移動、戰場掃描與日期更新先後；比較最早產生差異的局部資料流。
- 原版執行期指標與remake完整存檔紀錄的比較契約。保留全紀錄，不因資料名相同就將每個byte當成玩法差異。
- 正常S3相同日期／座位／參戰名單之後的索敵等待、正式renderer及完整存讀。

本規格不授權猜補正式產品。相同seed數字不能證明整段亂數條件可比，
不要求整段骰序／內部抽取次數一致，不以改seed使名單碰巧相同。
先追最早分歧，證據充分後才寫READY子契約並實作。

## 驗收

相同S3、事前固定seed、等價玩家輸入，原版獨立重跑。
正式bootstrap入口可到相同參戰名單、回合／座位及原池資料；完整640×480原圖、
remake圖、差分與逐區數字保存。診斷入口與RNG測試期覆蓋明示，不匯入原版RAM。
補正式SHIP／FLEET保存、受影響陣營與輸入最小抽樣、標準產品回歸。
未通過之前保持DRAFT；其他八步／AI增強／音樂驗收不由此代替。
