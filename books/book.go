package books

import (
	"sort"
	"strings"
)

type Book struct {
	Summary     Summary  `json:"summary"`
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	Description string   `json:"description"`
}

type Summary struct {
	Sections map[int]Section `json:"sections"`
}

type Section struct {
	Number   int             `json:"number"`
	Title    string          `json:"title"`
	Chapters map[int]Chapter `json:"chapters"`
}

type Chapter struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

type BuildableBook interface {
	BuildContent()
	GetTitle() string
	GetAuthors() []string
	GetAuthorsJoined() string
	GetDescription() string
	GetSortedSections() []Section
	GetSortedChaptersFromSection(sK int) []Chapter
}

func (b *Book) BuildContent() {
	panic("implement me, plz")
}

func (b *Book) GetTitle() string {
	return b.Title
}

func (b *Book) GetAuthors() []string {
	return b.Authors
}

func (b *Book) GetAuthorsJoined(separator string) string {
	return strings.Join(b.Authors[:], separator)
}

func (b *Book) GetDescription() string {
	return b.Description
}

// func sortSlice(input map[int]interface{}) []interface{} {
// 	var keys []int
// 	for k := range input {
// 		keys = append(keys, k)
// 	}
// 	sort.Ints(keys)
// 	sorted := make([]interface{}, len(keys))
// 	for k := range keys {
// 		sorted = append(sorted, input[k])
// 	}

// 	return sorted
// }

func (b *Book) GetSortedSections() []Section {
	var keys []int

	for k := range b.Summary.Sections {
		keys = append(keys, k)
	}

	sort.Ints(keys)
	sorted := make([]Section, len(keys))
	for k := range keys {
		sorted = append(sorted, b.Summary.Sections[k])
	}

	return sorted
}

func (b *Book) GetSortedChaptersFromSection(sK int) []Chapter {
	var keys []int
	sectionChapters := b.Summary.Sections[sK].Chapters

	for k := range sectionChapters {
		keys = append(keys, k)
	}

	sort.Ints(keys)
	sorted := make([]Chapter, len(keys))
	for k := range keys {
		sorted = append(sorted, sectionChapters[k])
	}

	return sorted
}

func NewBook(URL string) BuildableBook {
	switch URL {
	case "https://basecamp.com/gettingreal":
		return &GettingRealBook{Authors: []string{"Basecamp", "37Signals"}}
	case "https://basecamp.com/shapeup":
		return &Book{}
	default:
		return &Book{}
	}
}
