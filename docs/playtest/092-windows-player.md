# Windows 玩家入口驗證

2026-10-05，第104輪，沿 GitHub #69。產品 Go 程式未修改。
目前提交為 `54c643262765bb9ff2b2c64c6032e28416fe347f`；使用第103輪固定
`9e1d318` 來源包重建的玩家程式。本輪不建立正式版號或交付目錄。

## 已驗範圍

| 項目 | 結果 | 證據等級與限制 |
|---|---|---|
| Windows 預設片頭自然到標題 | 兩次4998更新，皆StateTitle、末格420、退出0 | proven，Wine GUI；無腳本、輸入或局面注入 |
| 標題原生畫布 | 兩次PNG及307200色號逐位元組符合來源Linux基準 | proven，同來源remake；未新增原版oracle |
| 一倍正常取進度路徑 | 新遊戲、防拷、月報、船艦訂造、命令、系統選單、取進度 | proven，7張完整X11視窗與原生畫布符合Linux公開輸入對照 |
| 一倍正常艦隊路徑 | 戰略命令、出擊命令、提督、編成、艦隊編號 | proven，6張完整X11視窗及原生畫布符合對照 |
| 1.5倍輸入與原生畫布 | 960×720視窗可走取進度，307200色號相同 | proven，7張畫面正規化結果相同；完整實體像素未達相同 |
| 整數畫布及一倍全圖統計 | 37組11366400像素零差，最大通道差0，bbox空 | proven，無裁切、遮罩或排除矩形 |
| 1.5倍兩平台實際視窗 | 7組4838400像素，共154407不同，最大通道差1 | proven，差異完整保存，不稱全圖零差 |
| 標準產品測試 | 18套件，13通過／5無測試 | remake回歸；不外推原版parity |

三條最終GUI路徑共87項X11滑鼠移動、按下、放開及鍵盤命令。Windows使用正式未修改
`cmd/gin3` 的來源重建執行檔，預設片頭透過真實左鍵跳過，沒有使用 `-intro=false`。
Linux對照以相同來源執行檔和公開輸入腳本產生，屬診斷對照，不當成正常Linux玩家驗收。
本輪未指定曲庫，未新增Windows配樂播放、人耳或實體音效裝置證據。

## 身份與輸入

- 原版bundle `org_game/银河英雄传说3SP.zip`，SHA-256
  `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
- Windows玩家程式13458944 bytes，SHA-256
  `6af940c72f340a5ecc49b2c6dc092e0e6dca077c1136abada3cad09b49e8ffa9`。
- Linux來源玩家程式SHA-256
  `48c89ba6e9b7bf069ccba7f6345228af4ffc23a7baf63fd87da02aa44ff70754`。
- 建置來源包SHA-256
  `89be259435d35d22f3db7e27ffe2658bdafb67183abd592c4bc8b13e2af5f954`，見[091](091-current-source.md)。
- Wine 9.0，既有image `colonization-windows-verification:20261001-r1`，ID
  `sha256:7d3c7ff554db39300e3e7a92a7b79c2d5fcd647289596d22fa10899baac61201`。
- Linux對照及標準測試沿 `logh3-ebiten:2`，ID
  `sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8`。
- [唯一來源收據](../internals/data/windows-player.json)保存完整原版目錄雜湊、實際參數、
  二進位、畫面、動作時間戳、失敗探針及工具來源。原版、資料與工具唯讀掛載。

這些路徑不實行戰術或戰略回合，不新增RND規則比較。未使用測試種子或挑選隨機局面。
兩次自然標題及正常選單均從相同唯讀輸入啟動。

## 1.5倍差異的追查

最初把來源PNG用ImageMagick point濾鏡放大，7張完整視窗共402496像素不同，
最大通道差182。此對照假設不符合產品所用Ebitengine濾鏡，失敗統計保留在
`player69-r104-PIXELS.json`。

改擷取Linux實際960×720視窗，差異減為154407像素，最大通道差1，仍未通過
全視窗相同。`player69-r104-PIXELS2.json`保存這個完整失敗結果。
再以最近鄰模型列舉來源候選仍失敗，兩側各402496像素超出候選，反例保存於
`player69-r104-PIXELS3.json`。沒有以提高容許差或裁掉邊界讓全圖通過。

固定Ebitengine來源的 `run.go` 明訂screen濾鏡為方框濾鏡，`gameforui.go`
預設啟用；`internal/builtinshader/shader.go` 的 `ScreenShaderSource` 以
`p0=pos-0.5/scale`、`p1=pos+0.5/scale` 及雙軸mix計算實體像素。
產品未呼叫 `SetScreenFilterEnabled` 或覆寫 `DrawFinalScreen`。

1.5倍時，每第三個實體像素的中間一點跨來源像素邊界，須混合相鄰兩點；列、欄皆
跨界時混合四點。兩平台共4838400實體像素逐通道均落在這個等權平均值的下／上整數界，
差異全部在跨界列或欄，其他位置零差。這是可回查的量化範圍，尚未證明是哪一層GPU
浮點運算或轉整數造成兩平台不同。實體全圖不同的事實不因此消失。

完整原尺寸統計、bbox、最大通道差及模型核對見 `player69-r104-PIXELS4.json`。
範例由完整來源收據綁定，不以散落本機檔名單獨作證。

## 保留的環境與驗證失敗

- GUI與Wine image均無Pillow；未安裝主機套件。畫面核對沿既有Go擷取器及Wine image的ImageMagick。
- 初版1.5倍驅動器直接連送兩個右鍵，最後仍在船艦訂造報告。`load15`完整圖與TRACE保留。
  補上第一個右鍵後的完整報告回授，再用同一image、同一玩家二進位重跑 `load15b` 通過。
  送鍵過早為strong inference；沒有改產品速度、計時或報告規則。
- 初版Linux放大擷取後又拿180更新的防拷畫布比4更新畫布，只有底線32像素不同，
  位於 `(240,283)` 至 `(255,284)`。正式防拷PIT閃爍造成相位不同，部分 `reference15` 保留。
  `reference15b`每次保存完整擷取與相位探針，選擇完整邏輯畫面符合固定對照的擷取時刻。
  不遮掉游標，也不把稍後180更新的32像素差稱為零差。

## 重生入口

下列產物已存在。標題／GUI重跑時使用新的 `--revision`；參考重跑使用新的
`--reference`，像素統計使用新的 `--output`。工具拒絕覆寫。
實際每個GUI批次的 `START.json`、`TRACE.json`、`driver.py` 及來源收據保存完整命令與版本。

```sh
bash tools/gui.sh xrun python3 tools/scripts/windows_player_check.py references
bash tools/player_wine.sh python3 /tools/scripts/windows_player_check.py startup --revision repeat
bash tools/player_wine.sh python3 /tools/scripts/windows_player_check.py gui --revision load1b --target load --scale 1
bash tools/player_wine.sh python3 /tools/scripts/windows_player_check.py gui --revision load15b --target load --scale 1.5
bash tools/player_wine.sh python3 /tools/scripts/windows_player_check.py gui --revision fleet1 --target fleet --scale 1
bash tools/gui.sh xrun python3 tools/scripts/windows_player_check.py scaled-references --revision reference15b
bash tools/player_wine.sh python3 /tools/scripts/windows_player_pixels.py
LOGH3_CPUS=2 bash tools/gui.sh xrun bash -c 'go test ./internal/... ./cmd/... -count=1 > workplace/round104-standard-gui.log 2>&1'
bash tools/gui.sh xrun python3 tools/scripts/windows_player_receipt.py --freeze
bash tools/gui.sh xrun python3 tools/scripts/windows_player_receipt.py --check
```

標準測試 `internal/game` 201.344秒、`cmd/gin3` 110.341秒。第103輪20469來源亦重新核對通過。
GUI輸出及工具擁有權為1000:1000；本輪沒有重建image。容器清理與commit／push記於WORKLOG。

## 尚未完成

#69保持OPEN。正式版號／dist-all、macOS工具、Windows正常曲庫播放與公開存讀抽測、
實機、人耳及循環仍未完成。1.5倍跨平台完整實體像素一致性未達成，這項差異未經使用者
裁決成正式容許差異。既有原版對拍與112／116規格狀態保持，不外推完整原版正常整回合同態。
