# 402：第④步交戰的係數、戰鬥機消耗與單擊戰損

第64輪，2026-10-03開工，2026-10-04收尾。工作依據為[#79](https://github.com/wicanr2/logh3/issues/79)。
父提交`5c2049510890f50b52986976a520fb6f5319499d`，正式玩法未改，092保持DRAFT。
沿原版GUI還原路由，以既有提交狀態追單擊callee，不追DOS硬體時序或整段骰序。

## 證據身份與範圍

| 項目 | 固定來源 |
|---|---|
| 原版bundle | `org_game/银河英雄传说3SP.zip`，SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673` |
| GIN3.COM | SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918` |
| dosgolem | 專用logh3-oracle副本，`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，clean，未改 |
| IDA輸入 | `normal79-r61-a/end-mem.bin`，SHA-256 `6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4` |
| IDA | 9.4；raw COM前置256零bytes，base `0x10200`；原版CS:IP、IDA EA及執行器線性位址分欄 |
| 固定資料 | [fleet-combat-damage.json](../internals/data/fleet-combat-damage.json)，SHA-256 `360277c8dcbf6214da429baead75d4a7bf20b09b03f1ca2af3c0313df4171b30` |

31組原版雙跑，62次執行；完整輸入／末端RAM、色號、色盤及PNG共155組檔案相同。
649份來源指紋、四份IDA匯出1179條原始指令及30項proven語意保存。
各案例完整旗標、父快照、祖先注入、seed、呼叫及返回值皆在固定JSON。
序列化state仍有差異，原因未裁決；不稱整機state逐byte相同。

`laser-interpolate`與`missile-interpolate`使用相同觀測點雙跑。
較早的兩組`observe-b`比a多六處唯讀觀測，全部旗標與多出的資料保留；共同觀測點相同。
這項觀測差異不當作畫面、輸入或時間差異。

## 單擊owner鏈

以下proven限所列原始bytes、caller及受控觀測。

| 原版CS:IP | 契約 |
|---|---|
| 1DD6:030F → 05D7 | 排序後逐攻擊者處理；自身或目標艘數0不進0C54 |
| 1DD6:0C68／0C7E | 依序求203A:0002攻擊係數及203A:0214防禦係數 |
| 1DD6:0D01／0D28／0D70 | 光線砲、飛彈、戰鬥機的射擊量與庫存更新 |
| 1DD6:0DE4–0E82 | 百分比、目標防禦與旗艦減傷 |
| 1DD6:0E82–0F39 | 艘數比例、亂數差額、負值夾零及零傷害重抽 |
| 1DD6:0F59 → 203A:0593 | 實扣及資料更新，回actual |
| 203A:06F5／0729 → 0743 | 分數與提督／副官功績 |
| 1DD6:0F6D–0FC1 | 損失數字顯示；不是計分owner |

本次508000000指令上限已涵蓋單擊返回，未表示0152／026B完整演出結束。
戰鬥機、零庫存、座標與能力的注入只驗傷害callee；不證明正式武器選單接受它們。

## 攻擊係數

艦種byte `+0A/+0E/+12`依武器0／1／2選類型，攻方艦隊相應能力取`+3D/+3E/+3F`。
類型表為DS:4FDA起、stride6的power／長半軸／短半軸；曲線遠指標表為DS:477C起、stride4。
曲線每項6bytes，含16.16的x與word的y。

```text
q = FixedDist(攻方目前XY, 目標目前XY) / [a × (1 + EllipseEccentricity(a,b))]
fraction = FixedDiv(q-left.x, right.x-left.x)
curveY = left.y + signed高word(fraction × (right.y-left.y))
weaponPower = curveY × power / 100
attack = weaponPower × 攻方對應能力 / 100
```

插值的負斜率取signed乘積高word，不可換成Fixed.Int的往零截斷。
光線砲基準距離100時，q=`0x00008593`、fraction=`0x0000164C`、curveY63、weaponPower113、attack114。
飛彈同距離的attack279；戰鬥機注入的attack34。
類型0／5／10完整六項曲線及原始bytes已保存，其他類型不外推。
q=0的前項指標契約仍unknown，不以任意fallback猜補。

## 防禦、射擊量與百分比

防禦沿既有定點橢圓規則：

```text
bearing = Atan256(攻方.X-目標.X, 攻方.Y-目標.Y)
delta = AngleDiff256(bearing, 目標艦隊朝向)
defence = EllipseRadius(目標艦種+16, +18, delta).Int × 目標對應能力 / 100
```

半軸來自目標艦種。ℓ保留`b²/a`的16.16小數，最後半徑才取整。
目標朝向72時光線砲防禦105、飛彈103、戰鬥機95；改128／0時光線砲為144／74。

| 武器 | damage用射擊量 | 攻方庫存 |
|---|---|---|
| 0 | 艦種+0B × 現有艘數 | 不扣 |
| 1 | min(艦種+0F × 現有艘數, 飛彈庫存) | 扣射擊量 |
| 2 | min(艦種+13 × 現有艘數, 戰鬥機庫存)；艦種3再除二 | 先扣原射擊量的20%，再將damage用射擊量減半 |

```text
percentage = lowword[(shots × attack / defence) / 750]
percentage = min(percentage,100)
目標武器3：percentage /= 2
目標分隊序0：
    ADML+12 > 100：percentage = percentage × 100 / ADML+12
    否則percentage >= 80且3982:022D返回1：percentage -= 10
```

3982:022D對陣營2回FF，其餘依DS:8F25選座位，再返回DS:8F21的操作模式。
本次模式0／1分別保留100／縮減至90。這不修改061的單人被動政策。
0E17的`or bx,bx; jnb`不構成負數夾零證據；probe保留低word及無號上限。

## requested、實扣與副作用

```text
noise = RND()%200 - 目標提督ADML+12
目標艘數 < 100：noise /= 2，往零截斷
requested = signed16(目標艘數 × percentage / 100 + noise)
requested < 0：設0
requested == 0：再取RND()%10
actual = 203A:0593(目標, requested, 攻方艦隊)
```

目標提督ADML+12基準為104，不是101。
既有`ApplyShipLoss`的艘數、修理上限及兩種庫存規則均符合本次原版案例。
兩個分數各加`actual/20`；攻方提督ADML+02 u32加`actual/2`，副官非FF再加`actual/4`。
FLEET士氣等完整raw前後記錄保存，但本probe不宣稱其所有欄位已由remake重建。

| 固定seed12345678的傷害入口案例 | actual |
|---|---:|
| 光線砲／飛彈 | 775／625 |
| 戰鬥機／艦種3戰鬥機 | 505／235 |
| 防禦3／一般分隊 | 385／805 |
| 目標99／100／1艘 | 36／50／1 |
| 艘數99、提督屬性130／131 | 19／17 |
| 零飛彈、提督屬性200 | 7，先夾零再重抽 |

原有提交後不重設seed的兩武器結果仍726／576；本次局部seed重設有完整注入記錄，兩者不混用。
零庫存被注入到callee仍可能由noise產生傷害；正式候選門檻尚未閉合，不能據此啟用無彈藥攻擊。

## 工具修正與限制

- 原版重播初稿把多個dump用逗號隔開，解析即失敗，未執行遊戲。改分號後以新名稱重跑。
- py工具將scripts唯讀掛載；第一次COM準備寫錯目錄，IDA未取得輸入。改在workplace準備新COM後重跑。
- `flag-high101`名稱沿準備時誤讀；舊收據保留，現行集合用`flag-high104`。
- `zero-fallback`在495600001、排序前注入。probe初稿漏計前段16次RND，曾預估7而原版為1。
  依實際plan修正，另以`fallback-fixed`在506948195固定seed，實扣7；不追整段骰序。
- 首次凍結誤要求早期兩組觀測點集合相同；完整旗標顯示b新增六個唯讀觀測。
  共同呼叫資料及五類輸出均相同，另補相同觀測點的飛彈雙跑，差異保留在JSON。
- 短功績匯出未含完整epilogue，保留原輸出，再用新COM匯出0743–07AC完整函式。

失敗stderr摘錄及原始準備／執行log留在workplace，未改遊戲或oracle以迎合探針。
[診斷054](../playtest/054-fleet-combat-damage.md)保存重跑方法與驗證結果。
下一個正式接線閘門是可用戰鬥機及武器候選完整條件、取消／多攻擊者與排序，再閉合演出、正常Game與保存鏈。
092仍DRAFT；沒有新增正式helper或Game接線，不稱完整第④步或remake單擊對拍完成。
