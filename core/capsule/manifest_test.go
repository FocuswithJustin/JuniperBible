package capsule

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewManifest(t *testing.T) {
	m := NewManifest()

	// Check version
	if m.CapsuleVersion != Version {
		t.Errorf("CapsuleVersion = %q, want %q", m.CapsuleVersion, Version)
	}

	// Check CreatedAt is valid RFC3339
	if _, err := time.Parse(time.RFC3339, m.CreatedAt); err != nil {
		t.Errorf("CreatedAt %q is not valid RFC3339: %v", m.CreatedAt, err)
	}

	// Check Tool info
	if m.Tool.Name != "capsule" {
		t.Errorf("Tool.Name = %q, want %q", m.Tool.Name, "capsule")
	}
	if m.Tool.Version != Version {
		t.Errorf("Tool.Version = %q, want %q", m.Tool.Version, Version)
	}

	// Check maps are initialized
	if m.Blobs.BySHA256 == nil {
		t.Error("Blobs.BySHA256 should not be nil")
	}
	if m.Artifacts == nil {
		t.Error("Artifacts should not be nil")
	}
	if m.Runs == nil {
		t.Error("Runs should not be nil")
	}
}

func TestManifestToJSON(t *testing.T) {
	m := NewManifest()
	m.Artifacts["test-artifact"] = &Artifact{
		ID:                "test-artifact",
		Kind:              "file",
		OriginalName:      "test.txt",
		PrimaryBlobSHA256: "abc123",
	}

	data, err := m.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Check it's valid JSON
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("ToJSON() produced invalid JSON: %v", err)
	}

	// Check expected fields
	if decoded["capsule_version"] != Version {
		t.Errorf("capsule_version = %v, want %q", decoded["capsule_version"], Version)
	}
}

func TestParseManifest(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid manifest",
			json: `{
				"capsule_version": "1.0.0",
				"created_at": "2024-01-01T00:00:00Z",
				"tool": {"name": "test", "version": "1.0"},
				"blobs": {"by_sha256": {}},
				"artifacts": {},
				"runs": {}
			}`,
			wantErr: false,
		},
		{
			name:    "empty json",
			json:    `{}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			json:    `{invalid}`,
			wantErr: true,
		},
		{
			name:    "empty string",
			json:    ``,
			wantErr: true,
		},
		{
			name: "with artifacts",
			json: `{
				"capsule_version": "1.0.0",
				"artifacts": {
					"art1": {
						"id": "art1",
						"kind": "file",
						"primary_blob_sha256": "deadbeef"
					}
				}
			}`,
			wantErr: false,
		},
		{
			name: "with runs",
			json: `{
				"capsule_version": "1.0.0",
				"runs": {
					"run1": {
						"id": "run1",
						"status": "success"
					}
				}
			}`,
			wantErr: false,
		},
		{
			name: "with ir extractions",
			json: `{
				"capsule_version": "1.0.0",
				"ir_extractions": {
					"ir1": {
						"id": "ir1",
						"source_artifact_id": "art1",
						"ir_blob_sha256": "abc123",
						"loss_class": "L0"
					}
				}
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := ParseManifest([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && m == nil {
				t.Error("ParseManifest() returned nil manifest without error")
			}
		})
	}
}

func TestManifestRoundtrip(t *testing.T) {
	// Create a comprehensive manifest
	original := NewManifest()
	original.Artifacts["art1"] = &Artifact{
		ID:                "art1",
		Kind:              "file",
		OriginalName:      "bible.zip",
		SourcePath:        "/path/to/bible.zip",
		PrimaryBlobSHA256: "sha256hash",
		Hashes: ArtifactHashes{
			SHA256: "sha256hash",
			BLAKE3: "blake3hash",
		},
		SizeBytes: 1024,
		Detected: &DetectionResult{
			FormatID:   "sword",
			Confidence: 0.95,
		},
	}

	original.Runs["run1"] = &Run{
		ID:     "run1",
		Status: "success",
		Engine: &Engine{
			EngineID: "nix-vm-1",
			Type:     "nix",
			Nix: &NixConfig{
				FlakeLockSHA256: "locksha",
				System:          "x86_64-linux",
				Derivations:     []string{"drv1", "drv2"},
			},
			Env: &EnvConfig{
				TZ:    "UTC",
				LCALL: "C.UTF-8",
				LANG:  "en_US.UTF-8",
			},
		},
		Plugin: &PluginInfo{
			PluginID:      "format-sword",
			PluginVersion: "1.0.0",
			Kind:          "format",
		},
		Inputs: []RunInput{
			{ArtifactID: "art1", Role: "primary"},
		},
		Command: &Command{
			Argv:    []string{"diatheke", "-b", "KJV"},
			Profile: "list-modules",
		},
		Outputs: &RunOutputs{
			TranscriptBlobSHA256: "transcriptsha",
			StdoutBlobSHA256:     "stdoutsha",
			Artifacts: []RunOutputArtifact{
				{ArtifactID: "out1", Label: "output"},
			},
		},
	}

	original.Blobs.BySHA256["sha256hash"] = &BlobRecord{
		SHA256:    "sha256hash",
		BLAKE3:    "blake3hash",
		SizeBytes: 1024,
		Path:      "blobs/sha256hash",
		MIME:      "application/zip",
	}

	original.IRExtractions = map[string]*IRRecord{
		"ir1": {
			ID:               "ir1",
			SourceArtifactID: "art1",
			IRBlobSHA256:     "irhash",
			IRFormat:         "ir-v1",
			IRVersion:        "1.0.0",
			LossClass:        "L1",
			ExtractorPlugin:  "format-sword",
		},
	}

	original.RoundtripPlans = map[string]*Plan{
		"plan1": {
			ID:          "plan1",
			Description: "Test plan",
			Steps: []PlanStep{
				{
					Type: "EXPORT",
					Export: &ExportStep{
						Mode:       "IDENTITY",
						ArtifactID: "art1",
					},
				},
			},
			Checks: []PlanCheck{
				{
					Type:  "BYTE_EQUAL",
					Label: "Check equality",
					ByteEqual: &ByteEqualCheck{
						ArtifactA: "art1",
						ArtifactB: "art2",
					},
				},
			},
		},
	}

	original.SelfChecks = map[string]*SelfCheck{
		"check1": {
			ID:                "check1",
			PlanID:            "plan1",
			TargetArtifactIDs: []string{"art1"},
			ReportBlobSHA256:  "reportsha",
			Status:            "pass",
		},
	}

	original.Exports = map[string]*Export{
		"export1": {
			ID:               "export1",
			Mode:             "IDENTITY",
			ArtifactID:       "art1",
			ResultBlobSHA256: "resultsha",
		},
	}

	// Serialize to JSON
	data, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Parse back
	decoded, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("ParseManifest() error = %v", err)
	}

	// Verify all fields
	if decoded.CapsuleVersion != original.CapsuleVersion {
		t.Errorf("CapsuleVersion = %q, want %q", decoded.CapsuleVersion, original.CapsuleVersion)
	}

	// Verify artifact
	art := decoded.Artifacts["art1"]
	if art == nil {
		t.Fatal("Artifact art1 not found")
	}
	if art.OriginalName != "bible.zip" {
		t.Errorf("Artifact.OriginalName = %q, want %q", art.OriginalName, "bible.zip")
	}
	if art.Detected == nil || art.Detected.FormatID != "sword" {
		t.Error("Artifact.Detected not preserved correctly")
	}

	// Verify run
	run := decoded.Runs["run1"]
	if run == nil {
		t.Fatal("Run run1 not found")
	}
	if run.Engine == nil || run.Engine.Nix == nil {
		t.Error("Run.Engine.Nix not preserved")
	}
	if run.Command == nil || run.Command.Profile != "list-modules" {
		t.Error("Run.Command not preserved")
	}
	if run.Outputs == nil || len(run.Outputs.Artifacts) != 1 {
		t.Error("Run.Outputs.Artifacts not preserved")
	}

	// Verify IR extractions
	if decoded.IRExtractions == nil {
		t.Fatal("IRExtractions is nil")
	}
	ir := decoded.IRExtractions["ir1"]
	if ir == nil || ir.LossClass != "L1" {
		t.Error("IRExtractions not preserved correctly")
	}

	// Verify plans
	if decoded.RoundtripPlans == nil {
		t.Fatal("RoundtripPlans is nil")
	}
	plan := decoded.RoundtripPlans["plan1"]
	if plan == nil || len(plan.Steps) != 1 {
		t.Error("RoundtripPlans not preserved correctly")
	}

	// Verify self checks
	if decoded.SelfChecks == nil {
		t.Fatal("SelfChecks is nil")
	}
	sc := decoded.SelfChecks["check1"]
	if sc == nil || sc.Status != "pass" {
		t.Error("SelfChecks not preserved correctly")
	}

	// Verify exports
	if decoded.Exports == nil {
		t.Fatal("Exports is nil")
	}
	exp := decoded.Exports["export1"]
	if exp == nil || exp.Mode != "IDENTITY" {
		t.Error("Exports not preserved correctly")
	}
}

func TestArtifactComponents(t *testing.T) {
	artifact := &Artifact{
		ID:   "container",
		Kind: "dir",
		Components: []ArtifactComponent{
			{Path: "file1.txt", ArtifactID: "comp1"},
			{Path: "file2.txt", ArtifactID: "comp2"},
		},
	}

	m := NewManifest()
	m.Artifacts["container"] = artifact

	data, err := m.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	decoded, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("ParseManifest() error = %v", err)
	}

	art := decoded.Artifacts["container"]
	if art == nil {
		t.Fatal("Artifact not found")
	}
	if len(art.Components) != 2 {
		t.Errorf("len(Components) = %d, want 2", len(art.Components))
	}
	if art.Components[0].Path != "file1.txt" {
		t.Errorf("Components[0].Path = %q, want %q", art.Components[0].Path, "file1.txt")
	}
}

func TestAttributesSerialization(t *testing.T) {
	m := NewManifest()
	m.Attributes = Attributes{
		"string":  "value",
		"number":  42.5,
		"boolean": true,
		"nested": map[string]interface{}{
			"key": "nested_value",
		},
	}

	data, err := m.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	decoded, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("ParseManifest() error = %v", err)
	}

	if decoded.Attributes["string"] != "value" {
		t.Errorf("Attributes[string] = %v, want %q", decoded.Attributes["string"], "value")
	}
	if decoded.Attributes["number"] != 42.5 {
		t.Errorf("Attributes[number] = %v, want 42.5", decoded.Attributes["number"])
	}
	if decoded.Attributes["boolean"] != true {
		t.Errorf("Attributes[boolean] = %v, want true", decoded.Attributes["boolean"])
	}
}

func TestTranscriptCheck(t *testing.T) {
	check := &PlanCheck{
		Type:  "TRANSCRIPT_EQUAL",
		Label: "Verify transcripts match",
		TranscriptEqual: &TranscriptCheck{
			RunA: "run1",
			RunB: "run2",
		},
	}

	m := NewManifest()
	m.RoundtripPlans = map[string]*Plan{
		"plan1": {
			ID:     "plan1",
			Checks: []PlanCheck{*check},
		},
	}

	data, err := m.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	decoded, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("ParseManifest() error = %v", err)
	}

	plan := decoded.RoundtripPlans["plan1"]
	if plan == nil {
		t.Fatal("Plan not found")
	}
	if len(plan.Checks) != 1 {
		t.Fatal("Check not preserved")
	}
	if plan.Checks[0].TranscriptEqual == nil {
		t.Fatal("TranscriptEqual is nil")
	}
	if plan.Checks[0].TranscriptEqual.RunA != "run1" {
		t.Errorf("TranscriptEqual.RunA = %q, want %q",
			plan.Checks[0].TranscriptEqual.RunA, "run1")
	}
}
