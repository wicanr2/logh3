# 065：艦隊交戰正式玩家鏈

日期：2026-10-04，第76輪。沿[#48](https://github.com/wicanr2/logh3/issues/48)與
[#79](https://github.com/wicanr2/logh3/issues/79)，[102](../spec/102-fleet-player-chain.md)先READY再實作。
父提交`02328c82b8fb27e21ac4a9e89b1aa87205c560b6`，工作樹修改的精確來源以索引雜湊為準。

## 來源與比較等級

| 收據 | SHA-256 | 範圍 |
|---|---|---|
| [原版名冊及UI](../internals/data/fleet-player-chain-native.json) | `f569a16f863f9e54d3cda5fe0a8fe2c39566ec74864469d05b374fa29017c7eb` | 五種有限排列，103份來源 |
| [完整Game索引](../internals/data/fleet-player-chain-game.json) | `4e980f1c150fcfb76f22f015cb1c0dccd7654b7da370f5f1d0fb361b19eca2b1` | 141份來源、雙跑、程式、完整命令及seed設定 |
| [原版場景](../internals/data/fleet-scene.json) | `7d7e5b388eebc8348825e91352ff10e9821ce2b0cf7912d6f9b9cf1cafc95379` | 第69輪四案，268檔雙跑相同；本輪沿用，不覆寫 |

原版bundle沿AGENTS.md的`de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
dosgolem為`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，乾淨且未修改。
Game使用`logh3-ebiten:2`；場景診斷從固定原版原表恢復，比較等級事前為same-state，僅限指定場景幀及資料。
正常操作使用remake公開輸入，不注入位置／命令／RAM；每場執行前固定測試seed12345678。
正常操作沒有原版整段同態收據，不宣稱正常全流程parity。

## 已通過

- 第③步完成後接命令、順位、093核心及099完整場景，結束後續第⑤步。正式入口不再呼叫舊行星對艦隊近似。
- 五種原版名冊順位相同；手動光線砲、飛彈、戰鬥機、防禦、自動及取消接原池，命令中拒絕保存。
- 光線砲、飛彈、同目標兩擊、死亡四案：艘數、庫存、修理上限、seed、功績、兩側分數與完整度符合原版。死亡後第二擊不扣庫存。
- 四案演出完成後續回Input，六份Slot保存重讀再寫相同；WORK比較正式導出值，未拿未導出的raw WORK作保存缺陷。
- 光線砲與飛彈的傷害前後四張完整640×480圖，各307200像素零差。每張八區的相同數、差異數、最大通道差與bounding box齊全，無排除區。
- Game最終雙跑`round76-chain-final-d/e`的36檔相同。正式bootstrap兩陣營各雙跑11檔相同，包含七份保存檔、場景圖、終態图與重播metadata。
- 正常帝國下令10次、交戰7擊，在第二場進入射程；正常同盟下令7次、交戰5擊，在首場完成。兩側CPU位置保持不動，六份保存與bootstrap重播逐位元組相同。
- 標準`tools/gui.sh test ./cmd/... ./internal/... -count=1`：13套件PASS、5套件無測試，零失敗。

原版／remake／差分PNG及逐區JSON留在`workplace/round76-chain-final-d/e`，
全部雜湊由Game索引保存。正常bootstrap輸出在`workplace/round76-bootstrap-third/fourth-0/1`，
實際目錄名與全部輸入／相對時間戳以索引為準。兩種比較不混用。
完整度95／94根因與原始bytes見[研究413](../re/413-fleet-player-chain.md)。

## 重播與限制

[凍結器](../../tools/scripts/fleet_player_chain_receipt.py)提供`--ui`、`--replay-overlay`與`--game`。
`--replay-overlay`只在workplace新增固定seed setter，正式產品來源不改。
[正常重播器](../../tools/scripts/fleet_player_chain_replay.go)從bootstrap與主選單開始，
讀Game收據的公開輸入；不匯入原版RAM。完整命令保存在Game索引。

正常測試早期因鏡頭目的地、星雲路線、右側清單誤點及Xvfb未就緒失敗，輸出保留。
這些失敗沒有改九回合、CPU被動或射程規則。測試玩家路線以既有通行資料規劃。
舊「未下戰鬥命令仍自動扣艘數」測試已換為無命令保持；戰損改用四種原版場景及正常下令驗。
強化戰術回歸比較實際位置與目的地；戰術傷害優勢未驗，戰略增強AI既有驗收保持。

102仍READY，完整056／092／062仍DRAFT。尚缺正常原版整段同態、交戰命令亮標與游標全圖、
真實GUI鍵鼠、音效接線與人耳驗收。本輪不關#48／#79，不外推macOS或完整發行驗收。
