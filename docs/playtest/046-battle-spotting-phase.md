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
