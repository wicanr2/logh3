# 第148輪：清理可重建的工作區副本

使用者要求先清理workspace以釋放硬碟空間。本輪只處理本專案`workplace/`；正式`dist-all/`、原始資料、來源、獨有音畫與研究證據保留。

## 目前狀態表

| 範圍 | 狀態 | 邊界 |
|---|---|---|
| 淨釋放 | 17131577344 bytes，約15.96GiB | workplace由66.17降至50.21GiB；不動其他專案或Docker映像 |
| Go／Wine快取 | 18個Go快取、5個Wine環境已清理 | Go模組下載保留；Wine無玩家資料，先備份registry及使用者分支978747 bytes |
| 正式交付的工作副本 | 14個解壓副本、10個common、27個壓縮副本已清理 | 逐檔與正式ZIP核對，保留還原對照；無相同正式包的49個候選保留 |
| 證據與原始資料 | 保留 | 原版oracle、decomp、音畫來源、截圖、收據、凍結source及未提交修改 |

[清理及還原工具](../../tools/scripts/workspace_cleanup.py)已產生`workplace/cleanup148/PLAN.json`，逐筆記錄原路徑、檔案雜湊、來源ZIP及還原方法，再依核對結果刪除。[清理收據](../internals/data/workspace-cleanup148.json)保存74項結果、原路徑及還原抽驗。計畫與刪除之間內容有變便停止，不憑檔名清理。

清理後，舊驗收中的安裝副本或common可能需要先從正式ZIP還原，才能重跑原驗證器；歷史收據記錄的是原驗收當下，不改寫其雜湊。Go快取與Wine環境由既有Docker工具鏈重建。正式ZIP、影片及校驗不變。

## 驗證及還原

清理前確認本專案沒有執行中容器，目標不含Git追蹤檔，逐筆檢查UID／GID與內容指紋；原版、decomp、模組下載、來源及不相同歷史包跳過。清理後74項目標均不存在，v1.5.2根SHA256SUMS全部通過，正式三平台ZIP、修正包、影片及manifest保持。

抽驗ID4壓縮副本及ID29解壓副本：從保留來源還原，逐檔雜湊通過後再次移除抽驗副本。ID及完整來源對照見`workplace/cleanup148/PLAN.json`。還原命令如下，數字改為所需項目的ID；這個範例為已抽驗的舊Windows引擎副本。

```sh
tools/gui.sh xrun python3 tools/scripts/workspace_cleanup.py restore --id 29
```

Go快取下次建置自動重生。Wine測試環境可用既有包裝器及映像重建，例如：

```sh
LOGH3_WINE_IMAGE=eob-remake-wine-verify:ubuntu-noble-20261008-r2 \
bash tools/player_wine.sh env WINEPREFIX=/out/full147/gui-wine-prefix wineboot -u
```

本次未重跑產品測試，避免立即重建剛清除的建置快取。產品來源沒有改動，開工282條worklist核對零失效。Python語法、差異、輸出擁有權及容器清理通過；未commit、push或發布。
