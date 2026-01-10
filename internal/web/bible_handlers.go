package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/FocuswithJustin/JuniperBible/core/ir"
)

// bibleCache caches Bible info to avoid re-parsing capsules on every request.
var bibleCache struct {
	sync.RWMutex
	bibles    []BibleInfo
	populated bool // true if cache has been populated (even if empty)
	timestamp time.Time
	ttl       time.Duration
}

// corpusCache caches parsed IR corpora for individual Bibles.
// This speeds up Bible detail, book, and chapter views significantly.
var corpusCache struct {
	sync.RWMutex
	corpora map[string]*corpusCacheEntry
	ttl     time.Duration
}

type corpusCacheEntry struct {
	corpus    *ir.Corpus
	capsuleID string
	timestamp time.Time
}

func init() {
	bibleCache.ttl = 5 * time.Minute // Cache for 5 minutes
	corpusCache.corpora = make(map[string]*corpusCacheEntry)
	corpusCache.ttl = 10 * time.Minute // Cache corpora longer since they're expensive to load
}

// getCachedBibles returns cached Bible list or rebuilds if expired.
func getCachedBibles() []BibleInfo {
	bibleCache.RLock()
	if bibleCache.populated && time.Since(bibleCache.timestamp) < bibleCache.ttl {
		bibles := bibleCache.bibles
		bibleCache.RUnlock()
		return bibles
	}
	bibleCache.RUnlock()

	// Rebuild cache
	bibleCache.Lock()
	defer bibleCache.Unlock()

	// Double-check after acquiring write lock
	if bibleCache.populated && time.Since(bibleCache.timestamp) < bibleCache.ttl {
		return bibleCache.bibles
	}

	start := time.Now()
	bibleCache.bibles = listBiblesUncached()
	bibleCache.populated = true
	bibleCache.timestamp = time.Now()
	log.Printf("[CACHE] Rebuilt Bible cache: %d bibles in %v", len(bibleCache.bibles), time.Since(start))

	return bibleCache.bibles
}

// invalidateBibleCache forces a cache rebuild on next access.
func invalidateBibleCache() {
	bibleCache.Lock()
	bibleCache.populated = false
	bibleCache.timestamp = time.Time{}
	bibleCache.Unlock()
}

// invalidateCorpusCache clears all cached corpora.
func invalidateCorpusCache() {
	corpusCache.Lock()
	corpusCache.corpora = make(map[string]*corpusCacheEntry)
	corpusCache.Unlock()
}

// getCachedCorpus returns a cached corpus or loads it from disk.
func getCachedCorpus(capsuleID string) (*ir.Corpus, string, error) {
	corpusCache.RLock()
	if entry, ok := corpusCache.corpora[capsuleID]; ok {
		if time.Since(entry.timestamp) < corpusCache.ttl {
			corpus := entry.corpus
			path := entry.capsuleID
			corpusCache.RUnlock()
			return corpus, path, nil
		}
	}
	corpusCache.RUnlock()

	// Load from disk
	capsules := listCapsules()
	var capsulePath string
	for _, c := range capsules {
		id := strings.TrimSuffix(c.Name, ".capsule.tar.xz")
		id = strings.TrimSuffix(id, ".tar.xz")
		id = strings.TrimSuffix(id, ".tar.gz")
		id = strings.TrimSuffix(id, ".tar")
		if strings.EqualFold(id, capsuleID) {
			capsulePath = c.Path
			break
		}
	}
	if capsulePath == "" {
		return nil, "", fmt.Errorf("capsule not found: %s", capsuleID)
	}

	irContent, err := readIRContent(filepath.Join(ServerConfig.CapsulesDir, capsulePath))
	if err != nil {
		return nil, "", err
	}

	corpus := parseIRToCorpus(irContent)
	if corpus == nil {
		return nil, "", fmt.Errorf("invalid IR content")
	}

	// Store in cache
	corpusCache.Lock()
	corpusCache.corpora[capsuleID] = &corpusCacheEntry{
		corpus:    corpus,
		capsuleID: capsulePath,
		timestamp: time.Now(),
	}
	corpusCache.Unlock()

	log.Printf("[CACHE] Loaded corpus for %s", capsuleID)
	return corpus, capsulePath, nil
}

// PreWarmCaches pre-populates caches on server startup.
// This runs in a goroutine so it doesn't block server startup.
func PreWarmCaches() {
	go func() {
		log.Println("[CACHE] Pre-warming Bible cache...")
		getCachedBibles()
		log.Println("[CACHE] Pre-warm complete")
	}()
}

// StartBackgroundCacheRefresh starts a goroutine that refreshes caches
// before they expire to ensure users never experience cache miss latency.
func StartBackgroundCacheRefresh() {
	go func() {
		// Refresh at 80% of TTL to ensure cache is always warm
		refreshInterval := time.Duration(float64(bibleCache.ttl) * 0.8)
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for range ticker.C {
			bibleCache.RLock()
			needsRefresh := bibleCache.populated && time.Since(bibleCache.timestamp) > refreshInterval
			bibleCache.RUnlock()

			if needsRefresh {
				log.Println("[CACHE] Background refresh starting...")
				invalidateBibleCache()
				getCachedBibles()
			}
		}
	}()
}

// BibleInfo describes a Bible for the index page.
type BibleInfo struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Abbrev        string   `json:"abbrev"`
	Language      string   `json:"language"`
	Versification string   `json:"versification"`
	BookCount     int      `json:"book_count"`
	Features      []string `json:"features"`
	CapsulePath   string   `json:"capsule_path"`
}

// BookInfo describes a book in a Bible.
type BookInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Order        int    `json:"order"`
	ChapterCount int    `json:"chapter_count"`
	Testament    string `json:"testament"`
}

// ChapterInfo describes a chapter.
type ChapterInfo struct {
	Number     int `json:"number"`
	VerseCount int `json:"verse_count"`
}

// VerseData represents a single verse.
type VerseData struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
}

// ChapterData contains the verses of a chapter.
type ChapterData struct {
	BibleID string      `json:"bible_id"`
	Book    string      `json:"book"`
	Chapter int         `json:"chapter"`
	Verses  []VerseData `json:"verses"`
}

// SearchResult represents a search match.
type SearchResult struct {
	BibleID   string `json:"bible_id"`
	Reference string `json:"reference"`
	Book      string `json:"book"`
	Chapter   int    `json:"chapter"`
	Verse     int    `json:"verse"`
	Text      string `json:"text"`
}

// BibleIndexData is the data for the Bible index page.
type BibleIndexData struct {
	PageData
	Bibles        []BibleInfo
	AllBibles     []BibleInfo // All bibles for compare tab
	Languages     []string
	Features      []string
	Tab           string
	Query         string
	BibleID       string
	CaseSensitive bool
	WholeWord     bool
	Results       []SearchResult
	Total         int
	// Pagination
	Page           int
	PerPage        int
	TotalPages     int
	PerPageOptions []int
}

// BibleViewData is the data for viewing a single Bible.
type BibleViewData struct {
	PageData
	Bible BibleInfo
	Books []BookInfo
}

// BookViewData is the data for viewing a single book.
type BookViewData struct {
	PageData
	Bible    BibleInfo
	Book     BookInfo
	Chapters []ChapterInfo
}

// ChapterViewData is the data for viewing a chapter.
type ChapterViewData struct {
	PageData
	Bible   BibleInfo
	Book    BookInfo
	Chapter int
	Verses  []VerseData
	PrevURL string
	NextURL string
}

// SearchData is the data for the search page.
type SearchData struct {
	PageData
	Bibles  []BibleInfo
	Query   string
	BibleID string
	Results []SearchResult
	Total   int
}

// handleBibleIndex shows all available Bibles and handles search.
func handleBibleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/bible" && r.URL.Path != "/bible/" {
		// Route to specific Bible
		handleBibleRouting(w, r)
		return
	}

	allBibles := getCachedBibles()

	// Collect unique languages and features
	langMap := make(map[string]bool)
	featMap := make(map[string]bool)
	for _, b := range allBibles {
		if b.Language != "" {
			langMap[b.Language] = true
		}
		for _, f := range b.Features {
			featMap[f] = true
		}
	}

	var languages, features []string
	for l := range langMap {
		languages = append(languages, l)
	}
	for f := range featMap {
		features = append(features, f)
	}
	sort.Strings(languages)
	sort.Strings(features)

	// Handle tab parameter
	tab := r.URL.Query().Get("tab")

	// Handle search if query parameter is present
	query := r.URL.Query().Get("q")
	bibleID := r.URL.Query().Get("bible")
	caseSensitive := r.URL.Query().Get("case") == "1"
	wholeWord := r.URL.Query().Get("word") == "1"

	var results []SearchResult
	var total int

	if query != "" {
		if bibleID != "" {
			// Search specific Bible
			results, total = searchBible(bibleID, query, 100)
		} else {
			// Search all Bibles
			for _, b := range allBibles {
				r, t := searchBible(b.ID, query, 100-len(results))
				results = append(results, r...)
				total += t
				if len(results) >= 100 {
					break
				}
			}
		}
	}

	// Pagination (only for browse tab, not when searching)
	perPageOptions := []int{11, 22, 33, 44, 55, 66}
	perPage := 11 // default
	page := 1

	if perPageStr := r.URL.Query().Get("perPage"); perPageStr != "" {
		fmt.Sscanf(perPageStr, "%d", &perPage)
		// Validate perPage is one of the allowed options
		valid := false
		for _, opt := range perPageOptions {
			if perPage == opt {
				valid = true
				break
			}
		}
		if !valid {
			perPage = 11
		}
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		fmt.Sscanf(pageStr, "%d", &page)
		if page < 1 {
			page = 1
		}
	}

	// Calculate pagination for browse tab (not search)
	var paginatedBibles []BibleInfo
	totalPages := 1
	if query == "" && tab != "compare" {
		totalItems := len(allBibles)
		totalPages = (totalItems + perPage - 1) / perPage
		if page > totalPages {
			page = totalPages
		}
		if page < 1 {
			page = 1
		}

		start := (page - 1) * perPage
		end := start + perPage
		if end > totalItems {
			end = totalItems
		}
		if start < totalItems {
			paginatedBibles = allBibles[start:end]
		}
	} else {
		paginatedBibles = allBibles
	}

	data := BibleIndexData{
		PageData:       PageData{Title: "Bible"},
		Bibles:         paginatedBibles,
		AllBibles:      allBibles,
		Languages:      languages,
		Features:       features,
		Tab:            tab,
		Query:          query,
		BibleID:        bibleID,
		CaseSensitive:  caseSensitive,
		WholeWord:      wholeWord,
		Results:        results,
		Total:          total,
		Page:           page,
		PerPage:        perPage,
		TotalPages:     totalPages,
		PerPageOptions: perPageOptions,
	}

	if err := Templates.ExecuteTemplate(w, "bible_index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleBibleRouting routes requests to the appropriate handler.
func handleBibleRouting(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/bible/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	switch {
	case len(parts) == 1 && parts[0] != "":
		// /bible/{capsule}
		handleBibleView(w, r, parts[0])
	case len(parts) == 2:
		// /bible/{capsule}/{book}
		handleBookView(w, r, parts[0], parts[1])
	case len(parts) >= 3:
		// /bible/{capsule}/{book}/{chapter}
		handleChapterView(w, r, parts[0], parts[1], parts[2])
	default:
		http.Redirect(w, r, "/bible", http.StatusFound)
	}
}

// handleBibleView shows a single Bible's books.
func handleBibleView(w http.ResponseWriter, r *http.Request, capsuleID string) {
	bible, books, err := loadBibleWithBooks(capsuleID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Bible not found: %v", err), http.StatusNotFound)
		return
	}

	data := BibleViewData{
		PageData: PageData{Title: bible.Title},
		Bible:    *bible,
		Books:    books,
	}

	if err := Templates.ExecuteTemplate(w, "bible_view.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleBookView shows a book's chapters.
func handleBookView(w http.ResponseWriter, r *http.Request, capsuleID, bookID string) {
	bible, books, err := loadBibleWithBooks(capsuleID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Bible not found: %v", err), http.StatusNotFound)
		return
	}

	var book *BookInfo
	for _, b := range books {
		if strings.EqualFold(b.ID, bookID) {
			book = &b
			break
		}
	}
	if book == nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	chapters := make([]ChapterInfo, book.ChapterCount)
	for i := 0; i < book.ChapterCount; i++ {
		chapters[i] = ChapterInfo{Number: i + 1}
	}

	data := BookViewData{
		PageData: PageData{Title: fmt.Sprintf("%s - %s", book.Name, bible.Title)},
		Bible:    *bible,
		Book:     *book,
		Chapters: chapters,
	}

	if err := Templates.ExecuteTemplate(w, "bible_book.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleChapterView shows a chapter's verses.
func handleChapterView(w http.ResponseWriter, r *http.Request, capsuleID, bookID, chapterStr string) {
	var chapter int
	fmt.Sscanf(chapterStr, "%d", &chapter)
	if chapter < 1 {
		http.Error(w, "Invalid chapter", http.StatusBadRequest)
		return
	}

	bible, books, err := loadBibleWithBooks(capsuleID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Bible not found: %v", err), http.StatusNotFound)
		return
	}

	var book *BookInfo
	for _, b := range books {
		if strings.EqualFold(b.ID, bookID) {
			book = &b
			break
		}
	}
	if book == nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	verses, err := loadChapterVerses(capsuleID, bookID, chapter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load chapter: %v", err), http.StatusInternalServerError)
		return
	}

	// Build prev/next URLs
	var prevURL, nextURL string
	if chapter > 1 {
		prevURL = fmt.Sprintf("/bible/%s/%s/%d", capsuleID, bookID, chapter-1)
	}
	if chapter < book.ChapterCount {
		nextURL = fmt.Sprintf("/bible/%s/%s/%d", capsuleID, bookID, chapter+1)
	}

	data := ChapterViewData{
		PageData: PageData{Title: fmt.Sprintf("%s %d - %s", book.Name, chapter, bible.Title)},
		Bible:    *bible,
		Book:     *book,
		Chapter:  chapter,
		Verses:   verses,
		PrevURL:  prevURL,
		NextURL:  nextURL,
	}

	if err := Templates.ExecuteTemplate(w, "bible_chapter.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleBibleCompare redirects to /bible?tab=compare
func handleBibleCompare(w http.ResponseWriter, r *http.Request) {
	// Build redirect URL preserving query parameters
	redirectURL := "/bible?tab=compare"
	if ref := r.URL.Query().Get("ref"); ref != "" {
		redirectURL += "&ref=" + ref
	}
	if bibles := r.URL.Query().Get("bibles"); bibles != "" {
		redirectURL += "&bibles=" + bibles
	}
	http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
}

// handleBibleSearch shows the search page and handles search requests.
func handleBibleSearch(w http.ResponseWriter, r *http.Request) {
	bibles := getCachedBibles()

	query := r.URL.Query().Get("q")
	bibleID := r.URL.Query().Get("bible")

	var results []SearchResult
	var total int

	if query != "" && bibleID != "" {
		results, total = searchBible(bibleID, query, 100)
	}

	data := SearchData{
		PageData: PageData{Title: "Search Bible"},
		Bibles:   bibles,
		Query:    query,
		BibleID:  bibleID,
		Results:  results,
		Total:    total,
	}

	if err := Templates.ExecuteTemplate(w, "bible_search.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleAPIBibles returns JSON list of Bibles.
func handleAPIBibles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/bibles")
	path = strings.Trim(path, "/")

	if path == "" {
		// List all Bibles
		bibles := getCachedBibles()
		json.NewEncoder(w).Encode(bibles)
		return
	}

	parts := strings.Split(path, "/")
	capsuleID := parts[0]

	switch {
	case len(parts) == 1:
		// Get Bible info with books
		bible, books, err := loadBibleWithBooks(capsuleID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"bible": bible,
			"books": books,
		})
	case len(parts) == 2:
		// Get book info
		bookID := parts[1]
		bible, books, err := loadBibleWithBooks(capsuleID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		for _, book := range books {
			if strings.EqualFold(book.ID, bookID) {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"bible": bible,
					"book":  book,
				})
				return
			}
		}
		http.Error(w, "Book not found", http.StatusNotFound)
	case len(parts) >= 3:
		// Get chapter verses
		bookID := parts[1]
		var chapter int
		fmt.Sscanf(parts[2], "%d", &chapter)

		verses, err := loadChapterVerses(capsuleID, bookID, chapter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(ChapterData{
			BibleID: capsuleID,
			Book:    bookID,
			Chapter: chapter,
			Verses:  verses,
		})
	}
}

// handleAPIBibleSearch handles search API requests.
func handleAPIBibleSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query().Get("q")
	bibleID := r.URL.Query().Get("bible")
	limitStr := r.URL.Query().Get("limit")

	limit := 100
	if limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	if query == "" || bibleID == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []SearchResult{},
			"total":   0,
			"error":   "query and bible parameters required",
		})
		return
	}

	results, total := searchBible(bibleID, query, limit)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"total":   total,
	})
}

// listBiblesUncached returns all Bible capsules without caching.
// Uses goroutines for parallel processing.
func listBiblesUncached() []BibleInfo {
	capsules := listCapsules()
	if len(capsules) == 0 {
		return nil
	}

	type result struct {
		bible BibleInfo
		ok    bool
	}

	// Create and start worker pool
	pool := NewWorkerPool[CapsuleInfo, result](maxWorkers, len(capsules))
	pool.Start(func(c CapsuleInfo) result {
		irContent, err := readIRContent(filepath.Join(ServerConfig.CapsulesDir, c.Path))
		if err != nil {
			return result{ok: false}
		}

		corpus := parseIRToCorpus(irContent)
		if corpus == nil || corpus.ModuleType != ir.ModuleBible {
			return result{ok: false}
		}

		capsuleID := strings.TrimSuffix(c.Name, ".capsule.tar.xz")
		capsuleID = strings.TrimSuffix(capsuleID, ".tar.xz")
		capsuleID = strings.TrimSuffix(capsuleID, ".tar.gz")
		capsuleID = strings.TrimSuffix(capsuleID, ".tar")

		bible := BibleInfo{
			ID:            capsuleID,
			Title:         corpus.Title,
			Abbrev:        corpus.ID,
			Language:      corpus.Language,
			Versification: corpus.Versification,
			BookCount:     len(corpus.Documents),
			CapsulePath:   c.Path,
		}

		// Check for Strong's numbers
		for _, doc := range corpus.Documents {
			for _, cb := range doc.ContentBlocks {
				for _, tok := range cb.Tokens {
					if len(tok.Strongs) > 0 {
						bible.Features = append(bible.Features, "Strong's Numbers")
						break
					}
				}
				if len(bible.Features) > 0 {
					break
				}
			}
			if len(bible.Features) > 0 {
				break
			}
		}

		return result{bible: bible, ok: true}
	})

	// Submit jobs
	for _, c := range capsules {
		pool.Submit(c)
	}
	pool.Close()

	// Collect results
	var bibles []BibleInfo
	for r := range pool.Results() {
		if r.ok {
			bibles = append(bibles, r.bible)
		}
	}

	sort.Slice(bibles, func(i, j int) bool {
		return bibles[i].Title < bibles[j].Title
	})

	return bibles
}

// loadBibleWithBooks loads a Bible and its books from a capsule.
// Uses corpus cache for better performance.
func loadBibleWithBooks(capsuleID string) (*BibleInfo, []BookInfo, error) {
	corpus, capsulePath, err := getCachedCorpus(capsuleID)
	if err != nil {
		return nil, nil, err
	}

	bible := &BibleInfo{
		ID:            capsuleID,
		Title:         corpus.Title,
		Abbrev:        corpus.ID,
		Language:      corpus.Language,
		Versification: corpus.Versification,
		BookCount:     len(corpus.Documents),
		CapsulePath:   capsulePath,
	}

	var books []BookInfo
	for _, doc := range corpus.Documents {
		testament := "OT"
		if isNewTestament(doc.ID) {
			testament = "NT"
		}

		chapterCount := countChapters(doc)

		books = append(books, BookInfo{
			ID:           doc.ID,
			Name:         doc.Title,
			Order:        doc.Order,
			ChapterCount: chapterCount,
			Testament:    testament,
		})
	}

	sort.Slice(books, func(i, j int) bool {
		return books[i].Order < books[j].Order
	})

	return bible, books, nil
}

// loadChapterVerses loads verses for a specific chapter.
// Uses corpus cache for better performance.
func loadChapterVerses(capsuleID, bookID string, chapter int) ([]VerseData, error) {
	corpus, _, err := getCachedCorpus(capsuleID)
	if err != nil {
		return nil, err
	}

	// Find the book
	var doc *ir.Document
	for _, d := range corpus.Documents {
		if strings.EqualFold(d.ID, bookID) {
			doc = d
			break
		}
	}
	if doc == nil {
		return nil, fmt.Errorf("book not found: %s", bookID)
	}

	// Extract verses for the chapter
	var verses []VerseData
	verseRe := regexp.MustCompile(`^` + regexp.QuoteMeta(doc.ID) + `\.(\d+)\.(\d+)`)

	for _, cb := range doc.ContentBlocks {
		// Try to parse verse reference from content block ID
		matches := verseRe.FindStringSubmatch(cb.ID)
		if len(matches) == 3 {
			var cbChapter, cbVerse int
			fmt.Sscanf(matches[1], "%d", &cbChapter)
			fmt.Sscanf(matches[2], "%d", &cbVerse)

			if cbChapter == chapter {
				verses = append(verses, VerseData{
					Number: cbVerse,
					Text:   cb.Text,
				})
			}
		}
	}

	// Sort by verse number
	sort.Slice(verses, func(i, j int) bool {
		return verses[i].Number < verses[j].Number
	})

	return verses, nil
}

// searchBible searches for text in a Bible.
// Uses corpus cache for better performance.
func searchBible(bibleID, query string, limit int) ([]SearchResult, int) {
	corpus, _, err := getCachedCorpus(bibleID)
	if err != nil {
		return nil, 0
	}

	var results []SearchResult
	total := 0
	queryLower := strings.ToLower(query)

	// Check if it's a phrase search (quoted)
	isPhrase := strings.HasPrefix(query, "\"") && strings.HasSuffix(query, "\"")
	if isPhrase {
		query = strings.Trim(query, "\"")
		queryLower = strings.ToLower(query)
	}

	// Check if it's a Strong's number search
	strongsRe := regexp.MustCompile(`^[HG]\d+$`)
	isStrongs := strongsRe.MatchString(strings.ToUpper(query))

	for _, doc := range corpus.Documents {
		for _, cb := range doc.ContentBlocks {
			var matched bool

			if isStrongs {
				// Search for Strong's number in tokens
				for _, tok := range cb.Tokens {
					for _, s := range tok.Strongs {
						if strings.EqualFold(s, query) {
							matched = true
							break
						}
					}
					if matched {
						break
					}
				}
			} else if isPhrase {
				// Exact phrase search
				matched = strings.Contains(strings.ToLower(cb.Text), queryLower)
			} else {
				// Word search
				matched = strings.Contains(strings.ToLower(cb.Text), queryLower)
			}

			if matched {
				total++
				if len(results) < limit {
					// Parse reference from content block ID
					chapter, verse := parseContentBlockRef(cb.ID, doc.ID)
					results = append(results, SearchResult{
						BibleID:   bibleID,
						Reference: fmt.Sprintf("%s %d:%d", doc.Title, chapter, verse),
						Book:      doc.ID,
						Chapter:   chapter,
						Verse:     verse,
						Text:      cb.Text,
					})
				}
			}
		}
	}

	return results, total
}

// parseIRToCorpus converts raw IR JSON to a Corpus.
func parseIRToCorpus(irContent map[string]interface{}) *ir.Corpus {
	data, err := json.Marshal(irContent)
	if err != nil {
		return nil
	}

	var corpus ir.Corpus
	if err := json.Unmarshal(data, &corpus); err != nil {
		return nil
	}

	return &corpus
}

// countChapters counts the number of chapters in a document.
func countChapters(doc *ir.Document) int {
	chapters := make(map[int]bool)
	re := regexp.MustCompile(`\.(\d+)\.`)

	for _, cb := range doc.ContentBlocks {
		matches := re.FindStringSubmatch(cb.ID)
		if len(matches) >= 2 {
			var ch int
			fmt.Sscanf(matches[1], "%d", &ch)
			chapters[ch] = true
		}
	}

	return len(chapters)
}

// isNewTestament checks if a book ID is from the New Testament.
func isNewTestament(bookID string) bool {
	ntBooks := map[string]bool{
		"Matt": true, "Mark": true, "Luke": true, "John": true,
		"Acts": true, "Rom": true, "1Cor": true, "2Cor": true,
		"Gal": true, "Eph": true, "Phil": true, "Col": true,
		"1Thess": true, "2Thess": true, "1Tim": true, "2Tim": true,
		"Titus": true, "Phlm": true, "Heb": true, "Jas": true,
		"1Pet": true, "2Pet": true, "1John": true, "2John": true,
		"3John": true, "Jude": true, "Rev": true,
	}
	return ntBooks[bookID]
}

// parseContentBlockRef parses chapter and verse from a content block ID.
func parseContentBlockRef(cbID, bookID string) (int, int) {
	re := regexp.MustCompile(regexp.QuoteMeta(bookID) + `\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(cbID)
	if len(matches) >= 3 {
		var chapter, verse int
		fmt.Sscanf(matches[1], "%d", &chapter)
		fmt.Sscanf(matches[2], "%d", &verse)
		return chapter, verse
	}
	return 1, 1
}
