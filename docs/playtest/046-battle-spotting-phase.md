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
