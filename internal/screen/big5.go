package screen

// toBig5 把一個 UTF-8 字轉成 Big5 位元組。
//
// 只做這個專案用得到的方向（UTF-8 → Big5），而且**查表來源是原版字型檔
// 對得上的那份對照**：規格 JSON 裡的字是給人讀的，真正決定畫面的是字碼。
// 目前的做法是靠 Go 標準庫沒有的 Big5 表，所以規格 JSON 另外帶字碼，
// 這支只在字碼缺席時當退路。
func toBig5(s string) []byte {
	if c, ok := big5Table[s]; ok {
		return []byte{byte(c >> 8), byte(c)}
	}
	return nil
}

// big5Table 只放目前畫面用得到的字。**不做完整的 Big5 表**——
// 完整轉碼是資料層的事，等文字資料格式解出來之後由那一層負責。
var big5Table = map[string]uint16{
	"帝": 0xABD2, "國": 0xB0EA, "同": 0xA650, "盟": 0xB7F9,
	"對": 0xB9EF, "戰": 0xBED4, "取": 0xA8FA, "進": 0xB669, "度": 0xABD7,
}
