# 090：正常戰鬥結束與公開存讀的60 Hz驗證

2026-10-05，第102輪，GitHub #79。
父提交b5ac935d2e77d45d82e1173d340e26fb116f03d8，正式產品無修改。
沿[109](../spec/109-battle-empty-groups.md)及[110 READY](../spec/110-battle-finish.md)。
來源由[完整索引](../internals/data/battle-lifecycle-60hz.json)記錄SHA-256及大小。

| 核對 | 結果 |
|---|---|
| 正常入口 | 正式bootstrap及cmd/gin3，預設片頭、主選單、防拷138、S3同盟、戰略實行與首戰 |
| 固定條件 | 首次戰場實行前設定seed12345678一次；無位置、艘數、功績、命令或勝負注入 |
| 正常結束 | 沿第95輪6745項公開輸入，50次下令／19次交戰；第九戰術回合先全滅再敗退，回命令階段 |
| 目前程式雙跑 | a／b各141檔逐byte相同，含計畫、63張色號及PNG、公開保存與末端重存 |
| 正式更新頻率 | main正文從目前程式重建，只附觀測；沒有SetTPS，10090次更新實際TPS均為60 |
| 真實GUI | 1032項操作，63張全圖共19353600像素；Game色號及X11原生RGB零差 |
| 鏡頭輸入 | 279次額外右鍵回授完整記錄；不稱相同實體輸入時序 |
| 結算資料 | 全滅與敗退報告的文字、頭像、完整六表、分隊及seed符合目前Game計畫 |
| 公開存讀 | 六檔44560 bytes符合計畫；真實公開保存、讀檔及末端診斷重存相同 |
| 抽圖 | 全滅及敗退X11全圖已目視，文字完整，頭像及游標可見 |

## 範圍

這是remake正常玩家流程驗證，沒有新增原版正常整回合同態收據。
原版單擊、回收及勝負契約沿既有來源；本輪不升格為原版same-state。
兩則報告不含占領。非旗艦自然死亡／空組搬移、自然升級及占領GUI仍缺。
單人戰術CPU沿061保持停留，既有增強戰略AI不改。
109／110保持READY、完整062 DRAFT、#79 OPEN。

## 重跑

使用既有logh3-ebiten:2，命令經tools/gui.sh，UID/GID1000、無網路及有界Xvfb。
輸出存在即拒絕覆寫。下列a／b是已完成批次，重跑要選新的英數revision。
先等計畫與建置完成，再啟動GUI。

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_lifecycle_60hz.py a
bash tools/gui.sh xrun python3 tools/scripts/battle_lifecycle_60hz.py b
bash tools/gui.sh xrun python3 tools/scripts/battle_lifecycle_60hz.py a --gui
bash tools/gui.sh xrun python3 tools/scripts/battle_lifecycle_60hz_receipt.py --check
```

輸入與擷取工具來源為tools/scripts/sound_tail_input.go及fleet_player_gui_capture.go。
容器內用go build分別建至workplace/order79-r93-input及workplace/round79-capture。
觀測器、計畫及驅動器的歷史來源都先核對第95輪固定SHA-256，來源重生入口見[084](084-battle-lifecycle-player.md)。
只附加觀測函式，正式Game及app正文使用目前工作樹，不以歷史產品覆蓋。

第一次GUI命令在計畫尚未完成時執行，缺plan.json即退出，沒有啟動遊戲。
保留已生成驅動器，待建置完成後執行同一驅動器成功；生成器已加前置檔案檢查。
這是相依命令排序錯誤，不是產品缺陷。兩次失敗的文件patch沒有部分寫入。
標準18套件13通過／5無測試。首次凍結器假設成功列為ok加tab，實際Go输出中間有空白；
核對器改讀第一個欄位後，以同一標準log及同一容器重核，不重跑已通過的產品測試。
