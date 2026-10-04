# 420：戰場升級的全局重算與星域擷取勘誤

沿[Issue #79](https://github.com/wicanr2/logh3/issues/79)與[108草稿](../spec/108-battle-promotion.md)。
本輪只補原版診斷與既有規則的組合驗算，未接正式8-4。

## 來源與位址

| 來源 | 身分 |
|---|---|
| GIN3.COM | SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918` |
| 第82輪父RAM | `workplace/oracle/ownership79-r82-entry-a/mem-in.bin`，SHA-256 `b905eda112fd3a78292bc0d1cdc10068703e0291968640cbaa0d150404404900` |
| 完整父保存 | `workplace/oracle/ownership79-r82-entry-a/entry.state`，SHA-256 `598e10e5f9457842275fa5639219636f322d8bb2e0f8393fe25b93bfbdef0cfe` |
| dosgolem專用副本 | `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，工作樹乾淨 |
| 觀測器 | `workplace/promotion79-r86-global.go`，SHA-256 `895f92e7ac5f3431569f2b84cf5bf342114e249f87051f6d07345fafd6573b8b` |
| 原版雙跑 | 各24871318 bytes，SHA-256 `2bc8fa7bf8dc4081fad414d23e7bfeabb76cd8df510d53f3442136283cb1a435` |
| 全局收據 | [battle-promotion-global.json](../internals/data/battle-promotion-global.json)，475742 bytes，21來源，SHA-256 `38db9c02ae4f4883f13d4cb5fc5316620b23b4e179eb5ecbfb4332cbb13d81b6` |
| remake | 父提交 `710d2d13ed7fa807c0be657710c82857b93da637`；沒有產品變更 |
| 工具 | `golang:1.24-bookworm`與`logh3-ebiten:2`，UID1000，無網路；本輪沒有新IDA匯出 |

原版執行期DS為54B3。以下是dosgolem的CS:IP／DS位移，不能混用IDA的COM線性位址。
STELLAR基底為DS:A79C，行星檔頭為DS:CFB8，行星payload遠指標為DS:5336／5338。
行星檔頭64 bytes與81筆82-byte記錄組成6706 bytes輸入。
路徑四個遠指標位於DS:D062、D066、D06A、D06E，本次共同EMS frame為D000:0000。

## 第85輪勘誤

[419](419-battle-promotion-recalc.md)的觀測器誤用DS:A7BC擷取STELLAR，起點晚32 bytes。
第一筆從防衛力開始，缺首32 bytes，尾端多讀32 bytes鄰接記憶體。
因此該輪的STELLAR4032 bytes不能稱為完整星域表，差分位移也不能直接當作記錄位移。
新雙跑的正確STELLAR後4000 bytes與舊擷取前4000 bytes逐byte相同，證實偏移原因。
舊工具與JSON保持，新增修正觀測器；三表、WORK費用與亂數的舊結論保持有效。

最初全局probe誤用舊STELLAR，53次有6624個差異bytes。
輸出`workplace/promotion79-r86-global-probe.json`保留，SHA-256
`a423024153a6aa77e45343908a99d65ba2c57c17765a7a9aca7d9a5a0380970c`。
這是擷取錯誤，不能當產品缺陷，也不能用欄位遮罩消除。

另一個觀測限制是EMS映射。8-4入口的frame仍是圖像資料，不能把`paths`欄位直接當路徑矩陣。
模型沿[042](../spec/042-region-unrest-recompute.md)，從重算前星域歸屬獨立建表。
只在重算返回後核對6144-byte EMS資料，53次四矩陣全部相同。
未把原版返回後路徑作為模型輸入。

## 已核的資料鏈

所有結果為`proven`，範圍限固定8-4入口診斷。
沿用[419](419-battle-promotion-recalc.md)已審查的成長與三表模型，
新舊66案的完整ADMIRAL、FLEET、SHIP、WORK、呼叫、輸入、frame、注入與seed相同。
53次三表雜湊逐項符合第85輪v2的獨立模型雜湊。

| 既有規則與來源 | 本輪核對 |
|---|---|
| `2703:02E5`，042／研究299、304 | 星域經濟力由重算前行星加總，依歸屬累加預估稅，再算32格騷動 |
| `ResumRegionEconomy`、`RegionTax` | 每次完整STELLAR4032 bytes相同，不只核+1A／+7D |
| `UnrestState`、`StarGraph.UnrestPaths` | Turn、Ruler、Blend、Rate取重算前WORK；Table／Base取固定父RAM；四矩陣與返回後EMS相同 |
| `297B:10DE`／`116C`／`12FB`，393／419 | 使用已獨立核過的全艦隊成本低word，signed16加總維持費及狀態非2／5的移動費 |
| `090A:0E88`，419 | 自行算出的稅額參與signed32可用款，核造艦預估分支；每次完整WORK230 bytes相同 |

模型53次返回共比較213696個STELLAR bytes及12190個WORK bytes，沒有遮罩，零差異。
66案最終STELLAR、WORK及行星全表也相同，包括沒有成功升級的案例。
行星全表在本批始終不變。637個星域改變bytes只落在記錄+1A、+1B、+7D。
WORK有218個改變bytes；ADMIRAL280、FLEET344、SHIP243，與第85輪一致。
模型兩次輸出完全相同，76521 bytes，SHA-256
`a18d643f57080ec63e8581f7469b03f7566e06a94b2a1d60513fc0406b466292`。

## 停止線

全局公式已有最小充分證據，不再重查通用硬體或費用helper。
下一步是列明typed欄位窄回寫、逐則報告等待與軍階UI、保存時的覆寫邊界，
完成108證據審查後接正式8-3返回點。
正常玩家／GUI及完整原版整段同態仍缺，108與完整062保持DRAFT。
