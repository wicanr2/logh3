# 058：艦隊完整場景的原版重播

2026-10-04，沿#48／#79。[研究406](../re/406-fleet-scene.md)保存原始位址與證據等級；
[092](../spec/092-fleet-combat-entry.md)仍DRAFT。
分類為原版已提交命令的受控重播，沒有remake畫面對拍，不標same-state／pixel-perfect。

## 固定條件

dosgolem-logh3 f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c／clean。
原版bundle de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673。
640×480，未縮放；完整state恢復CPU／DOS／時鐘。每案由同一a父state開始兩次。
原seed及祖先注入／點擊保存在第63／66輪凍結收據，不稱父狀態自然、未受控。

| 案例 | 父state SHA-256 | 新輸入／注入 |
|---|---|---|
| 光線砲 | af0019ce8a29ea2e0026a0ba15ac4f37d870f848000247a07250ee10190de19d | 無 |
| 飛彈 | 1b84dd2f0bf76cfd5b526d4df18d7dcbc1f7330478c35e2d97844f6dfaf0457a | 無 |
| 同目標兩擊 | 5722e421d007a050c37788577d31119b19876b8e27e3405a674f355a39a30d82 | 無 |
| 死亡後停止 | 同上 | step507999999，lin:687ce由B8 0B改01 00；不換seed |

## 重播命令與收據

以下每列各跑a及b；封裝器核專用副本HEAD／乾淨狀態，拒絕覆寫輸出。
父state重建方法沿研究401／404，命令、祖先及完整注入在凍結JSON可回查。

```bash
bash tools/fleet_scene_oracle.sh laser a
bash tools/fleet_scene_oracle.sh laser b
bash tools/fleet_scene_oracle.sh missile a
bash tools/fleet_scene_oracle.sh missile b
bash tools/fleet_scene_oracle.sh same-target a
bash tools/fleet_scene_oracle.sh same-target b
bash tools/fleet_scene_oracle.sh death a
bash tools/fleet_scene_oracle.sh death b
bash tools/py.sh tools/scripts/fleet_scene_receipt.py
bash tools/gui.sh test ./internal/rules -run TestFleetCombatDeathStopsFollowingCommand -count=1 -v
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
```

輸出在workplace/oracle/scene48-r69-v2-<case>-<a|b>/。
每個邊界保存1MiB RAM、307200色號、完整色盤及未縮放PNG。
report記相對step、CPU原始入口、AX、stack參數、7-byte COMANI記錄、seed及clock。

| 案例 | trace事件 | 畫面邊界 | 加首尾的雙跑相同檔 | 執行步數 |
|---|---:|---:|---:|---:|
| 光線砲 | 75 | 12 | 56 | 38,746,670 |
| 飛彈 | 75 | 12 | 56 | 40,346,452 |
| 同目標兩擊 | 124 | 22 | 96 | 82,146,748 |
| 死亡後停止 | 76 | 13 | 60 | 42,946,712 |

共268個RAM／色號／色盤／PNG檔，兩側逐byte相同；四份report扣除a/b命令名稱後相同。
565份來源指紋與1,980條IDA原bytes凍結，獨立重建完全相同。
JSON SHA-256 7d7e5b388eebc8348825e91352ff10e9821ce2b0cf7912d6f9b9cf1cafc95379。

死亡核心回歸無skip通過：從原版逐命令seed開始，首擊實扣1，第二擊Skipped，
無第二次庫存／目標／RND變動；末端原版艘數、庫存與修理上限相同。
此測試不另稱排序已驗；排序由第67輪既有全鍵測試核對。

標準18產品套件13PASS／5無測試，正常結束；log SHA-256
a1970596c46e854f0d5923cade6e6126b66d891797b350b5e14f6d49ea0e2b87。

完整remake場景、正常玩家入口、下方buffer、正式色盤、GUI與保存待接。
原版圖沒有remake對應圖，因此本輪沒有差分圖或未核准像素差異為零的聲明。
#48／#79保持OPEN，093 READY、092 DRAFT。
