# 062：完整艦隊交戰單幀對拍

2026-10-04，沿#48／#79，父提交3b4c54c5219c93f2779ad141e965351c45556f7d。
remake含本輪未提交改動，精確提交由本輪Git紀錄定位。
[099 READY](../spec/099-fleet-frame.md)授權無時間單幀；完整056／092保持DRAFT。

## 比較等級與來源

`same-state`限定第70輪固定光線砲交戰的兩個06FC完整幀。
相同typed編制、目前座標、觀看控制、選取分隊80／102、圖段及傷害前後狀態。
原版完整RAM與畫面由dosgolem f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c產生；
完整命令與父state沿[059](059-fleet-blocks.md)。本輪只核對既有a／b來源，不宣稱新執行器雙跑。
原版輸入GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
bundle SHA-256沿AGENTS.md。原版seed、輸入、注入及兩側14分隊紀錄沿059／061。
本輪未增加原版輸入、注入或RND；未修改dosgolem。

原版與remake均原生640×480，logical色盤相同；未縮放、裁切或遮罩。
Go1.24.13。正式測試用logh3-ebiten:2，image SHA-256
`6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8`。
探針用golang:1.24-bookworm，image SHA-256
`1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`。
IDA用ida-pro-9.4-idapython:locked-v1，image SHA-256
`6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780`。
remake從空白畫布與玩家素材產生底圖、文字、兩量表、圖段及下方，沒有借原版像素。
原版期望畫面與17項來源雜湊保存於[fleet-frame.json](../internals/data/fleet-frame.json)，
SHA-256 `17d968aa06150de3f67e98f8ead12f8636c2624b78a08802ab1e8bd203a4a027`。
此處是合成器診斷入口，沒有正常Game／GUI路徑聲明；不把原版像素當remake輸入。

## 重現

既有原版收據重生沿`tools/fleet_blocks_oracle.sh`與059；固定RAM切段以IDA9.4匯出：

```sh
bash tools/gui.sh xrun python3 tools/scripts/fleet_frame_receipt.py
bash tools/go.sh gofmt -l internal/screen/fleet_frame.go internal/screen/fleet_frame_test.go internal/screen/fleet_presentation.go
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
```

凍結器確認既有來源與輸出完全相同，異內容拒絕覆寫。
全圖產生需新輸出目錄，在gui容器以`LOGH3_FLEET_FRAME_OUTPUT`指定，執行
`go test ./internal/screen -run 'TestFleetFrame|TestFleetPresentationNative' -count=1 -v`。
正式產物在`workplace/frame48-r73-remake/`，比較JSON、原版／remake／差分六張PNG齊全。
傷害後圖已目視。此輪沒有新增正式交付或公開素材。

三份IDA輸入由`workplace/oracle/projection48-r72-a/case-00-before-mem.bin`切取，
每份為256零byte加上段位址起的RAM，段1DD6長030F、3958長0041、348A長0500，長度均為十六進位。
輸入RAM SHA-256 `9dfbb2cbec5d93849e0f100511cad65b342b41f68198378cd9cdc103ebeb9279`。
使用`tools/ida.sh raw idat -A -B`及`engagement_opening_ida.py`；
範圍依序0152:030F、0005:0041、01DC:0228及0281:ret。
輸入命名為`workplace/ida/opening48-r73-<段>.com`，輸出同名.json，全部262條原指令核實際RAM。
不從exit code或IDA自動名稱判定原版語意。

## 結果

| 範圍 | 每幀像素 | 兩幀不同像素 |
|---|---:|---:|
| 全圖 | 307200 | 0 |
| 左／右圖視窗，各自 | 26112 | 0 |
| 左／右分隊文字，各自 | 9600 | 0 |
| 中間文字 | 3840 | 0 |
| 左／右量表，各自 | 726 | 0 |
| 下方 | 139776 | 0 |

兩幀最大RGB通道差均0、差異範圍空，無未核准像素差。
比較JSON SHA-256 `deede76d21a4df1ef706e630c657778976fbd125d61e0673e013276823ae6bf4`。
scene-1原版／remake PNG SHA-256
`aa033dfdaa1875fdcc3d248bb15736f7ab9362f24f21cf81e7ef9e1286480d5f`。
scene-2原版／remake PNG SHA-256
`296ecd6d7767bf42e898e34f9f479485df4ce3a7b0510b340ed48aecf2dce058`。
兩張差分PNG SHA-256 `b63c42743ea731077221620de1aa9508a3dc1b90c401f179ce6e49df12442769`。

標準18套件13PASS／5無測試，Game111.267秒、gin3 93.599秒；
`workplace/round73-standard.log` SHA-256 `62ceb9d2fd6c4463df7348da722e8f4131796013381f0a67bfc5a3cef7a60c89`。
完整兩圖、92個文字原版樣本、合法裁切域及錯誤來源回歸無skip通過。
補超編拒絕後targeted log SHA-256 `04b604860446866642881a0a519a162b3e9fa3b315bf6e7237d2a0b88a1553d0`。
既有等待／傷害／保存與其他套件回歸保持。

## 差異分類與限制

- 靜圖名稱A0不存在，屬來源產生器錯誤，草稿另存。修正為原指令D0後重跑。
- 傷害後266像素字色差，屬組合caller錯誤；結果須用目標陣營，不能用攻方武器色。
  失敗差分保留`frame48-r73-probe-v2/`，正確探針另存`-v3/`。
- 099只接受目前0..4000艘編成。157種可見Y、9種堆疊數與兩端X的全部操作檢查通過，
  原版095零高度未返回案例仍保留，外部修改存檔未外推。
- 原版動畫關閉時八tick包含繪圖CPU耗時；開場沒有額外固定八tick契約。
  保留Game計時差異，不以步數換成固定延遲，不深入硬體driver。
- Game尚未呼叫完整合成器及093核心；正常玩家、GUI、音效、時間線與保存仍由056／092驗收。
  本輪不關#48／#79，不升格完整規格。

遠端48／79逐字回讀且保持OPEN；89張worklist／9張OPEN，269條核對零失效。
核對log SHA-256 `462b0b88f9cf92c207bc5cba6b088764e2b247884dcb11d27fafdef253aaff7b`。
擁有權及容器清理記錄於本輪WORKLOG；工作根無root-owned或誤建.md目錄，輸出1000:1000。
本輪容器均--rm退出，logh3清單空；未清理其他專案資源。
