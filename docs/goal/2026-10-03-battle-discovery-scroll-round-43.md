# 第四十三輪：索敵來源標頭與逐格移鏡

工作項 [#79](https://github.com/wicanr2/logh3/issues/79) 保持開啟。
父提交 649bef601b3ed46f976e9cfc5627a729ca41310b。
開工遠端88張worklist、11張open，268條核實零不成立。

沿087 READY及第39至42輪成果，固定原版入口與seed，補獨立艦隊／已放出偵察隊標頭、
近距來源及界外候選的每次完整重畫。探索與正式a／b收據分目錄，不覆寫舊收據。
新增行為先補原版證據與規格審查，再改正式Game；不猜補鏡頭或來源。

入口：[計畫](../internals/data/battle-discovery-scroll-plan.json)、
[收據](../internals/data/battle-discovery-scroll.json)、
[擷取器](../../tools/battle_discovery_scroll_oracle.sh)、
[計畫與核對工具](../../tools/scripts/battle_discovery_scroll.py)。
結果追加至[046](../playtest/046-battle-spotting-phase.md)與[研究381](../re/381-battle-spotting-display.md)。

## 本輪結果

五組原版62邊界／201份輸出雙跑相同。五個寫字前入口不當成正式UI幀；
其餘57幀／3135筆SHIP78／399筆FLEET122與正式Game及有效RGB零差異。
正式來源置鏡保留14次完整重畫，界外候選保留4／2次；沒有一次更新走完而漏圖。
父提交在相同修正後fixture三案失敗，正式Tick、正常兩回合／存讀及18標準套件通過。
舊94幀維持，087仍READY；下一閘門補正常玩家原版／remake同狀態配對，#79保持OPEN。
