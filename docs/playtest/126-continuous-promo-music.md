# 126：推廣片連續配樂版

依使用者2026-10-14要求，重新製作約125秒推廣片音軌，讓背景音樂不因剪接而換曲或重啟。沿用既有開場BGM_0F，交火音效獨立混音。影片位於本機`dist-all/v.1.2.0-20261012/promo/LOGH3-v.1.2.0-20261012-promo-continuous-music.mp4`。
本輪狀態見[計畫137](../goal/2026-10-14-continuous-promo-music-round-137.md)，全部來源與輸出雜湊集中於[唯一索引](../internals/data/promo-continuous-music.json)。

## 原因與修正

舊合成器[hero_story_promo.py](../../tools/scripts/hero_story_promo.py)在每個影片段落保留來源原聲，再混入另一段BGM_0F；該配樂又依`start_at % 70`重新定位。因此各段遊戲或動畫原聲中的背景音樂會跟著切換。
新版從原片保留完整視訊串流，重新製作整條聲音。BGM_0F由0秒連續播放，長127.134853秒，足夠覆蓋全片，沒有循環、分段定位、剪接點淡出或切曲。整片只在片頭1秒淡入、片尾2.5秒淡出，增益固定。

| 素材 | 身份與處理 |
|---|---|
| 原片 | SHA-256 `1974f95bf87b05f1d8059d21e3a04605f840f67077c157a11798cbcf6b062bea` |
| 單首配樂 | 同版遊戲曲庫BGM_0F.ogg，SHA-256 `48c22ed0b5d9fec5d0b9bed70e7a06ae1ab3a07faef99da80ddfd21b4644739c` |
| 遊戲音效 | 同版SFX素材，依實錄中第一次觀測到`played`的時間重混；71個剪輯音效片段，來源及取樣區間完整記錄 |
| 影片來源原聲 | 不混入，避免帶回其他背景音樂；動畫原聲亦不混入 |
| 新片 | SHA-256 `cb1557f811366cd1cae06d38f7a000db422b5bba8964db94e1789fd9b03cbfe0` |

遊戲音效的開始時間精度受frame trace取樣影響，不宣稱與原錄音逐取樣相同。既有39段畫面、完整交火、字幕與片尾保持，沒有重新錄製遊戲操作，也沒有以本片新增遊戲功能驗收聲明。其畫面來源及原版／強化比較範圍沿[119](119-hero-ai-release-promo.md)。

## 驗收結果

- 全片124.956315秒，1920×1080、60fps、AAC雙聲道44.1kHz。
- 全7496個視訊封包內容及PTS／DTS／時長／大小與原片相同，沒有重編碼畫面或移動剪接點。
- 全音樂底軌與原曲連續取樣核對，38個剪接點均無重啟或切曲。AAC解碼混音與原PCM全片訊噪比33.86dB，剪接點最低27.26dB，高於預定25dB全片／20dB局部門檻。
- 實際成片平均音量-19.9dB、峰值-3.8dB，沒有削波、長靜音或黑段，片尾完整。靜止標題、選單與結果等待沿既有通過分類；全視訊與時間相同可證分類仍適用。
- 實際抽取8格總覽已目視，字幕及片尾未裁切。本輪技術驗收不宣稱新混音已通過使用者人耳驗收。

首次封裝使用MP4預設毫秒時基，造成全視訊起點固定少15/15360秒；驗收拒絕。改用視訊與音訊取樣率的最小公倍數11289600作movie timescale，同一輸入重跑後全部封包時間完全相同，未降低比較門檻。失敗輸出保留於`workplace/promo-music137/`，正式批次為`workplace/promo-music137-b/`。
舊TRACE.json在程序關閉前記錄的雜湊與最終frame trace不同，已依原有FRAME-TRACE-CLOSE.json核對，全部來源一致；沒有把正常收尾差異列作遊戲缺陷。

## 重生與交付

[製作器](../../tools/scripts/promo_continuous_music.py)在既有Docker音訊工具鏈執行。來源唯讀，工作輸出只寫workplace，工具拒絕覆寫。更換批次名稱即可另存重生結果。

```sh
LOGH3_AUDIO_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_MEM=3g \
bash tools/audio.sh python3 /src/tools/scripts/promo_continuous_music.py render --out promo-music137-b

LOGH3_AUDIO_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_MEM=3g \
bash tools/audio.sh python3 /src/tools/scripts/promo_continuous_music.py audit --out promo-music137-b
```

目視CONTACT後，使用既有可寫工作樹容器交付：

```sh
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh LOGH3_CPUS=2 LOGH3_MEM=3g \
tools/gui.sh xrun python3 /src/tools/scripts/promo_continuous_music.py publish \
  --out promo-music137-b --visual-reviewed
```

新版另存同版音軌修訂檔，不覆寫原片、來源索引、已發布tag、Release或遊戲封包。交付同目錄的`continuous-music-`前綴檔保存聲畫檢查、PLAN、來源、目視總覽與SHA-256清單。含原版與衍生音畫，只留本機，不提交影片或上傳。
