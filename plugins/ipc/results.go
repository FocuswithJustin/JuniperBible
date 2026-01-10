package ipc

// Result types for IR conversion operations.
// These extend the base protocol types with IR-specific fields.

// ExtractIRResult is the result of an extract-ir command.
// Plugins can return either IRPath (file-based) or IR (inline).
type ExtractIRResult struct {
	IRPath     string      `json:"ir_path,omitempty"`     // Path to IR file (for large corpuses)
	IR         interface{} `json:"ir,omitempty"`          // Inline IR data (for small corpuses)
	LossClass  string      `json:"loss_class"`            // L0-L4 classification
	LossReport *LossReport `json:"loss_report,omitempty"` // Detailed loss information
}

// EmitNativeResult is the result of an emit-native command.
type EmitNativeResult struct {
	OutputPath string      `json:"output_path"`           // Path to generated file
	Format     string      `json:"format"`                // Output format name
	LossClass  string      `json:"loss_class"`            // L0-L4 classification
	LossReport *LossReport `json:"loss_report,omitempty"` // Detailed loss information
}

// LossReport describes any data loss during conversion.
// Follows the L0-L4 classification system:
// - L0: Byte-for-byte round-trip (lossless)
// - L1: Semantically lossless (formatting may differ)
// - L2: Minor loss (some metadata/structure)
// - L3: Significant loss (text preserved, markup lost)
// - L4: Text-only (minimal preservation)
type LossReport struct {
	SourceFormat string        `json:"source_format"`
	TargetFormat string        `json:"target_format"`
	LossClass    string        `json:"loss_class"`
	LostElements []LostElement `json:"lost_elements,omitempty"` // Specific elements lost
	Warnings     []string      `json:"warnings,omitempty"`      // General warnings
}

// LostElement describes a specific element that was lost during conversion.
type LostElement struct {
	Path          string      `json:"path"`                     // XPath or location
	ElementType   string      `json:"element_type"`             // Type of element
	Reason        string      `json:"reason"`                   // Why it was lost
	OriginalValue interface{} `json:"original_value,omitempty"` // Original value if captured
}

// EmitResult is a general result for emit commands that can return multiple files.
type EmitResult struct {
	Files      []EmittedFile `json:"files"`                 // List of emitted files
	LossReport *LossReport   `json:"loss_report,omitempty"` // Detailed loss information
}

// EmittedFile describes a file produced by an emit command.
type EmittedFile struct {
	Path   string `json:"path"`             // Path to the emitted file
	Format string `json:"format,omitempty"` // Format of the file
	Size   int64  `json:"size,omitempty"`   // Size in bytes
}
