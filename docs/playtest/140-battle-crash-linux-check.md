# Windows交火閃退：Linux實包複驗

2026-10-30，第169輪。使用者確認閃退發生在Windows，要求先以Linux驗證。使用者實際版號、海報設定與觸發存檔尚未確認。本輪測現行`v.1.6.3-20261026`，來源main `5918a5c0177bd135071ddcb914905bbac1e0f103`，正式遊戲程式與素材沒有修改。

## 結果與限制

| 測試 | 實際結果 | 限定範圍 |
|---|---|---|
| Linux AppImage、HD、v1海報、音效開啟 | 正常讀檔進戰場，完成四回合，最後第5回合待命且子程序仍存活 | 69個不同交火識別，含行星攻擊及艦隊三種武器；未重現閃退 |
| Windows正式EXE、Wine 9.0、同設定但音效關閉 | trace也到第5回合待命，69個交火識別；驅動器900秒逾時，未完成最後截圖 | 逾時時子程序仍存活；不可記為完整劇本通過或Windows真機驗收 |
| Windows診斷啟動檔 | 兩份以實際EXE的`-version`通過，保留版本與`ExitCode=0` | 只驗證輸出收集，尚未取得使用者的閃退堆疊 |

Linux的實際播放音效識別為0、1、2、3、5、6、7、8、17、21、22、23、24、25。最後畫面已目視，正常回到戰場，海報與上下旗艦仍在。另一次無海報測試只走到行星攻擊，不能拿它當艦隊交火證據。

Linux完成判定時仍存活，最終日誌含ALSA I/O錯誤，但沒有該行時間戳。不能確定它發生在收尾哪個時點，也不能據此歸因Windows閃退。軟體繪圖、虛擬音訊裝置、Go記憶體與GC參數的限制見收據；沒有驗證實機顯示卡、AppImage FUSE或Windows音訊。

**根因仍為unknown，尚無產品修正。** 沒有修改版號、正式封包、tag或Release。下一步需要使用者的閃退日誌及相同存檔，再重播同一條玩家路徑。

## 重播與證據

完整輸入、實際命令、工具版本、封包／驅動器／trace／截圖SHA-256保存於[機器收據](../internals/data/battle-crash169.json)。正常路徑是主選單138答案、讀取隨包S1存檔、選同盟、執行戰略指令、進戰場並操作四回合。起始戰略回合93，日期796年10月10日；不是新建艦隊後的首次出擊。沒有注入狀態、種子或時間，也沒有使用script或direct-entry。

本機收據目錄：

- `workplace/battle169/linux-poster-f1-before/`：`RESULT.json`、`EVENTS.json`、`FRAMES.jsonl`、8張截圖及音訊日誌。
- `workplace/music69-windows/battle169/windows-before/`：相同格式，7張截圖，測試逾時保留失敗狀態。
- `workplace/music69-windows/battle169/cmd-test/`：兩份啟動檔版本測試日誌。

沿[圖形工具Dockerfile](../../tools/Dockerfile.gui-debug)建立`logh3-ebiten:20261030-r169`，基底為既有`dosgolem-go-ebiten:1.24-r1`，替代本機已不存在的舊GUI工具映像。由[gui.sh](../../tools/gui.sh)與[有界Xvfb](../../tools/with-xvfb.sh)執行[正常操作驅動器](../../tools/scripts/battle_crash_gui.py)。`--out`必須是尚不存在的目錄，收據內`plan`可匯出為JSON重播；所有工作在Docker內、UID/GID1000、network none執行。

```sh
LOGH3_GUI_IMAGE=logh3-ebiten:20261030-r169 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh tools/gui.sh xrun \
  python3 tools/scripts/battle_crash_gui.py \
  --package /src/workplace/poster166/extracted/squashfs-root/usr/bin \
  --binary /src/dist-all/v.1.6.3-20261026/full/LOGH3-v.1.6.3-20261026-linux-x86_64.AppImage \
  --plan /src/workplace/battle169/fleet-f1-plan.json \
  --out /src/workplace/battle169/linux-replay-new \
  --poster v1 --graphics hd --locale zh-TW --audio on
```

## Windows保留閃退訊息

把符合原啟動模式的[原版AI診斷啟動檔](../../tools/capture-battle-original.cmd)或[增強AI診斷啟動檔](../../tools/capture-battle-enhanced.cmd)放在`gin3.exe`旁執行，照原操作進入交火。兩份分別沿原版與增強啟動參數，不替換EXE。日誌寫在`%TEMP%\logh3-battle-*.log`，非零退出時會顯示日誌並暫停。日誌與交戰前存檔是下一輪重現輸入。

本輪不宣稱原版同態，也不關閉#48正式戰鬥讀回重建。狀態入口為[第169輪](../goal/2026-10-30-battle-crash-round-169.md)。
