# 原版診斷056：多分隊交戰與距離曲線

比較等級為受控原版重播，不作remake same-state聲明。
完整來源、原始bytes與勘誤見[研究404](../re/404-combat-sequence.md)。
[凍結收據](../internals/data/combat-sequence.json)SHA-256：199f328b4463d49a817e54c2ade49dd148a24d45e94e94d135488d699c3205a8。

## 固定條件

原版bundle de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673；
GIN3.COM 770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918。
dosgolem f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c／clean，remake父提交b61764b4e524e9885ffb799b0be870a726e405b9／本輪dirty。
機器設定沿原版svga_s3、16MiB、SB16；原始資料唯讀。
自然父狀態與第63輪已提交父狀態均綁SHA，所有祖先注入與完整flags在收據中。

## 重播

使用既有Docker工具鏈。每個名稱只可寫一次，既有輸出不覆寫。

```sh
tools/combat_sequence_oracle.sh sequence79-r66-<案例>-a
tools/combat_sequence_oracle.sh sequence79-r66-<案例>-b
tools/go.sh run tools/scripts/combat_sequence_probe.go > workplace/round66-sequence-probe.json
LOGH3_EXTRA_RO="$PWD" tools/py.sh /work/tools/scripts/combat_sequence_receipt.py /work/docs/internals/data/combat-sequence.json
```

先取得既有parent，再按[計畫器](../../tools/scripts/combat_sequence_plan.py)的依賴順序跑。
曲線案例使用11種實際映射type的zero／near，加curve0-stack-first，共23組。
多擊依序multi-entry、multi-command、multi-first、multi-second或same-target-second，再跑末端／排序觀測。
same-target-ranked的排序RAM錯相位，保留為擷取反例；正式模型只用same-target-measured。
全部點擊含步數、hold與move lead；RAM注入一律lin前綴，完整寫入plan。

## 結果與限制

| 驗證 | 結果 |
|---|---|
| 原版雙跑 | 35組70次，189組完整檔案相同 |
| 距離曲線 | 23案指定係數、庫存、戰損、分數、功績與seed符合probe |
| 多分隊排序 | 8鍵及完整順序符合；未下令分隊仍消耗排序RND |
| 兩目標 | 636、527艘，分數57，seed95190066 |
| 同目標 | 726、374艘，末端1900艘，分數54，seed95190066 |
| 中間狀態 | 第二次傷害入口承接前次艘數、彈藥及修理上限 |
| 正式產品回歸 | 18套件正常結束，13通過、5無測試檔 |
| 正常remake路徑 | 本輪未新增收據 |

原版PNG均640×480，滑鼠虛擬640×400；本鏡頭世界座標加176／80、PNG縱座標另加40。
沒有裁切、縮放、遮罩或逐區pixel parity聲明。兩個交戰末端PNG已目視；原圖及雜湊在完整收據sources。
序列化state只作固定父快照，不宣稱兩次state檔逐byte相同。
零距離已核准取曲線首點，但本輪未改production；原版stack模型不當正式規則。
死亡、其他取消、完整演出、正式保存及兩主要路徑仍待驗，092 DRAFT。

## 第67輪純核心驗證

remake父提交627c73bae07c29b4d3d8662cf31d75147a43110e／本輪dirty；dosgolem版本保持。
原版輸入、完整flags與祖先注入沿第64至66輪已凍結索引，本輪不重生或改寫原版收據。
先審查[093](../spec/093-fleet-combat-core.md)至READY，再實作
[核心](../../internal/rules/fleetcombat.go)與[測試](../../internal/rules/fleetcombat_test.go)。

```sh
tools/go.sh test ./internal/rules -run '^TestFleetCombat' -count=1 -v > workplace/round67-fleet-core-tests.log 2>&1
tools/gui.sh test ./cmd/... ./internal/... -count=1 > workplace/round67-standard-tests.log 2>&1
LOGH3_EXTRA_RO="$PWD" tools/py.sh /work/tools/scripts/fleet_combat_core_receipt.py /work/docs/internals/data/fleet-combat-core.json
```

| 核對 | 結果 |
|---|---|
| 單擊直接原版係數及末端 | 31個第64輪案例、11個近距離符合 |
| 合法戰鬥機 | 一擊呼叫、分隊／庫存、分數／功績及seed符合；沒有中途係數觀測聲明 |
| 八筆排序與兩組連打 | 全部鍵、完整排列、攻守指標、中間狀態、逐擊係數及末端符合 |
| 已核准零距離 | 11種type及一個stack反例取原版表首點；這是刻意差異驗證 |
| 原版數值表 | 12武器各六點、五排序基值、11機動係數與兩陣營艦種符合 |
| 錯誤原子性 | 無效命令／資料及零防禦不扣庫存、不取RND；排序先驗全部條目 |
| 核心回歸 | 六組測試無skip通過 |
| 標準回歸 | 18產品套件正常結束，13通過、5無測試檔 |

[純核心收據](../internals/data/fleet-combat-core.json)保存程式、工具image、test log與原版來源雜湊。
SHA-256為2afa559c97106a4e702307b0ceeef8889073d170a37cca9009f3811bb60b21b7，共300份來源。
凍結器拒絕覆寫；可另寫workplace內的新檔，與凍結JSON逐byte核對。
正式Game尚未呼叫核心，沒有新的畫面或正常玩家路徑收據。093 READY、092 DRAFT。
