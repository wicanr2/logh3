# 第65輪：武器候選與可用戰鬥機

沿[#79](https://github.com/wicanr2/logh3/issues/79)及[#48](https://github.com/wicanr2/logh3/issues/48)。
父提交124f4ee33f4593a3605574152d251ad155cbb925，89張worklist、9張OPEN、269核對零失效。
目前狀態只看CONTEXT現況表及遠端Issue；第62至64輪固定收據不重生。

本輪先核1F6F候選、射程及光線砲遮蔽，再以原版命令UI選可用戰鬥機並續到返回戰場。
局面與seed在唯一入口固定，注入、完整旗標與父狀態雜湊全部保存，不重擲。
診斷不替代正常玩家路徑。092仍DRAFT，正式Game、UI及保存未接線。
完整多攻擊者及演出owner尚待閉合；061單人CPU被動政策保持。
