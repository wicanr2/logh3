# 045：索敵回收與移動畫面

## 目前續證：第三十五輪隨附存檔正常路徑

2026-10-02，#79保持open。父提交`1a7c5e5b1aef3eee08baf75aa4d5ca9c9aa71087`加本輪測試，測試時dirty。
沿079／080／083的既有READY範圍，未改產品規則或原版收據期望。
本輪是remake正常玩家抽樣，沒有雙方同狀態比較，不設對拍通過等級。
第34輪固定35邊界的same-state結論與舊nearby-state保持。

[正常測試](../../internal/game/battlesamples_test.go)走主選單、防拷、系統指令載入隨附存檔、
戰略實行、戰術實行、Input還原及正式SaveSlot／LoadSlot。
不注入艦隊、損傷、座標、資源或戰場；僅在首次戰略／戰術實行前固定各自seed。
戰略`D3DC167E`與產品預設相同，戰術`12345678`；不重擲，不宣稱整流程骰序與原版相同。
正式cmd/gin3的GUI抽樣另走產品預設seed，不當成受控Game同狀態比較。

| 存檔 | 首戰星域／戰略回合 | 艦隊／分隊 | 移動中間幀 | 補給／修理結果幀 | 完整存讀 |
|---|---|---|---:|---|---|
| S1，DT8 | 8／94 | 16／103 | 47 | 0／0 | SHIP 27456及FLEET 3904 bytes相同 |
| S2，DT9 | 13／347 | 6／45 | 0 | 0／0 | 同上 |
| S3，DTA | 26／381 | 4／31 | 0 | 0／0 | 同上 |

三份均回戰術第2回合與Input。S1有28支受損分隊，全部屬電腦陣營，玩家分隊未受損；
來源庫存存在也不表示玩家結果窗必須出現。三份首戰都沒有種類5補給艦。
正常資源結果幀仍未驗。S1固定再實行七回合的探索亦未出現玩家修理結果，不以重擲找局面。
這只覆蓋三份隨附存檔，不表示全部玩家存檔已驗。

正式S1移動與S2第2回合640×480 PNG已檢視：
`workplace/battle79-round35-s1-gui-movement.png` SHA-256
`94decd8af5671ad2d6745c3fda62ad32cc60d80e7442d8a989a06b484c65c148`；
`workplace/battle79-round35-s2-gui-input-corrected.png` SHA-256
`1119b99bec78d93388b138665a9b80217e4dbd5d562c13a191ee55f1e8be92a9`。
首份S2腳本誤把兩次點擊當成戰術實行，實際只關OccupationReport與TurnReport。
延長等待不能越過模態報告；唯讀逐段觀測確認後，補兩次正常點擊才到第2回合。
舊第1回合圖與失敗log保留，不分類成產品缺陷。探測誤以Int讀uint8、收據誤認48-byte DAC為768 bytes亦只屬驗證工具失敗；改正後同命令通過。

完整收據`workplace/battle79-round35-normal-receipt.json` SHA-256
`551a2d2b187a3e67f1bc95a0a49a4597abce690920e6055d83ad259ef54d6960`，
由[收據工具](../../tools/scripts/battle_samples_receipt.py)重生，保存七張Game原生PNG、兩張正式GUI、
色號／DAC、18份存檔及5份必要原版素材、來源與14份log的雜湊。目標4個通過事件、零skip；
18個標準產品套件通過或無測試，count=1。工具鏈沿用第34輪身份，沒有重建image。

```sh
bash tools/gui.sh xrun sh -c 'LOGH3_MOVEMENT_OUTPUT=battle79-round35-normal go test -count=1 -v -run TestBattleShippedSavesNormalExecuteAndSave ./internal/game'
bash tools/gui.sh test -count=1 ./internal/... ./cmd/...
# S1正常移動中；S2需先關兩張報告，再選戰術實行。
bash tools/gui.sh run ./cmd/gin3 -orig /orig -data /src/docs/internals/data -script 'move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 544 125; click 560 145; click 300 282; click 318 208; click 544 125; click 560 125; tick 2000; move 600 240; click; click; tick 10' -shot-bin /src/workplace/battle79-round35-s1-gui-movement.bin -shot /src/workplace/battle79-round35-s1-gui-movement.png -shot-after 2
bash tools/gui.sh run ./cmd/gin3 -orig /orig -data /src/docs/internals/data -script 'move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 544 125; click 560 145; click 300 302; click 318 208; click 544 125; click 560 125; tick 15000; move 600 240; click; click; click; click; tick 4000' -shot-bin /src/workplace/battle79-round35-s2-gui-input-corrected.bin -shot /src/workplace/battle79-round35-s2-gui-input-corrected.png -state-out /src/workplace/battle79-round35-s2-gui-input-corrected.state -shot-after 2
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 bash tools/py.sh tools/scripts/battle_samples_receipt.py 1a7c5e5b1aef3eee08baf75aa4d5ca9c9aa71087 dirty
```

原版探索另存`workplace/oracle/battle79-round35-s1-explore/`，dosgolem完整HEAD為
`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`、工作樹clean，旗標與所有輸出雜湊在收據。
父snapshot為`dosgolem-logh3/workplace/fcmd/fcmd-base.state`，279725 bytes，SHA-256
`c55de581a6393e0e39e304dd0f191e9d1cfa8457b86b93f2a9e3368b5751ba26`。
沒有新增poke；初態seed由固定snapshot而來，收尾仍D3DC167E。
750000000步仍停戰略艦隊過場，四個戰術入口命中均0；不是第二戰場收據，未與remake比較。
下一步固定續過原版過場取戰場初態，才做第二戰場同狀態；正常補給／修理結果亦仍缺。
080／083保持READY，062完整八步未完成，不重做已驗三份remake路徑。

收尾另由[#88](https://github.com/wicanr2/logh3/issues/88)修共用Issue證據解析，
第一個反引號區段才是路徑，章節附註保留於原Issue；缺前綴、結尾或空路徑均拒絕。
舊解析八案中五案失敗，修正後八案及21個頂層工具測試通過、零skip。
完整快照另外抓到#32的原版收據敘述與三條縮寫；將敘述移出路徑段、補實際檔名，
保留原判準及狀態。#88曾在完整快照驗證前過早關閉，已重開補驗再收尾。
782條既有及新增相對連結全部有效，沒有手改受控worklist或矩陣。
工具修正前的正常玩家收據保留為`workplace/battle79-round35-before-worklist-receipt.json`，
SHA-256 `c5dfd9e98f38efda406a19e89b64a0b432570c244a15101e454af5a6cb284789`；
最終收據另記工具修正及最終標準測試，不以舊收據宣稱目前工具通過。

## 目前續證：第三十四輪階段面板與35邊界

2026-10-02，#79保持open。父提交38794ab1f36a5e94be624b587ca2eded66c4fe8c加本輪修改，測試時dirty。
[研究375](../re/375-battle-phase-panel.md)審查原bytes、字串、清窗及呼叫順序後，
[083](../spec/083-battle-phase-panel.md)先READY再接正式標籤及Input還原。
非Input不再印日期、分數或星域，也不讓unit情報覆蓋階段文字。

新收據事前宣告same-state，只限第6回合固定局面、7艦隊／55分隊、鏡頭(66,30)及35個指定邊界。
原版雙跑105檔相同，來源snapshot、完整父序列／seed注入、新旗標及各圖／RAM／報告雜湊見
[逐圈全紀錄](../internals/data/battle-phase-frames.json)，SHA-256
`e1a149fed1eb5767a8011a71ea52f356444cdbfe55da29543774f554f4025941`。
Game由正常入口後明示只注入一次原版entry及鏡頭，之後正式規則自行逐圈執行。
原版輸出只供比較，沒有每幀餵入引擎。7×122＋55×78 bytes逐幀相同，
兩側seed固定0x12345678且35幀不變，鏡頭亦不變。原色號與有效RGB均相同。
診斷同狀態不代替正常原版玩家路徑，不外推第二戰場或其他回合；舊nearby-state不升格。

| 每一幀區域 | (x,y,w,h) | 比較像素 | 相同 | 不同 | 最大通道差 | 差異bounding box |
|---|---|---:|---:|---:|---:|---|
| 全畫面 | (0,0,640,480) | 307200 | 307200 | 0 | 0 | 無 |
| 戰場視窗 | (16,60,448,360) | 161280 | 161280 | 0 | 0 | 無 |
| 小地圖 | (480,75,128,128) | 16384 | 16384 | 0 | 0 | 無 |
| 階段面板 | (464,220,160,200) | 32000 | 32000 | 0 | 0 | 無 |

35幀全畫面合計10752000像素，零差異、零未核准差異。不縮放、裁切或遮罩。
entry仍顯示修理，loop-00為初始化後移動，loop-01..32為各圈完成後，end為第33圈返回0後。
只驗圈邊界，不宣稱逐CPU指令、圈內重畫或硬體wall-clock一致。

完整收據workplace/battle79-round34-after-render-receipt.json，SHA-256
`34c6330dac14d0df1790d7a788dad462e855ca91924bd48c2f44bf793030cf72`。
它保存35組原版／remake／差分PNG、來源及輸出雜湊、逐區統計、有效RGB，
12張正常／診斷抽樣PNG各自標明範圍。原版與工具身份沿用下方身份表，執行器沒有改動。
新增IDA輸入及位址空間見[17筆分級索引](../internals/data/battle-phase-semantics.json)。
修正前第33輪收據ce3370d8保留，完整雜湊見下節，兩端面板差3971／3951；
本輪新收據均0，沒有改舊原圖、期望或比較等級。

正常主選單→S3→戰術命令，長程／隊形／索敵／分隊實行均顯示移動標籤，
完成後回Input，完整SHIP／FLEET正式存讀相同。正常索敵回收再放出維持原池81／單位31。
正常S3此回合補給／修理結果均0幀，不能宣稱正常資源結果畫面已驗。
fixture的正式結果面板與階段底圖通過，與正常證據分列。
080／083保持READY；第二戰場／全部存檔、正常資源結果幀及其餘八步仍待接續。

正式cmd/gin3正常第2回合PNG已目視，workplace/battle79-round34-normal-gui.png，SHA-256
`eb41879ca7d87b48fd1276365ed7a4b333542e4e596a7a850f548b971385b65b`。
中間loop-16的正式畫面亦已檢視，右側只留移動階段。
目標log有13個通過事件、零skip；標準18套件通過或無測試，count=1；正常S3 GUI逐位元組相同。
三份log的完整檔名與SHA-256均在完整收據logs欄。

```sh
bash tools/gui.sh xrun sh -c 'LOGH3_MOVEMENT_OUTPUT=battle79-round34-after LOGH3_PREPARATION_OUTPUT=battle79-round34-after go test -count=1 -v -run "TestMovementAllBoundariesMatchOriginal|TestMovementCommandsExecuteAndSave|TestGameMovementPhaseMatchesOriginal|TestBattlePreparationWholePhaseMatchesOriginal|TestBattlePreparationNormalPlayerExecute|TestScoutRecoverAndRelaunchNormal" ./internal/game'
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 bash tools/py.sh tools/scripts/battle_phase_render_receipt.py 38794ab1f36a5e94be624b587ca2eded66c4fe8c
bash tools/gui.sh test -count=1 ./internal/... ./cmd/...
LOGH3_GUI_PREFIX=fcmd-s3-found bash tools/gui_check.sh
```

正式cmd/gin3輸入沿用第32輪命令，輸出prefix換battle79-round34。
原版重生入口為研究375及tools/battle_phase_oracle.sh。
普通資源測試改寫自己的暫存目錄；保存時明設LOGH3_PREPARATION_OUTPUT，避免覆寫第28輪收據。
首次格式化誤用go run入口，改tools/go.sh gofmt後成功；go fmt碰到兩個無關測試已還原。
這些屬工具操作，不登記成產品缺陷。

## 第三十三輪歷史收據：小地圖標記

2026-10-02，工作項[#79](https://github.com/wicanr2/logh3/issues/79)，父提交
`be187b1bff7fa32f8831302886c2ec69362901a8`加本輪修改，測試時dirty。
本節是現行結果，下方第32輪數字與失敗保留為歷史。來源見
[研究374](../re/374-battle-marker-lifecycle.md)、[規格082](../spec/082-battle-minimap-markers.md)。
原版bundle、dosgolem HEAD、兩個端點局面、seed、鏡頭、畫布、素材及image與下方身份表相同。
原版既有entry／end雙跑雜湊重新核對相同；另以正常S3輸入新跑兩份標記初始化收據，
完整命令、計數、seed注入及55筆原池記錄在[標記JSON](../internals/data/battle-markers.json)。

正式renderer改讀+19旗標及低4位顏色、+1A／+1C快照；StartBattle重設3F，
正常輸入恢復時依viewer刷新。五個bytes獨立寫回原池，不覆蓋其他資源與移動欄位。
整組比較維持nearby-state，未遮罩、裁切或縮放；零差異只限指定區域與端點。

| 邊界／區域 | (x,y,w,h) | 比較像素 | 相同 | 不同 | 最大通道差 | 差異bounding box |
|---|---|---:|---:|---:|---:|---|
| entry全畫面 | (0,0,640,480) | 307200 | 303229 | 3971 | 243 | [480,241,607,395] |
| end全畫面 | (0,0,640,480) | 307200 | 303249 | 3951 | 243 | [480,241,607,395] |
| entry／end戰場視窗 | (16,60,448,360) | 各161280 | 各161280 | 0 | 0 | 無 |
| entry／end小地圖 | (480,75,128,128) | 各16384 | 各16384 | 0 | 0 | 無 |
| entry階段面板 | (464,220,160,200) | 32000 | 28029 | 3971 | 243 | [480,241,607,395] |
| end階段面板 | (464,220,160,200) | 32000 | 28049 | 3951 | 243 | [480,241,607,395] |

修正前收據`workplace/battle79-round33-before-render-receipt.json`，SHA-256
`f163725fcf193aa9d36e1798796676e6e3f01bc65b9df922f50323dfda18c3b5`，小地圖各差35。
修正後收據`workplace/battle79-round33-after-render-receipt.json`，SHA-256
`ce3370d8b4717cc58a6b61b64aaf6faaf0652d32a2141db3b466e7fe9dbd3c33`，小地圖各差0。
兩份保存完整原圖、remake、差分、逐區統計與產品來源雜湊；下列PNG均在本機workplace。
剩餘全畫面差異全部屬於未對齊的階段面板，未核准豁免。

| 第33輪PNG | SHA-256 |
|---|---|
| `battle79-round33-after-entry-original.png` | `53348a02f06d8725dbcbf4bf31539261cab8f4642dc27349e1c6eb5920f45a56` |
| `battle79-round33-after-entry-remake.png` | `2b5f6e071d67f8b66cd6fd520a2ed13b9f8103ea056e8ef3e450d3596cbe4eea` |
| `battle79-round33-after-entry-difference.png` | `2abd759ddc5a28a6a750b0d9fe2d10feb8bb0bc4613f70ea4aa5866757b3d1ae` |
| `battle79-round33-after-end-original.png` | `339606f012b1018966074099d8a04e83d25c1bb7830d42a007f031a091fd4c48` |
| `battle79-round33-after-end-remake.png` | `b8931ad2237f93b8844a12f63ebd22221bed2a0fb53a343ee61b0865da35fa44` |
| `battle79-round33-after-end-difference.png` | `6cc92a741c01f39258c5f1d5f3ec6fb8cab9d56896a148b3ba140f724bc23a48` |

原版96案完整78-byte記錄與正式UpdateBattleMarker／ApplyBattleMarker全數相同。
`TestBattleMarkersNormalEntryAndSave`由主選單到S3正常實行，無單位、旗標或座標注入；
55筆原池旗標及8筆已畫快照符合原版，首次戰術實行前完整SHIP 27456／FLEET 3904 bytes存讀相同。
這是標記欄位比較，不外推完整進場狀態或畫面same-state。
正常索敵放出、移動、歸隊、回收及再放出仍通過，原池81／單位31保持；
回收後完整SHIP／FLEET正式存讀相同，四張正常階段圖的雜湊保存在修正後收據。
正常正式cmd/gin3到第2回合的PNG已檢視，
`battle79-round33-normal-gui.png` SHA-256
`eb41879ca7d87b48fd1276365ed7a4b333542e4e596a7a850f548b971385b65b`。

```sh
bash tools/gui.sh xrun sh -c 'go test -count=1 -v -run TestBattleMarkerMatchesOriginal ./internal/rules && go test -count=1 -v -run "TestMinimapMatchesOracle|TestBattleMatchesOracle" ./internal/screen && LOGH3_MOVEMENT_OUTPUT=battle79-round33-after go test -count=1 -v -run "TestGameMovementPhaseMatchesOriginal|TestScoutRecoverAndRelaunchNormal|TestBattleScrollMatchesOracle" ./internal/game'
bash tools/gui.sh xrun go test -count=1 -v -run TestBattleMarkersNormalEntryAndSave ./internal/game
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 bash tools/py.sh tools/scripts/battle_movement_render_receipt.py be187b1bff7fa32f8831302886c2ec69362901a8 dirty battle79-round33-after battle79-round33-after
bash tools/gui.sh test -count=1 ./internal/... ./cmd/...
LOGH3_GUI_PREFIX=fcmd-s3-found bash tools/gui_check.sh
```

正常cmd/gin3輸入與下方第32輪命令相同，只將輸出prefix換成battle79-round33。

| 第33輪本機log | 結果／SHA-256 |
|---|---|
| `battle79-round33-target-fixed.log` | 96案、兩份舊小地圖、完整移動、正常回收及12組鏡頭通過，零skip；`379d0716355abc2f29cde12edf356e77c5a3b7bfcb512257e2252cbba3ba56df` |
| `battle79-round33-normal-entry.log` | 正常初始化與首次實行前完整存讀通過；`f0e4c67fc471ab0186239bdd5ee197ff5e8aedf270cfb0d27fc2183bb9f4aa1f` |
| `battle79-round33-standard.log` | 18個標準產品套件通過或無測試，count=1；`94c8cb1df791c9cfee674f587f5e10b302a970e1d5790075c4f5e921d705c80c` |
| `battle79-round33-gui.log` | 正常S3 GUI逐位元組相同；`3b9b080e7ecf23ca565e065ed04b3eb6c986bbef1c4c29534c4474dde14faec6` |

初次舊收據測試因誤用艦艇+02篩星域而漏掉全部標記，失敗log
`battle79-round33-target.log` SHA-256
`4dc96ff06a64e6e605ee330dd95c87e8e3c3e4989655af2dcb26edab87ad6831`。
原版該欄為0，改讀DS:9898艦隊名單後同容器同命令通過，沒有改原版期望畫面。
082只完成正常單人標記範圍；080維持READY，#79及062八步仍未完成。
後續驗階段面板、中間動畫、第二戰場及其餘八步；不重開已驗標記鏈。

## 第三十二輪歷史收據

工作項：[#79](https://github.com/wicanr2/logh3/issues/79)。父提交32af819。
正常索敵隊放出、移動、歸隊、回收及再放出，原池81與單位31保持。
回收後攔截點選已修正；畫面另依原版完整資料與鏡頭做診斷，不宣稱完整第二階段完成。
原始證據及限定呈現審查見[研究373](../re/373-battle-interleaved-movement.md)，契約見080。

原版來源與完整旗標由[研究371](../re/371-battle-movement-phase.md)及
[三邊界JSON](../internals/data/battle-movement-phase.json)管理。
畫面比較固定nearby-state：完整移動資料與鏡頭已相同，但階段面板等尚未全部對齊。
新畫面、全畫面與逐區差分統計由[收據工具](../../tools/scripts/battle_movement_render_receipt.py)產生，
不裁切、不縮放、不遮罩。正式完整對拍門檻保持。

## 身份與重播條件

| 項目 | 固定身份／方法 |
|---|---|
| remake | `32af81959dd82a4e473f1eafb7aa98019a50cfc8`加本輪修改，測試時工作樹dirty；產品與測試來源雜湊均在下述收據 |
| bundle | `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673` |
| 原版GIN3.COM | `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918` |
| dosgolem | 專用副本`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，乾淨、未修改 |
| 原版局面 | 第6回合、7艦隊／55分隊；初態及全部輸入由三邊界JSON管理，不以正常S3初態替代 |
| 原版重現 | 沿用371入口／終態各a、b兩份完整RAM、色號、色盤及PNG，重核全部SHA-256相同；本輪沒有重新執行原版 |
| 亂數 | 原版與remake在移動前固定`0x12345678`；移動階段不消費亂數，後續階段不作整段骰序一致聲明 |
| 鏡頭 | 原版DS:9BF4／9BF6，dosgolem線性DS基底54B30，讀lin:5E724／5E726；世界(1056,600)，格(66,30) |
| 診斷注入 | 正常S3入口後明示載入原版entry資料、顯示等級及鏡頭；明設移動邊界與滑鼠(240,240)。使用正式產品renderer，不把診斷入口當正常原版路徑 |
| 畫布／色盤 | 原生640×480、16色；原版PNG有效RGB與remake的48-byte DAC依正式Palette.RGB8展開相同；沒有裁切、縮放或排除區 |
| 建置／測試 | Go1.24.13，`logh3-ebiten:2`：`sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8` |
| oracle Go image | `golang:1.24-bookworm`：`sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac` |

三邊界JSON SHA-256：`b920eb42ccc70c8c447a9a2bd53a5e40f9efae5e807929fe4972fa3cc5fbe516`。
本輪完整收據：`workplace/battle79-round32-final-render-receipt.json`，
SHA-256：`89918162ffc028e848606744a13ca7221dbcedf4f1e4505807a3f97182afd1c6`。
它保存原版完整旗標、步數、各項注入、來源與輸出雜湊、產品來源、原版素材及所有區域數字。
entry停在dosgolem CS:IP `181B:00FB`、計數1213409021；end停在`181B:0263`、計數1226638239。
原版可重生方法仍以371與`tools/battle_movement_oracle.sh`為準。
本輪比較圖的原版PNG只是從既有原始色號重編碼，另記雜湊，沒有登錄成新oracle收據。

## 全畫面與區域結果

區域從產品[screen-battle.json](../internals/data/screen-battle.json)推導。
bounding box為含端點的`[最小x,最小y,最大x,最大y]`，通道差使用RGB8。
整組保持nearby-state，零差異限表內區域與兩個端點。

| 邊界／區域 | (x,y,w,h) | 比較像素 | 相同 | 不同 | 最大通道差 | 差異bounding box |
|---|---|---:|---:|---:|---:|---|
| entry全畫面 | (0,0,640,480) | 307200 | 303194 | 4006 | 243 | [480,134,607,395] |
| end全畫面 | (0,0,640,480) | 307200 | 303214 | 3986 | 243 | [480,127,607,395] |
| entry／end戰場視窗 | (16,60,448,360) | 各161280 | 各161280 | 0 | 0 | 無 |
| entry小地圖 | (480,75,128,128) | 16384 | 16349 | 35 | 162 | [551,134,568,164] |
| end小地圖 | (480,75,128,128) | 16384 | 16349 | 35 | 162 | [551,127,567,152] |
| entry階段面板 | (464,220,160,200) | 32000 | 28029 | 3971 | 243 | [480,241,607,395] |
| end階段面板 | (464,220,160,200) | 32000 | 28049 | 3951 | 243 | [480,241,607,395] |

其餘像素零差異；全部剩餘差異等於小地圖與階段面板差異之和。
上下外區51200像素原版為黑色色號3，修正前錯用藍色色號0。
圖示9應不畫，修正前被貼成藍色方塊；回收Out亦不畫、不攔截點選。
移動端點原版沒有游標，正式移動階段藏游標；尚未核所有中間動畫相位。

小地圖差異分類為標記資料來源／呈現差異，研究373記錄原版標記旗標、快照及顏色的
可丟棄重播可令兩個端點零差異。正常初始化／更新鏈未核，正式產品尚未改。
階段面板分類為階段狀態尚未對齊，未核准豁免或以遮罩排除。兩項均阻塞整階段畫面完成。

下列檔案均在`workplace/`，只留本機；完整原版來源PNG雜湊另在三邊界JSON。

| 本輪PNG | SHA-256 |
|---|---|
| `battle79-round32-final-entry-original.png` | `53348a02f06d8725dbcbf4bf31539261cab8f4642dc27349e1c6eb5920f45a56` |
| `battle79-round32-final-entry-remake.png` | `de29911bc5a5e2139446e6df4e8d84d113b72084f7a826b930f5f24d81eb48f3` |
| `battle79-round32-final-entry-difference.png` | `19faf1128f2468433563202b7fc55a6cf5576c166c8c0b071c503a8df8cf0146` |
| `battle79-round32-final-end-original.png` | `339606f012b1018966074099d8a04e83d25c1bb7830d42a007f031a091fd4c48` |
| `battle79-round32-final-end-remake.png` | `3435d47184192c65bb15517945349495ecbe01c8eef68aebe0b93b1fd82eddf8` |
| `battle79-round32-final-end-difference.png` | `74dbc95bad86e963d68f775f97fd0be5f0a33a873e223a37b4c8c4dc12860813` |

## 正常玩家與存讀鏈

`TestScoutRecoverAndRelaunchNormal`走主選單、防拷、S3、戰略實行到達希利戰場。
由滑鼠選旗艦「放出」、選目的地、實行，再點索敵隊「歸隊」、實行後回收。
沒有注入Out或回收旗標，種子只在首次實行前固定一次。
本局面一次歸隊回合即回收，Action=FF、標記bit7清、ScoutMode=FF／ScoutState=0。
回收單位不攔截旗艦點選，再放出重用原池81與單位31，不增配第二個單位。

回收後正式SaveSlot與正常LoadSlot完整SHIP 27456 bytes、FLEET 3904 bytes相同。
這是保存回收資料，不宣稱載入後可繼續中途戰鬥。
四張正常放出／移動後／回收／再放出PNG及原色號、DAC雜湊均在完整收據。
另用正式`cmd/gin3`正常入口到第2回合，保存並目視
`workplace/battle79-round32-normal-gui.png`，SHA-256
`eb41879ca7d87b48fd1276365ed7a4b333542e4e596a7a850f548b971385b65b`。
這些只作remake正常路徑抽樣，沒有升格為原版同狀態畫面。

## 命令與測試收據

```sh
bash tools/gui.sh xrun sh -c 'LOGH3_MOVEMENT_OUTPUT=battle79-round32-final go test -count=1 -v -run "TestGameMovementPhaseMatchesOriginal|TestScoutRecoverAndRelaunchNormal|TestBattleScrollMatchesOracle" ./internal/game && go test -count=1 -v -run TestBattleUnitsMatchOracle ./internal/screen'
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 bash tools/py.sh tools/scripts/battle_movement_render_receipt.py 32af81959dd82a4e473f1eafb7aa98019a50cfc8 dirty battle79-round32-final battle79-round32-final
bash tools/gui.sh test -count=1 ./internal/... ./cmd/...
LOGH3_GUI_PREFIX=fcmd-s3-found bash tools/gui_check.sh
bash tools/gui.sh run ./cmd/gin3 -orig /orig -data /src/docs/internals/data -script 'move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 544 125; click 560 145; click 300 322; click 318 208; click 544 125; click 560 125; tick 2000; move 600 240; click; click; tick 4000' -shot-bin /src/workplace/battle79-round32-normal-gui.bin -shot /src/workplace/battle79-round32-normal-gui.png -shot-after 2
```

| 本機測試log | 結果／SHA-256 |
|---|---|
| `battle79-round32-final.log` | Game全紀錄、正常回收／再放出／存讀、12案捲動與原版分隊圖示通過；`7474b7931a6265d9d862975a6e2410cff92d16d524b977d8fbbdac9e87398d7f` |
| `battle79-round32-standard-release.log` | 18個標準產品套件通過或無測試；`5fd37eba1296e3cdd103ae6d7269ea1e8188bf0e3e2ce0e0671aee692f211b65` |
| `battle79-round32-gui.log` | 正常S3 GUI既有收據逐位元組相同；`3b9b080e7ecf23ca565e065ed04b3eb6c986bbef1c4c29534c4474dde14faec6` |

本輪原版資料存在，目標驗證無skip。GUI篩選由環境變數設定，只跑指定收據。
綠色測試不取代全畫面同狀態驗收、第二戰場、其他八步階段或音樂聆聽。

## 保留的失敗與限制

- 正常回收首次失敗是產品點選缺陷，log SHA-256：
  `f1455f6e4161e3c8b68c6b913294692eaa85ae328e8014a35e04e37b0d530dbe`。
  依080排除Out後同命令通過，不改原版規則。
- 初次畫面診斷缺正式BLOCK_2.BLK及原版顯示等級；修資源與注入後才判定產品差異。
  原診斷收據SHA-256：`ef05dd5c11ee9bcdab2b34c22837d9fd667a10760295091b3a26730cdf5ce207`。
  補齊後尚未修呈現的收據為`ed5a8c489ea1469d69d0ffce5d8eb0ebccb76f083eb216519f1c2e02e22bdbc8`。
- 圖示與游標修正後，外區藍底仍造成51200像素差異；當時收據
  `175439f04010704b86b6bce7b5b746b0ea6bfea82030c8055fab3742d1b57f87`保留，不充當最後結果。
- 原版標記重播的可丟棄收據SHA-256：
  `506a6b538e66d737879d584984fb4ce481c28316b95f8f4c59ad28a159a884d7`。
  未接正式小地圖，不以模型零差異宣稱玩家路徑完成。

080保持READY，062八步與#79保持未完成。完整同狀態畫面仍須正常標記初始化／更新、
階段面板與中間動畫，第二戰場及全部存檔尚未抽驗。
