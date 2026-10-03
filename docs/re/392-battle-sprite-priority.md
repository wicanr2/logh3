# 392 戰場圖示順位初始化

沿[#89](https://github.com/wicanr2/logh3/issues/89)及[第54輪](../goal/2026-10-03-battle-sprite-priority-round-54.md)。
父提交1ce5c23ef2bab576eb21971fa81e1896dbe49d6a，第51至53輪收據保持。

## 原版與位址

GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
dosgolem `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`／clean，沿既有S3載入snapshot、
完整正常輸入及實行前D3DC167E。新增兩份獨立正常入口收據，RAM／色號／PNG／DAC相同。
入口RAM SHA-256 `ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97`。
完整旗標由[原版重播工具](../../tools/battle_sprite_priority_oracle.sh)保存。

IDA Pro 9.4切段COM SHA-256 `d1004845670387a15986f585658051965a2cb899f1f499d062b7b7ba8bd34b7a`，
IDA基準10200，原版CS:IP與dosgolem線性位址另列。7..952共873條原始指令全部對回RAM；
只審查0007／0170／0915三個函式邊界及19項附加語意，其餘保持unknown。

## 已證實的生命週期

| 原CS:IP | 原始bytes | 語意 |
|---|---|---|
| 25DB:00A5 | `9a7001db25` | 一般分隊呼叫0170建立圖示 |
| 25DB:00B8 | `26ff7738` | 一般組之後取FLEET+38偵察池 |
| 25DB:00CC | `9a7001db25` | 偵察池也呼叫0170，不由出擊旗標排除 |
| 25DB:01AC | `26884430` | AL80寫SHIP+30 |
| 25DB:01C9 | `26c6470380` | 對應顯示物件+3同設80 |
| 25DB:01E7 | `26894704` | word+4高byte為分隊序號、低byte為艦隊索引 |
| 25DB:0922 | `268a4730` | 後續刷新讀SHIP+30 |
| 25DB:092C | `26c437` | 經SHIP+38遠指標取得顯示物件 |
| 25DB:0939 | `268804` | 順位改變才覆寫物件+3，並將dirty設FF |

0007正常呼叫一次，407194657步進入。0178／01A8／01B0各55次，順序與池身份兩跑相同。
01B0是01AC寫入完成後的IP，不能當寫入指令位址；每次ES:SI對應完整SHIP池，AL均80。
末態55筆+30及各自物件+3皆80，物件word+4也符合原SHIP+01／+03。
0915捕獲前80次，這是記錄上限，不稱整段只呼叫80次。

SHIP+30除補給／修理候選排序外，也是戰場顯示順位，兩用途共用同一格。
補給排序的遞減寫入沿[研究131](131-supply-and-repair.md)與既有079範圍，這輪不推測未捕獲的清除時點。
renderer比較器沿[085](../spec/085-battle-unit-draw-order.md)：物件+3無號byte升序，word+4降序。
原085已驗的全部128情形保持，這輪僅接已證實的typed順位來源與正常入口初始化。

## 探針限制

一次性`normal89-r54-low-first-explore`明示注入兩池及物件位置／順位後純移動，
未觸發完整物件重畫，畫面仍等同既有空白Input。它不能驗不同順位的原版像素結果。
探針旗標與輸出留在workplace/explore，不收入正式雙跑或升格為正常玩家證據。
不同順位的renderer回歸只能支持已證實比較器的remake實作，不稱新增原版same-state對拍。
完整088、FF相等順位字鍵及首戰剩餘一像素保持待驗，不造DOS指標或像素常數。

## 限定驗收

[完整收據](../internals/data/battle-sprite-priority.json)SHA-256
`4540cb1002bf37cf7f927cd91349d3910cf59b66b4d06ee24c893a84768b5e29`。
父snapshot SHA-256 `0155f63e373582b76a3b7ee54eccbef5ae2626481e09a68550b5fddc580146d7`，
載入路徑、大小265888 bytes、完整命令與兩側輸出SHA均保留。

- 正常Game55檔a／b相同，相較第53輪只改55個SHIP+30 bytes，載入前快照及其他表保持。
  兩AI55筆快取／順位、48可點選單位及完整SHIP存讀通過；typed順位不增加既有JSON／存檔格式。
- SHIP269減214，剩餘皆+38..3B；WORK14／FLEET27／STELLAR0／PLANET0及seed保持。
- 五張全圖與第52／53輪相同，原PNG／差分圖重核後沿用；初始仍1 pixel，其餘四張0。
  正常cmd/gin3 X11一倍／兩倍整張邏輯畫面與Game相同。
- 舊renderer忽略順位，不同順位重疊子案例失敗；現行四分支通過，包含無號FF及相等順位鍵。
  色塊只辨覆蓋，不是原版素材代用品；沒有聲稱非128順位原版畫面通過。
- 148限定PASS零skip、18標準產品套件通過，含既有索敵、準備、增強AI及保存回歸。
  完整088 DRAFT，#89／#79 OPEN；未知欄位、FF字鍵與原Goal保持。

重播需本輪來源及已保存的父snapshot，各輸出拒絕覆寫：

```sh
bash tools/battle_sprite_priority_oracle.sh normal89-r54-init-a
bash tools/battle_sprite_priority_oracle.sh normal89-r54-init-b
bash tools/gui.sh xrun python3 tools/scripts/battle_normal_probe_seed.py --prefix round54-normal
bash tools/gui.sh xrun go run -overlay workplace/round54-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round54-pass-a -input-display
bash tools/gui.sh xrun go run -overlay workplace/round54-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round54-pass-b -input-display
bash tools/gui.sh xrun env LOGH3_TURN_GUI_PREFIX=round54-gui LOGH3_TURN_PLAYER_DIR=round54-pass-a bash tools/turn_counter_gui.sh
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
bash tools/gui.sh xrun python3 tools/scripts/battle_sprite_priority_receipt.py
```

一次性重疊探針另由工具的low-first-explore分支保存，不能替代上列正常init-a／b收據。
873條中的函式服務及清除helper維持unknown，不擴大逆向深度；目前證據足以接玩家入口。

## 2026-10-03探索旗標勘誤

[研究394](394-battle-object-keys.md)核對probe的-poke是單值旗標。
本頁多筆-poke探索只會保留最後一筆，不能證明全部位置／順位注入生效。
上文未重畫探針原本就標unknown，沒有正式通過聲明；正常init-a／b僅有一筆seed注入，範圍保持。
新反例改為同一參數的分號腳本，report逐筆驗實際寫入位置；不改執行器或重生舊JSON。
