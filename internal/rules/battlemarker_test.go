package rules

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestBattleMarkerMatchesOriginal(t *testing.T) {
	b, err := os.ReadFile("../../docs/internals/data/battle-markers.json")
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(b)
	if hex.EncodeToString(h[:]) != "6f67e3c5107f249b70280b8115d795492c7c823648d7ed9ccb935f1d66873424" {
		t.Fatal("標記證據身份不同")
	}
	var data struct {
		Rows []struct {
			Name            string
			Enabled, Forced int
			Before          string `json:"before_hex"`
			After           string `json:"after_hex"`
		} `json:"primitive_rows"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 96 {
		t.Fatal("缺原版案例")
	}
	for _, row := range data.Rows {
		t.Run(row.Name, func(t *testing.T) {
			before, err := hex.DecodeString(row.Before)
			if err != nil {
				t.Fatal(err)
			}
			want, err := hex.DecodeString(row.After)
			if err != nil {
				t.Fatal(err)
			}
			u := BattleUnit{Faction: int(before[0]), Detail: int(before[0x17]), MarkerFlags: int(before[0x19]),
				MarkerX: int(binary.LittleEndian.Uint16(before[0x1A:])), MarkerY: int(binary.LittleEndian.Uint16(before[0x1C:])),
				X: int(binary.LittleEndian.Uint16(before[0x34:])), Y: int(binary.LittleEndian.Uint16(before[0x36:]))}
			UpdateBattleMarker(&u, row.Forced != 0, row.Enabled != 0)
			s := &Slot{Ship: bytes.Clone(before)}
			s.ApplyBattleMarker(0, u)
			if !bytes.Equal(s.Ship, want) || u.Marked != (want[0x19]&0x80 != 0) {
				t.Fatalf("完整78-byte記錄不同：%X／%X", s.Ship, want)
			}
		})
	}
}
