# 070：正式音效完整尾端

2026-10-04。父提交1993f1a04b43dab7e9b5372dc6175d88d62dbab6。
沿[#48](https://github.com/wicanr2/logh3/issues/48)、[#69](https://github.com/wicanr2/logh3/issues/69)、[#79](https://github.com/wicanr2/logh3/issues/79)。
[規格105](../spec/105-sound-completion.md)補104既有完成等待；
[工作快照](../goal/2026-10-04-sound-tail-round-81.md)與[索引生成器](../../tools/scripts/sound_tail_receipt.py)保存重生方法。

## 修正與證據

正式app在Game.Tick前回報音效播放器Busy。名義樣本時長歸零但後端尚未完成時，
後一音效及圖段首格共同再等一個演出tick，之後重查。
既有名義下限、音效選擇、圖段次序、滑鼠排隊與保存格式保持。

修正前帝國第5次SFX21及第13次SFX08分別遺失81／974個取樣，約1.84／22.09毫秒。
Close前Oto緩衝量648／7792 bytes，恰與缺少的float32雙聲道PCM相同。
來源已讀到EOF不表示緩衝已送出；Position只用來記錄，不能代替波形證據。
固定版依據見105的Ebitengine／Oto來源索引。

| 正常關配樂GUI | 音效次數 | 完整PCM相同 | 完整來源保留 | 關閉前尚有緩衝 |
|---|---:|---:|---:|---:|
| 修正前帝國 | 14 | 12 | 12 | 2 |
| 修正前同盟 | 10 | 9 | 10 | 0 |
| 修正後帝國 | 14 | 14 | 14 | 0 |
| 修正後同盟 | 10 | 10 | 10 | 0 |

同盟修正前一段插入304個靜音取樣，完整尾端仍保留。
單獨播放28段另有兩段插入204／326個靜音取樣；26段全PCM相同，全28段來源保留。
這三筆全段比較仍FAIL，插入靜音成因unknown，不能由尾端保留豁免。
固定原始Game的961.184毫秒請求間隔局部替換19→4沒有截短，沒有調時點製造反例。

## 正常玩家路徑

兩陣營皆由正式cmd/gin3的預設片頭、主選單、138、S3、系統指令、設定畫面到達。
在公開設定選「關配樂」再返回命令，音效保持啟用。
熱區取正式commandBar及設定規格，不另複製座標。實際操作皆真實X11鍵鼠。
隨後照既有正常艦隊下令、戰術實行、交戰命令及演出返回。
每場實行前固定remake seed12345678，不注入位置、鏡頭、命令或原池。

每案三張完整X11視窗與獨立Game計畫相同，六份診斷SaveSlot相同。
與第77輪相比，只有WORK+0x1F為公開設定的MusicOff2，其餘完整bytes相同。
保存用診斷API，不能稱玩家存檔選單通過。

模型一次RightClick只捲一格，實體保持可跨數個Update。
沒有命令或快照的連續模型捲動合併為公開手勢；下一操作前仍精確核鏡頭。
原模型索引、額外手勢、時間與每Update狀態全部記錄，不宣稱相同實體時序。
帝國95個合併動作、同盟73個；正式程式的速度及右鍵規則未改。

前四批鏡頭失敗保留。舊控制器對任何重訪位置反向，會把單軸逼近目標的視野送回邊界。
修正為單軸持續朝目標，兩軸重訪時走誤差較大的一軸，仍保留40次上限。
常駐XTEST工具在正式Update確認按下後立即放開，並核X server已處理送出事件。
沒有增加重試上限、降低像素門檻或改遊戲去遷就操作腳本。

## 固定工具與重生

GUI3為`sha256:ecb3612ac7c18d0cd225fd55b7b186052a8bf0029e82aa786a23f4b88f196fa8`。
Ebitengine2.8.8、Oto3.3.3、PulseAudio16.1、ALSA default、44100 Hz雙聲道float32有時鐘虛擬輸出。
聲音由真正正式播放器產生；monitor與正式解碼signed16除32768轉float32比整段，不設容差或遮罩。
28段、22首曲庫及所有生成OGG皆既有真實來源，沒有合成測試音源。

before來源綁父提交1993及凍結overlay，不能拿目前修正後程式重生修正前結果。
局部探針入口為`tools/scripts/sound_tail_probe_prepare.py baseline`；
正常GUI來源為`tools/scripts/sound_tail_gui_prepare.py <新revision> off|mixed`。
兩者先在`tools/gui.sh xrun`內生成，再用同一容器工具鏈建置overlay。
原版只讀掛載，輸出必須為新批次，播放器wrapper拒絕覆寫。

```sh
bash tools/playback.sh <新批次> python3 workplace/round80-driver-<revision>.py \
  --plan workplace/round81-before-plan-0/plan.json \
  --out /src/workplace/<新批次>/gui --scale 1
bash tools/gui.sh xrun python3 tools/scripts/sound_tail_compare.py \
  workplace/<新批次> --sources workplace/round81-pcm-complete-v1 --gui
bash tools/gui.sh xrun python3 tools/scripts/sound_tail_gaps.py workplace/<新批次>
```

正式重生命令、兩側輸出、實際裝置連線、源碼及二進位身份由完整播放索引保存。
批次名含before／after與side；mixed用既有round80-plan-0，保留原配樂設定。
原版dosgolem仍f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c／clean，未改執行器或重生原版收據。

## 驗收範圍

新增後端未完成、名義時間已到的首格／事件／資料保持與停用回歸通過。
既有四案×四模式原版音效編號、戰損及seed回歸無skip通過。
標準18套件13通過、5無測試。

正常remake GUI及正式後端完成回報已驗；不升格原版整段same-state、實體音效卡或人耳。
#89一像素裁決仍pending，未改順位或加遮罩。完整056／092／062仍DRAFT，104維持READY。
105已限定CONFORMED。配樂共播樣本完成14次音效及五次換曲，每次音效Close前IsPlaying為false、buffer為零。
修正前後及共播共15張原生全圖各307200像素零差，最大通道差0、bbox null。
修正後關配樂帝國69.463秒／同盟50.935秒；共播56.890秒、峰值0.438782。
三個monitor皆無NaN／Inf或絕對值大於等於1的樣本。
修正後帝國95動作／85額外鏡頭手勢，同盟73／77；共播91／83。
所有完整輸入、差分、狀態、PCM及來源身份另存本輪播放索引。
完整[播放索引](../internals/data/sound-tail.json)共521份來源，SHA-256為
`a5541a1b654e2c5a218ef3aa1394e177f956dad62749973299075e09520de8e7`。

共播monitor保持原音量轉為signed16 WAV，供實際聆聽：
`workplace/round81-after-gui-mixed-0/listen.wav`，雙聲道44100 Hz，56.890秒。
WAV為正式虛擬輸出的格式轉換，來源monitor SHA-256為
`abf183935fd456815fbf97f98f8e5ade24894a3ec9f1e7d628f2e8e416e81022`。
實體音效卡、人耳、循環接點、Windows／macOS GUI仍待驗，不由此升格完整發行。
