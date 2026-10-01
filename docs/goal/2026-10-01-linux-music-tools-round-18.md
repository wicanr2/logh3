# 第十八輪 — Linux 隨包音樂工具

日期：2026-10-01。已開工作項目：[遠端 #69](https://github.com/wicanr2/logh3/issues/69)。
前輪完整匯入已提交；本輪 87 張 Issue、12 張 open，267 條斷言零失效。

沿用已確認的工具與音色庫，封裝 Linux 原生載入器與完整依賴。
先以可丟棄探針證明乾淨容器能解析所有依賴，再審查 READY 規格，實作原生啟動入口與
可重現建置器。完成後在沒有系統音樂工具的容器抽驗 FM／GM 實際匯入與既有基準。
完整 Windows／macOS 依賴封裝、對應來源碼交付與聆聽仍按 #69 判準接續。

證據：[研究 360](../re/360-linux-bundled-music-tools.md)；契約：[規格 072](../spec/072-linux-music-tools.md)。
