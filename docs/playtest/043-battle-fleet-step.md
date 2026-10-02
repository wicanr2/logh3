# 043 — 艦隊單步、通行與保存鏈

日期：2026-10-02；[#79](https://github.com/wicanr2/logh3/issues/79)；
[規格081](../spec/081-battle-fleet-step.md)、[研究372](../re/372-battle-step-and-collision.md)。
父提交`3a4b92f9f47efd2ac7de525141b97211b88ea047`；修改中的工作樹，提交身份以Git紀錄為準。
Go1.24.13、`logh3-ebiten:2`，image ID
`sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8`。

## 有限同狀態資料對拍

原版固定`GIN3.COM`、dosgolem f7523f9與seed／完整初態／注入／呼叫參數均由
[分級JSON](../internals/data/battle-step.json)及研究372保存。102案雙跑只證明原版可重現。
remake用相同前態重建正式型別、原版地形表與注入地圖；單步規則不接收或消費RNG。
驗38筆艦隊收據（含7筆重複Action輸入）的完整122 bytes、一般分隊追隊bit、AL及全部候選順序；
另驗9筆15／16像素及陣營／種類／出局碰撞。此範圍屬**same-state資料**，沒有原版／remake全畫面比對。

```sh
bash tools/gui.sh test -count=1 -v ./internal/rules -run '^(TestFleetSingleStepMatchesOriginal|TestBattleMovementCollisionMatchesOriginal)$'
bash tools/gui.sh test -count=1 -v ./internal/game -run '^(TestFleetMoveNormalExecuteAndSave|TestOrderedFleetAdvances)$'
bash tools/gui.sh test -count=1 ./internal/... ./cmd/...
```

| 收據 | SHA-256 | 結果 |
|---|---|---|
| `workplace/battle79-fleet-step-before.log` | `d64d700101feaf90e56a53b246ebdcdbb8962de2ee1225c4d2f65b067f217e46` | 舊規則16案3失敗 |
| `workplace/battle79-fleet-step-after.log` | `e886dd84a50209181190372d019810d3a9d7b4d81c27c59fab861905e2fb0f76` | 38艦隊＋9碰撞通過 |
| `workplace/battle79-fleet-step-player-fixed.log` | `f870276c9002c587cbaa555980ec3fc917fc883743b2cb5d3d0957e1c16e10d2` | 正常移動與正式存讀通過 |
| `workplace/battle79-round30-standard-fixed.log` | `734f8f9c27cd58803d8323f751b2bb389df225c550cdb26a898926e2f59b1eac` | 18套件通過或無測試 |

首次正常存讀兩次失敗，實際差僅+5C／+5E；保存前同步以戰略Fleet舊座標覆寫。
補同步後以同一容器與命令乾淨重跑。標準測試首次也因此失敗，保留
`workplace/battle79-round30-standard.log`；不是放寬比較條件後過關。

## 正常玩家路徑與畫面

`battleTestGame`從主選單、防拷138、正常載入S3及實行遭遇進達希利；沒有direct-entry。
`TestFleetMoveNormalExecuteAndSave`點可見己方旗艦→移動→可通行目的地→抵達朝向→
正常戰術「實行」，等待已接資源／報告及演出續接。進實行前固定remake seed=0x12345678；
後序戰鬥可消費亂數，此抽樣不宣稱全回合seed或骰序相同。
艦隊從(416,280)移至(472,288)，到第2回合；畫面正常顯示達希利。
完整122-byte艦隊記錄經正式`SaveSlot`及讀檔相同。

| 本機正常抽樣 | SHA-256 |
|---|---|
| 原生640×480色號 `workplace/battle79-round30-normal.bin` | `8b63b30e23eead448a8d4f1a17bdb7cdfff3ce95a8ed3b2f16102be42fc36b1c` |
| 16色6-bit DAC `workplace/battle79-round30-normal.pal` | `3c12c8aa1e110e7839984929c9dd62fafb9ea88f0edd7f4ecdeaece9100da511` |
| RGB8展開PNG `workplace/battle79-round30-normal.png` | `220a8ae61f07268ad644b9a50f6659851e46531d21f6cdefbecca7e464629fe0` |

PNG依`Palette.RGB8`的`(v<<2)|(v>>4)`展開，原尺寸、未裁切縮放；已目視第2回合、
小地圖與右側文字。它是正常remake路徑畫面，沒有對應原版同幀收據，故不產生像素差分或全畫面parity聲明。

另依序執行`LOGH3_GUI_PREFIX=main1e bash tools/gui_check.sh`與
`LOGH3_GUI_PREFIX=fcmd-s3-found bash tools/gui_check.sh`，兩份既有正式Ebiten正常入口收據逐位元組相同。
這些入口抽樣不替代第二階段完整原版／remake同狀態、UI中間幀或完整分隊存讀。

## 停止線

規格081只涵蓋艦隊單步與這條保存鏈；規格080完整交錯移動仍DRAFT，#79不關閉。
原版材料、完整RAM、原圖、IDA資料庫均留本機。新輸出為1000:1000，一次性Docker均以
`--rm`退出；本批沒有重建image，收尾容器清理結果另記WORKLOG。
