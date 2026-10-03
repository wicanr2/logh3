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
