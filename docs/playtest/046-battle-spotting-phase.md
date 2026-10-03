# 046 第三階段原版重現與座標子契約

日期：2026-10-02；Issue #79，原版來源與位址證據見[379](../re/379-battle-spotting-phase.md)。
唯一目前狀態為[CONTEXT](../../CONTEXT.md)，完整第三階段驗收由[087](../spec/087-battle-spotting-phase.md)管理。

第44輪正常入口已找到上游差異，接續[#89](https://github.com/wicanr2/logh3/issues/89)與[088 DRAFT](../spec/088-normal-battle-player-entry.md)。
完整新增紀錄見本檔末尾；151個受控UI幀不升格為正常玩家same-state。

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

## 第42輪正式Game接線與限定同狀態

來源：[正式收據](../internals/data/battle-spotting-game.json)、[087](../spec/087-battle-spotting-phase.md)。
GIN3.COM SHA-256 770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918，
dosgolem f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c、乾淨工作樹；原版與執行器未改，本輪沒有重跑原版。
第39／40輪各兩份色號、PNG與1 MiB RAM逐檔重核固定雜湊及相同內容。

比較先限為same-state診斷UI邊界：唯一入口轉接SHIP／FLEET／WORK／PLANET、seed，
正式parser、ObserveSpotting、Game續接與renderer計算其餘邊界，沒有每幀回注期望。
原版完整命令與注入保存在收據original_commands；continue對應固定UI呼叫邊界，
實際Tick及左右鍵另驗，不聲稱相同CPU時間或硬體wall-clock。

camera要求鏡頭0,0與最後已畫鏡頭66,30不同。入口顯示物件+6/+8可重建最後鏡頭，
用正式Frame畫入口後恢復要求鏡頭；正式實作保存兩者到真正置鏡。
不複製原版framebuffer，也不遮罩。S1游標100,150，其餘240,240。

| 檢查 | 結果 | 限制 |
|---|---|---|
| 九案可比UI邊界 | 94幀全307200色號及有效RGB零差異 | 116停點中22個寫字前／內部邊界逐項排除，不稱全部覆蓋 |
| 完整資料紀錄 | 5330筆SHIP78、680筆FLEET122相同 | 不忽略+2F／+6A；不是整張原池每幀都驗 |
| 原版／remake／差分PNG | 94組及field／minimap／status逐區數字齊全 | 原生640×480、無縮放／遮罩，bbox=null、最大通道差0 |
| 正常玩家報告 | 主選單至S3戰場，實行見艦隊13號標頭；兩次實行 | 既有入口fixture，未注入情報／報告；remake-normal-player，非原版同態 |
| 等待診斷 | MsgSpeed90、逐筆6 ticks；左右鍵僅結束當次等待 | 固定S3診斷入口，不代替正常玩家收據 |
| 正式存讀 | SHIP27456／FLEET3904 bytes相同；下輪Seen同步 | 不稱中途讀檔可繼續戰鬥 |
| 標準產品回歸 | 18套件通過或無測試，第一／第二階段保持 | 不外推其他八步 |

重跑正式對拍及正常圖：
tools/gui.sh xrun env LOGH3_MOVEMENT_OUTPUT=round42-normal go test ./internal/game -run '^TestBattleSpotting' -count=1 -v。
再用tools/gui.sh xrun python3 tools/scripts/battle_spotting_display_evidence.py phase-render生成PNG／區域統計。
標準回歸為tools/gui.sh xrun go test ./cmd/... ./internal/... -count=1。

正式收據SHA-256 2ef3932429a368eefc85e723639683b8ba891bc83007198f5afe352cd3165a32。
限定log workplace/round42-spotting-final-tests.log，SHA-256
ef6b4ca5ef4317c895d0977767253b4d0d711d4ea71b2521ea27fba323ffc228。
標準log workplace/round42-standard-confirmed-tests.log，SHA-256
d09fb4ba0330d988ad2fef43f5e55e74348ca6cda975635b16aeac7bd7748054。

已目視行星標頭、Detail3第一筆標亮、正常艦隊13號標頭。
編號沿3982:0073與研究338：一位全形、兩位半形；2526:04C2加2是buffer byte偏移。
正常來源FLEET12已顯示兩位數，但還沒有此來源的獨立原版標頭畫面。

首次游標隱藏及已畫鏡頭誤用已修。轉接鍵名／游標座標、RightHold參數是驗證腳本問題。
舊雙擊測試未等正式續接，十回合測試越過九回合結束後再實行，
依正常流程修測試並同容器重跑；沒有改原版期望或遮罩。
首次標準失敗log workplace/round42-standard-tests.log，SHA-256
de820e7a4f963fb484c878ae1cfc31a6a9dc4c044fa655907c9931c44e70e9b9。

087保持READY：艦隊／已放出偵察隊來源獨立原版標頭與近距來源／界外候選移鏡待驗。
15案可見性常式不代替正式動態畫面，完整八步、#79與Goal未完成。

## 第43輪續證：來源標頭與逐格移鏡

日期2026-10-03，父提交649bef601b3ed46f976e9cfc5627a729ca41310b，工作項目#79。
執行前定為same-state受控UI比較；正常玩家產品路徑另列，不升格為原版配對。
原版來源、bundle、兩個image及固定dosgolem f7523f9完整HEAD／clean沿381與JSON。
父machine.state SHA-256 39ed892ae729c2b2bd1a0b208439d106e4913867e1b50134396fabd704b39b55。
雙方seed12345678只在入口固定，索敵全程不變，沒有重擲。

| 案例 | 原版停點 | 可比正式UI | 預先固定條件 |
|---|---:|---:|---|
| fleet13 | 7 | 6 | 唯一候選109距旗艦100，其他Seen bit0固定 |
| fleet3-cache13 | 7 | 6 | 只改艦隊編號3，完整保留舊13號圖示快取 |
| scout13 | 11 | 10 | 池81已放出；候選距240，旗艦200查不到；界外候選4次移鏡 |
| near-source | 28 | 27 | 來源鏡頭55,26，舊背景66,30；14次重畫到69,32 |
| offscreen | 9 | 8 | 旗艦200半徑邊界，來源置鏡後候選界外；2次移鏡 |

五組原版a／b的201份PNG、framebuffer、完整1MiB RAM與終態色盤相同。
五個寫索敵字前入口明示不算正式UI，57幀完整640×480及有效RGB、3135筆SHIP78、399筆FLEET122零差異。
每幀保存原版／remake／差分及修正前診斷PNG，逐區含full／field／minimap／status、比較數、相同數、
不同數、最大RGB通道差與bbox。沒有遮罩、裁切、縮放或原版framebuffer當背景。
唯一入口注入後，原版所有紀錄與鏡頭只作斷言；每個scroll固定只走一次正式更新。

完整固定計畫：[battle-discovery-scroll-plan](../internals/data/battle-discovery-scroll-plan.json)。
完整第43輪收據SHA-256 a7169b381e06e1532e4ff3887d0c75ea08b9fa751b7a7b6ee72faf9c96d0d53d。
原版命令、所有lin注入及readback、trace、full records、渲染／正常抽樣與source SHA：
[battle-discovery-scroll](../internals/data/battle-discovery-scroll.json)。工具由[第43輪](../goal/2026-10-03-battle-discovery-scroll-round-43.md)及internals索引進入。
原版輸出分workplace/oracle；所有探索分workplace/explore，舊收據不覆寫。

原版重播使用新a-replay1／b-replay1目錄，五案依表中名稱各執行：

```sh
bash tools/battle_discovery_scroll_oracle.sh battle79-scroll-r43-near-source-a-replay1
bash tools/battle_discovery_scroll_oracle.sh battle79-scroll-r43-near-source-b-replay1
```

五案兩側全部完成後，collect a-replay1 b-replay1可核對新批次；
以下命令使用本輪既有a／b。不刪除舊收據，重播序號不可重用。

```sh
tools/py.sh /work/tools/scripts/battle_discovery_scroll.py collect
tools/gui.sh xrun env LOGH3_MOVEMENT_OUTPUT=round43-normal go test ./internal/game -run '^TestBattleSpotting' -count=1 -v
tools/gui.sh xrun go test ./internal/... ./cmd/... -count=1
tools/gui.sh xrun python3 tools/scripts/battle_discovery_scroll.py render
tools/gui.sh xrun python3 tools/scripts/battle_spotting_display_evidence.py phase-render
```

父提交鏡頭實作以唯讀Go overlay對同一修正後fixture，scout13／near-source／offscreen三案失敗，
第一個移鏡幀各差4073／10426／3732像素。修正保留來源精確目標與候選續接，每次最多一步。
修正前後GUI命令與輸入相同；父提交對照是漏幀診斷，不將不一致中間狀態稱同態通過。
log round43-corrected-fixture-before-tests SHA-256 484ae10305b0b603b98fc665bb7fab078d4d9b636f41b30d5f198a4a095d0382。

正式Tick逐格、移鏡期間左右鍵／實行阻擋、MsgSpeed90與6 ticks、僅續接一次通過。
正常主選單／防拷138／S3／戰略實行到戰場，再兩回合與SHIP27456／FLEET3904 bytes存讀通過。
normal-spotting-header／completed原生PNG保存，正常圖不是原版同態收據。
限定25個通過事件，零跳過；18標準套件通過或無測試，含原有第一／第二階段、AI與音訊產品回歸。
final-tests log SHA-256 98fc9e202c65cd0655b16031db847f6966b8c821cc0d87cd7e3b699e81af5cf0。
standard-tests log SHA-256 03f16dcc39532fe76017ba01a27e39ad90aefa11ebdbbc91d4995b30bb27e3af。
完整路徑、日誌及PNG雜湊均在JSON，不靠檔名單獨定位。

舊九案94幀／5330筆SHIP／680筆FLEET維持，全畫面及有效RGB仍零差異。
刷新舊收據的目前來源身份SHA-256 5a7f9d65dc8c640c33fd1c68c3e6bf15f045f369298994a56dc3baca6ffd4824；
第42輪原先2ef393身份保留在歷史，只是工具來源更新，原版期望未改。
合計151可比UI幀；原有22個內部停點與本輪5個寫字前入口仍不算正式UI。

行星半徑外兩份探索、一位數圖示注入不完整的反例及未凍結的兩份輸出全部保留，原因見381與plan。
新指令語意與018／062／087回鏈由工具核對。087保持READY，正常玩家原版配對仍缺，#79 OPEN。
擁有權與Docker清理核對記錄於本輪WORKLOG；沒有發佈或完整八步聲明。

## 第44輪正常玩家入口比較

工作項[#89](https://github.com/wicanr2/logh3/issues/89)，[088 DRAFT](../spec/088-normal-battle-player-entry.md)，
[研究382](../re/382-normal-battle-player-entry.md)，[本輪goal](../goal/2026-10-03-battle-normal-entry-round-44.md)。
完整收據[battle-normal-entry.json](../internals/data/battle-normal-entry.json) SHA-256
`d9e719ca5ffca52447c7476154deec1023c87d057987af9ae68d5f163036181e`。

```sh
bash tools/battle_normal_entry_oracle.sh battle79-normal-r44-a
bash tools/battle_normal_entry_oracle.sh battle79-normal-r44-b
tools/gui.sh xrun python3 tools/scripts/battle_normal_probe_seed.py
tools/gui.sh xrun go run -overlay workplace/round44-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round44-normal-fixed-a
tools/gui.sh xrun go run -overlay workplace/round44-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round44-normal-fixed-b
tools/gui.sh xrun go run -overlay workplace/round44-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round44-normal-nosave -skip-loaded-save
tools/gui.sh xrun python3 tools/scripts/battle_normal_entry_receipt.py
tools/gui.sh xrun go test ./internal/game -run '^TestBattleSpotting' -count=1 -v
```

原版批次、probe輸出與seed覆蓋拒絕覆寫。未改dosgolem f7523f9或原始遊戲；固定前後實讀remake D3DC167E。
原版5份輸出獨立雙跑相同；snapshot全檔不同，不列為通過。remake17份輸出雙跑相同，
省略實行前存檔的六表／畫面對照亦同。所有SHA、完整argv、source與工具鏈身份在JSON。

載入後FLEET3904／SHIP27456／STELLAR4032全表零差異；WORK+0C／+0E／+10不同。
進場原版7支、remake3支艦隊，回合380／381；全圖5863、Field73像素不同。
整組分類nearby-state，不能進第三階段same-state裁決。原版／remake／並排／差分PNG已目視，不裁切或遮罩。
影像診斷沿原版PNG有效PLTE色盤，不直接把DAC當色號映射；原版CPU／日期／座位根因仍unknown。

目標索敵回歸25個PASS事件、零skip，151個受控UI幀及正常兩回合／完整SHIP／FLEET存讀保持。
log `workplace/round44-spotting-tests.log` SHA-256
`1236704ae68c9ec24a250ec98b12079db46888a3c3d99cc477db3e9f9744e247`。
正式產品未改，18套件基線沿第43輪；不以此次探測宣稱AI根因、完整索敵或八步完成。
087保持READY、#79 OPEN；下一步沿#89／088追正常讀檔與首次實行最早分歧。

## 2026-10-03第45輪：正常首戰AI一次性跳過

[研究383](../re/383-normal-ai-once-gate.md)、[固定方法](../goal/2026-10-03-normal-ai-gate-round-45.md)、
[完整收據](../internals/data/normal-ai-once-gate.json)補正常入口最早分歧。
原版自然1與受控0各四份完整RAM／畫面／色盤雙跑相同；相同初始RAM只差WORK+32。
正式修正後17份正常輸出雙跑相同，省略診斷存檔的七份表／畫面相同。
原版與remake首戰7支次序、55筆池及結尾seed B6E4B78A相同。
WORK回合380／381、座位控制及全局其他bytes仍不同，未成完整same-state。

完整640×480仍差5863像素，Field161280差73；有效RGB同口徑。
本機完整並排round45-normal-entry-pair.png、差分及原色盤Game圖已目視，SHA在JSON。
沒有裁切、縮放、遮罩；第44輪正常冷啟動及第43輪151幀成果保持，兩者不升格或互換。

38個限定PASS、零skip，包含原版／增強第二次AI、非零邊界、消費後正式存讀與正常雙陣營增強AI。
S1／S2／S3正常戰術實行與完整SHIP27456／FLEET3904 bytes存讀通過，只聲明remake抽樣。
標準18套件初跑S1／S3舊數量失敗，更新實測抽樣後同命令乾淨重跑通過；沒有放寬原版golden。

| 本機log | SHA-256 |
|---|---|
| round45-gate-tests.log | fd4c680d768283077999b9d5cf6b4e0cc2c0cac5276230816ffc92c7a63899bc |
| round45-shipped-tests.log | ea555bd3e01f849dbd1a177afa4ce0b41705d4d63eab2b9838c8ed9ed08b07a9 |
| round45-standard-tests.log，初跑失敗 | 8bc7e605492ba117373996988fed513152846baff950b6a9822b9ef7a8d6546a |
| round45-standard-retest.log | ee29de685f02cd2d3422bf1f244a218c152d187ebad3a715b73bf0089855f57e |

088一次性跳過子契約已驗，整體保持DRAFT。#89與#79 OPEN，正常索敵配對及其他八步未完成。

## 第46輪讀檔陣營上游

[研究384](../re/384-load-faction-selection.md)與[完整座位收據](../internals/data/load-seat-selection.json)
SHA-256 `f8ba6dacc701418f1fe88c0c785fd871af8677916e4a54ed0ad1f9c19debf129`。
原版9組完整RAM／色號／PNG／DAC雙跑相同；十個命中邊界另標單次。
正式bootstrap／Game三選擇55份輸出雙跑相同，九個WORK bytes、兩方財政及完整存讀通過。
載入雙人這輪帝國／同盟確認直接命令畫面，交接保持全域FLEET／SHIP。
真實X11五案含右鍵反白、三選擇與兩倍視窗命中通過，未使用script或direct-entry。

四種選單6076 pixels均零差，但各全圖仍16333差。帝國／同盟載入及雙人兩確認後全圖零差；
雙人未確認仍2841差及STELLAR+7D共22 bytes未知，WORK活動陣營／座位仍差。
比較預定nearby-state，完整畫面零差的幀不事後升格。原圖、有效色盤、Game圖及完整差分均保留。

正常S3首次實行17份輸出雙跑相同；事前固定D3DC167E，7支次序／55筆池／結尾B6E4B78A保持。
WORK差36減33，FLEET120／SHIP865／STELLAR44不變，日期380／381及5863 pixels差待續。
21限定PASS零skip，含正常兩陣營增強AI及第二次AI旗標；標準18套件同命令乾淨重跑通過。
初跑舊正常讀檔測試少選陣營，補正式步驟；原有玩法及保存判準沒有放寬。

| 本機log | SHA-256 |
|---|---|
| round46-load-final.log | 0370dd1662f19e065087e355d283b58863f6981897f959061d15a03f1deb0c52 |
| round46-standard-tests.log，初跑失敗 | 29e89e7b80bdeac71a1258ad7ed9d2c21889ef4a49c11809626f1c8cb6d60bb7 |
| round46-standard-retest.log | dad502f9124555fa8f0e659a5f12681a6656481187ccb9606689140e0743082a |

088兩個READY子契約已驗、整份仍DRAFT，087及151個受控索敵幀保持。#89／#79仍OPEN。

## 第47輪回合進位上游

[研究385](../re/385-turn-counter-boundary.md)、[完整計數器收據](../internals/data/turn-counter-boundary.json)
SHA-256 `09d28a9e6d2b5c5cc49fd4f1ffa8d7c057170e4eb7f9e022baf6c9593d3a5a7b`。
固定dosgolem f7523f9、原版S3與D3DC167E，無戰鬥進位及首戰保持各雙跑完整RAM／全圖／色盤相同。
0374:03D9 `ff06188f`才是進位指令，watch03DD是執行後IP，251追加勘誤。

088子契約READY後修正正式Game：全部戰場及尾段報告完成且未終局後進位，移除turn-1補償。
正常S3十九份輸出雙跑相同，首戰與正式WORK保存380／799/3/1；7艦隊／55筆池及B6E4B78A保持。
WORK差33減32，完整640×480差5863減5853，FLEET120／SHIP865／STELLAR44不變。
原圖、使用有效原版色盤的前後圖、差分與逐區統計全部保存，nearby-state不升格、不遮罩。

48限定PASS零skip，18標準套件通過；包含正常三份隨附存檔、#85全表／全圖、兩AI雙陣營、
兩個戰場與費沙模態保持、9→10及終局不進位。正式X11兩尺寸全圖與正常Game相同，日期已目視。
外部resize兩次未通過，改用產品既有-scale及PID選視窗後同容器／命令重跑，不當產品缺陷。

| 本機log | SHA-256 |
|---|---|
| round47-counter-limited.log | 0e70196a99ce698b6a23c905e0f14c45db9d86f5e1faa350ba386b1abcdea0be |
| round47-standard-tests.log | 4a65ff8cdd39d980cd001326b2d02736f6dd4d58d78581bc3ea660ef9ada7657 |

088三個READY子契約已驗，整份仍DRAFT。#89／#79 OPEN，活動座位與剩餘全表差異、正常索敵同態另續。

## 第48輪活動控制上游

[研究386](../re/386-active-seat-lifecycle.md)、[本輪goal](../goal/2026-10-03-active-seat-round-48.md)、
[完整收據](../internals/data/active-seat.json) SHA-256 `40ce3fbedda34a073f80f0799a820e27f567b9cb3210cf6c282ef67d66f87bac`。

第48輪活動陣營／座位子契約已驗。原版三組各雙跑，完整RAM／色號／PNG／DAC相同；
四個IDA原名／邊界、102條原始指令及7項proven語意保存。WORK+09／+0A隨正常控制邊界更新，
固定玩家視角與AI選擇保持，+0B語意unknown，沒有填入首都。
正式讀檔／交接55檔及帝國正常首戰19檔各雙跑相同，活動控制及正式WORK保存符合原版。
44限定PASS零skip，18標準套件及七案真實X11通過；兩AI在兩陣營仍會建隊／下令。
首戰380／799/3/1、7艦隊／55筆池／B6E4B78A保持。全圖5853 pixels及WORK32／FLEET120／SHIP865／STELLAR44差保持。
同盟435M仍處理FF／1；650M首戰1／0為單次輔助，不宣稱完整同盟首戰雙跑。
雙人未確認2841 pixels、STELLAR+7D22 bytes及剩餘首戰資料待續；nearby-state不升格。
088完整入口仍DRAFT，#89／#79 OPEN；087及151個受控UI幀保持，完整八步未完成。

```sh
bash tools/active_seat_oracle.sh normal89-r48-clock-a
bash tools/active_seat_oracle.sh normal89-r48-clock-b
bash tools/active_seat_oracle.sh normal89-r48-alliance-a
bash tools/active_seat_oracle.sh normal89-r48-alliance-b
bash tools/active_seat_oracle.sh normal89-r48-versus-a
bash tools/active_seat_oracle.sh normal89-r48-versus-b
tools/gui.sh xrun go run tools/scripts/load_seat_player.go -out workplace/round48-player-final-a
tools/gui.sh xrun go run tools/scripts/load_seat_player.go -out workplace/round48-player-final-b
tools/gui.sh xrun go run -overlay workplace/round44-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round48-normal-final-a
tools/gui.sh xrun go run -overlay workplace/round44-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round48-normal-final-b
tools/gui.sh xrun env LOGH3_ACTIVE_GUI_PREFIX=round48-seat-checked-gui bash tools/active_seat_gui.sh
tools/gui.sh xrun env LOGH3_TURN_GUI_PREFIX=round48-final-gui bash tools/turn_counter_gui.sh
tools/gui.sh xrun go test ./internal/... ./cmd/... -count=1
tools/gui.sh xrun python3 tools/scripts/active_seat_receipt.py
```

各批次拒絕覆寫；重跑時必須先用獨立工作目錄保存當輪來源與收據。
原版battle資料沿第47輪固定版本收據，沒有執行器改動；不重生舊46／47輪JSON來替換歷史來源身份。
新局模板陣營2已按原版初始命令修正；無移動命令的同步回合不會停在處理畫面。
舊測試前提的限定／標準失敗log保留，改正常移動抽驗後同容器／命令重跑通過。
完整畫面與所有表差異照實保存，沒有排除未知+0B或雙人+7D。

| 本機log | SHA-256 |
|---|---|
| round48-active-limited.log | 533e607a97f4a264a49877338d33b5e38c5361a0fefe8e596e0d431b5d34c59a |
| round48-standard-tests.log | 042fa5a848621b2747dcd96f49e03d0ad219564160d59e8b1af19f2ae6605be7 |
| round48-active-limited-fixture-failed.log | 2e207f1a2f0ad89434df49e3b98de1096ee242558758dc88075ac71fbc3e11fe |
| round48-standard-fixture-failed.log | c76468a27a305b2c1c88933aa3e295cf6d62aa2d0217e54ba831592d9e72b179 |

最終GUI全圖核對另抓到雙人交接7526 pixels差，PNG仍為帝國命令，StateCommandPhase不足以確認陣營。
正常Game的55／19檔仍與初次各74檔相同；未改產品以遷就錯圖。
保留short-hold腳本與失敗圖。按下、放開與移動間隔調為0.2秒，900ticks後擷取，腳本每案加cmp全圖。
哪一次短按未被處理未定位，保持unknown；不把這次序列未完成當成原版或正式規則差異。

## 第49輪月底旗標上游

[研究387](../re/387-month-seat-gate.md)、[本輪計畫](../goal/2026-10-03-month-seat-gate-round-49.md)及
[完整收據](../internals/data/month-seat-gate.json) SHA-256 `1a06468e6cf408a3e4bbae057bf6a7111cc1ba1fca548d7e025b9d21d3939217`。
原版自然S3旗標0與只改同盟旗標1各雙跑完整RAM／全圖／色盤相同；載入只差線性5DB25。
原版旗標0不實際收稅／造船，旗標1才執行並清0。IDA84條bytes與主迴圈重設重新核對。

088子契約READY後接正式單人raw旗標閘門、清除／重設與保存。十九份正常S3輸出雙跑相同，
首戰STELLAR4032／PLANET6642 bytes全同，WORK32減14；軍事費／總稅收／建造符合自然原版。
FLEET120／SHIP865與全圖5853 pixels差保留，原生全圖及逐區色號／有效RGB完整記錄，nearby-state不升格。
380／799/3/1、7艦隊／55筆池／B6E4B78A保持。完整088仍DRAFT，#89／#79 OPEN，087受控UI幀不重開。

37限定PASS零skip、18標準產品套件與正式X11一倍／兩倍正常首戰通過，日期／階段已目視。
兩AI與兩陣營的正常存讀、旗標0／1／FF、非月初、9→10與不重複結算通過。
首次測試錯把正常讀檔重算未安定度當重複結算，修正比較停點後同命令通過，產品未因這次失敗改動。
失敗與最終log雜湊在收據；雙人跨月、同盟首戰全圖、任意人類旗標1讀檔月報與實機聆聽未新增通過聲明。

```sh
bash tools/month_seat_gate_oracle.sh normal89-r49-natural-a
bash tools/month_seat_gate_oracle.sh normal89-r49-natural-b
bash tools/month_seat_gate_oracle.sh normal89-r49-due-a
bash tools/month_seat_gate_oracle.sh normal89-r49-due-b
tools/gui.sh xrun go run -overlay workplace/round44-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round49-normal-a
tools/gui.sh xrun go run -overlay workplace/round44-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round49-normal-b
tools/gui.sh xrun env LOGH3_TURN_GUI_PREFIX=round49-gui LOGH3_TURN_PLAYER_DIR=round49-normal-a bash tools/turn_counter_gui.sh
tools/gui.sh xrun go test ./internal/... ./cmd/... -count=1
tools/gui.sh xrun python3 tools/scripts/month_seat_gate_receipt.py --final
```

原版與玩家輸出拒絕覆寫，重建請用保留同版本來源的乾淨工作副本。
測試期seed覆蓋由[既有工具](../../tools/scripts/battle_normal_probe_seed.py)產生，只固定實行前D3DC167E。
完整命令列、父snapshot與工具映像身份、來源SHA、控制注入及未解全紀錄保存在JSON。

## 第50輪到達配置及正常首回合無報告

[研究388](../re/388-fleet-arrival-placement.md)、[計畫](../goal/2026-10-03-fleet-arrival-placement-round-50.md)、
[完整收據](../internals/data/fleet-arrival-placement.json) SHA-256
`58fd315475e4756f671078e71044d319078689a53b2c6f7d25ce2f24a8516a2a`。

原版到達六份及正常第一戰術實行五份輸出各雙跑相同，固定dosgolem f7523f9／clean。
九艦隊含兩支未參戰者、七十筆池現在／集結座標及航向驗證後接088 READY子契約。
正常首回合兩座位查詢都完成，標頭及展示常式零次；舊近似入場的發現報告不再作玩家路徑判準。
受控151幀／來源／等待／保存保持原局面，未換名成正常S3。

十九份正常bootstrap／Game輸出雙跑相同；380／799/3/1、7艦隊／55戰場池／B6E4B78A保持。
FLEET全3904 bytes差120減27，SHIP全27456差865減276；WORK14、STELLAR4032／PLANET6642全同。
全圖5853、Field73、右側5780 pixels仍差，有效RGB相同口徑；底部區零差。
原圖／前後圖／並排與差分已目視，不裁切、縮放或遮罩。剩餘指標、選取與sprite差異保留。
真實cmd/gin3一倍／兩倍正常首戰，全圖與Game相同且日期／輸入階段已目視。

110限定PASS零skip、18標準套件通過。原版逐圈fixture補originalBattleAI，停用061刻意偏離；
普通單人CPU留原地／不傷害與原版正對照都通過，規則公式未因政策改動。
初跑直接入口、缺資源、CPU歸隊及原版fixture未關政策的失敗log在JSON，原golden不放寬。
完整088仍DRAFT，#89／#79 OPEN；同盟／雙人全流程、其他八步與平台音樂驗收未新增通過聲明。

```sh
bash tools/fleet_arrival_oracle.sh normal89-r50-arrival-a
bash tools/fleet_arrival_oracle.sh normal89-r50-arrival-b
bash tools/fleet_arrival_oracle.sh normal89-r50-arrival-spot-a
bash tools/fleet_arrival_oracle.sh normal89-r50-arrival-spot-b
tools/gui.sh xrun python3 tools/scripts/battle_normal_probe_seed.py --prefix round50-normal
tools/gui.sh xrun go run -overlay workplace/round50-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round50-final-a
tools/gui.sh xrun go run -overlay workplace/round50-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round50-final-b
tools/gui.sh xrun env LOGH3_TURN_GUI_PREFIX=round50-gui-final LOGH3_TURN_PLAYER_DIR=round50-final-a bash tools/turn_counter_gui.sh
tools/gui.sh xrun go test ./internal/... ./cmd/... -count=1
tools/gui.sh xrun python3 tools/scripts/fleet_arrival_receipt.py
```

各批次拒絕覆寫；在保留本輪來源的乾淨工作副本重建。測試期seed覆蓋讀本輪產品的rebellion.go，
不能沿用第44輪舊覆蓋，否則會把正常入口接回舊StartBattle。完整來源SHA與argv在JSON。

## 第51輪上游正常Input摘要

沿#89及088 READY補純移動摘要與兩個SPRITE 10反白物件；
[研究389](../re/389-battle-input-display.md)及[完整收據](../internals/data/battle-input-display.json)
SHA-256 `c1f387c1e3729450f4b7b3f9f8babb21b0b8f49d1b0762b1f6e0cca648cfb5c6`。
原版五組RAM／畫面／色盤雙跑相同；正常Game55檔雙跑與兩尺寸X11通過，61限定PASS、18套件。
初始首戰5853→2 pixels；旗艦／分隊各1、空白／右面板全圖0，五案文字區0／20480。
全資料仍WORK14／FLEET27／SHIP276差，完整入口nearby-state，不由單側全圖或局部同圖升格。
151受控索敵幀及087 READY保持；正常玩家完整索敵配對仍待#89，完整八步未完成。

## 第52輪上游摘要小地圖

沿#89與088第八個READY子契約接旗艦集結航線及一般分隊集結點，XOR色號4。
[研究390](../re/390-battle-hover-marks.md)、[完整收據](../internals/data/battle-hover-marks.json)
SHA-256 `91ec1901d3232702c3612d4edd459d72ff60c4633483bc578339b0e3f5c56250`。
四案原版各獨立雙跑，完整RAM／色號／PNG／DAC相同；兩個非零集結點案是明示注入診斷。
105條原始IDA指令及12項語意保存，所有FF物件的word+4初始化仍unknown，不猜改排序。

正常Game55檔雙跑相同，五張小地圖各0／16384；初始全圖2→1，其他四張全圖0。
WORK14／FLEET27／SHIP276／STELLAR0／PLANET0與正常799/3/1、回合380、7艦隊／55池保持。
45限定PASS零skip、18產品套件及一倍／兩倍正常X11通過。重畫不累積XOR，純移動完整保存及seed不變。
已目視完整並排圖；剩餘(255,255)由方向物件覆蓋，未遮罩或填像素常數。
完整088 DRAFT、#89／#79 OPEN、nearby-state及151受控幀門檻保持。

```sh
for kind in flagship squad fleet-target squad-target; do
  for side in a b; do
    bash tools/battle_hover_marks_oracle.sh "normal89-r52-$kind-$side"
  done
done
tools/gui.sh xrun python3 tools/scripts/battle_normal_probe_seed.py --prefix round52-normal
tools/gui.sh xrun go run -overlay workplace/round52-normal-overlay.json tools/scripts/battle_normal_probe.go -input-display -out workplace/round52-pass-a
tools/gui.sh xrun go run -overlay workplace/round52-normal-overlay.json tools/scripts/battle_normal_probe.go -input-display -out workplace/round52-pass-b
tools/gui.sh xrun env LOGH3_TURN_GUI_PREFIX=round52-gui-final LOGH3_TURN_PLAYER_DIR=round52-pass-a bash tools/turn_counter_gui.sh
tools/gui.sh test ./cmd/... ./internal/... -count=1
tools/gui.sh xrun python3 tools/scripts/battle_hover_marks_receipt.py
```

各輸出拒絕覆寫；重建需本輪來源及保留第51輪父snapshot，雜湊與完整旗標在JSON。
原版無注入兩案及正常Game五案與primitive診斷分列，不能將靜態相同升為完整資料或八步完成。

## 第53輪：未出擊偵察隊快取

[研究391](../re/391-battle-scout-cache.md)、[第53輪](../goal/2026-10-03-battle-scout-cache-round-53.md)及
[結構化收據](../internals/data/battle-scout-cache.json)，SHA-256
`1203dc687f6f2481ac53acfaa3a2da8866d8d82638ebeda6a9dd452d875bf3d9`。

原版雙跑沿用第51輪固定f7523f9，全部來源及輸出雜湊重核；本輪沒有更動執行器或重新擷取。
既有IDA209條原指令逐byte符合入口RAM，十項附加語意proven。
正常Game55檔a／b相同，修正僅七個偵察池+2F bytes，SHIP276→269。
WORK14／FLEET27／STELLAR0／PLANET0及380／799/3/1／7艦隊／55池／B6E4B78A保持。
五張完整畫面與第52輪相同，原PNG／差分圖沿用並核雜湊；初始全圖仍1 pixel，其餘四張0。
137限定PASS零skip、18產品套件及正常cmd/gin3 X11兩倍率通過。
兩AI全部55筆快取、七筆未出擊池及完整SHIP存讀已驗，可點選單位仍48筆。
重播命令在研究391；完整088 DRAFT、#89／#79 OPEN，nearby-state與151受控幀範圍保持。
