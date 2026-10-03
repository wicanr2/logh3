# 054：第④步原版單擊戰損診斷

第64輪，2026-10-03至2026-10-04。分類為受控原版重播及可丟棄計算probe。
本項沒有正式remake same-state或完整戰鬥UI聲明。

來源與契約見[研究402](../re/402-fleet-combat-damage.md)、[092草稿](../spec/092-fleet-combat-entry.md)。
[完整收據](../internals/data/fleet-combat-damage.json)SHA-256為
`360277c8dcbf6214da429baead75d4a7bf20b09b03f1ca2af3c0313df4171b30`。

## 固定條件

- 原版bundle與GIN3.COM的SHA、固定dosgolem f7523f9／clean、Go／Python／IDA／GUI image身份均保存。
- 父狀態為第63輪光線砲或飛彈提交，保存步數495599999；同一S3受控局面及固定a父狀態。
- 常規局部注入在506948195，輸入RAM在506948196；零艘數與早期zero-fallback在495600001注入。
- 每項lin位址、bytes、理由及步數保存。原始XY為416、280及516、280，目標池102。
- limit508000000、tick20000；31組62次原版執行，沒有新增滑鼠或鍵盤輸入。
- 原版Mode12h及remake邏輯畫布640×480；本輪只比較原版雙跑，未裁切、縮放或排除像素。

## 可重播方法

工具拒絕覆寫既有原版run及凍結JSON。新工作樹需備妥固定父state，所有祖先plan已嵌入JSON。
原始state、RAM及PNG是本機oracle，不進Git；缺來源即失敗，不自行替代。

```sh
bash tools/fleet_combat_damage_oracle.sh damage79-r64-laser-fixed-a
bash tools/fleet_combat_damage_oracle.sh damage79-r64-laser-fixed-b
bash tools/go.sh run tools/scripts/fleet_combat_damage_probe.go
LOGH3_EXTRA_RO="$PWD" bash tools/py.sh \
  /work/tools/scripts/fleet_combat_damage_receipt.py \
  workplace/round64-damage-rebuilt.json
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
```

全部31個案例名、兩跑旗標及父狀態SHA均在JSON，以上只示入口用法。
probe沿既有Fixed／Ellipse／Rand／ApplyShipLoss，讀完整原版RAM及暫存器，逐項核對：

| 比較項 | 結果 |
|---|---|
| 五類原版檔案 | 31×5＝155組逐byte相同；畫面每組307200像素，差異0、最大通道差0、bbox空 |
| 呼叫及返回 | 相同觀測點的兩跑資料相同；早期observe多出的六處唯讀觀測完整保留 |
| 29個有傷害callee案例 | 攻擊、防禦、射擊量、百分比、requested、actual、seed及指定庫存／修理上限符合 |
| 2個零艘數案例 | 不進0C54；艘數、庫存、修理上限、分數保持；只有前段16次排序亂數 |
| 計分／功績 | 兩個u32分數各actual/20，提督actual/2、副官actual/4符合 |
| IDA9.4 | 四份1179條指令均逐byte重查固定RAM，30項附加proven；其餘保持unknown |
| 來源 | 649份指紋，UID/GID1000:1000；完整注入回查RAM及report實際線性位址 |
| 標準回歸 | 13有測試套件通過，5套件無測試檔；原有18產品套件均正常結束 |

戰鬥機每艘架數×艘數先夾庫存，再扣20%；艦種3把damage用架數除二。
99／100艘分支、負奇數／偶數往零除二、旗艦操作模式及零傷害第二次RND均有實際原版收據。

## 尚未通過的範圍

- 戰鬥機及零庫存案例以注入進傷害callee，不能取代正式可選旗標、UI提交或演出驗收。
- probe只重建指定欄位，未宣稱完整SHIP／FLEET或正常保存逐byte相同。
- 單擊返回後的508M畫面不代表完整0152／026B演出結束；多攻擊者與正常玩家入口仍缺。
- 其他武器類型、q=0曲線前項與極端溢位不外推；序列化state差異未裁決。
- 092 DRAFT、#79／#48 OPEN；061被動政策及#89待裁決忠實度差異保持。

檔案存在、原版兩跑相同或probe通過都不單獨作正式功能完成聲明。
