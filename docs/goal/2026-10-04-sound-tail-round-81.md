# 第81輪：音效尾端與正式完成等待

2026-10-04，父提交1993f1a04b43dab7e9b5372dc6175d88d62dbab6。
沿open [#48](https://github.com/wicanr2/logh3/issues/48)、[#69](https://github.com/wicanr2/logh3/issues/69)及[#79](https://github.com/wicanr2/logh3/issues/79)。
開工89張worklist／9張open，269條核實零失效。上一輪有實際進展，正式Linux音訊已輸出。

先以固定Ebitengine／Oto及同一有時鐘PulseAudio虛擬sink核播放器buffer與完整PCM。
28段單獨正式播放建立完整輸出基準；正常玩家經設定畫面關配樂，分離實際音效。
查第80輪四次IsPlaying=true時的替換是否真的丟樣本，不用Position估計當波形證據。
已證實缺陷才依104的「前樣本完成才開始後樣本及首格」契約修正；缺口先補規格審查。
保留原版硬體規格近似、正式輸入排隊、RND及保存格式，不深入DOS硬體driver。
正常兩陣營、混音／關配樂、正式完成等待與保存都需最小充分驗證。
不把虛擬sink、解碼或技術波形等同實體音效卡、人耳或原版逐波形。
#89排序裁決仍pending；不改順位、不加遮罩、不重問。完整remake與發行Goal保持。
