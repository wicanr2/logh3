# 069：正式音訊後端與正常交戰

2026-10-04，第80輪。沿[#48](https://github.com/wicanr2/logh3/issues/48)、
[#69](https://github.com/wicanr2/logh3/issues/69)、[#79](https://github.com/wicanr2/logh3/issues/79)及
[104 READY](../spec/104-fleet-sound.md)。父提交134dcae343fd08b6931b78c63957f34433673a37。

## 裝置誤判與修正

主機及容器都沒有`/dev/snd`。固定PulseAudio虛擬輸出下，ALSA探針成功開啟default PCM。
未修改的正式遊戲仍停用配樂及全28段音效，2.18秒monitor的192276個樣本全為零。
根因是[修正前music.go](https://github.com/wicanr2/logh3/blob/134dcae343fd08b6931b78c63957f34433673a37/cmd/gin3/music.go)僅憑實體目錄判斷Linux裝置。
修正前程式與執行檔另存，完整身份見JSON，不覆寫歷史來源。

[Linux判定](../../cmd/gin3/audio_device_linux.go)先沿用實體裝置檢查。
存在已配置PulseAudio端點或標準使用者socket時，實際試開及關閉ALSA default。
只有成功才啟用；其他平台維持原本後端。main只做一次判定。
沒有裝置及失效PulseAudio default的兩個反例均通過，沒有用空目錄冒充`/dev/snd`。

同一入口、素材及第30次更新，修正前後PNG全檔相同。
修正後正式Ebitengine／Oto輸出有18256個非零樣本。
短片頭樣本仍接近前奏靜音，另錄600次更新，14.119秒、913477個非零樣本，峰值0.942322。
沒有注入遊戲狀態或測試用音源。

## 正常X11與實際輸出

正常路徑仍由預設片頭、主選單、138、S3、實行、命令、交戰及返回。
真實X11鍵鼠沿第79輪；只讀overlay記正式播放器呼叫後的狀態。
固定測試seed12345678，沒有位置、命令或鏡頭注入，沒有`-script`。
素材為既有Linux22首配樂與28段音效，組合前後逐檔雜湊相同。

| 驗證 | 帝國 | 同盟 |
|---|---:|---:|
| 計畫公開動作 | 183 | 159 |
| 額外鏡頭回授手勢 | 67 | 244 |
| 正式音效開始／停止 | 14／14 | 10／10 |
| 完整X11畫面 | 3張零差 | 3張零差 |
| 每張原生像素 | 307200 | 307200 |
| 診斷保存 | 六檔相同 | 六檔相同 |
| monitor時長 | 67.114秒 | 71.710秒 |
| monitor峰值 | 0.456879 | 0.390381 |

兩側的實際配樂開始序列為OPEN_M、BGM_0F、BGM_3F、BGM_6F、BGM_3F。
音效編號與第78輪正常Game／app錄製器相同。
PulseAudio實際連線標示正式執行檔及ALSA Playback，44100 Hz、float32、立體聲、未靜音。
monitor全樣本皆有限數，沒有絕對值大於等於1的樣本。
這證明正式後端有輸出，不能單靠混音monitor證明每段音效的尾端完整。

視窗原生全圖、Game色號及獨立公開輸入計畫相同；六檔保存亦與第77輪相同。
保存由診斷SaveSlot API，不稱玩家存檔選單驗收。
額外鏡頭手勢與每Update均記錄，仍只主張等價鏡頭，不主張相同實體時序。
本輪沒有新增原版收據，dosgolem仍f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c且乾淨。

## 音效收尾限制

帝國一次、同盟三次換音效時，舊播放器仍回報`IsPlaying=true`。
其`Position`距解碼樣本名義末端約21.962、17.302、17.302及4.286毫秒。
播放器位置與裝置buffer有關，目前未以分離PCM證實monitor是否少了這些樣本。
將此記為已證實的播放器狀態與unknown的實際尾端，不宣稱音效完整收尾或聽感通過。
下一窄任務應核對正式後端的音效尾端與完成等待，不追DOS硬體driver。

實體音效卡、Windows／macOS GUI、真人聆聽及循環接點仍未驗。
兩側配樂同為BGM_6F，只驗當前預設選擇；BGM_7仍沿既有選曲測試，未新增實際輸出聲明。
配樂關閉與音效獨立的實際GUI抽樣也仍缺，不能由這輪混音推導。
104／102／103保持READY，完整056／092／062仍DRAFT。
#89一像素及排序裁決仍pending，#48／#69／#79保持OPEN。

## 來源及重現

[完整索引](../internals/data/fleet-playback.json)保存293份來源，SHA-256
`f3a37d50fc86d2485113a2e20f84e6c2c2c07e6b12e6697d30d5a2c3341fb209`。
包含正式修正前後程式、執行檔、素材雜湊、映像、完整套件清單、輸入與每Update、PNG／差分／統計、PCM及播放器事件。
原版衍生音訊、二進位及收據留本機，沒有放進Git或公開發行包。

GUI2映像6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8保持。
新增GUI3為ecb3612ac7c18d0cd225fd55b7b186052a8bf0029e82aa786a23f4b88f196fa8，
來源為[Dockerfile.gui-playback](../../tools/Dockerfile.gui-playback)。
固定PulseAudio及utils 16.1+dfsg1-2+b1、ALSA plugins 1.2.7.1-1；實際所有相依版本另存清單。
虛擬sink有時鐘，monitor錄正式ALSA輸出；unix socket僅存在於network-none隔離容器。
套件語意參考[PulseAudio 16.1 null sink原始碼](https://raw.githubusercontent.com/pulseaudio/pulseaudio/v16.1/src/modules/module-null-sink.c)，
正式後端依鎖定[Oto 3.3.3 ALSA實作](https://raw.githubusercontent.com/ebitengine/oto/v3.3.3/driver_unix.go)。
未複製第三方實作到產品。

```sh
timeout 30m docker build --network default --memory 2g --cpu-quota 200000 \
  -t logh3-ebiten:3 -f tools/Dockerfile.gui-playback tools
bash tools/playback.sh my-device env LOGH3_AUDIO_DEVICE_TEST=pulse \
  go test ./cmd/gin3 -run '^TestAudioDeviceActualDefault$' -count=1 -v
```

批次名須尚未存在。曲庫來源沿合法原版匯入；本輪本機組合在workplace/round80-library。
新觀測版及計畫由[prepare](../../tools/scripts/fleet_playback_prepare.py)產生：

```sh
bash tools/gui.sh xrun python3 tools/scripts/fleet_playback_prepare.py myrev
bash tools/gui.sh build -overlay workplace/round79-fleet-gui-overlay-myrev.json \
  -o workplace/round80-gin3-myrev ./cmd/gin3
bash tools/gui.sh xrun go run -overlay workplace/round79-fleet-gui-overlay-myrev.json \
  workplace/round80-plan-source-myrev.go \
  -input workplace/round78-chain-a/normal-0/inputs.json -out workplace/my-plan
bash tools/playback.sh my-gui python3 workplace/round80-driver-myrev.py \
  --plan workplace/my-plan/plan.json --out /src/workplace/my-gui/gui --scale 1
```

同盟改用normal-1；原生全圖沿第79輪native checker，PCM統計用
[playback_stats.py](../../tools/scripts/playback_stats.py)。
既有歷史素材或二進位若不存在，先依067／068重建，不用合成檔替代。
[帝國完整實際錄音](../../workplace/round80-gui-0/listen.wav)為monitor轉signed16 WAV，保留原音量及全時長，供人耳驗收。
來源雜湊見完整索引；真人未回報前仍維持未驗。

標準18套件13通過、5無測試。三個實際ALSA目標測試均通過、零跳過。
同盟首次GUI失敗是第40次鏡頭手勢已抵達目標後，迴圈沒有最後一次成功檢查。
只修本輪驗證工具的判斷順序，保留舊批次及每Update，同一產品乾淨重跑通過。
