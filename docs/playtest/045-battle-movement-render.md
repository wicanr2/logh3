# 045：索敵回收與移動畫面

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
