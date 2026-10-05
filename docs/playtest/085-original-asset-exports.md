# 085：五類原版素材保存

2026-10-05，第96輪，GitHub #70及#68的圖形解碼子範圍。
來源：[研究430](../re/430-title-image.md)、[規格111](../spec/111-title-image.md)、
[素材收據](../internals/data/remaining-assets.json)、[匯出器](../../tools/scripts/remaining_assets_export.go)。

| 類別 | PNG數量 | 尺寸／來源 |
|---|---:|---|
| ANI_PNL0–2.A98 | 3 | 640×400，正式AMG單圖解碼，Used等於完整來源 |
| BLOCK_0–2.BLK | 384 | 各128張16×20，正式planar圖塊解碼 |
| MAP00–31.CMP | 32 | 2048×2040，128×102格；圖塊索引驗實際BLOCK_3 archive count |
| IBM16.350 | 374 | 640×227來源格；每格消耗與IBMFLS16.350相同 |
| SPTIT.350 | 1 | 640×350，224000色號及16色與固定原版雙跑相同 |

共794張、5853987 bytes，來源44個原版檔及一個正式palette JSON。
兩個乾淨批次逐檔PNG與完整來源／輸出JSON相同。
每張PNG重讀的尺寸、色號及palette符合typed圖；圖像SHA、色號SHA、尺寸與來源SHA都記在收據。
保存到private repo的 `assets/original/{ani_pnl,block_0,block_1,block_2,map,intro,title}/`，不進一般發行包。
全部目錄已掛入assets_manifest.py的KINDS；README及MANIFEST.tsv由腳本重生。

原版動畫玩家略過349格、350–373重播三次；保存匯出用完整374個來源格。
此差別有意保留，不修改020播放器。
除SPTIT完整圖以外，本輪沒有新增其他素材逐圖原版畫面對拍；沿既有格式規格與正式解碼器保存。
不把保存PNG稱為正常玩家畫面驗收。#68完整玩家轉場與最後等待仍未完成。

## 重跑

```sh
bash tools/go.sh run tools/scripts/remaining_assets_export.go \
  -orig /orig -out workplace/assets-remaining-<新批次> \
  -manifest workplace/assets-remaining-<新批次>.json
bash tools/go.sh test ./internal/asset \
  -run '^(TestTitle350Original|TestRemainingAssetExports)$' -count=1 -v
```

匯出器拒絕覆寫。第二次以新批次重建並比較完整JSON及PNG，再複製各分類目錄到assets/original。
`TestRemainingAssetExports`核794項數量、實際PNG長度／SHA、PNG重讀尺寸及色號SHA。
標題測試核所有33346種截斷長度、四個分隔byte、零／超長run與尾端多資料。

標準18套件13PASS／5無測試以既有logh3-ebiten:2、Xvfb通過。
新增catalog回歸另跑通過；111只限圖形子範圍CONFORMED。
完整收據由[核對器](../../tools/scripts/title_asset_receipt.py)的`--check`核對，
來源及工具版本在[來源索引](../internals/data/title-asset-receipts.json)。
