# 047：四十列資料視窗的完整畫面對拍

2026-10-03，第57輪，完成[#67](https://github.com/wicanr2/logh3/issues/67)。
比較等級為same-state，限定正常S3同盟的資料表偏移0／1／26三張靜態畫面。
沒有裁切、縮放、遮罩或允許差異；各比完整640×480色號與有效RGB。

## 狀態與來源

- 主選單→138→S3→同盟→實行→首戰→戰術選單→資料。六支同盟艦隊自然產生40列，顯示14列，上界26。
- 日期799/3/1，戰略回合380，戰術Input；鏡頭世界像素1312／1660，兩側相同。
- 原版實行前固定D3DC167E；remake使用正式預設同值，不使用測試覆蓋。首戰原版seed B6E4B78A。不注入艦隊或分隊。
- 原版滑鼠y加40對應畫面y。正式開窗用(400,80)／(400,100)，表頭不觸發表列置中；最後游標(440,390)。
- 原版固定dosgolem工作副本f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c，clean，未改執行器。
- remake父提交d070f01ee90ff7c4bf914f62edf80b657a8be633，加本輪呈現修正；最終提交身份由Git紀錄回查。來源雜湊全部封存。
- bundle SHA-256：de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673。
- GIN3.COM SHA-256：770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918。

完整旗標、父snapshot雜湊、原版RAM／PNG／色號／色盤、兩側保存、GUI與測試log SHA-256均在
[固定收據](../internals/data/battle-data-scroll.json)，SHA-256為
`0da250722c37fa5c51518f7b932d29e7a800fd60bab54ac2d7bbb8066a279819`。
工具鏈為logh3-ebiten:2、golang:1.24-bookworm、ida-pro-9.4-idapython:locked-v1；image ID亦在收據。
原版素材及本機dump不加入Git。

## 結果

| 偏移 | 總列數 | 修正前全圖差 | 修正後全圖差 | 原版screen.bin SHA-256 |
|---|---:|---:|---:|---|
| 0 | 40 | 1189 | 0／307200 | 8fdb9d1b89c905dc50ed638aa19dd69a63d3949c5519cc8e35790aa1e154eecb |
| 1 | 40 | 1094 | 0／307200 | c62da2e5eb1b406bb6b5d418a75d795b9252ac63c82801a86244c197a7941550 |
| 26 | 40 | 962 | 0／307200 | ed8e1a575a412db0fc95acffbe124c773bbd7cf3978c32278de4459159db0782 |

每張原版各獨立跑兩次，RAM、色號、PNG與色盤相同。remake正常808 tick到首戰，兩跑27檔全部相同。
分區為全圖(0,0,640,480)、表窗(16,60,448,360)、小地圖(480,75,128,128)、狀態(464,220,160,200)。
各區不同色號／有效RGB為0，最大通道差0，差異邊界空；完整數字見收據。
原圖在workplace/oracle/data67-r57-{zero,one,last}-{a,b}/screen.png；
修正前／後與差分在workplace/round57-comparison/，每檔雜湊已封存。

本輪修正全形單位數艦隊編號、兩個捲軸箭頭、視窗右側陰影、同盟摘要排列與顏色、小地圖視野框換算。
原始指令及字元平面見[研究395](../re/395-battle-data-scroll.md)。六份IDA匯出1836條原指令與原版RAM相符；16項直接語意proven，未審查項保留unknown。
純呈現修正前／後五表逐byte保持。原版與remake各張均為WORK0／FLEET27／SHIP214／STELLAR0／PLANET0。
後兩表只差顯示物件指標，原始差分及位移全數保留；沒有宣稱整份RAM或DOS指標一致。

正式cmd/gin3鍵盤138、滑鼠載入／實行／開表／短按捲動，1倍及2倍各張畫面與正常Game偏移1收據相同。
標準18個產品套件通過，13個含測試、5個無測試檔；完整畫面回歸無skip。

## 重現入口

原版以[固定包裝器](../../tools/battle_data_scroll_oracle.sh)執行，輸出目錄已存在會拒絕覆寫。
完整主選單冷啟動在parent，後續由固定父snapshot接續；每條命令仍保存至flags.txt及meta.txt。

```sh
for mode in parent zero-a zero-b one-a one-b last-a last-b; do
  bash tools/battle_data_scroll_oracle.sh "$mode"
done
bash tools/gui.sh test ./internal/game -run '^TestBattleDataScrollWholeFrameMatchesOracle$' -count=1 -v
bash tools/gui.sh xrun bash tools/battle_data_scroll_gui.sh
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
```

[正常入口測試](../../internal/game/battledata_wholeframe_test.go)使用產品bootstrap與Game，不匯入原版RAM。
原版不在本機或收據不存在時明示skip，已存在但雜湊漂移則失敗。額外短按驗上界夾住。
測試輸出可用容器內LOGH3_DATA_SCROLL_OUT指定新的既存空目錄，保存三份完整畫面及正常存檔。
[收據產生器](../../tools/scripts/battle_data_scroll_receipt.py)固定本輪雙跑／GUI／log路徑，封存後拒絕覆寫，不拿新來源重生歷史JSON。

## 限制

本輪不含表列指向置中、按住連捲#60或其他頁籤驗收。
資料窗蓋住首戰殘留排序鍵的一像素，不能據此把第56輪忠實度取捨視為已決定。
完整088仍DRAFT，#79／#89未完成；本輪same-state不升格其他nearby-state或受控索敵收據。
Linux Xvfb鍵鼠不代表Windows／macOS實機，音訊聆聽與正式交付仍按#69原門檻。
