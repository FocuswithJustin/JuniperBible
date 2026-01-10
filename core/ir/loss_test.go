package ir

import (
	"encoding/json"
	"testing"
)

func TestLossClassConstants(t *testing.T) {
	tests := []struct {
		lc   LossClass
		want string
	}{
		{LossL0, "L0"},
		{LossL1, "L1"},
		{LossL2, "L2"},
		{LossL3, "L3"},
		{LossL4, "L4"},
	}

	for _, tt := range tests {
		if string(tt.lc) != tt.want {
			t.Errorf("LossClass = %q, want %q", tt.lc, tt.want)
		}
	}
}

func TestLossClassValidation(t *testing.T) {
	tests := []struct {
		lc    LossClass
		valid bool
	}{
		{LossL0, true},
		{LossL1, true},
		{LossL2, true},
		{LossL3, true},
		{LossL4, true},
		{LossClass("L5"), false},
		{LossClass("INVALID"), false},
		{LossClass(""), false},
	}

	for _, tt := range tests {
		if got := tt.lc.IsValid(); got != tt.valid {
			t.Errorf("LossClass(%q).IsValid() = %v, want %v", tt.lc, got, tt.valid)
		}
	}
}

func TestLossClassLevel(t *testing.T) {
	tests := []struct {
		lc    LossClass
		level int
	}{
		{LossL0, 0},
		{LossL1, 1},
		{LossL2, 2},
		{LossL3, 3},
		{LossL4, 4},
		{LossClass("INVALID"), -1},
	}

	for _, tt := range tests {
		if got := tt.lc.Level(); got != tt.level {
			t.Errorf("LossClass(%q).Level() = %d, want %d", tt.lc, got, tt.level)
		}
	}
}

func TestLossClassIsLossless(t *testing.T) {
	tests := []struct {
		lc       LossClass
		lossless bool
	}{
		{LossL0, true},
		{LossL1, false},
		{LossL2, false},
		{LossL3, false},
		{LossL4, false},
	}

	for _, tt := range tests {
		if got := tt.lc.IsLossless(); got != tt.lossless {
			t.Errorf("LossClass(%q).IsLossless() = %v, want %v", tt.lc, got, tt.lossless)
		}
	}
}

func TestLossClassIsSemanticallyLossless(t *testing.T) {
	tests := []struct {
		lc       LossClass
		semantic bool
	}{
		{LossL0, true},
		{LossL1, true},
		{LossL2, false},
		{LossL3, false},
		{LossL4, false},
	}

	for _, tt := range tests {
		if got := tt.lc.IsSemanticallyLossless(); got != tt.semantic {
			t.Errorf("LossClass(%q).IsSemanticallyLossless() = %v, want %v", tt.lc, got, tt.semantic)
		}
	}
}

func TestLostElementJSON(t *testing.T) {
	elem := LostElement{
		Path:          "Gen.1.1/strongs[0]",
		ElementType:   "strongs",
		Reason:        "Target format does not support Strong's numbers",
		OriginalValue: "H430",
	}

	// Marshal to JSON
	data, err := json.Marshal(elem)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded LostElement
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.Path != elem.Path {
		t.Errorf("Path = %q, want %q", decoded.Path, elem.Path)
	}
	if decoded.ElementType != elem.ElementType {
		t.Errorf("ElementType = %q, want %q", decoded.ElementType, elem.ElementType)
	}
	if decoded.Reason != elem.Reason {
		t.Errorf("Reason = %q, want %q", decoded.Reason, elem.Reason)
	}
	if decoded.OriginalValue != elem.OriginalValue {
		t.Errorf("OriginalValue = %v, want %v", decoded.OriginalValue, elem.OriginalValue)
	}
}

func TestLossReportJSON(t *testing.T) {
	report := &LossReport{
		SourceFormat: "SWORD",
		TargetFormat: "IR",
		LossClass:    LossL1,
		LostElements: []LostElement{
			{
				Path:        "Gen.1.1/format",
				ElementType: "formatting",
				Reason:      "Custom formatting not preserved",
			},
		},
		Warnings: []string{
			"Module encryption not supported",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded LossReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.SourceFormat != report.SourceFormat {
		t.Errorf("SourceFormat = %q, want %q", decoded.SourceFormat, report.SourceFormat)
	}
	if decoded.TargetFormat != report.TargetFormat {
		t.Errorf("TargetFormat = %q, want %q", decoded.TargetFormat, report.TargetFormat)
	}
	if decoded.LossClass != report.LossClass {
		t.Errorf("LossClass = %q, want %q", decoded.LossClass, report.LossClass)
	}
	if len(decoded.LostElements) != 1 {
		t.Errorf("len(LostElements) = %d, want 1", len(decoded.LostElements))
	}
	if len(decoded.Warnings) != 1 {
		t.Errorf("len(Warnings) = %d, want 1", len(decoded.Warnings))
	}
}

func TestLossReportHasLoss(t *testing.T) {
	tests := []struct {
		name    string
		report  *LossReport
		hasLoss bool
	}{
		{
			name: "L0 with no lost elements",
			report: &LossReport{
				LossClass: LossL0,
			},
			hasLoss: false,
		},
		{
			name: "L0 with lost elements",
			report: &LossReport{
				LossClass: LossL0,
				LostElements: []LostElement{
					{Path: "test", ElementType: "test", Reason: "test"},
				},
			},
			hasLoss: true,
		},
		{
			name: "L1 with no lost elements",
			report: &LossReport{
				LossClass: LossL1,
			},
			hasLoss: true,
		},
		{
			name: "L2 with lost elements",
			report: &LossReport{
				LossClass: LossL2,
				LostElements: []LostElement{
					{Path: "test", ElementType: "test", Reason: "test"},
				},
			},
			hasLoss: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.report.HasLoss(); got != tt.hasLoss {
				t.Errorf("HasLoss() = %v, want %v", got, tt.hasLoss)
			}
		})
	}
}

func TestLossReportAddLostElement(t *testing.T) {
	report := &LossReport{
		SourceFormat: "SWORD",
		TargetFormat: "IR",
		LossClass:    LossL1,
	}

	report.AddLostElement("Gen.1.1/strongs", "strongs", "Not supported")

	if len(report.LostElements) != 1 {
		t.Fatalf("len(LostElements) = %d, want 1", len(report.LostElements))
	}

	elem := report.LostElements[0]
	if elem.Path != "Gen.1.1/strongs" {
		t.Errorf("Path = %q, want %q", elem.Path, "Gen.1.1/strongs")
	}
	if elem.ElementType != "strongs" {
		t.Errorf("ElementType = %q, want %q", elem.ElementType, "strongs")
	}
	if elem.Reason != "Not supported" {
		t.Errorf("Reason = %q, want %q", elem.Reason, "Not supported")
	}
}

func TestLossReportAddWarning(t *testing.T) {
	report := &LossReport{
		SourceFormat: "SWORD",
		TargetFormat: "IR",
		LossClass:    LossL1,
	}

	report.AddWarning("Module may have issues")
	report.AddWarning("Another warning")

	if len(report.Warnings) != 2 {
		t.Fatalf("len(Warnings) = %d, want 2", len(report.Warnings))
	}

	if report.Warnings[0] != "Module may have issues" {
		t.Errorf("Warnings[0] = %q, want %q", report.Warnings[0], "Module may have issues")
	}
}

func TestEmptyLossReport(t *testing.T) {
	report := &LossReport{}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded LossReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Empty report should not have loss (no lost elements and L0/empty class)
	if decoded.HasLoss() {
		t.Error("empty report should not have loss")
	}
}
