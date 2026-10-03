# 第70輪：艦隊交戰下方場景

沿[#48](https://github.com/wicanr2/logh3/issues/48)與[#79](https://github.com/wicanr2/logh3/issues/79)。
父提交ee5561916311c8445a2434fa6d831bf9d88a73ed，89張worklist／9張OPEN，269核對零失效。
093／094 READY；完整056／092仍DRAFT，Game尚未接艦隊交戰核心。

沿第69輪已保存的完整scene owner，補2376:07FA透明堆疊／裁切、0AA5背景buffer及0B43呈現。
從321F:06FC整個下方場景的前後畫面裁決，不只驗素材解碼。
用固定完整state、唯讀原版與專用dosgolem取證；保留位址／原名／bytes／分級語意。
先建立可重現的完整下方場景，再審查READY與接正式renderer；未知不猜補。
完整Game命令／演出／保存與正常玩家路徑仍是目標，不把probe通過當整項完成。

本輪勘誤：0B43只有retf，無背景buffer搬移。40層及20可返回裁切閉合，095先READY再實作。
兩次直接入口未返回保留；文字／正式色盤／等待輸入及Game仍待092，不升格完整交戰。
