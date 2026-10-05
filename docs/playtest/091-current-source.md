# 091：目前玩家程式及音樂工具的來源重建

2026-10-05，第103輪，GitHub #69。沿[116](../spec/116-current-source.md)及[077](../spec/077-music-source-delivery.md)。
logh3父提交9e1d31830f3d5555cfc313e06f83b0d159186ff1，dosgolem固定f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c。
完整身份由[current-source.json](../internals/data/current-source.json)逐項保存。

| 核對 | 結果 |
|---|---|
| 來源包 | d／e各2810項，全tar相同；固定Git來源、完整vendor、Go及第三方授權、真實dosgolem Git物件 |
| 來源版本 | 目前logh3的cmd／internal／tools、模組及FM鎖；當輪兩個封裝／重建工具附件另列，不偽稱已在父提交 |
| 離線依賴 | 正式cmd／internal獨立模組產生vendor；來源與目前程式另建vendor全同，空GOMODCACHE及GOPROXY=off |
| 完整二進位 | Linux／Windows玩家程式、兩個music-import、Linux music-tool及兩個dosgolem-probe共7種符合對照 |
| dosgolem身份 | 單一固定提交、真正根目錄、fsck及clean；兩平台buildinfo為f7523f9／modified=false |
| 原生CLI | 6項通過、零跳過，來源重建probe實際通過身份檢查 |
| 正常標題 | 421格含分段停留，標題起點75573918781 ns；中點4998次更新，無腳本或狀態注入 |
| 標題圖 | 來源與目前程式307200像素及PNG全同，已目視完整標題，不以最初黑幀冒稱整段完成 |
| 正式FM入口 | 中文／空白路徑，無系統ffmpeg、PATH=/none；來源重建匯入器正常生成BGM_0F |
| 音訊閘門 | 完整VGM／WAV／增益後PCM符合固定鎖，Linux OGG全檔亦相同；不改跨平台OGG裁決 |
| Windows入口 | 來源玩家程式在Wine以-h成功，原版／音樂／存檔參數存在；未稱Windows GUI通過 |
| 邊界 | 既有來源包拒絕覆寫且SHA不變；錯誤HEAD沒有建立輸出 |
| 標準測試 | 18套件13通過／5無測試，Game199.225s、gin3 110.727s |

最終本機來源包：
`workplace/source69-r103-d/source.tar`及`workplace/source69-r103-e/source.tar`。
2810項，35,461,120 bytes；SHA-256：
`89be259435d35d22f3db7e27ffe2658bdafb67183abd592c4bc8b13e2af5f954`。
大小由最終BUNDLE重核；暫存b／c另有不同附件及manifest欄位，不能換名當最終包。

## 技術限制

這是來源重建與啟動抽驗，沒有正式版號、dist-all交付或公開發行。
來源包不含原版、解碼素材、生成曲庫或所有合成器／編碼器來源；GUI原版及畫面規格為分開的本機輸入。
Linux／Windows工具及共享程式庫來源仍沿033／035／036；macOS、實機、人耳及循環接點仍待完成。
Windows玩家程式只驗編譯及Wine命令入口，不能由此宣稱完整Windows遊戲流程。
標題只比目前程式，不新增原版整回合同態。116限定CONFORMED，#69 OPEN。

## 重生

既有image及Docker包裝器，UID/GID1000、無網路，有界Xvfb及外層逾時。
輸出已存在就拒絕覆寫。d／e為已完成批次，重跑須換新的英數revision。

```sh
bash tools/gui.sh xrun python3 tools/scripts/remake_source_kit.py \
  --head 9e1d31830f3d5555cfc313e06f83b0d159186ff1 --revision d
bash tools/gui.sh xrun python3 tools/scripts/remake_source_kit.py \
  --head 9e1d31830f3d5555cfc313e06f83b0d159186ff1 --revision e
bash tools/gui.sh xrun python3 tools/scripts/remake_source_rebuild.py --revision d
bash tools/gui.sh xrun python3 tools/scripts/remake_source_runtime.py --revision b --mode title --attempt second
bash tools/py.sh /work/tools/scripts/remake_source_runtime.py --revision b --mode fm
bash tools/gui.sh xrun python3 tools/scripts/remake_source_receipt.py --check
```

標題及FM先由b的重建程式執行。最終d全部7個二進位SHA與b相同，因此執行證據綁同一程式。
實際完整命令、絕對路徑、來源Git archive、逐檔manifest、工具及產物SHA均在本輪索引。
來源抽驗讀[037](037-music-source-delivery.md)已凍結來源包及其對應工具，不搬運主機runtime。
單獨來源包的手動Go建置命令在包內README；自附重建器的對照驗收須使用本專案workplace批次配置。

## 失敗及訂正

a在儲存庫根目錄執行vendor，掃到研究tools的dosgolem/internal及workplace的oracle，缺模組而拒絕。
後續重建器缺MANIFEST即退出。沒有建置產品，也沒有改go.mod或引入原版探針為玩家依賴。
改在正式cmd／internal獨立模組產生vendor後，b／c成功；最終d／e另保留077的全部既有必填manifest欄位。

最初60秒截圖仍為動畫372格，判準錯把374個來源格當完整播放時程。
依正式421格、各格停頓與112末端等待算出4998次更新，保留舊圖後重跑，兩邊自然標題通過。
這是驗證計畫錯誤，未改產品或增加原版逆向深度。硬體時間仍屬既有規格近似。
