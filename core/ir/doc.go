// Package ir provides the Intermediate Representation (IR) for lossless Bible format conversion.
//
// The IR uses stand-off markup to handle overlapping structures common in Bible texts,
// where verses can span poetry lines, quotations cross chapter boundaries, and
// annotations attach to arbitrary text ranges.
//
// # Core Types
//
// The IR is organized hierarchically:
//
//   - Corpus: Top-level container for a complete Bible module
//   - Document: A single book, article, or dictionary entry
//   - ContentBlock: Contiguous text unit (paragraph or section)
//   - Token: Word or whitespace unit for linguistic annotation
//
// # Stand-off Markup
//
// Instead of inline markup (which forces a tree structure), the IR uses:
//
//   - Anchor: Position markers within content blocks
//   - Span: Regions defined by start/end anchors (can overlap freely)
//   - Annotation: Metadata attached to spans
//
// # Loss Classification
//
// Every format conversion is classified by fidelity:
//
//   - L0: Lossless - byte-for-byte round-trip possible
//   - L1: Semantically Lossless - all content preserved, formatting may differ
//   - L2: Minor Loss - some formatting lost
//   - L3: Significant Loss - annotations lost
//   - L4: Plain Text - only raw text preserved
//
// # Content Addressing
//
// All IR content is hashed using SHA-256 for deduplication, change detection,
// and verification of round-trip fidelity.
//
// # Example
//
//	corpus := &ir.Corpus{
//	    ID:            "KJV",
//	    Version:       "1.0.0",
//	    ModuleType:    ir.ModuleBible,
//	    Versification: "KJV",
//	    Language:      "en",
//	    Title:         "King James Version",
//	}
//
//	doc := &ir.Document{
//	    ID:    "Gen",
//	    Title: "Genesis",
//	    Order: 1,
//	}
//	corpus.Documents = append(corpus.Documents, doc)
package ir
