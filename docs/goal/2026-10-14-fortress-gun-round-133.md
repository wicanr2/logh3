# 第133輪：觸發要塞砲與補演出證據

2026-10-14。使用者要求先commit／push現有HD與decomp，再繼續#48並設法觸發要塞砲。

| 項目 | 目前狀態 | 本輪閘門 |
|---|---|---|
| HD／decomp保存 | `b36b950f`已提交並推送main | 完成，保留原證據等級 |
| 要塞砲規則 | 028 READY；同次原版前後RAM已取得，148支有效分隊驗10支／10637艘及seed／庫存／分數相同 | 本輪規則回歸完成 |
| 要塞砲演出 | 原版命中／空射／動畫關閉三案雙跑828檔相同；PFAM首發一次、逐分隊中彈已有收據 | 正式場景仍未接，#48保持OPEN |
| 正常操作 | 正常片頭→S1同盟→伊謝爾倫選單瞄準及實行；無狀態／seed注入，無被動覆寫；2支友軍4932艘 | 本輪正常開火抽樣完成，不升格動畫或存讀 |

原版採專用dosgolem副本`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，原版唯讀，固定seed一次。先沿[028](../spec/028-fortress-gun-fire.md)、[研究145](../re/145-fortress-gun-aim-and-fire.md)與[研究148](../re/148-fortress-gun-fire-receipt.md)的已知條件：有要塞、GunState=3，在第五步開火。未知圖序列不猜補進正式程式。

工作歷程追加[WORKLOG](../../WORKLOG.md)，目前入口為[CONTEXT](../../CONTEXT.md)，Issue仍是未完成項權威。

成果為[研究446](../re/446-fortress-gun-scene.md)、[驗證122](../playtest/122-fortress-gun-trigger.md)、[來源索引](../internals/data/fortress-scene.json)。[核對器](../../tools/scripts/fortress_scene_receipt.py)重核828檔、IDA778指令及正常GUI。[正式規則回歸](../../internal/rules/fortress_scene_oracle_test.go)缺原版收據便skip。標準19套件15PASS／4無測試；未建立新READY呈現規格或新發行包。

原版有界觸發入口為`bash tools/fortress_scene_oracle.sh <hit|empty|off> <a|b>`，探針[fortress_scene_native.go](../../tools/scripts/fortress_scene_native.go)記錄每筆線性位址注入、完整輸入、呼叫與逐格畫面，拒絕覆寫輸出。此為診斷工具，不能作正式玩家路徑聲明。

[IDA匯出器](../../tools/scripts/fortress_scene_ida.py)只在既有資料庫副本核對2264的場景caller；原始DB與執行檔唯讀，保留原始命名與位址，並逐指令核對本輪原版RAM。

[正常GUI觸發器](../../tools/scripts/fortress_player_gui.py)由正式片頭載入S1同盟，先撤退非要塞戰場，再從真實要塞行星選單瞄準及實行。觀測只讀取現有行星／戰損；不注入局面、seed、CPU政策或鏡頭。測試射擊允許朝最近友軍方向，原版候選本來就含兩陣營，結果如實記錄。
