# 042 — 正式自動第一階段

日期：2026-10-02；工作項[#79](https://github.com/wicanr2/logh3/issues/79)。
契約：[規格079](../spec/079-battle-preparation.md)；證據與停止線：[研究370](../re/370-battle-preparation-reset.md)。
本檔記錄完整資源階段的限定驗收；完整八步仍由規格062管理。

## 重生入口

所有命令使用既有Docker包裝器，原版只讀，UID/GID 1000:1000。
原版重生模式／命令與完整注入見研究369／370與其JSON；不得改執行器或覆寫原收據。

```sh
bash tools/gui.sh test ./internal/game -run TestBattlePreparation -count=1 -v
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 bash tools/py.sh tools/scripts/battle_preparation_render_receipt.py "$(git rev-parse HEAD)" "$(git status --porcelain)"
bash tools/gui.sh test -count=1 ./internal/... ./cmd/...
```

## 已驗結果與範圍

**remake已完成**：完整自動第一階段已接正式戰術「實行」。先運補艦／行星補彈藥，
再修護艦／行星補艦、重算士氣並等待報告，最後回復防衛力才續後面的移動。
結果每幀自行續接，點擊不能重跑；正式修理面板補回目標艦隊與提督／部隊兩列。

**原版指定狀態已驗**：72案完整排序與16案資源邊界直接核對原版輸出；
三個第6回合固定局面各核55筆艦艇的艘數／滿編／修理上限／32位彈藥、
3行星防衛力／庫存、7艦隊士氣／旗標及兩來源池身份。兩側種子在執行前固定`0x12345678`，
首尾相同。未放出的索敵隊仍在來源清單，沒有用可見組數替代原版清單。
原版端六組面板／終點均已雙跑完整記憶體／畫面／色盤相同，來源為研究369的固定收據。

**正常入口及存讀已驗**：`TestBattlePreparationNormalPlayerExecute`走主選單、防拷、月初報告、
正常S3讀檔、艦隊命令與實行到達希利，再正式點選戰術實行推到下一回合。
正式`cmd/gin3`同一正常路徑已抓完整GUI並目視第2回合，沒有狀態注入。
三個資源診斷局面則明示注入，正式撤退／戰後回寫、關閉月初報告後存讀整份SHIP／PLANET相同。
額外防衛力降為0的士氣等待診斷證明報告關閉前不移動／回復，關閉後只續接一次；
此合成順序診斷不是新增原版同狀態局面。

## 同狀態面板與全畫面差分

原版`GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`，
原bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
dosgolem專用副本`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`、乾淨；Go image為`golang:1.24-bookworm`。
原版完整旗標／前後記憶體／輸入／游標及步數：`battle-preparation-lifecycle.json`，
SHA-256 `f8d4848a3e19232a66f92b18c2477647ea375a36379d2b5c792e3f5f48f1e650`。
原版地址為dosgolem實體線性，IDA資料獨立基底與分級見研究370。

remake擷取父提交`58e02df905cec946eb7be40f53bfe0ceed54a6cf`、本輪程式未提交；
`logh3-ebiten:2`／Go1.24.13。來源逐檔SHA-256與工作樹狀態在本機
`workplace/battle79-round28-render-receipt.json`，SHA-256
`e179dbfb04602767430001a93e56d1c2de500b1374dc01915b8c420581dd5001`。
正式`Game.Frame`與實際`Game.Palette`，沒有另外離線拼畫面。

兩側原生640×480、16色色號及有效色盤；不裁切、不縮放、不遮罩。
比較等級：**完整結果面板same-state；其餘畫面不列同狀態聲明**。
規格的8×8文字格區域由`SupplyPanelSpec`和`FrameCellOriginY`推導為`(480,240,128,160)`。

| 局面 | 面板比較像素 | 相同 | 不同／最大通道差／bbox | 全畫面不同 | 全畫面最大差／bbox |
|---|---:|---:|---|---:|---|
| 補給 | 20480 | 20480 | 0／0／無 | 56263 | 243／(0,0)..(639,479) |
| 修理 | 20480 | 20480 | 0／0／無 | 56247 | 243／(0,0)..(639,479) |

兩種面板有效色盤相同。面板外鏡頭、單位、日期／動態狀態及底板不列本輪同狀態範圍，
其差異保留unknown，不掩蓋也不把面板通過升格為全畫面通過。
雙方PNG與差分位於既有`workplace/`，完整雜湊在收據；主要檔案：

| 檔案 | SHA-256 |
|---|---|
| `battle79-round28-supply-original.png` | `8baa7329a665191465abb63c404b80b702eb0fd5c1fa885c22c76a883cc5062a` |
| `battle79-round28-supply-remake.png` | `3c6939cc6bb92357dba95951e9b82c33b763ab1ff85017803a93c4df9bb5138a` |
| `battle79-round28-supply-difference.png` | `3b8bb1fab15bbc2480356386519b8cf8aee9702a05bb8a40848a1918bf055a61` |
| `battle79-round28-repair-original.png` | `09dc24e460e6a483781d219ca6b347010b19a5dff094a2deeb0391cedb891132` |
| `battle79-round28-repair-remake.png` | `f771fdc19be919ab01e2a9204257e2953a7145cda81272c831e8d2880576ffe8` |
| `battle79-round28-repair-difference.png` | `828df70380288a34a27a6907cf0af591996d5b51e2f7bd70ff830a291f448577` |
| `battle79-round28-normal-gui.png` | `e6cc2d6ffca7dbf8acd76313cf6bb7b1222f3bab3aec43852651d60e155e9e5e` |

正常GUI的完整命令在收據`normal_gui.command`；主遊戲`main1e`及S3遭遇`fcmd-s3-found`
依`LOGH3_GUI_PREFIX=<批次> bash tools/gui_check.sh`單獨重跑皆逐位元組相同。
標準18個產品套件通過或無測試，日誌`workplace/battle79-round28-product-tests.log`的
SHA-256 `3da13a05f020a32fd342f7e9b7de30ff7fdeab0db94948a93304bce5619470da`。
其後新增首尾seed核對與完整文字格原點後，重跑Game全部新流程與88案規則皆通過，零跳過。

## 尚未涵蓋

原版oracle未知：所有存檔、第二戰場、完整戰場鏡頭／其他面板及八步其餘階段。
不把上述未知當成這個已驗第一階段必須再次深挖的完成門檻；#79仍開啟。
可選polish：資源結果幀的顯示時長未做原版CPU／wall-clock一致聲明。
本輪不新增AI、音樂人耳／macOS、跨平台打包或正式發行完成聲明。
