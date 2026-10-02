# 第47輪：正式回合計數器的進位邊界

工作[#89](https://github.com/wicanr2/logh3/issues/89)，父提交3c7d2126562c2a81e2519fe0bab8ac09496647a9。
開工89張worklist／12張OPEN，269條核對零不成立。第46輪完成的讀檔選擇及九個控制bytes不重開。
命中remake／正式spec流程，已載入規格閘門與文件職責。

[研究251](../re/251-turn-loop-and-battle-nesting.md)顯示所有戰場與回合尾段返回後才在0374:03DD進位。
第47輪原始bytes核對：watch的03DD是執行後IP；實際寫入指令0374:03D9，bytes `ff06188f`。
目前正式Game在beginTurn就進位，用多個turn-1補償；正常S3入口原版380、remake381。
以固定原版檢查點重驗無交戰回合寫入及S3戰場保持，IDA核對原始bytes，建立088 READY子契約後才改正式產品。

驗收覆蓋正常主選單／讀檔／首次實行、所有戰場結束與模態尾段、月底9→10、兩AI與正式存讀。
不只改Date文字或保存偏移，保留完整畫面／全表差分。#89完整入口仍DRAFT，活動座位與其他全表差異另續。

本輪結果：[研究385](../re/385-turn-counter-boundary.md)、[完整收據](../internals/data/turn-counter-boundary.json)。
088回合進位子契約已驗，正常S3與保存380／799/3/1；48限定PASS、18套件及真實X11兩尺寸通過。
全圖5853 pixels及其餘全表差保留，#89／#79保持OPEN，不宣稱正常入口整段同態。
