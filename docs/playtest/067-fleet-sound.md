# 067：交戰音效接線

2026-10-04，第78輪。沿[#48](https://github.com/wicanr2/logh3/issues/48)、
[#79](https://github.com/wicanr2/logh3/issues/79)、[104 READY](../spec/104-fleet-sound.md)及
[研究415](../re/415-fleet-sound.md)。父提交e927750a782af344d77e6a3dd866e18af753ad28。

## 原版來源及比較範圍

GIN3SP.VOC是MicroWaveSVF，前16 bytes為`MicroWaveSVFv1.0`，不是Creative Voice。
28段各含word長度、byte時間常數、byte DSP命令及payload。
本版時間常數166、命令14。原版載入N bytes，交DSP的長度為N-2。
依Creative規格導出播放前N-1樣本、每樣本90微秒，標hardware-spec approximation。
只追遊戲參數、編號及完成等待，不深入ISR、DMA或實機時鐘。

四案音效編號直接取既有dosgolem完整COMANI記錄，不以remake自行生成期望值。
原版來源及原池／seed比較沿第69輪固定收據；執行器
f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c未修改且乾淨。
此輪沒有新增原版PCM播放收據，不能稱逐波形或實機耗時一致。

## 驗證結果

- 28段payload及EOF核完整原檔雜湊。截斷、尾端多餘bytes、不支援命令明示拒絕。
- 四案×三速度／動畫關閉共16項，編號與原版記錄相符；不提前重播，原池戰損／庫存及seed保持。
- 先前音效未完成時不播下一段、不畫攻擊首格、不提前戰損或取RND；完成後續接原版編號8。
- 零損失守方不播另做局部診斷，明示注入目標0艘；原版死亡第二命令跳過由四案回歸核對。
- 28段轉44100 Hz立體聲OGG，固定序列號、-1 dB增益，時長、峰值及非靜音通過。獨立雙跑OGG與有效PCM雜湊相同。
- 音效13的FFmpeg尾端補樣本按OGG有效結束位置裁切後核時長，5毫秒門檻保持。正式播放器全28段完整解碼，長度與有效結束位置相同。
- 正常帝國10次下令／7擊／14個音效，同盟7次下令／5擊／10個音效。兩側固定測試seed12345678，無位置、命令或RAM注入。
- 正常Game雙跑產物相同；正式bootstrap與app錄製器收到相同音效與相對時間。六份保存與正常Game及第77輪無音效保存逐位元組相同，CPU位置保持061政策。

## 重現及使用

原版須由玩家自備。輸出批次名必須尚未存在，轉檔器拒絕覆寫。

```sh
bash tools/sound_to_ogg.sh my-sfx
tools/gui.sh run ./cmd/gin3 -orig /orig -music /src/workplace/my-sfx
```

`-music`指定的目錄同時接受配樂及`SFX_00.ogg`至`SFX_27.ogg`。
要同時播放配樂，將合法匯入的曲庫與這28檔放同一個素材目錄。
正式啟動先完整解碼28段，只有素材及裝置都可用才啟用音效。
配樂關閉不影響音效；無裝置、缺檔或錯誤留下明確訊息，Game不等待無法播放的樣本。
音效不循環，播放完成後在下一次更新釋放播放器，與配樂換曲互不干擾。

```sh
tools/gui.sh xrun python3 tools/scripts/fleet_sound_game_receipt.py --overlay
tools/gui.sh xrun env LOGH3_FLEET_SOUND=1 LOGH3_FLEET_CHAIN_OUTPUT=/src/workplace/round78-chain-a go test ./internal/game -run '^TestNormalFleetPlayerChain$' -count=1 -v
tools/gui.sh xrun env LOGH3_SOUND_APP=1 go test -overlay workplace/round78-sound-replay.json ./cmd/gin3 -run '^TestNormalSoundApp$' -count=1 -v
tools/gui.sh test ./cmd/... ./internal/... -count=1
```

重跑另用新批次名，保留歷史輸出。overlay只固定RND，不改位置、命令或正常玩家入口。
首格與零損失診斷在`TestFleetSoundFirstFrameWait`／`TestFleetSoundZeroLoss`。
正式解碼在`TestSoundEffectPlayerDecode`，原版序列在`TestFleetSoundNative`。

[原始來源](../internals/data/fleet-sound-source.json) SHA-256為
`5d3b31b8d16c4fc34b655b52a9caadfa40cdbecbf2d87f21bf1942ec75db2d66`；
[兩側表](../internals/data/fleet-sound-spec.json)為
`e4c40427f3edbb6b245b1cf4150fe5faf66476feb981eaa806b7150319fb5a50`。
[Game完整索引](../internals/data/fleet-sound-game.json)由
[凍結器](../../tools/scripts/fleet_sound_game_receipt.py)記錄原始命令、各輸入與產物雜湊。
SHA-256為`4388203c2d6d9c69f69c38f1107b871b92c1397923f38a9eb2b7fe72ed39b9e3`，164份來源。
標準18套件13通過、5套件無測試；原版序列、正常路徑與正式解碼的目標測試均無跳過。

104仍READY，完整056／092／062仍DRAFT。裝置、人耳、真實X11交戰及正常原版整段同態仍缺。
#89剩餘一像素未核准，排序裁決pending；不由本輪聲音接線升格完整對拍。
#48／#79保持OPEN，原版音效及衍生OGG只留本機或獲授權私有位置，不放公開發行包。
