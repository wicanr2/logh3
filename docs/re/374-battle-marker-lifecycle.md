# 374：正常戰場小地圖標記生命週期

日期：2026-10-02。工作項：[#79](https://github.com/wicanr2/logh3/issues/79)。
父提交`be187b1bff7fa32f8831302886c2ec69362901a8`。
接續[研究154](154-minimap-marker-not-a-trail.md)、[研究373](373-battle-interleaved-movement.md)
及[第三十三輪](../goal/2026-10-02-battle-marker-round-33.md)。
標記子契約見[規格082](../spec/082-battle-minimap-markers.md)，完整移動仍依080。

## 來源與位址基準

- `GIN3.COM`：`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
- dosgolem專用副本：`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，乾淨，未改執行器。
- 常式輸入完整RAM：`9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d`。
- 正常S3檢查點`fcmd-s3-base.state`：`0155f63e373582b76a3b7ee54eccbef5ae2626481e09a68550b5fddc580146d7`。
- IDA Pro9.4經`tools/ida.sh`，一次性切段加256-byte前綴；原位移0對應IDA `0x10200`。
  下表為原版CS:IP；dosgolem線性位址及IDA位址在分級索引分列。
  原始名稱、bytes、運算元不改，以[分級索引](../internals/data/battle-marker-semantics.json)附語意。

## 已證實的資料流

| 原版定位 | 契約 | 證據等級 |
|---|---|---|
| `297B:0974..0984` | 一般分隊Action=7F、Detail=0、標記旗標=3F，低byte觀測旗標清0 | proven，IDA原始寫入指令 |
| `297B:09D2..09E8` | 未放出索敵隊Action=FF、Detail=0、標記旗標=3F | proven，IDA原始寫入指令 |
| `26E5:00AC` | 按戰場清單逐一般分隊與索敵隊，更新圖示後叫標記開關 | proven，IDA及正常進場55次呼叫 |
| `26E5:0141..015C` | 一般分隊陣營等於viewer或全域8F1C bit0設起時force=1，否則0 | proven，原始分支 |
| `26E5:0188..01C4` | 未放出索敵隊force=0；放出後與一般分隊相同 | proven，原始分支 |
| `1727:02E2` | force=0且Detail=0先擦再清bit7；其餘設bit7並嘗試點上 | proven，96案完整78-byte記錄雙跑 |
| `1727:0207` | bit7設起、bit6未設、地圖啟用才點上；顏色0→5、1→9、其他→4；寫C0與顏色、現在位置快照 | proven，IDA及96案 |
| `1727:0298` | bit7／bit6均設且地圖啟用才擦，沿用快照與低4位顏色並清bit6 | proven，IDA及96案 |
| `1727:0324` | 像素為(480+x/16,76+y/16)，底層XOR | proven，IDA及154既有底層證據 |

原版用標記旗標與快照畫小地圖，與現在位置、圖示可見性及Out分開。
已點著時重複開啟不覆蓋顏色或快照；地圖停用時開關仍改bit7，卻不改bit6／快照。
第32輪正式renderer無條件取所有分隊現在位置，是當時35像素差異的來源。
本輪不推測8F1C其餘位元、兩席切換、原版硬體時長或其他八步順序。

## 原版受控與正常重播

常式入口[探針](../../tools/scripts/battle_marker_probe.go)每案Restore、tick0、另建堆疊、
DS=ES=54B3。啟用／停用、force0／1、Detail0／2、三種陣營及四種舊旗標共96案。
原版兩次完整JSON相同，SHA-256：
`966fba906abade89332678243984ab11eb41976def7e0cc5a1d797d0d515371d`。
事前seed固定`0x12345678`，每案seed不變；所有注入與前後完整記錄保存。
這是常式診斷，不驗呼叫時機或VGA狀態。

正常入口由[重播工具](../../tools/battle_marker_oracle.sh)從S3檢查點實行，
確認遭遇訊息後進場。seed在第300000000指令前固定；不是挑種子。
正常佈陣／策略流程仍消費亂數，終態seed為`FBE6A524`，不能冒稱整段無RNG。
兩次至第435000000指令的完整RAM、307200-byte畫面、色盤與PNG相同。
完整RAM SHA-256：`256a6fe846c6e75746f690be38f3b53a2c0e6af6e417c495d78e9a33ef399e03`；
色號：`a26be89bd8de293e5832ec1a1208e0a5f9c53567121dedfc5f4188f69fd966e8`；
PNG：`c6ddfb988cf15898626fa50b040b1f2423baabae0aa27f57f3fd153abdbc00cb`。
正常清單7艦隊／55分隊、mapEnabled=1、8F1C bit0=0，8筆己方一般組C5，
其餘47筆3F。這是原版正常來源，正式remake比較另存驗證045，不以原版雙跑宣稱產品完成。

## 保留的定位／工具失敗

- 專用Go／GUI image沒有rg，改容器grep／Python，不在主機搜尋分析。
- `1727:0000`實際是前函式尾端；00B0／00F0是指令中間。初次導覽均unknown，
  沒有用這些位置宣稱呼叫者為零；真正配置入口0004與已證實的0207等分列。
- 線性26FAC／27014的開關呼叫屬於`26E5:015C／01C4`，不是2583。
  初步切段2583只作定位線索，不進已證實索引。
- 第一個正常探索停在遭遇訊息，沒有26E5呼叫。補等價確認輸入後實際進戰場，
  再用固定seed雙跑；探索輸出沒有換名成正式收據。

詳細來源、全部旗標、切段與匯出雜湊、分級索引及96案由
[產生器](../../tools/scripts/battle_marker_receipts.py)與
[資料JSON](../internals/data/battle-markers.json)保存。

## 審查匯出與舊定位勘誤

受版控索引自動合併到IDA匯出，逐項核對6個函式及10個指令的原始定位、bytes與分級相同。
其餘指令保留unknown，不因整支函式已審查而批次升格。

| 本機IDA匯出 | SHA-256 |
|---|---|
| `re79-marker-r33-reviewed-1727.json` | `67fa176c634e7947304b1900384c0ba22dac33d56c32d9f66bc7a2c89d77d508` |
| `re79-marker-r33-reviewed-26E5.json` | `67c9e061d877dcbc91d59500450d7f651e0521aa023b1e072e29e0839e29bd2a` |
| `re79-marker-r33-reviewed-297B.json` | `4dd86ac9a044e45acc891d887f16af0fb5ad4b7d7b094cc47aa849ff1dc58771` |

[研究85](85-units-are-sprite-tiles.md)的圖示常式起點`25DB:03E8`須訂正為`25DB:03E9`。
原版RAM上述`9c4a4c28…`，dosgolem線性`0x26198`起8 bytes為`cb 55 56 57 83 ec 06 89`。
03E8的CB是前支retf，03E9的55才是push bp；26E5:00AC內兩個far call均指03E9。
這是proven定位勘誤，未改舊圖示分支結論。

## 正式實作續證

正常StartBattle初始化、輸入恢復刷新及五個標記bytes回寫已接正式程式。
原版96案完整78-byte記錄全數相同；正常主選單／S3／實行入口55筆旗標、
8筆已畫快照符合原版，首次戰術實行前完整SHIP／FLEET正式存讀相同。
固定原版第6回合資料與鏡頭的兩個端點，小地圖各16384像素零差異。
全畫面仍差3971／3951像素，均在階段面板；保持nearby-state，不改八步門檻。
完整前後圖、數字、正常索敵鏈與測試見既有[驗證045](../playtest/045-battle-movement-render.md)。

第一次舊收據測試漏掉全部標記，原因是測試誤用艦艇+02當星域26篩選，
兩份原版記錄該byte均為0。改由DS:9898戰場艦隊名單取資料後，同容器同流程通過。
保留失敗log SHA-256：`4dc96ff06a64e6e605ee330dd95c87e8e3c3e4989655af2dcb26edab87ad6831`；
未修改原版screen.bin或以fallback補標記。
