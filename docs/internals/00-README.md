# 原版內部構造參考

第58輪沿#60：[左鍵保持工作快照](../goal/2026-10-03-left-hold-round-58.md)、
[研究396](../re/396-left-button-repeat.md)、[089契約](../spec/089-left-button-repeat.md)與
[原始指令匯出](../../tools/scripts/left_hold_ida.py)、[固定原版保持探針](../../tools/left_hold_oracle.sh)、
[速度操作探針](../../tools/scripts/left_hold_prototype.html)、[補證凍結工具](../../tools/scripts/left_hold_receipt.py)與
[完整保持補證](data/left-button-repeat.json)。補證凍結於裁決前；現行契約見089。

第58輪已確認350／100ms清單速度，限定CONFORMED子契約與[正式左鍵保持](../../internal/game/lefthold.go)接線依089。
[保持回歸](../../internal/game/lefthold_test.go)核時間邊界、列選擇、上下界、模態及保持取消。
[正常首戰保持全圖](../../internal/game/lefthold_normal_test.go)沿主選單138／S3同盟到資料窗，保持到26及0，不遮罩。
[真實GUI保持](../../tools/left_hold_gui.sh)用正常鍵盤與滑鼠按住五秒到上界，驗一倍／兩倍與全圖。
[GUI就緒觀測](../../tools/scripts/left_hold_gui_ready.py)唯讀比對首戰輸入標籤，再送正常滑鼠操作。
[月初提督清單保持補驗](../../tools/monthly_admiral_hold_oracle.sh)從正常建造計劃取得獨立父狀態與雙跑。
[月初框線游標清理](../../tools/scripts/monthly_admiral_cursor.py)由雙跑證據剔除25個舊量測混入的游標像素，產生器同時接線。
[保持成果凍結工具](../../tools/scripts/left_hold_game_receipt.py)、[正式保持收據](data/left-button-repeat-game.json)與
[對拍048](../playtest/048-left-button-repeat.md)保存明示時間差異、正常GUI、全圖及保存範圍。
[新原版保持結果比較](../../internal/game/lefthold_oracle_test.go)以明示等價輸入映射核右下清單／一覽表／稅率／百分比全圖。

《銀河英雄傳說III SP》PC DOS 中文版（1994，亨和企業．微波軟體發行）的
結構化逆向資料。這裡放**目前為止查證出來的事實本身**；查證過程與逐次
調查紀錄在 `docs/re/`，執行器規格在 dosgolem 的 `docs/spec/`。

這個專案除了 remake 之外，還有一個目的是保存這款中文版的技術細節——
1994 年台灣中文化作品的封裝方式、文字編碼取捨、資料格式與製作團隊，
在原始碼與開發資料都不存在的情況下，只剩執行檔本身可以問。所以逆向出來
的東西一律整理成可讀的 markdown，不停留在對話紀錄或一次性的輸出檔裡。

- [索敵顯示最小契約研究381](../re/381-battle-spotting-display.md)
- [索敵顯示完整原版資料](data/battle-spotting-display.json)
- [索敵顯示分級定位與固定61案計畫](data/battle-spotting-display-semantics.json)
- [索敵顯示收據工具](../../tools/scripts/battle_spotting_display_evidence.py)


第57輪入口：[資料視窗捲動工作快照](../goal/2026-10-03-battle-data-scroll-round-57.md)。沿#67，以自然S3同盟局面補超過十四列的完整畫面收據。

[資料視窗依賴原指令匯出](../../tools/scripts/battle_data_scroll_ida.py)保留原始定位及bytes，未審查項保持unknown。
[資料視窗四十列原版重播](../../tools/battle_data_scroll_oracle.sh)固定正常同盟路徑與seed，不注入局面。
[四十列及摘要／視野框研究395](../re/395-battle-data-scroll.md)保存本輪呈現根因與原始定位。
[正常入口全畫面回歸](../../internal/game/battledata_wholeframe_test.go)與[正式GUI鍵鼠驗證](../../tools/battle_data_scroll_gui.sh)覆蓋三個偏移及兩種倍率。
[收據產生器](../../tools/scripts/battle_data_scroll_receipt.py)、[完整收據](data/battle-data-scroll.json)與[對拍047](../playtest/047-battle-data-scroll.md)保存原圖、差分、逐區統計及驗證限制。

第56輪入口：[工作快照](../goal/2026-10-03-battle-object-keys-round-56.md)、
[研究394](../re/394-battle-object-keys.md)、
[IDA匯出](../../tools/scripts/battle_object_keys_ida.py)、[原版鍵反例](../../tools/battle_object_keys_oracle.sh)、
[收據產生器](../../tools/scripts/battle_object_keys_receipt.py)與[結構化證據](data/battle-object-keys.json)。限定顯示物件指標與排序鍵，不改完整入口門檻。


## 這裡的東西怎麼寫

- [第五十輪移動後戰術座標](../goal/2026-10-03-fleet-arrival-placement-round-50.md)
- [移動後座標原版擷取器](../../tools/fleet_arrival_oracle.sh)
- [移動後配置IDA定位工具](../../tools/scripts/fleet_arrival_ida.py)
- [移動後原池配置規則](../../internal/rules/fleetarrival.go)
- [移動過場後的正式配置接線](../../internal/game/fleetarrival.go)
- [正常S3配置與保存驗證](../../internal/game/fleetarrival_test.go)

- [第四十四輪正常玩家戰場入口配對](../goal/2026-10-03-battle-normal-entry-round-44.md)
- [正常原版冷啟動擷取器](../../tools/battle_normal_entry_oracle.sh)
- [正式資源正常玩家探測](../../tools/scripts/battle_normal_probe.go)
- [僅固定種子的測試期Go覆蓋](../../tools/scripts/battle_normal_probe_seed.py)
- [正常入口完整資料](data/battle-normal-entry.json)
- [正常入口雙跑與差異核對](../../tools/scripts/battle_normal_entry_receipt.py)
- [正常首戰入口研究382](../re/382-normal-battle-player-entry.md)
- [正常首戰入口草稿088](../spec/088-normal-battle-player-entry.md)
- [第四十五輪正常AI一次性閘門](../goal/2026-10-03-normal-ai-gate-round-45.md)

第三階段可見性／快取與正式續接：[第四十一輪](../goal/2026-10-03-battle-spotting-integration-round-41.md)，工作項目 #79。
- [第四十二輪正式索敵接線](../goal/2026-10-03-battle-spotting-integration-round-42.md)
- [第四十三輪來源標頭與逐格移鏡](../goal/2026-10-03-battle-discovery-scroll-round-43.md)
- [來源標頭／逐格移鏡固定計畫](data/battle-discovery-scroll-plan.json)
- [來源標頭／逐格移鏡完整原版及正式Game收據](data/battle-discovery-scroll.json)
- [原版擷取器](../../tools/battle_discovery_scroll_oracle.sh)
- [計畫凍結與雙跑／全畫面核對工具](../../tools/scripts/battle_discovery_scroll.py)

第三階段發現報告與正式續接：[第四十輪](../goal/2026-10-02-battle-discovery-round-40.md)，工作項目 #79。

第三階段索敵與戰鬥前續接：[第三十九輪](../goal/2026-10-02-battle-spotting-round-39.md)，工作項目 #79。
固定原版雙跑入口：[battle_spotting_oracle.sh](../../tools/battle_spotting_oracle.sh)，
邊界與父狀態身份：[battle_spotting_plan.py](../../tools/scripts/battle_spotting_plan.py)。
完整前後紀錄與分級索引：[battle-spotting-phase.json](data/battle-spotting-phase.json)，
重生工具：[battle_spotting_evidence.py](../../tools/scripts/battle_spotting_evidence.py)。
有限狀態探測：[battle_spotting_probe.go](../../tools/scripts/battle_spotting_probe.go)，不接正式玩家路徑。
第三階段規格：[087](../spec/087-battle-spotting-phase.md)，研究：[379](../re/379-battle-spotting-phase.md)。
原版重現範圍：[046](../playtest/046-battle-spotting-phase.md)。
六案現在位置／目的地分離與半徑邊界：[battle-spotting-coordinates.json](data/battle-spotting-coordinates.json)。

第二戰場殘留列來源：[第三十八輪](../goal/2026-10-02-battle-residue-source-round-38.md)，工作項目 #79。
小地圖bitmap契約：[086](../spec/086-battle-minimap-bitmap.md)，[研究378](../re/378-minimap-bitmap-bottom.md)。
分級來源及45份核對：[battle-minimap-bitmap.json](data/battle-minimap-bitmap.json)，
重生入口：[battle_minimap_evidence.py](../../tools/scripts/battle_minimap_evidence.py)。

第二戰場剩餘呈現來源：[第三十七輪](../goal/2026-10-02-battle-render-order-round-37.md)，工作項目 #79。
分隊繪製次序契約：[085](../spec/085-battle-unit-draw-order.md)。
調查紀錄：[研究377](../re/377-battle-render-order.md)。
分隊鍵分級索引：[battle-draw-order.json](data/battle-draw-order.json)，重生入口[原版核對工具](../../tools/scripts/battle_order_evidence.py)。

戰場傷害與方位規則接續範圍：[第二十四輪](../goal/2026-10-02-battle-rules-round-24.md)，工作項目 #53。

三項戰場規則勘誤：[規格 078](../spec/078-battle-rule-corrections.md)、
[研究 366](../re/366-battle-rule-corrections.md)、[對拍 038](../playtest/038-battle-rule-corrections.md)。
解出的事實：[欄位摘要](data/battle-rule-corrections.json)、
[IDA 自動合併分級索引](data/battle-rule-semantics.json)。

八步戰鬥的補給／修理階段接續範圍：[第二十五輪](../goal/2026-10-02-battle-preparation-round-25.md)，工作項目 #79。
本輪原版事實：[65 — 自動補給與修理](65-battle-preparation.md)、
[研究367](../re/367-battle-preparation.md)、[規格079](../spec/079-battle-preparation.md)、
[驗證039](../playtest/039-battle-preparation.md)。
欄位與分級索引：[battle-preparation.json](data/battle-preparation.json)、
[battle-preparation-semantics.json](data/battle-preparation-semantics.json)。
跨艦隊排序接續：[第二十六輪](../goal/2026-10-02-battle-preparation-round-26.md)、
[研究368](../re/368-battle-preparation-sort.md)。
原版排序欄位：[battle-preparation-sort.json](data/battle-preparation-sort.json)，
審查索引：[battle-preparation-sort-semantics.json](data/battle-preparation-sort-semantics.json)；
重驗入口：`tools/scripts/battle_preparation_sort_receipt.py`，方法與限制見研究368。
驗證範圍：[040 — 自動候選排序](../playtest/040-battle-preparation-sort.md)。
來源生命週期接續：[第二十七輪](../goal/2026-10-02-battle-preparation-round-27.md)、
[研究369](../re/369-battle-preparation-lifecycle.md)。
來源及結果驗證：[041](../playtest/041-battle-preparation-lifecycle.md)，
解出欄位：[battle-preparation-lifecycle.json](data/battle-preparation-lifecycle.json)，
審查索引：[battle-preparation-lifecycle-semantics.json](data/battle-preparation-lifecycle-semantics.json)。
重播／匯出工具的入口與限制見研究369，完整第一階段仍由規格079管理。
來源重設及正式接線接續：[第二十八輪](../goal/2026-10-02-battle-preparation-round-28.md)、
[研究370](../re/370-battle-preparation-reset.md)。
修理目標身分的[分級索引](data/battle-preparation-target-semantics.json)由同一IDA匯出工具合併。
正式第一階段的資源、面板、等待、正常入口與存讀驗收見[對拍042](../playtest/042-battle-preparation-first-phase.md)。
第二階段移動接續：[第二十九輪](../goal/2026-10-02-battle-movement-round-29.md)、
[研究371](../re/371-battle-movement-phase.md)、[草稿080](../spec/080-battle-movement-phase.md)、
[分級索引與局面摘要](data/battle-movement-phase.json)；原版重播入口與限制由研究371管理。
單步／避障接續：[第三十輪](../goal/2026-10-02-battle-movement-round-30.md)、
[研究372](../re/372-battle-step-and-collision.md)。
艦隊單步與碰撞的限定正式契約：[規格081](../spec/081-battle-fleet-step.md)、
[分級原版102案](data/battle-step.json)；重生入口為`tools/scripts/battle_step_receipts.py`。
限定同狀態資料與正常保存鏈驗收：[驗證043](../playtest/043-battle-fleet-step.md)。

音色庫來源與完整授權索引：[`player-music-library.json`](data/player-music-library.json)，
用途與原版素材邊界見 [`17-mds.md`](17-mds.md)。
原生 GM 移植與真實轉檔驗收索引：[`native-gm-import.json`](data/native-gm-import.json)，
方法及限制見[驗證 030](../playtest/030-native-gm-import.md)。
原生 FM 的固定版本／曲目／封裝索引：[`fm-import-lock.json`](data/fm-import-lock.json)，
原版觸發與重生限制見[規格 071](../spec/071-native-fm-import.md)、[研究 359](../re/359-native-fm-import.md)。
完整原生曲庫匯入與三路徑重跑收據：[`native-fm-import.json`](data/native-fm-import.json)，
正式解碼、重生命令與平台限制見[驗證 031](../playtest/031-native-fm-import.md)。
Linux 隨包工具、共享依賴與授權索引：[`linux-music-tools.json`](data/linux-music-tools.json)，
原生啟動、乾淨容器抽驗與交付限制見[驗證 032](../playtest/032-linux-music-tools.md)。
Linux 工具固定來源集合：[`linux-music-source-lock.json`](data/linux-music-source-lock.json)，
完整取得／離線重驗與本機封裝收據：[`linux-music-sources.json`](data/linux-music-sources.json)，
命令與正式交付限制見[驗證 033](../playtest/033-linux-music-sources.md)。
Windows FM 反例與完整 GM 探針收據：[`windows-music-probes.json`](data/windows-music-probes.json)，
重生命令及限制見[研究 362](../re/362-windows-music-tools.md)；
跨平台 FM 無損雜湊契約見[規格 074](../spec/074-cross-platform-fm-pcm.md)。
其十一首正式原生匯入與三路重跑收據由 `tools/scripts/music_windows_fm_receipt.py`
產生 [`windows-fm-pcm.json`](data/windows-fm-pcm.json)，未完成批次不得產生通過收據。
Windows 固定建置來源鎖：[`windows-music-source-lock.json`](data/windows-music-source-lock.json)，
十四組來源與兩組額外來源完整核對，方法與正式交付限制見[驗證 035](../playtest/035-windows-gm-sources-unicode.md)。
Windows GM 的 Unicode 命令列與載入失敗拒絕契約：[規格 075](../spec/075-windows-gm-unicode.md)。
Windows FM／GM 同包工具與預設完整入口：[規格 076](../spec/076-windows-combined-music-tools.md)，
候選依賴與驗證過程見[研究 364](../re/364-windows-combined-music-tools.md)。
同包二十二首、Unicode OPL 與中文 ZIP 搬移收據：[`windows-combined-music-tools.json`](data/windows-combined-music-tools.json)，
由 `tools/scripts/music_windows_bundle_receipt.py` 重驗產生；完整命令、反例與限制見
[驗證 036](../playtest/036-windows-combined-music-tools.md)。
音樂工具的本作品授權與可重建來源契約：[規格 077](../spec/077-music-source-delivery.md)，
本輪範圍見[第二十三輪](../goal/2026-10-02-music-source-delivery-round-23.md)。
來源重建正式 FM 抽驗入口：`tools/scripts/music_source_import_check.py`，方法見[驗證 037](../playtest/037-music-source-delivery.md)。
完整來源、授權、重建程式與首批反例的唯一收據：[`music-source-delivery.json`](data/music-source-delivery.json)，
由 `tools/scripts/music_source_receipt.py` 重驗產生；命令與限制見
[驗證 037](../playtest/037-music-source-delivery.md)，調查過程見[研究 365](../re/365-music-source-delivery.md)。
其固定來源工具與原生曲庫驗收收據：[`windows-gm-source-tools.json`](data/windows-gm-source-tools.json)，
由 `tools/scripts/music_windows_gm_source_receipt.py` 重驗產生；過程與勘誤見[研究 363](../re/363-windows-music-source-build.md)。
remake 增強策略的正式入口驗證索引：[`enhanced-ai-player-path.json`](data/enhanced-ai-player-path.json)，
範圍及原版差異見[規格 068](../spec/068-enhanced-ai-player-path.md)。

**每一項都要標推論等級**：`confirmed`（有實測或反組譯位元組證據）、
`強證據`、`假說`、`未知`。推測不得因為「看起來很合理」就升格，
升格要補證據並註明來源。

**什麼是穩定的、什麼不是**——這決定資料該怎麼記：

| 穩定 | 不穩定 |
|---|---|
| 結構欄位的**位移**（`rec[+0x0B]`）| **段值**（`6B93:0068`）|
| 資料段內的**變數位移**（`ds:94E0`）| 執行期的**線性位址** |
| **位元組簽章**（一段沒有絕對位址的機器碼）| 步數（`#25757331`）|
| 演算法、編碼、檔案格式 | 暫存器快照 |

段值取決於程式被載到哪裡，而那取決於 DOS 的記憶體佈局。實例：修好
dosgolem 的 EXEC 記憶體回收之後，`GIN3PS.EXE` 的 PSP 從 `0x3922` 降到
`0x0339`，所有模組跟著搬家，`docs/re/05` 之前記下的段值全部失效
（結構位移與演算法則完全不受影響）。

**所以定位程式碼一律用位元組簽章。** dosgolem 的 probe 可以一次驗一批：

```
tools/go.sh run ./cmd/probe -exe /orig/GIN3.COM -root /orig ... \
  -find "26 8A 47 0B 30 E4 26 8A 4F 0C 30 ED F7 E1, 83 3E E6 94 FF 75 1A"
```

## 產出格式

**每一項解出來的東西都要有機器可讀版本。** markdown 給人看，
`data/*.json` 給後續的工具與 remake 用；兩邊由同一份證據產生，不得只更新
一邊。素材（字型、圖庫、地圖）解出來之後匯出成 **PNG** 放
`workplace/assets/`——那是原版素材，**不進版控**，但要讓下一輪能直接拿來用。

挑簽章時避開含絕對位址的位元組（`9A xx xx ss ss` 的 far call 目標、
`A1 xx xx` 之類指到別段的存取），那些會隨載入位址改變。

## 工具

| 包裝器 | 用途 |
|---|---|
| `tools/py.sh` | Python 分析（`python:3.12-slim`）。原版素材唯讀掛在 `/orig`，倚天字型之類的外部參照用 `LOGH3_EXTRA_RO` 掛在 `/ref`。|
| `tools/ida.sh` | IDA Pro 9.4 headless。`analyze <exe>` 建 `.i64`，`script <exe> <py>` 跑 IDAPython。**exit code 不能當證據，只看輸出檔**；腳本一律寫檔，一支腳本一次處理多個目標（對同一個 `.i64` 連續開關會弄壞它）。|

`GIN3PS.EXE` 是六層封裝（GLUE／TINYPROG／PKLITE 系，見 `02`），IDA 只看得到最外層的解壓殼——本體要在執行期逆向
（dosgolem 的 `-peek`／`-find`／trace／檔案存取記錄）。`OPEN.EXE` 與
`ENDING.EXE` 沒有封裝，適合 IDA。

## 目錄

| 檔 | 內容 |
|---|---|
| [`01-assets-inventory.md`](01-assets-inventory.md) | 原版素材清單 |
| [`02-executables-and-credits.md`](02-executables-and-credits.md) | 執行檔、封裝器與製作團隊 |
| [`03-startup-and-memory.md`](03-startup-and-memory.md) | 啟動鏈、記憶體佈局與堆積配置器 |
| [`04-text-windows-and-big5.md`](04-text-windows-and-big5.md) | 文字視窗、儲存格格式與中文編碼 |
| [`05-palette-and-fade.md`](05-palette-and-fade.md) | 調色盤與淡入淡出 |
| [`06-font-graph-img.md`](06-font-graph-img.md) | 中文字型：`GRAPH.IMG` |
| [`07-window-frame.md`](07-window-frame.md) | 視窗外框 |
| [`08-admiral-dt.md`](08-admiral-dt.md) | ADMIRAL.DT[89AX] — 提督能力表 |
| [`09-face-amg.md`](09-face-amg.md) | FACE.AMG — 人物頭像（.AMG 圖庫格式） |
| [`10-sprite-spr.md`](10-sprite-spr.md) | SPRITE.SPR — 艦隊編號圖塊 |
| [`11-copy-protection-screen.md`](11-copy-protection-screen.md) | 防拷問答畫面 |
| [`12-strmap-bit.md`](12-strmap-bit.md) | STRMAP.BIT — 星域地圖底圖 |
| [`13-strmap-rgn.md`](13-strmap-rgn.md) | STRMAP.RGN — 星域區域遮罩 |
| [`14-screen-strings.md`](14-screen-strings.md) | 畫面文字表與熱區 |
| [`15-keireki-dat.md`](15-keireki-dat.md) | KEIREKI.DAT — 人物傳 |
| [`16-ending-exe.md`](16-ending-exe.md) | ENDING.EXE — 結局文字與職員表 |
| [`17-mds.md`](17-mds.md) | `.MDS` — 樂曲 |
| [`18-vo5.md`](18-vo5.md) | `.VO5` — 結局圖 |
| [`19-ibm16-350.md`](19-ibm16-350.md) | `IBM16.350` ＋ `IBMFLS16.350` — 開頭動畫 |
| [`20-comani-bmg.md`](20-comani-bmg.md) | `COMANI.BMG` — 戰鬥動畫封存檔 |
| [`21-3dblock-spr.md`](21-3dblock-spr.md) | `3DBLOCK.SPR` — 等距方塊 |
| [`22-stellar-dt.md`](22-stellar-dt.md) | `STELLAR.DT[89AX]` — 星域資料 |
| [`23-hotspot-lists.md`](23-hotspot-lists.md) | 熱區清單與按鈕標籤 |
| [`24-window-list-and-right-button.md`](24-window-list-and-right-button.md) | 視窗串列、命中測試與右鍵 |
| [`25-text-rendering.md`](25-text-rendering.md) | 文字繪製：粗體、陰影、底線與網點 |
| [`26-star-region-graph.md`](26-star-region-graph.md) | 星域之間的相鄰關係與路徑表 |
| [`27-treasury-globals.md`](27-treasury-globals.md) | 財政的全域變數 |
| [`28-screen-specs.md`](28-screen-specs.md) | 畫面規格 JSON：`data/screen-*.json` |
| [`29-planets-and-economy.md`](29-planets-and-economy.md) | `PLADAT.DAT` 與經濟力 |
| [`30-save-format.md`](30-save-format.md) | 存檔：`*.DT[89A]` 就是執行期的表 |
| [`31-text-plane.md`](31-text-plane.md) | 全螢幕字元平面（`ds:D81E`）|
| [`32-ship-build-costs.md`](32-ship-build-costs.md) | 船艦建造：單價（兩個陣營不同）、一單位幾艘與捨去規則 |
| [`data/`](data/) | 機器可讀版本（JSON），給後續工具與 remake 用 |

- [`34-battle-screen.md`](34-battle-screen.md)：戰術（戰鬥）畫面的版面、命令選單與怎麼走到。
- [`35-block3-blk.md`](35-block3-blk.md)：戰場的圖塊集（128 塊 16×20）。
- [`36-map-cmp.md`](36-map-cmp.md)：每個星域的戰場地圖（128×102 格，兩條規則的編碼）。
- [`37-ani-pnl-a98.md`](37-ani-pnl-a98.md)：戰鬥畫面的底圖（`.AMG` 編碼、畫在 y=40）。
- [`38-battle-units.md`](38-battle-units.md)：戰場上的單位是「分隊」，不是艦隊。
- [`39-battle-entry-points.md`](39-battle-entry-points.md)：戰場入口表。
- [`40-ship-kind-table.md`](40-ship-kind-table.md)：艦種屬性表。
- [`data/live-fleet-save.json`](data/live-fleet-save.json)：現況艦隊存讀、速度寫入的分級語意與原版／GUI 雜湊；入口為[規格 067](../spec/067-live-fleet-save.md)。
- [`41-sin-table.md`](41-sin-table.md)：原版的正弦表。
- [`42-screen-tile-table.md`](42-screen-tile-table.md)：畫面的圖塊表 `ds:D81E`。
- [`43-battle-terrain-and-rally.md`](43-battle-terrain-and-rally.md)：戰場地形的通行規則與集結點。
- [`44-battle-command-menus.md`](44-battle-command-menus.md)：戰場的命令選單（五張表、可選旗標、版面公式）。
- [`45-battle-unit-panel.md`](45-battle-unit-panel.md)：戰場右側的分隊面板（八列、值的來源、三個顏色規則）。
- [`46-battle-formation-chooser.md`](46-battle-formation-chooser.md)：戰場「編隊」的選格子畫面。
- [`47-window-frame-tiles.md`](47-window-frame-tiles.md)：視窗外框的圖塊（`FRAME.BLK`）。
- [`48-message-inventory.md`](48-message-inventory.md)：訊息框說了什麼。
- [`49-text-engine.md`](49-text-engine.md)：印字進視窗。
- [`50-fortress-gun.md`](50-fortress-gun.md)：要塞砲的射程、瞄準與開火。
- [`51-borland-long-helpers.md`](51-borland-long-helpers.md)：Borland 的 32 位元輔助函式（`3CC9` 那一段）。
- [`52-plot-primitives-and-canvas.md`](52-plot-primitives-and-canvas.md)：畫點的三支常式與那 40 列的畫布偏移。
- [`53-strategic-planet-assault.md`](53-strategic-planet-assault.md)：戰略回合裡的打行星（`20B4:1512` → `20B4:1662`）。
- [`54-strategic-ai-decision.md`](54-strategic-ai-decision.md)：戰略 AI 決定一支艦隊要不要動。
- [`56-fleet-composition-panel.md`](56-fleet-composition-panel.md)：艦隊組成畫面的工作區（編隊那一行的字母與格盤的圖塊）。
- [`57-phezzan-event.md`](57-phezzan-event.md)：費沙事件——第一次易主時觸發、經濟力流向對手首都、邊境叛離、人類座位受罰。
- [`58-strategic-fleet-window.md`](58-strategic-fleet-window.md)：戰略地圖上的艦隊視窗與命令（清單點列 → 作戰 → 目的地 → 確認）。
- [`59-list-screens.md`](59-list-screens.md)：一覽表四張（星系／艦隊／提督／敵將資料）：點列開什麼、游標反白、位移與分頁的變數。
- [`60-mouse-roles.md`](60-mouse-roles.md)：滑鼠在每張畫面的作用總表（左鍵、右鍵、按住、游標經過），逐格掛證據；原版有、remake 沒有的列在最後。
- [`61-turn-transition-windows.md`](61-turn-transition-windows.md)：回合處理中的畫面（「戰略模式處理中」、清單清空、地圖不看座位）與每支艦隊走一格的過場視窗（出發前往、被擋、經費不足、問話）、等待怎麼結束；報告框（拿下行星、士氣、已全軍覆沒）。
- [`62-strategic-attrition.md`](62-strategic-attrition.md)：戰略階段士氣低落停在別人星域的艦隊每回合扣兵、第 0 組 0 艘就全滅（報告、艦艇作廢、提督失蹤、星域與陣營艦隊數減一），沒全滅就移掉 0 艘的組。
- [`63-info-screen-tabs.md`](63-info-screen-tabs.md)：情報收集的三個分頁（治安／勢力／敵艦）與「確認」。
- [`64-ship-pool.md`](64-ship-pool.md)：艦艇池一筆 78 bytes 的欄位、配置與作廢（分隊就是它）。
- [完整交錯移動審查](../re/373-battle-interleaved-movement.md)：[第三十一輪](../goal/2026-10-02-battle-movement-round-31.md)依既有三邊界及102案核對完整第二階段。
- [交錯移動驗證044](../playtest/044-battle-interleaved-movement.md)：完整7艦隊／55分隊資料、正式Game轉接、正常命令中間幀及存讀；整階段畫面仍待對拍。
- [回合後距離／座標語意索引](data/post-turn-runtime-fields.json)：研究 354 的原始定位與證據分級，由 IDA 匯出自動合併；[全表收據摘要](data/post-turn-runtime-receipts.json) 的驗收入口是[對拍 027](../playtest/027-post-turn-runtime-fields.md)。
- [第三十二輪：索敵回收與移動畫面](../goal/2026-10-02-battle-movement-round-32.md)：接續正常回收、原池重用與產品畫面收據。
- [移動畫面收據工具](../../tools/scripts/battle_movement_render_receipt.py)：依371固定收據及正式Game.Frame記錄全畫面差分，不排除未對齊區域。
- [移動畫面與正常回收驗證045](../playtest/045-battle-movement-render.md)：隱藏圖示、點選及游標修正，原池再放出與原生差分。
- [第三十三輪：小地圖標記生命週期](../goal/2026-10-02-battle-marker-round-33.md)：接續正常初始化、更新及正式呈現。
- [小地圖標記原版探針](../../tools/scripts/battle_marker_probe.go)：固定原版RAM與seed，驗開關及座標快照；只限受控常式。
- [正常標記原版重播](../../tools/battle_marker_oracle.sh)：固定S3檢查點、輸入與seed，確認佈陣及顯示開關的玩家呼叫鏈。
- [標記生命週期研究374](../re/374-battle-marker-lifecycle.md)、[規格082](../spec/082-battle-minimap-markers.md)：正常佈陣、顯示開關與已畫快照的呈現契約。
- [標記收據產生器](../../tools/scripts/battle_marker_receipts.py)、[標記資料](data/battle-markers.json)及[語意索引](data/battle-marker-semantics.json)：96案、正常55筆與原始指令分級。
- [第三十四輪：階段面板與移動中間幀](../goal/2026-10-02-battle-phase-panel-round-34.md)、[階段面板規格083](../spec/083-battle-phase-panel.md)：接續#79固定第6回合逐圈畫面與階段文字。
- [逐圈原版重播](../../tools/battle_phase_oracle.sh)及[固定計數計畫](../../tools/scripts/battle_phase_plan.py)：唯讀snapshot、35個完整RAM／planar PNG／色號。
- [階段文字研究375](../re/375-battle-phase-panel.md)、[語意索引](data/battle-phase-semantics.json)、[逐圈全紀錄](data/battle-phase-frames.json)：呼叫順序、三字串與35邊界雙跑；由[證據工具](../../tools/scripts/battle_phase_evidence.py)重生。
- [階段面板正式畫面收據](../../tools/scripts/battle_phase_render_receipt.py)：35邊界全畫面及逐區有效RGB差分，不裁切或遮罩。
- [第三十五輪：第二戰場及正常資源結果](../goal/2026-10-02-battle-second-sample-round-35.md)：接續#79現行READY範圍，抽正式載入路徑及另一戰場。
- [隨附存檔正常玩家收據](../../tools/scripts/battle_samples_receipt.py)：三份存檔進場、實行、面板還原與完整存讀；[驗證045](../playtest/045-battle-movement-render.md)限定為remake正常路徑，不宣稱原版第二戰場對拍。
- [第三十六輪：原版第二戰場同狀態](../goal/2026-10-02-battle-second-oracle-round-36.md)：接續#79原版S1正常過場與固定戰場邊界，不重做三份remake玩家抽樣。
- [S1第二戰場原版重播](../../tools/battle_second_oracle.sh)與[固定計畫／匯出](../../tools/scripts/battle_second_plan.py)：43個邊界、18艦隊／135分隊，入口seed與完整旗標固定。
- [S1第二戰場正式Game驗證](../../internal/game/battlesecond_test.go)與[逐圈紀錄](data/battle-second-frames.json)：唯一入口注入，後續全部使用正式移動引擎。
- [S1第二戰場畫面差分](../../tools/scripts/battle_second_render_receipt.py)：原生全畫面與逐區色彩統計，保留修正前後差分，不遮罩。
- [第二戰場研究376](../re/376-second-battle-render.md)與[方向圖示規格084](../spec/084-battle-fleet-directions.md)：方向位置、條件及圖塊來源；同順位重疊與203列差分另待補證。
- [方向圖示分級索引](data/battle-directions.json)由[證據工具](../../tools/scripts/battle_second_render_evidence.py)核原始指令、八向位移及素材bytes。

第40輪索敵呈現收據入口：[固定七分支計畫](data/battle-discovery-plan.json)、[原版收據包裝器](../../tools/battle_discovery_oracle.sh)、[計畫旗標工具](../../tools/scripts/battle_discovery_plan.py)。
第40輪呈現事實：[分級指令索引](data/battle-discovery-semantics.json)、[完整原版證據](data/battle-discovery-evidence.json)、[證據重生工具](../../tools/scripts/battle_discovery_evidence.py)。
第40輪研究入口：[380發現標頭／暫時標亮](../re/380-battle-discovery-presentation.md)，目前事實同步於[戰場畫面](34-battle-screen.md)及[戰場分隊](38-battle-units.md)。

第42輪正式索敵：[Game](../../internal/game/battlespotting.go)、[087](../spec/087-battle-spotting-phase.md)。

第42輪可比UI邊界收據：[battle-spotting-game.json](data/battle-spotting-game.json)，由[索敵收據工具phase-render](../../tools/scripts/battle_spotting_display_evidence.py)重生。
- [第45輪AI閘門證據](../re/383-normal-ai-once-gate.md)：[完整收據](data/normal-ai-once-gate.json)，原版雙跑由[工具](../../tools/normal_ai_gate_oracle.sh)重生，[IDA匯出](../../tools/scripts/normal_ai_gate_ida.py)與[收據產生器](../../tools/scripts/normal_ai_gate_receipt.py)保存原始定位。
- [第46輪正常讀檔座位／日期](../goal/2026-10-03-load-seat-date-round-46.md)，接續#89與088未完成範圍。
- [讀檔陣營研究384](../re/384-load-faction-selection.md)、[完整座位收據](data/load-seat-selection.json)：[原版雙跑](../../tools/load_seat_oracle.sh)、[IDA匯出](../../tools/scripts/load_seat_ida.py)、[可丟棄選單](../../tools/scripts/load_seat_probe.go)與[收據重生](../../tools/scripts/load_seat_receipt.py)。正式入口見[Game讀檔選單](../../internal/game/loadfaction.go)及[正常玩家探針](../../tools/scripts/load_seat_player.go)，驗收範圍由088限定。
- [讀檔陣營正常回歸](../../internal/game/loadfaction_test.go)與[真實X11輸入抽驗](../../tools/load_seat_gui.sh)：依088驗正常主選單、三選擇、右鍵反白與放大命中。
- [第47輪日期進位](../goal/2026-10-03-turn-counter-round-47.md)：接續#89，核對0374:03D9在所有戰場與模態尾段之後進位；watch03DD為執行後IP，前輪座位子契約保持。
- [回合計數器原版重播](../../tools/turn_counter_oracle.sh)及[IDA原始指令匯出](../../tools/scripts/turn_counter_ida.py)：固定無戰鬥進位與正常S3戰場保持，輸入、工具版本與限制見第47輪。
- [回合計數器研究385](../re/385-turn-counter-boundary.md)：主迴圈原始定位、watch執行後IP勘誤及正常首戰日期修正依據。
- [正常回合與兩AI計數驗證](../../internal/game/turncounter_test.go)：正式玩家完整回合／WORK保存、兩種AI與兩陣營的建隊／命令時點；全戰場與模態尾段抽樣另由既有regionturn／victoryreport測試驗收。
- [回合修正真實X11輸入](../../tools/turn_counter_gui.sh)：正式cmd/gin3正常主選單至S3首戰，兩種視窗尺寸；不使用script或direct-entry。
- [計數器收據工具](../../tools/scripts/turn_counter_receipt.py)及[完整收據](data/turn-counter-boundary.json)：原版雙跑／IDA／19份正式資料／全圖差分、GUI與來源SHA，研究385及088限定驗收。
- [第48輪活動座位](../goal/2026-10-03-active-seat-round-48.md)：接續#89，查正常讀檔、實行與雙人交接的WORK控制生命週期。
- [活動座位原版重播](../../tools/active_seat_oracle.sh)：固定正常無戰鬥回合、同盟S3處理與雙人首次交接，全部RAM與watch保留，語意未證實欄位不補值。
- [活動座位IDA匯出](../../tools/scripts/active_seat_ida.py)：04F8四個座位函式的原始名稱、範圍、bytes與附加分級，不將自動名稱當證據。
- [活動座位研究386](../re/386-active-seat-lifecycle.md)與[事實27](27-treasury-globals.md)：兩層座位迴圈、思考／處理／命令的原始控制值與同盟435M實際停點，088限定驗收。
- [正式活動控制更新](../../internal/game/activeseat.go)：依088在載入／交接、思考／AI、處理與首戰更新WORK+09／+0A，不改固定玩家視角。
- [活動控制正常驗證](../../internal/game/activeseat_test.go)：原版捕獲值、正常回合／雙人交接／正式保存、兩AI實際命令與缺WORK／非法上下文邊界。

- [活動座位正常GUI](../../tools/active_seat_gui.sh)：正式三選擇、雙人同盟交接與兩倍視窗，使用產品scale及PID選視窗。
- [活動座位收據生成](../../tools/scripts/active_seat_receipt.py)與[結構化收據](data/active-seat.json)：原版控制邊界、正式保存與完整全圖比較。
- [第49輪月底座位閘門](../goal/2026-10-03-month-seat-gate-round-49.md)與[原版重播](../../tools/month_seat_gate_oracle.sh)：接續#89，隔離S3載入的WORK+E0／+E1旗標與錯誤重複結算。
- [月底閘門IDA匯出](../../tools/scripts/month_seat_gate_ida.py)：0654:0034的原始範圍、兩條閘門及共同清旗標尾巴。
- [月底閘門完整收據](data/month-seat-gate.json)由[驗證工具](../../tools/scripts/month_seat_gate_receipt.py)重生，保存原版自然與受控雙跑、IDA原始bytes及未解差異。
- [月底旗標研究387](../re/387-month-seat-gate.md)：重複結算根因、原版受控變因及088子契約的證據審查。
- [正式月底旗標](../../internal/game/monthseat.go)：單人raw WORK閘門、座位清除與回合進位重設，雙人仍依既有016範圍。
- [月底旗標正常驗證](../../internal/game/monthseat_test.go)：原版正常旗標、兩AI／兩陣營的正式存讀、旗標非零與跨月一次結算。
- [到達配置研究388](../re/388-fleet-arrival-placement.md)、[完整收據](data/fleet-arrival-placement.json)及[產生器](../../tools/scripts/fleet_arrival_receipt.py)：九艦隊／七十筆池的到達配置、完整表差分與正常首回合無報告的原版雙跑。
- [第51輪正常首戰輸入顯示](../goal/2026-10-03-battle-input-display-round-51.md)：接續#89，核對游標／右側摘要與選取、sprite快取，完整088仍DRAFT。
- [正常首戰純移動原版重播](../../tools/battle_input_display_oracle.sh)：固定435M輸入邊界，分隊、空白及面板位置的單次純移動與完整RAM／畫面。
- [純移動摘要研究389](../re/389-battle-input-display.md)與[IDA匯出](../../tools/scripts/battle_input_display_ida.py)：摘要快取、退出與兩個SPRITE 10物件的原始定位；088限定正常Input範圍。
- [純移動完整收據產生器](../../tools/scripts/battle_input_display_receipt.py)：保存五組獨立原版雙跑、正常入口全圖及未解像素，不改第50輪收據。
- [純移動結構化收據](data/battle-input-display.json)：五個IDA函式535條原指令與22項分級定位、55份正常入口雙跑及兩尺寸X11；全圖2／1／1／0／0 pixels差保留。
- [第52輪正常首戰摘要標記](../goal/2026-10-03-battle-hover-marks-round-52.md)：沿#89追剩餘小地圖與位置反白像素，既有第51輪收據保持。
- [摘要標記IDA探針](../../tools/scripts/battle_hover_marks_ida.py)：由固定原版RAM切段保存原始指令，未審語意保持unknown。
- [摘要小地圖原版重播](../../tools/battle_hover_marks_oracle.sh)：空白Input移回分隊及明示集結點注入，保存新RAM／畫面與完整旗標。
- [摘要小地圖研究390](../re/390-battle-hover-marks.md)：旗艦航線、一般分隊集結點與剩餘方向物件覆蓋的分級證據。
- [摘要小地圖收據產生器](../../tools/scripts/battle_hover_marks_receipt.py)與[結構化收據](data/battle-hover-marks.json)：八組原版雙跑、正常入口五個完整畫面與剩餘差異，原JSON保持。
- [第53輪隱藏偵察隊快取](../goal/2026-10-03-battle-scout-cache-round-53.md)：沿#89核對完整戰場池的圖示刷新與存讀，不重開既有畫面收據。
- [隱藏偵察隊研究391](../re/391-battle-scout-cache.md)：既有IDA資料庫的209條原始bytes與原版七筆快取；規格088限定接線。
- [偵察隊快取收據產生器](../../tools/scripts/battle_scout_cache_receipt.py)與[結構化收據](data/battle-scout-cache.json)：核對固定原版雙跑、正常Game雙跑、七個byte與未改畫面，保留原始位址與分級。
- [第54輪戰場圖示順位](../goal/2026-10-03-battle-sprite-priority-round-54.md)與[IDA探針](../../tools/scripts/battle_sprite_priority_ida.py)：沿#89核+30初始化及顯示消費端，未審查語意保持unknown。
- [圖示順位原版重播](../../tools/battle_sprite_priority_oracle.sh)：固定正常S3狀態、輸入及seed，追完整分隊與偵察池的建立／刷新，輸出拒絕覆寫。
- [圖示順位研究392](../re/392-battle-sprite-priority.md)：55池初始化、原指令及顯示消費端，分開正常收據與未重畫的診斷探針。
- [圖示順位收據產生器](../../tools/scripts/battle_sprite_priority_receipt.py)與[結構化收據](data/battle-sprite-priority.json)：原版55筆物件、完整指令索引、正常資料／保存／畫面及比較器回歸，前輪JSON保持。
- [第55輪首戰WORK控制與經費保存](../goal/2026-10-03-battle-work-control-round-55.md)：沿#89追剩餘控制與經費的原版寫入生命週期，既有收據不重生。
- [首戰WORK原版重播](../../tools/battle_work_control_oracle.sh)：固定S3父狀態與seed，監看控制、兩座位鏡頭與兩方移動經費；完整旗標與原始輸出分批保存。
- [首戰WORK的IDA匯出](../../tools/scripts/battle_work_control_ida.py)：原始定位、指令與附加語意分欄，未審查的掃描範圍保持unknown。
- [首戰WORK研究393](../re/393-battle-work-control.md)：控制與經費的寫入生命週期、座位鏡頭及未驗續戰的限制。
- [正式戰場WORK同步](../../internal/game/battlework.go)：088限定控制、鏡頭及經費更新；原版存檔格式保持。
- [WORK邊界回歸](../../internal/game/battlework_test.go)：兩座位鏡頭、Input保存邊界、狀態2／5及signed經費；完整正常WORK對拍沿既有fleetarrival_test。
- [WORK完整收據產生器](../../tools/scripts/battle_work_control_receipt.py)與[結構化收據](data/battle-work-control.json)：原版雙跑／寫入監看、785條IDA原始指令、完整230-byte保存／重讀及未變畫面；前輪JSON保持。
