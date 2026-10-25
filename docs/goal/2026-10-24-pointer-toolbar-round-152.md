# 第152輪：滑鼠回應、hover與工具列字型／AI

使用者試玩v1.6.0回報滑鼠移動遲鈍、選項及按鈕沒有hover變色，並要求上方工具列直接調字型及電腦戰略AI。字型已明確選擇現有Noto／思源黑體，不新增圓體。

## 目前狀態表

| 項目 | 狀態 | 證據及閘門 |
|---|---|---|
| 開工 | 完成 | Issue102張、#48 OPEN；282條verify零失效，保留既有未提交修改及44a8cae1 README提交 |
| 滑鼠根因與修正 | 完成 | 游標參與背景鍵已證實；最上層游標獨立GPU繪製。24次移動上傳20至20，Linux Draw中位32.85至11.93ms；不宣稱固定硬體輸入延遲 |
| hover | 完成 | 主指令座標少40像素、空角色框及系統下拉列已接線。正式不含游標ROI的HD變色916／398，原版1512／912，基準0／0 |
| 工具列字型 | 完成 | 現有Noto／思源黑體與其他已載入字型直接可選，設定同步；設定頁專用黑體及原版字形保持 |
| 工具列AI | 完成 | 直接選原版／強化0–3，沿既有SetAILevel；設定與存讀後當前值同步，不改AI規則或六表格式 |
| 實作與驗收 | 154限定CONFORMED | 16 PASS／4無測試，Linux兩陣營與最終Windows Wine各31張正常操作；五語、F1／F7、取消、海報中央及六表通過。見[134](../playtest/134-pointer-toolbar.md) |
| Windows更新包 | 完成 | v.1.6.1-20261025 ZIP10766673 bytes，實際解出EXE驗證；1616項來源含638項非vendor，24日候選與25日來源全同 |
| 收尾 | 完成 | 最新Issue102張、#48 OPEN，282條verify零失效。獨立收據、擁有權及Docker清理核對，未commit、push或發布 |

本輪工作區沿`workplace/pointer152/`；低風險probe只用來找原因，不能代替正常玩家路徑。最終交付為`dist-all/v.1.6.1-20261025/release/LOGH3-v.1.6.1-20261025-windows-amd64-patch.zip`，替換既有v1.6.0的gin3.exe即可。前版故事推廣片、公開README與三平台包保持。

機器收據：[pointer-toolbar152.json](../internals/data/pointer-toolbar152.json)。重現入口：[打包器](../../tools/scripts/pointer_windows_patch.py)、[正常GUI](../../tools/scripts/pointer_toolbar_gui.py)、[獨立驗證器](../../tools/scripts/pointer_toolbar_receipt.py)。Windows採Wine，未驗真機；本次不新增DOS parity，字型及AI沿既有本次執行設定。

入口：[目前脈絡](../../CONTEXT.md)、[136時鐘與顯示](../spec/136-live-graphics-and-cadence.md)、[135呈現與既有人事快取](../spec/135-original-style-hd-ui.md)。
