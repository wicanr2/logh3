package main

import (
	"reflect"
	"strings"
	"testing"
)

// 每一條畫成 issue 本文再讀回來，要與原來**完全相同**。
//
// ⚠ 這支測試是搬移的前提。GitHub 變成權威之後，JSON 那一份就不再是
// 真相——讀回來少一個欄位（證據少一行、判準的換行被吃掉），那一條就
// 在搬移的那一刻安靜地變了，而 issue 看起來完全正常。所以拿**現有的每
// 一條**驗，不是拿幾條手寫的樣本驗。
func TestIssueRoundTrip(t *testing.T) {
	doc, err := Load("../../" + dataPath)
	if err != nil {
		t.Skip("讀不到 worklist.json：", err)
	}
	for i := range doc.Items {
		it := doc.Items[i]
		if it.Issue > 0 {
			continue // 佔位：內容在 GitHub，它本身只是指標，不會再被畫成 issue
		}
		body, err := IssueBody(&it)
		if err != nil {
			t.Fatalf("%s：%v", it.ID, err)
		}
		state := "OPEN"
		if it.Status == StatusDone {
			state = "CLOSED"
		}
		got, err := ParseIssue(IssueRecord{Number: i + 1, Title: it.Title, Body: body, State: state})
		if err != nil {
			t.Fatalf("%s：%v", it.ID, err)
		}
		if !reflect.DeepEqual(got, it) {
			t.Errorf("%s 讀回來不一樣：\n原本 %+v\n讀回 %+v", it.ID, it, got)
		}
	}
}

// GitHub 網頁編輯會把換行存成 CRLF。讀的時候要當成同一件事。
func TestIssueParseToleratesCRLF(t *testing.T) {
	it := Item{ID: "x", Status: StatusOpen, Layer: "re", Title: "標題",
		Body: []string{"第一行", "", "第三行"}, Accept: "判準",
		Evidence: []string{"docs/re/1.md"}, Verify: Verify{Kind: "manual", Note: "n"}}
	body, _ := IssueBody(&it)
	got, err := ParseIssue(IssueRecord{Title: "標題", Body: strings.ReplaceAll(body, "\n", "\r\n"), State: "OPEN"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, it) {
		t.Errorf("CRLF 讀回來不一樣：%+v", got)
	}
}

// 開頭標記與機器區塊的 id 不同時要拒絕——那通常是複製了別張的本文。
func TestIssueParseRejectsMismatchedID(t *testing.T) {
	it := Item{ID: "a", Status: StatusOpen, Layer: "re", Title: "t", Verify: Verify{Kind: "manual"}}
	body, _ := IssueBody(&it)
	body = strings.Replace(body, issueMarkerPrefix+"a", issueMarkerPrefix+"b", 1)
	if _, err := ParseIssue(IssueRecord{Number: 7, Title: "t", Body: body, State: "OPEN"}); err == nil {
		t.Error("id 不一致應該要被拒絕")
	}
}

// 同一個 id 以 issue 為準，而且關掉的 issue ＝ done。
func TestMergeIssuesPrefersTheIssue(t *testing.T) {
	doc := &Doc{Items: []Item{{ID: "a", Status: StatusOpen, Layer: "re", Title: "舊標題", Verify: Verify{Kind: "manual"}}}}
	it := Item{ID: "a", Status: StatusOpen, Layer: "re", Title: "新標題", Verify: Verify{Kind: "manual"}}
	body, _ := IssueBody(&it)
	rec := IssueRecord{Number: 1, Title: "新標題", Body: body, State: "CLOSED"}
	rec.Labels = append(rec.Labels, struct {
		Name string `json:"name"`
	}{labelAll})
	if err := MergeIssues(doc, []IssueRecord{rec}); err != nil {
		t.Fatal(err)
	}
	if len(doc.Items) != 1 || doc.Items[0].Title != "新標題" || doc.Items[0].Status != StatusDone {
		t.Errorf("合併結果不對：%+v", doc.Items)
	}
}

// 佔位合併之後要**原地**換成 issue 的內容：矩陣列的順序照 JSON 排，
// 搬家不能讓那幾列掉到章節尾巴。
func TestStubIsReplacedInPlace(t *testing.T) {
	a := Item{ID: "a", Status: StatusDone, Layer: "matrix", Section: "s1", Title: "甲", Verify: Verify{Kind: "manual"}}
	mid := Item{ID: "b", Status: StatusOpen, Layer: "matrix", Section: "s1", Title: "乙", Verify: Verify{Kind: "manual"}}
	c := Item{ID: "c", Status: StatusDone, Layer: "matrix", Section: "s1", Title: "丙", Verify: Verify{Kind: "manual"}}
	doc := &Doc{Items: []Item{a, Stub(&mid, 12), c}}
	body, _ := IssueBody(&mid)
	rec := IssueRecord{Number: 12, Title: "乙", Body: body, State: "OPEN"}
	rec.Labels = append(rec.Labels, struct {
		Name string `json:"name"`
	}{labelAll})
	if err := MergeIssues(doc, []IssueRecord{rec}); err != nil {
		t.Fatal(err)
	}
	if len(doc.Items) != 3 || doc.Items[1].ID != "b" || doc.Items[1].Title != "乙" || doc.Items[1].Issue != 12 {
		t.Errorf("佔位沒有被原地換掉：%+v", doc.Items)
	}
}
