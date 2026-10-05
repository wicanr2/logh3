# 第107輪：正常戰場占領玩家流程

日期：2026-10-05。沿 [#79](https://github.com/wicanr2/logh3/issues/79)，保持 OPEN。
父提交 `9b5c4b451022286c30b9731c17dba4cf71fbe257`，工作樹乾淨；遠端90張worklist、6 OPEN／84 CLOSED，270條verify零失效。
前輪有實際進度：公開退出缺漏#90已修正並推送，正常Windows存讀完成。

命中正常玩家路徑／存讀與remake驗收路由；入口為
`~/.codex/knowledge-base/sources/claude/retro-cht/retro-game-playtest.md`。
文件依 `~/.codex/knowledge-base/local/project-document-responsibilities.md`。
沿106／110 READY；完整062 DRAFT保持。

- 既有S3自然勝利在第九戰術回合，沒有占領報告；不重新聲稱該項已完成。
- 使用出貨劇本正常公開載入，先做可丟棄玩家操作探針；第一戰首次實行前固定種子12345678一次。
- 不注入位置、艘數、士氣、功績、歸屬、勝負或鏡頭，不啟用被動戰術CPU，不換種子求通過。
- 找到可重播操作後，以正式bootstrap獨立重播、正常60 Hz真實GUI、完整保存／讀回驗證占領及換手。
- 原版已量測的callback與報告規則沿既有證據；新正常remake路徑不自動升格原版整回合同態。
- 保存所有失敗與診斷輸出，另建本輪收據，不改舊凍結文件或正式規則。

原Goal保持ACTIVE，MIDI與#89父層忠實度裁決仍待答；人耳、完整循環、macOS及正式交付保持未完成。

執行後追加：S1／S2正常艦隊戰鬥尚未產生110勝後占領。沿同一#79及019／078／106既有READY，改抽正常第⑥步行星攻擊與8-2換手。
S3同盟第六回合拿下雷奧米爾，已補獨立bootstrap、60 Hz真實GUI及公開存讀，見[驗證095](../playtest/095-battle-occupation-player.md)。
這條玩家鏈不代替110勝後占領；完整062與#79保持未完成。
