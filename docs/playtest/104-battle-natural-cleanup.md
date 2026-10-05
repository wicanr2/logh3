# 第116輪：正常非旗艦死亡與空分隊清理

目前分類：正常remake自然死亡、清理、下一回合及公開存讀已驗。
沿#79及109限定驗收；原版局部清理證據保持，062 DRAFT。
正式Go、AI、存檔格式、dosgolem與原版素材未修改。

父提交為`e0db0df50673b1972d238573604fa41ab3c14a7b`。
第115輪[正常升級來源](../internals/data/battle-natural-promotion.json)SHA-256為
`0790049faea28eb077b82c05ce8cf3a947f65e62d6b9bad564762e972e2e0824`。
正常S1公開存檔、兩個人類座位及公開移動路線保持，僅改測試玩家的公開選敵順序。
第一次戰術實行前固定一次seed`12345678`，不注入死亡或其他局面，不外推跨存讀骰序連續。

只讀觀測層保存109提交前後的完整六表、分隊與引用映射。
原版35案／77空組已驗證據沿109，不把本輪正常remake流程稱為新原版整回合對拍。
正式Go、AI、資料格式及dosgolem未修改。只讀觀測層不更動Game狀態、鏡頭、命令或RND。

## 正常死亡與完整資料鏈

採公開右鍵循環選取重疊的一般分隊，公開左鍵確認及下攻擊命令。
`middle`策略先排除末槽，第二輪在TURN99、星域16、第3戰術回合自然擊毀同盟艦隊16的原池95。
旗艦原池93仍有3000艘，死亡分隊原Squadron為1。

| 核對項目 | 實際結果 |
|---|---|
| 有效組數 | 7→6 |
| 有效引用 | 93、95、96、97、98、99、100→93、96、97、98、99、100 |
| 索引映射 | `[0,-1,1,2,3,4,5]` |
| 完整四byte引用 | 後方五組逐byte前移，尾槽保持 |
| 死亡SHIP+01／+19 | 16→FF／C9→89 |
| 死亡SHIP+03／+05 | 1／19保持，不重編原池欄位 |
| 死亡SHIP+38..3B | 前後皆零；非零清除沿原版35案證據 |
| 其他四表 | ADMIRAL、WORK、STELLAR與含64-byte表頭的PLANET完整相同 |
| 清理seed | 前後2016853485相同 |
| BattleUnit | 移除死亡組並依映射重編，其他欄位逐項保持 |

正常18次命令、42擊後回到命令畫面，只有一件清理、一件副官升級，零勝利或占領報告。
助手c／d各196檔相同；正式bootstrap b／c各242檔相同。
清理前後完整六表、所有分隊、所有艦隊與seed，助手與正式入口逐項相同；只除去助手時間戳。
清理後第4回合可正常輸入，正式入口與實際GUI完整圖相同，後續仍正常下令至戰術結束。
六檔44560 bytes公開保存、診斷及公開讀回重存皆相同，存讀不重新計費或重編SHIP+03。

## 工具失敗與乾淨重跑

- a只選單一命中，重疊一般分隊無法被選，多次退回旗艦。讀正式選敵後改公開右鍵循環。
- b最弱分隊策略只刪末槽，證明自然清理；中間引用搬移另用c／d驗證。
- replay a生成器語法錯誤，未執行遊戲。b／c修正後獨立重跑。
- GUI b有15張全圖零差，之後右鍵在兩側往返。Python確認按下再放開的往返讓正式程式多執行RightHold，屬測試時序。
- GUI c沿用[既有XTEST工具](../../tools/scripts/sound_tail_input.go)，由同一程序觀測正式Update後放開。正常60 Hz更新、正式RightClick／RightHold及玩家鏡頭規則保持。

舊批次、輸入及失敗log均保留，不覆寫為通過。

來源核對初稿把原版`cases`清單誤認為計數，改用清單長度。
另一個助手／正式快照不相等已定位到取圖時點：助手`observe`在第一個公開右鍵後執行，
正式`collectCleanup`在Tick後立即執行。前者鏡頭40／47、游標17／61，後者鏡頭41／48。
原圖5535像素不同，不遮罩或改寫。相同公開輸入補取右鍵後全圖，與助手逐byte相同；
新增兩張診斷圖之外，重播242個原檔全部保持。首批診斷把`view`的`[]int`誤用為`[2]int`，
修正型別後重新生成批次；兩份輸出皆保留。

## 實際GUI與驗收範圍

正常主選單、公開S0雙人讀取、艦隊命令、戰術回合、公開保存及讀回共615項操作。
108張原生640×480全圖，共33177600像素全部零差，沒有裁切或遮罩。
15055次正式更新TPS均60，176次額外鏡頭回授及75次選敵右鍵循環完整保存。
末筆實際輸入251.320364秒；助手Tick總和不作GUI時長。
清理前後完整六表、全部分隊與艦隊、映射及seed，GUI與正式重播逐項相同。
六檔44560 bytes在助手、正式診斷、公開保存、讀回重存與GUI全同。
已目視第4回合輸入及讀回星圖。標準18套件13PASS／5無測試。

限定完成空組清理的規則、Game、正常GUI與存讀鏈，不外推原版整回合八步或整場對拍。
原版35案為same-state局部診斷；本輪GUI全圖是正式remake與同態重播比較。
109限定CONFORMED，062 DRAFT、#79 OPEN；#48直接確認、MIDI／#89、人耳及平台交付限制保持。
[唯一來源索引](../internals/data/battle-natural-cleanup.json)保存原始文件、輸出與工具SHA-256。

## 重播入口

使用既有`logh3-ebiten:2`，Go1.24.13、UID／GID1000、4 CPU、4 GiB、PID512、無網路。
原版目錄掛為`/orig:ro`。每次換未使用的revision；腳本拒絕覆寫既有收據。

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_natural_cleanup_probe.py REV --target middle
bash tools/gui.sh xrun env LOGH3_FLEET_CHAIN_OUTPUT=/src/workplace/cleanup79-r116-REV \
  go test -overlay=/src/workplace/cleanup79-r116-REV-overlay.json \
  -run '^TestBattleFinishNormalProbe$' -count=1 -v ./internal/game
bash tools/gui.sh xrun python3 tools/scripts/battle_natural_cleanup_replay.py REV
LOGH3_TIMEOUT=45m bash tools/gui.sh xrun python3 tools/scripts/battle_natural_cleanup_replay.py REV --gui
bash tools/gui.sh xrun python3 tools/scripts/battle_natural_cleanup_replay.py TIMINGREV --snapshot-timing
bash tools/gui.sh xrun python3 tools/scripts/battle_natural_cleanup_receipt.py --check-only
```

探針與重播的預設父資料、seed、正常S0存檔及來源雜湊由腳本核對。
`--write`只能在文件與原始輸出定稿後執行一次；日後不帶參數核對凍結來源。
源資料仍屬本機oracle，已解摘要與雜湊保存於版控。
