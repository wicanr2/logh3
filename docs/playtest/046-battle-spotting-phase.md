# 046 第三階段原版重現與座標子契約

日期：2026-10-02；Issue #79，原版來源與位址證據見[379](../re/379-battle-spotting-phase.md)。
唯一目前狀態為[CONTEXT](../../CONTEXT.md)，完整第三階段驗收由[087](../spec/087-battle-spotting-phase.md)管理。

## 原版重生

```sh
bash tools/battle_spotting_oracle.sh battle79-spot-r39-s1-a
bash tools/battle_spotting_oracle.sh battle79-spot-r39-s1-b
bash tools/battle_spotting_oracle.sh battle79-spot-r39-s3-a
bash tools/battle_spotting_oracle.sh battle79-spot-r39-s3-b
bash tools/py.sh /work/tools/scripts/battle_spotting_evidence.py
bash tools/go.sh run ./tools/scripts/battle_spotting_probe.go
```

包裝器核對父狀態、固定dosgolem完整HEAD與clean，禁止覆寫既有批次。
全命令列在[資料JSON](../internals/data/battle-spotting-phase.json)的runs.command_argv。
缺少父狀態時先按既有[第36輪](../goal/2026-10-02-battle-second-oracle-round-36.md)
及[第29輪](../goal/2026-10-02-battle-movement-round-29.md)來源鏈重建，不能自製替代品。

| 檢查 | 結果 | 限制 |
|---|---|---|
| S1四邊界／S3十四邊界 | 54檔雙跑相同 | 原版決定性，不是remake通過 |
| 兩分支終態畫面、PNG、色盤 | 6檔雙跑相同 | 原生640×480，無裁切／縮放 |
| 完整分隊紀錄探測 | 1,310筆78 bytes相同 | 可丟棄模型，正式UI／保存未驗 |
| 自然局面情報規則 | 正確轉接117／48筆通過 | 現在位置等於目的地，不構成座標反例 |
| 六案原版座標／半徑反例 | 完整雙跑相同，舊規則五案失敗 | 原版常式診斷，非正常UI |
| 正式座標修正 | 六案全紀錄與自然情報回歸通過 | 不冒充整階段 |

探測log `workplace/round39-spotting-probe.json`，SHA-256
`1c973a549e72d82f8a6dacbf6d81c7eaa35f7ac32c21645c0ae20396c96a550e`。
這是已勘誤的首次轉接，不能作座標反例。修正後log `workplace/round39-spotting-fixed-probe.json`，SHA-256
`62f32576b967d6869565200aad9f58e01329d792e069cf55b01e8cadb8641516`。
限定測試log `workplace/round39-spotting-fixed-tests.log`，SHA-256
`8db8633dcec851ebfac08a6cdecf7535297e5da26b3de7356cad938446bdff64`。
舊規則六案對照log `workplace/round39-controlled-before.log`，SHA-256
`c573f5f123333679ccec867cb6668d8b6cebbb3659760ba8a7b8aa47187dcbc2`。
標準產品回歸log `workplace/round39-standard-fixed-tests.log`，SHA-256
`67ea3bd6ceb190f51fcbbeecc75e4f320bcd94e6a2d084c5ae8add6b5bc3bffa`。
18個產品套件通過或無測試，目標14個PASS事件零跳過。綠色產品回歸不證明完整第三階段parity。
每份原版畫面、完整RAM、metadata及來源匯出SHA均保存在JSON，不能只引用本機檔名。
header.png與unit-0.png已目視：前者為行星索敵標頭，後者開始展示新圖示。

沒有原版／正式remake／差分PNG與逐區數字，整階段不能標通過。
Detail3內容、需移動的鏡頭、提前點擊、雙人、正常玩家發現報告及正式保存仍待驗。
第38輪已完成的第二戰場43幀全畫面不重開，062完整八步保持未完成。

## 六案受控證據重生

```sh
bash tools/py.sh /work/tools/scripts/battle_spotting_evidence.py prepare-coordinate
DOSGOLEM_ORIG="$PWD/org_game/extracted" \
DOSGOLEM_EXTRA_MOUNT="$PWD/workplace:/out:ro" \
bash /home/anr2/cht/dosgolem-logh3/tools/go.sh run /out/round39-coordinate-probe.go \
  -plan /out/round39-coordinate-plan.json \
  -output /src/workplace/logh3-round39-coordinate-a.json
```

第二次改輸出檔名為`logh3-round39-coordinate-b.json`，不改任何輸入。
兩份收據先從專用dosgolem工作目錄唯讀掛載複製到本專案workplace，再跑evidence工具。
計畫與來源工具身份由[座標資料JSON](../internals/data/battle-spotting-coordinates.json)固定。
既有計畫或輸出存在時拒絕覆寫，不另選seed。


## 第40輪呈現分支續證

來源：[研究380](../re/380-battle-discovery-presentation.md)、[分支計畫](../internals/data/battle-discovery-plan.json)、
[分級索引](../internals/data/battle-discovery-semantics.json)及[完整資料](../internals/data/battle-discovery-evidence.json)。
這是原版受控分支重現，不是正式remake第三階段通過。

```sh
for scenario in detail3 camera cpu twohuman swapped left right; do
  bash tools/battle_discovery_oracle.sh "battle79-discovery-r40-$scenario-a"
  bash tools/battle_discovery_oracle.sh "battle79-discovery-r40-$scenario-b"
done
bash tools/gui.sh xrun python3 tools/scripts/battle_discovery_evidence.py
```

包裝器拒絕覆寫，逐次核對固定父狀態、完整dosgolem HEAD及clean，全argv在完整資料。
camera／swapped正式注入用空格分隔每個byte，錯誤探索不計入。

| 檢查 | 結果 | 範圍 |
|---|---|---|
| 七分支原版雙跑 | 315份輸出相同 | 294份邊界輸出及21份終態，原生640×480 |
| 98邊界資料 | 5390筆SHIP、686筆FLEET、294筆PLANET全紀錄 | 不排除快取+2F／+6A |
| 種子 | 七分支均12345678，未變 | 事先固定，無重擲 |
| Detail3／鏡頭／CPU／兩人／交換座位／左右鍵 | 逐分支結果已驗 | 兩人原映射第二席此局面沒有新發現 |
| 原始指令 | 八份IDA匯出、80項分級定位 | 保留原名／bytes／位址空間，未審查helper仍unknown |
| 標準產品測試 | 18套件通過或無測試 | 正式產品沒有新增行為，不證明第三階段parity |

原版完整資料SHA-256 383965f538fdf01ffad4a78b32957ad0372fd5f9a08440fe7d26210ac50507e1。
計畫SHA-256 02d6b05cfd031d14668ef80dea8d3a093462d4a1565a373954d38ecc224163d5，
指令索引SHA-256 63a4c95a79f538c8e6de4f15ba7efcf1c6a781c8595bcbc017bf0fad544ea15e。
證據log workplace/round40-evidence.log，SHA-256
a87b1b917b5e827903fb97db98b711306573c46ca86fc8213dbf8e44683b017c。
標準產品log workplace/round40-standard-tests.log，SHA-256
c7be5ae111fd1532de1b0ca0fb15e32f26a01d29af5f2f5d6d686716f19d0804。

標頭與Detail3第一筆PNG已目視；這不是人耳／音訊、正常玩家或正式renderer驗收。
camera注入後原版保留舊背景至置鏡，不拿它豁免remake差異。
下一閘門為物件+1E最小可見性契約、03E9快取寫入鏈、整步READY與正式等待續接／保存。

## 第41輪最小顯示契約與規則修正

來源：[研究381](../re/381-battle-spotting-display.md)、[完整證據](../internals/data/battle-spotting-display.json)
及[固定計畫／分級索引](../internals/data/battle-spotting-display-semantics.json)。
前節第40輪的+1E與03E9未知已解，不再列為接線前的研究阻塞。
正式完整第三階段仍DRAFT；本節只驗已升READY的情報／圖示子契約。

| 檢查 | 結果 | 範圍 |
|---|---|---|
| 61案原版 | 36圖示／15可見性／10情報，完整雙跑相同 | 常式診斷，非正式VGA畫面 |
| 兩案回收後刷新 | 完整SHIP／FLEET雙跑相同 | 原版26E5正式caller，Detail1可重新有標記 |
| 原始指令 | 七份IDA匯出、26項分級定位 | 原名／bytes／位址保留，不深挖通用renderer |
| 舊索敵條件對照 | 十案中七案失敗 | 等價單觀測者，不額外執行敵方偵察隊觀測 |
| 正式規則 | 36圖示全78 bytes、10情報雙方全78 bytes、兩案標記全78 bytes相同 | 15可見性僅原版已驗，正式鏡頭未接 |
| 正常命令／renderer／存讀 | 主選單既有fixture輸入至戰場，放出／移動／歸隊一回合／回收／再放出通過 | 原池81、單位31保持；SHIP／FLEET完整存讀相同 |
| 標準產品測試 | 18套件通過或無測試 | 不證明完整第三階段parity |

### 固定原版重生

```sh
tools/gui.sh xrun python3 tools/scripts/battle_spotting_display_evidence.py prepare
DOSGOLEM_ORIG="$PWD/org_game/extracted" \
DOSGOLEM_EXTRA_MOUNT="$PWD/workplace:/out:ro" \
/home/anr2/cht/dosgolem-logh3/tools/go.sh run /out/explore/battle79-spot-r41-probe.go \
  -plan /out/explore/battle79-spot-r41-final-plan.json \
  -output /src/workplace/battle79-spot-r41-final-a.json
```

第二次只改輸出尾碼b。兩案補充先用`prepare-recovered`，改計畫為
`battle79-spot-r41-recovered-plan.json`，輸出為`battle79-spot-r41-recovered-a.json`及b。
輸入存在即拒絕覆寫；來源缺失不以自製資料替代。
執行前核對專用dosgolem完整HEAD f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c與乾淨工作樹。
兩份收據從dosgolem的workplace唯讀掛載複製回本專案workplace/explore後，執行evidence工具。
IDA每份完整命令使用`tools/ida.sh raw idat -A`與`battle_rules_ida.py`，
entry查詢、切段身份、原始指令及固定輸入列在資料JSON的exports／functions；
來源鏈沿381，26E5補充caller保留研究374既有匯出。

原版61案SHA-256 f03cc58211076e8bc0aa5a2a2e09414ae84716d68080fa044af62169ada01471；
兩案回收SHA-256 d9cb659cc3bffced7dd514d0ba1d3cba507f3d75f4ed4c549657763fa06152f9。
完整證據SHA-256 2d1f98931dae20e7de7759df1525a5fa1b12ce515f7c202783fd57002f25c288；
分級索引SHA-256 595fb99e03fbd8e083178ba1d222f997926b1218540abb38432549082bc2d301。
限定測試log workplace/round41-final-target-tests.log，SHA-256
3d30bc82308f63313423a9c2c4c2a688a45cc217d0d3b47dafeba6c4143cfd1b。
舊條件唯讀覆蓋log workplace/round41-before-equivalent-tests.log，SHA-256
93e41937b69fbbf81a6e9012a03ba18572a738813ac1ed124062f9c3a3d89ba1。
標準產品log workplace/round41-standard-final-tests.log，SHA-256
198dcd762e2b49df10d493dac9025ebd79c43157ccb507103564b45ed62148bc。

### 正式renderer修正前後抽樣

```sh
tools/gui.sh xrun env LOGH3_MOVEMENT_OUTPUT=round41-scout go test ./internal/game \
  -run '^TestScoutRecoverAndRelaunchNormal$' -count=1 -v
tools/gui.sh xrun python3 tools/scripts/battle_spotting_display_evidence.py render
```

修正前以父提交的`BattleUnitSprite`函式唯讀編譯覆蓋，其餘規則、fixture與輸入相同，
輸出前綴`round41-scout-before`。覆蓋檔及hash存在畫面收據，不回退正式程式或原版期望。
四組原生640×480前後PNG／完整色號／色盤／差分保存在workplace。
只有scout-away的(312,248)..(327,263)改變216像素，最大RGB通道差60；
其他三組全307200像素相同，小地圖四組均相同，無裁切、縮放或遮罩。
逐區幾何來自正式screen-battle.json；輸入序列由既有battleTestGame及正常命令測試保存。
兩張移動後／回收後PNG已目視。此比較標為remake-before-after，沒有原版同狀態畫面，
不升格成原版UI對拍通過。

畫面收據workplace/round41-scout-render.json，SHA-256
987a14e81d99556e77a803ce8d3be5d66252f4a96462597f4750dd9f6eded533。
畫面測試log workplace/round41-scout-render-tests.log，SHA-256
ae22eb0ef862e9a0f8846c04e805c554af03bc6cd089ac1cadb5e0d4844c0f0f。

首次回歸的單觀測者轉接及舊回收標記斷言已勘誤；首次PNG檢查誤要求256色色盤，
改驗正式16色色盤足夠涵蓋全部色號後同流程通過。兩者不列產品缺陷。
下一步為正式逐觀測者報告、增量圖示快取、候選置鏡、等待、戰鬥前續接與情報保存，
再做第39／40輪18＋98邊界的正式整幀差分。完整八步與Goal保持未完成。
