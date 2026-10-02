# 375：階段文字與移動逐圈收據

2026-10-02，工作項[#79](https://github.com/wicanr2/logh3/issues/79)，父提交38794ab。
接續[研究339](339-battle-turn-eight-steps.md)及[第三十四輪](../goal/2026-10-02-battle-phase-panel-round-34.md)。

## 身份

GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
bundle `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
dosgolem專用副本HEAD `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，乾淨且未修改。
IDA Pro 9.4只讀一次性段切片，原段offset0對應IDA 0x10200，CS:IP及線性位址分列。
181B切片SHA-256 `87892eec94becbde534474afa69e1fadea6afb27444d5efe8764b1046b35958f`；
3617切片 `0d8149ca4814af6f3de69670b9cec52e46aa973a6a3e86a00ae5c981def4abe5`。
審查範圍181B:003F..0082、00FA..021B及3617:000A..0039。
原bytes、name／operand、匯出雜湊與17筆分級定位見[語意索引](../internals/data/battle-phase-semantics.json)。
未審查指令仍unknown。原始及自動合併索引的reviewed匯出留在workplace/ida。

## 已證實的順序

| 原CS:IP／bytes | 結論 | 等級 |
|---|---|---|
| 181B:0045 bb d4 3c；0051 call3617；0059 call1930:019D | 先印補給階段，再補給 | proven |
| 181B:005F bb e2 3c；006B call3617；0073 call1AAB:006D | 先印修理階段，再修理 | proven |
| 181B:0112 bb f0 3c；011E call3617；0126 call1C03:0006 | 先印移動階段，再初始化額度 | proven |
| 3617:001F call0059；002E call341B:02E5 | 寫文字緩衝後重畫視窗 | proven |

本次DS線性基底0x54B30。DS:3CD4／3CE2／3CF0各14 bytes：0C、四個Big5字、四空白、00。
三串為「補給階段」「修理階段」「移動階段」。
0C清內容與重設文字游標沿用事實49，其入口由internals索引可查。
移動入口尚顯示修理，初始化前已顯示移動，不能延後切換。
清窗後不留日期、分數或星域，正常輸入重畫完整狀態。

## 重播與審查

入口snapshot SHA-256 `39ed892ae729c2b2bd1a0b208439d106e4913867e1b50134396fabd704b39b55`。
不加seed、pokes或新輸入；父正常序列與seed注入見[三邊界證據](../internals/data/battle-movement-phase.json)。
tick=20000、停止計數1226638240；追蹤181B:0135得33圈邊界。
兩次各保存入口、33圈、終態的完整RAM、原生640×480色號及PNG，35×3=105檔逐位元組相同。
各RAM長1048576、seed78563412、鏡頭世界(1056,600)。入口／終態RAM雜湊亦與371相同。
完整旗標、報告／metadata雜湊及7艦隊／55分隊全紀錄見[逐圈證據](../internals/data/battle-phase-frames.json)。
新收據事前定same-state，限定此局面及邊界；第32／33輪nearby-state保持。

```sh
bash tools/battle_phase_oracle.sh battle79-phase-r34-a
bash tools/battle_phase_oracle.sh battle79-phase-r34-b
bash tools/py.sh tools/scripts/battle_phase_evidence.py
```

工具拒絕覆寫既有a／b；使用-dump-at planar路徑，不用mode13的-shots。
Go image golang:1.24-bookworm digest `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`。
可丟棄probe的entry／loop-00／end面板各32000像素零差異，只證明既有框、字庫、位置與色號。
3617過寬查詢退出1且未生成JSON，原因unknown；縮至完整000A..0039成功。
這是取證查詢問題，不作產品缺陷或文字引擎完成聲明。

原bytes、字串、呼叫順序、雙跑畫面與清窗契約足以讓[083](../spec/083-battle-phase-panel.md)進READY。
不加未知階段文字、不比CPU耗時；正常產品抽樣另驗，第二戰場及其餘八步仍未完成。
