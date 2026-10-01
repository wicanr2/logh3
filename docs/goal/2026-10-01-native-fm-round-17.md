# 第十七輪 — 原生 FM 擷取與可重現 OGG

日期：2026-10-01。工作項目：[遠端 #69](https://github.com/wicanr2/logh3/issues/69)。
上一輪原生 GM 階段已提交；本輪開工 87 張 Issue、12 張 open，267 條斷言零失效。

核對目前 dosgolem 版本是否重生原有 VGM；固定 Ogg 封裝序號／CRC，
保持音樂封包不變並重生原有完整 OGG SHA-256，不放寬 SOURCES.tsv 閘門。
再接原生 FM 匯入，三條既有觸發路徑必須實跑及重現，未 READY 不進正式路徑。
來源與調查：[研究 359](../re/359-native-fm-import.md)。
固定 Ogg 契約：[規格 070](../spec/070-deterministic-ogg-framing.md)。
三平台第三方依賴封裝、實際聆聽／loop 仍是 #69 必要項目。
