# 第63輪：艦隊戰鬥目標與武器提交

依據[#79](https://github.com/wicanr2/logh3/issues/79)與[#48](https://github.com/wicanr2/logh3/issues/48)。
父提交3873bf01dfc15508c1395cd98074e5d59dd3b21a；開工89張Issue、9張OPEN、269核對零失效。
沿[092 DRAFT](../spec/092-fleet-combat-entry.md)及[研究400](../re/400-fleet-combat-entry.md)的固定命令父快照。

本輪核1ED2:02AB目標命中、03E9選擇、+2C／+2E寫入與返回演出／損失的控制鏈。
先以固定局面及明示輸入取得原版雙跑，確認UI幾何、可選旗標、確認／取消及提交順序。
有證據才升READY及實作；完整第④步未READY，不直接替換正式ResolveBattleTurn。

原版固定專用dosgolem f7523f9，seed及唯一入口注入沿第62輪完整記錄，不重擲。
新run使用combat79-r63獨立名稱，不覆寫原收據、凍結JSON或第62輪工具。
061單人CPU被動、正常同態門檻及#89忠實度裁決保持；不追完整骰序或硬體週期。
