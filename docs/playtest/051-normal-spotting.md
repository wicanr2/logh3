# 051：正常S3首次索敵

2026-10-03，第61輪，工作[#79](https://github.com/wicanr2/logh3/issues/79)。
方法、原始定位、輸入與工具身份見[研究399](../re/399-normal-spotting.md)。
[固定收據](../internals/data/normal-spotting.json)SHA-256
`4f15269fcedd71292a16a68c5bb6ec76c4303f4b0edc8d1b768a2d82909209a4`。

| 項目 | 結果與範圍 |
|---|---|
| 原版重現 | 固定正常父狀態與玩家輸入，兩次各16份輸出完全相同 |
| remake正常路徑 | 正式bootstrap／主選單／138／S3帝國／實行／首次戰術實行，兩次各24檔相同 |
| 限定情報 | 352筆Seen bit0／Detail／Hold，標籤與第四階段前均符合原版 |
| 續接與保存 | 無發現報告，只續接一次；戰術回合1→2，完整SHIP／FLEET正式保存及重讀相同 |
| 五表增量 | 兩側各自的索敵階段完整五表零變動 |
| 完整畫面 | 標籤與結束兩圖各0／307200，色號與有效RGB及三區零差；差分、並排圖已目視 |
| 完整狀態限制 | Slot raw WORK32／FLEET147／SHIP699差異，seed AFCFD1BC／12345678；nearby-state，不升格 |
| 標準回歸 | 18產品套件通過或無測試，新增限定測試兩跑均PASS，零SKIP |

這次正常場景沒有發現報告。正常來源報告與完整同狀態索敵仍缺，087維持READY，#79維持OPEN。
061被動電腦、顯示指標與FF字鍵契約保持；本輪不新增真實X11、硬體時序、八步或發行聲明。

```sh
bash tools/normal_spotting_oracle.sh normal79-r61-a
bash tools/normal_spotting_oracle.sh normal79-r61-b
bash tools/gui.sh xrun env LOGH3_NORMAL_SPOTTING_OUT=/src/workplace/round61-normal-checked-a go test ./internal/game -run '^TestNormalSpottingInformationAndSave$' -count=1 -v
bash tools/gui.sh xrun env LOGH3_NORMAL_SPOTTING_OUT=/src/workplace/round61-normal-checked-b go test ./internal/game -run '^TestNormalSpottingInformationAndSave$' -count=1 -v
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
bash tools/gui.sh xrun python3 tools/scripts/normal_spotting_receipt.py
```

原版與正常Game輸出拒絕覆寫。產生器核固定檔名與本機log；乾淨重建應保存這輪來源並依JSON完整命令執行。
新的重播批次不能覆蓋歷史收據。標準log SHA-256：
`405efe6c2d458c35f9d88860dd7f74787a4216ea848a16c5bcbd61e4776f73f0`。
