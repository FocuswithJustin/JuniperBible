package ir

// LossClass represents the fidelity level of a format conversion.
type LossClass string

// Loss class constants, from most to least fidelity.
const (
	// LossL0 indicates lossless conversion - byte-for-byte round-trip possible.
	LossL0 LossClass = "L0"

	// LossL1 indicates semantically lossless - all content preserved, formatting may differ.
	LossL1 LossClass = "L1"

	// LossL2 indicates minor loss - some formatting lost (e.g., custom fonts).
	LossL2 LossClass = "L2"

	// LossL3 indicates significant loss - annotations lost (e.g., Strong's numbers).
	LossL3 LossClass = "L3"

	// LossL4 indicates plain text only - only raw text preserved.
	LossL4 LossClass = "L4"
)

// validLossClasses is the set of valid loss classes.
var validLossClasses = map[LossClass]bool{
	LossL0: true,
	LossL1: true,
	LossL2: true,
	LossL3: true,
	LossL4: true,
}

// IsValid returns true if the loss class is valid.
func (l LossClass) IsValid() bool {
	return validLossClasses[l]
}

// Level returns the numeric level (0-4) of the loss class.
func (l LossClass) Level() int {
	switch l {
	case LossL0:
		return 0
	case LossL1:
		return 1
	case LossL2:
		return 2
	case LossL3:
		return 3
	case LossL4:
		return 4
	default:
		return -1
	}
}

// IsLossless returns true if this loss class indicates no data loss.
func (l LossClass) IsLossless() bool {
	return l == LossL0
}

// IsSemanticallyLossless returns true if content is fully preserved.
func (l LossClass) IsSemanticallyLossless() bool {
	return l == LossL0 || l == LossL1
}

// LostElement describes a specific piece of data that was lost during conversion.
type LostElement struct {
	// Path is the location in the source (e.g., "Gen.1.1/strongs[0]").
	Path string `json:"path"`

	// ElementType describes what was lost (e.g., "strongs", "morphology").
	ElementType string `json:"element_type"`

	// Reason explains why the element was lost.
	Reason string `json:"reason"`

	// OriginalValue is the value that was lost (optional).
	OriginalValue interface{} `json:"original_value,omitempty"`
}

// LossReport documents the fidelity of a format conversion.
type LossReport struct {
	// SourceFormat is the format being converted from (e.g., "SWORD").
	SourceFormat string `json:"source_format"`

	// TargetFormat is the format being converted to (e.g., "IR", "OSIS").
	TargetFormat string `json:"target_format"`

	// LossClass is the overall fidelity classification.
	LossClass LossClass `json:"loss_class"`

	// LostElements lists specific pieces of data that were lost.
	LostElements []LostElement `json:"lost_elements,omitempty"`

	// Warnings contains non-fatal issues encountered during conversion.
	Warnings []string `json:"warnings,omitempty"`
}

// HasLoss returns true if any elements were lost.
func (r *LossReport) HasLoss() bool {
	return len(r.LostElements) > 0 || r.LossClass.Level() > 0
}

// AddLostElement adds a lost element to the report.
func (r *LossReport) AddLostElement(path, elementType, reason string) {
	r.LostElements = append(r.LostElements, LostElement{
		Path:        path,
		ElementType: elementType,
		Reason:      reason,
	})
}

// AddWarning adds a warning to the report.
func (r *LossReport) AddWarning(warning string) {
	r.Warnings = append(r.Warnings, warning)
}
