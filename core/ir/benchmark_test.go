package ir

import (
	"encoding/json"
	"testing"
)

// BenchmarkCorpusSerialization benchmarks JSON serialization of an IR Corpus.
func BenchmarkCorpusSerialization(b *testing.B) {
	corpusSizes := []struct {
		name       string
		numDocs    int
		numBlocks  int
		numAnchors int
	}{
		{"Small_1Doc_10Blocks", 1, 10, 5},
		{"Medium_5Docs_50Blocks", 5, 50, 25},
		{"Large_66Docs_200Blocks", 66, 200, 100},
	}

	for _, cs := range corpusSizes {
		b.Run(cs.name, func(b *testing.B) {
			// Create a corpus with the specified size
			corpus := &Corpus{
				ID:            "BENCHMARK",
				Version:       "1.0.0",
				ModuleType:    ModuleBible,
				Versification: "KJV",
				Language:      "en",
				Title:         "Benchmark Bible",
				Documents:     make([]*Document, 0, cs.numDocs),
			}

			// Generate documents
			for d := 0; d < cs.numDocs; d++ {
				doc := &Document{
					ID:            "Book" + string(rune('0'+(d%10))),
					Title:         "Test Book " + string(rune('A'+(d%26))),
					Order:         d + 1,
					ContentBlocks: make([]*ContentBlock, 0, cs.numBlocks),
				}

				// Generate content blocks
				for i := 0; i < cs.numBlocks; i++ {
					block := &ContentBlock{
						ID:       "Block" + string(rune('0'+(i%10))),
						Sequence: i,
						Text:     "This is verse " + string(rune('0'+(i%10))) + " with some sample text content for benchmarking purposes.",
						Anchors:  make([]*Anchor, 0, cs.numAnchors),
					}

					// Generate anchors
					offset := 0
					for a := 0; a < cs.numAnchors; a++ {
						anchor := &Anchor{
							ID:         "anchor-" + string(rune('0'+(a%10))),
							CharOffset: offset,
						}
						block.Anchors = append(block.Anchors, anchor)
						offset += 10
					}

					doc.ContentBlocks = append(doc.ContentBlocks, block)
				}

				corpus.Documents = append(corpus.Documents, doc)
			}

			// Reset timer after setup
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				data, err := json.Marshal(corpus)
				if err != nil {
					b.Fatalf("failed to marshal corpus: %v", err)
				}

				// Report size on first iteration
				if i == 0 {
					b.ReportMetric(float64(len(data)), "bytes")
				}
			}
		})
	}
}

// BenchmarkCorpusDeserialization benchmarks JSON deserialization of an IR Corpus.
func BenchmarkCorpusDeserialization(b *testing.B) {
	corpusSizes := []struct {
		name       string
		numDocs    int
		numBlocks  int
		numAnchors int
	}{
		{"Small_1Doc_10Blocks", 1, 10, 5},
		{"Medium_5Docs_50Blocks", 5, 50, 25},
		{"Large_66Docs_200Blocks", 66, 200, 100},
	}

	for _, cs := range corpusSizes {
		b.Run(cs.name, func(b *testing.B) {
			// Create a corpus with the specified size
			corpus := &Corpus{
				ID:            "BENCHMARK",
				Version:       "1.0.0",
				ModuleType:    ModuleBible,
				Versification: "KJV",
				Language:      "en",
				Title:         "Benchmark Bible",
				Documents:     make([]*Document, 0, cs.numDocs),
			}

			// Generate documents
			for d := 0; d < cs.numDocs; d++ {
				doc := &Document{
					ID:            "Book" + string(rune('0'+(d%10))),
					Title:         "Test Book " + string(rune('A'+(d%26))),
					Order:         d + 1,
					ContentBlocks: make([]*ContentBlock, 0, cs.numBlocks),
				}

				// Generate content blocks
				for i := 0; i < cs.numBlocks; i++ {
					block := &ContentBlock{
						ID:       "Block" + string(rune('0'+(i%10))),
						Sequence: i,
						Text:     "This is verse " + string(rune('0'+(i%10))) + " with some sample text content for benchmarking purposes.",
						Anchors:  make([]*Anchor, 0, cs.numAnchors),
					}

					// Generate anchors
					offset := 0
					for a := 0; a < cs.numAnchors; a++ {
						anchor := &Anchor{
							ID:         "anchor-" + string(rune('0'+(a%10))),
							CharOffset: offset,
						}
						block.Anchors = append(block.Anchors, anchor)
						offset += 10
					}

					doc.ContentBlocks = append(doc.ContentBlocks, block)
				}

				corpus.Documents = append(corpus.Documents, doc)
			}

			// Serialize once to get JSON data
			data, err := json.Marshal(corpus)
			if err != nil {
				b.Fatalf("failed to marshal corpus: %v", err)
			}

			// Reset timer after setup
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				var result Corpus
				err := json.Unmarshal(data, &result)
				if err != nil {
					b.Fatalf("failed to unmarshal corpus: %v", err)
				}
			}

			b.StopTimer()
			b.ReportMetric(float64(len(data)), "bytes")
		})
	}
}

// BenchmarkIRValidation benchmarks validation of an IR Corpus.
func BenchmarkIRValidation(b *testing.B) {
	validationSizes := []struct {
		name       string
		numDocs    int
		numBlocks  int
		numAnchors int
	}{
		{"Small_1Doc_10Blocks", 1, 10, 5},
		{"Medium_5Docs_50Blocks", 5, 50, 25},
		{"Large_66Docs_200Blocks", 66, 200, 100},
	}

	for _, vs := range validationSizes {
		b.Run(vs.name, func(b *testing.B) {
			// Create a valid corpus
			corpus := &Corpus{
				ID:            "BENCHMARK",
				Version:       "1.0.0",
				ModuleType:    ModuleBible,
				Versification: "KJV",
				Language:      "en",
				Title:         "Benchmark Bible",
				Documents:     make([]*Document, 0, vs.numDocs),
			}

			// Generate documents
			for d := 0; d < vs.numDocs; d++ {
				doc := &Document{
					ID:            "Book" + string(rune('0'+(d%10))),
					Title:         "Test Book " + string(rune('A'+(d%26))),
					Order:         d + 1,
					ContentBlocks: make([]*ContentBlock, 0, vs.numBlocks),
				}

				// Generate content blocks with references
				for i := 0; i < vs.numBlocks; i++ {
					block := &ContentBlock{
						ID:       "Block" + string(rune('0'+(i%10))),
						Sequence: i,
						Text:     "This is verse " + string(rune('0'+(i%10))) + " with some sample text content for benchmarking purposes.",
						Anchors:  make([]*Anchor, 0, vs.numAnchors),
					}

					// Generate anchors
					offset := 0
					for a := 0; a < vs.numAnchors; a++ {
						anchor := &Anchor{
							ID:         "anchor-" + string(rune('0'+(a%10))),
							CharOffset: offset,
						}
						block.Anchors = append(block.Anchors, anchor)
						offset += 10
					}

					doc.ContentBlocks = append(doc.ContentBlocks, block)
				}

				corpus.Documents = append(corpus.Documents, doc)
			}

			// Reset timer after setup
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				errs := ValidateCorpus(corpus)
				if len(errs) > 0 {
					b.Fatalf("validation failed with %d errors: %v", len(errs), errs[0])
				}
			}
		})
	}
}

// BenchmarkRefParsing benchmarks parsing scripture references.
func BenchmarkRefParsing(b *testing.B) {
	refStrings := []string{
		"Gen.1.1",
		"Matt.5.3-12",
		"John.3.16",
		"Rom.8.28",
		"Ps.119.105",
	}

	for _, refStr := range refStrings {
		b.Run(refStr, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ref, err := ParseRef(refStr)
				if err != nil {
					b.Fatalf("failed to parse ref: %v", err)
				}
				if ref == nil {
					b.Fatal("ref is nil")
				}
			}
		})
	}
}

// BenchmarkRefFormatting benchmarks formatting scripture references.
func BenchmarkRefFormatting(b *testing.B) {
	refs := []*Ref{
		{Book: "Gen", Chapter: 1, Verse: 1},
		{Book: "Matt", Chapter: 5, Verse: 3, VerseEnd: 12},
		{Book: "John", Chapter: 3, Verse: 16},
		{Book: "Rom", Chapter: 8, Verse: 28},
		{Book: "Ps", Chapter: 119, Verse: 105},
	}

	for _, ref := range refs {
		b.Run(ref.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ref.String()
			}
		})
	}
}

// BenchmarkContentBlockHashing benchmarks computing content block hashes.
func BenchmarkContentBlockHashing(b *testing.B) {
	blockSizes := []struct {
		name     string
		textSize int
	}{
		{"Short_50chars", 50},
		{"Medium_500chars", 500},
		{"Long_5000chars", 5000},
	}

	for _, bs := range blockSizes {
		b.Run(bs.name, func(b *testing.B) {
			// Create text of specified size
			text := make([]byte, bs.textSize)
			for i := range text {
				text[i] = 'A' + byte(i%26)
			}

			block := &ContentBlock{
				ID:       "test-block",
				Sequence: 0,
				Text:     string(text),
			}

			// Reset timer after setup
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				block.ComputeHash()
			}
		})
	}
}

// BenchmarkParallelProcessing benchmarks parallel processing of verses.
func BenchmarkParallelProcessing(b *testing.B) {
	// Create a sample corpus
	corpus := &Corpus{
		ID:            "TEST",
		Version:       "1.0.0",
		ModuleType:    ModuleBible,
		Versification: "KJV",
		Language:      "en",
		Title:         "Test Bible",
		Documents:     make([]*Document, 0, 10),
	}

	// Generate multiple documents
	for d := 0; d < 10; d++ {
		doc := &Document{
			ID:            "Book" + string(rune('0'+d)),
			Title:         "Test Book",
			Order:         d + 1,
			ContentBlocks: make([]*ContentBlock, 0, 100),
		}

		// Generate 100 content blocks per document
		for i := 0; i < 100; i++ {
			block := &ContentBlock{
				ID:       "Block" + string(rune('0'+(i%10))),
				Sequence: i,
				Text:     "Sample verse text for parallel processing benchmark. This text is repeated to simulate realistic content.",
			}
			doc.ContentBlocks = append(doc.ContentBlocks, block)
		}

		corpus.Documents = append(corpus.Documents, doc)
	}

	// Reset timer after setup
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		// Process all blocks in parallel
		results := make([]string, 0)
		for _, doc := range corpus.Documents {
			for _, block := range doc.ContentBlocks {
				// Simulate some processing work
				hash := block.ComputeHash()
				results = append(results, hash)
			}
		}

		if len(results) != 1000 {
			b.Fatalf("expected 1000 results, got %d", len(results))
		}
	}
}

// BenchmarkVersificationMapping benchmarks versification system mapping.
func BenchmarkVersificationMapping(b *testing.B) {
	// Create a mapping table
	mappingTable := &MappingTable{
		ID:         "KJV-to-NRSV",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationNRSV,
		Mappings: []*RefMapping{
			{
				From: &Ref{Book: "Gen", Chapter: 1, Verse: 1},
				To:   &Ref{Book: "Gen", Chapter: 1, Verse: 1},
				Type: MappingExact,
			},
			{
				From: &Ref{Book: "Gen", Chapter: 1, Verse: 2},
				To:   &Ref{Book: "Gen", Chapter: 1, Verse: 2},
				Type: MappingExact,
			},
			// Add more mappings for realistic benchmarking
		},
	}

	// Add 100 mappings
	for i := 3; i <= 100; i++ {
		mappingTable.Mappings = append(mappingTable.Mappings, &RefMapping{
			From: &Ref{Book: "Gen", Chapter: 1, Verse: i},
			To:   &Ref{Book: "Gen", Chapter: 1, Verse: i},
			Type: MappingExact,
		})
	}

	// Reset timer after setup
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		// Validate the mapping table
		errs := ValidateMappingTable(mappingTable)
		if len(errs) > 0 {
			b.Fatalf("validation failed: %v", errs[0])
		}
	}
}
