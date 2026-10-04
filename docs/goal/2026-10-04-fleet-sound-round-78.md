# 第78輪：交戰音效來源與正式接線

日期：2026-10-04。父提交e927750a782af344d77e6a3dd866e18af753ad28。
沿[#48](https://github.com/wicanr2/logh3/issues/48)與[#79](https://github.com/wicanr2/logh3/issues/79)。
遠端89張worklist、9張OPEN，269條核對零失效。

既有406保存COMANI記錄+6與播放入口。先核GIN3SP.VOC的自訂標頭、
音效編號到樣本的資料鏈與取樣參數，再審查規格，足以READY才接正式播放。
原始素材唯讀，不把未初始化的靜圖byte當音效。
硬體依公開規格近似，不追driver／ISR逐週期同步。
正常GUI沿既有鍵鼠工具補驗，不能用Game內部重播取代實際視窗。
#89的排序取捨保持pending，不以音效或GUI通過升格完整畫面驗收。
