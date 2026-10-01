# 第十三輪 — 正常回合完整執行期表格

2026-10-01；開工 HEAD `f180e571b6f44048ccef0fedc69ae9ba5571eb0b`。
本輪以遠端 [#85](https://github.com/wicanr2/logh3/issues/85) 為權威，不替換完整 remake Goal。
上一輪完成 #84，正常回合庫存與 GUI 已一致，完整表尚有 33 bytes 差異。

1. 追星域 `+0x78` word 的原始寫入與排序消費者，核對 AI 距離。
2. 核對行星 `+0x46／+0x48` word 的座標算法與載入時機。
3. 由 [研究 354](../re/354-post-turn-runtime-fields.md) 審查
   [規格 066](../spec/066-post-turn-runtime-fields.md)，達 READY 才補正式接線。
4. 正常編成／實行回合比較全 4032／6642 bytes，另驗存讀檔與 GUI。
5. 核對遠端清單、全產品套件、輸出擁有權及 Docker 清理，再提交推送。

知識入口：`reverse-engineer-retro-game-remake`／`use-ida-pro-9-4`；
證據規則與文件職責已載入。全部調查、測試在既有 Docker 工具鏈執行。
