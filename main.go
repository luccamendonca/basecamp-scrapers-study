package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/bmaupin/go-epub"
	"github.com/gocolly/colly"
	"golang.org/x/exp/slices"
)

const baseURL = "https://basecamp.com"

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

var sel = map[string]string{
	"mainContent":             "main#main > div.content",
	"section":                 "div.toc__part",
	"sectionTitle":            "h2.toc__part-title",
	"chapterItem":             "li.toc__chapter",
	"chapterListNumber":       "p.toc__chapter-number",
	"chapterListAnchor":       "h3.toc__chapter-title > a",
	"chapterPageRemoveNav":    "nav.pagination",
	"chapterPageRemoveFooter": "footer.footer",
}

func createEpub(b *Book, filename string) error {
	e := epub.NewEpub(b.Title)
	e.SetAuthor(strings.Join(b.Authors[:], ", "))
	e.SetDescription(b.Description)

	coverImagePath, _ := e.AddImage("getting_real_cover.png", "cover.png")
	e.SetCover(coverImagePath, "")

	var sKeys []int
	for k := range b.Summary.Sections {
		sKeys = append(sKeys, k)
	}
	sort.Ints(sKeys)

	for _, sk := range sKeys {
		section := b.Summary.Sections[sk]
		sFilename := convertToFilename(section.Title)
		// sectionBody := fmt.Sprintf("<h1 style='text-align: center; margin-top: 50%%;'> %d. %s </h1>", section.Number, section.Title)
		sectionBody := fmt.Sprintf("<h1 style='text-align: center;'> %d. %s </h1>", section.Number, section.Title)
		e.AddSection(sectionBody, section.Title, sFilename, "")
		fmt.Printf(">> Add section %d: %s\n", sk, section.Title)

		var cKeys []int
		for k := range section.Chapters {
			cKeys = append(cKeys, k)
		}
		slices.Sort(cKeys)

		for _, ck := range cKeys {
			chapter := section.Chapters[ck]
			cFilename := convertToFilename(chapter.Title)
			chapterTitleH1 := fmt.Sprintf("<h1>Chapter %d: %s</h1>", chapter.Number, chapter.Title)
			chapterBody := fmt.Sprintf("%s%s", chapterTitleH1, chapter.Content)
			e.AddSubSection(sFilename, chapterBody, chapter.Title, cFilename, "")
			fmt.Printf(">>>> Add chapter %d: %s\n", ck, chapter.Title)
		}
	}

	return e.Write(filename)
}

func convertToFilename(s string) string {
	return strings.ReplaceAll(s, " ", "_")
}

func chapterTitleToNumber(title string) int {
	splitTitle := strings.Split(title, " ")
	converted, _ := strconv.Atoi(splitTitle[1])
	return converted
}

func (s *Section) buildChapters(c *colly.Collector, e *colly.HTMLElement) {
	s.Chapters = make(map[int]Chapter)
	colChapter := c.Clone()
	e.ForEach(sel["chapterItem"], func(_ int, e *colly.HTMLElement) {
		chapter := Chapter{
			Title:  e.ChildText(sel["chapterListAnchor"]),
			Number: chapterTitleToNumber(e.ChildText(sel["chapterListNumber"])),
			URL:    fmt.Sprintf("%s%s", baseURL, e.ChildAttr(sel["chapterListAnchor"], "href")),
		}
		colChapter.OnHTML(sel["mainContent"], func(e *colly.HTMLElement) {
			e.ForEach(sel["chapterPageRemoveNav"], func(_ int, e *colly.HTMLElement) {
				e.DOM.Remove()
			})
			e.ForEach(sel["chapterPageRemoveFooter"], func(_ int, e *colly.HTMLElement) {
				e.DOM.Remove()
			})
			chapter.Content, _ = e.DOM.Html()
		})
		colChapter.Visit(chapter.URL)
		s.Chapters[chapter.Number] = chapter
		fmt.Printf("-------- BUILD CHAPTER %d: %s\n", chapter.Number, chapter.Title)
	})
	colChapter.OnRequest(func(r *colly.Request) {
		fmt.Println("CHAPTER -- Visiting", r.URL.String())
	})
}

func buildContentFromCrawler(book *Book) {
	var c = colly.NewCollector(
		colly.AllowedDomains("basecamp.com"),
	)
	book.Summary = Summary{}
	book.Summary.Sections = make(map[int]Section)

	c.OnHTML(sel["mainContent"], func(rootElm *colly.HTMLElement) {
		book.Title = rootElm.ChildText("h1.landing-title")
		book.Description = rootElm.ChildText("p.landing-subtitle")
		rootElm.ForEach(sel["section"], func(sectionCount int, e *colly.HTMLElement) {
			sectionCount++
			section := Section{}
			section.Title = e.ChildText(sel["sectionTitle"])
			section.Number = sectionCount
			fmt.Printf("----- BUILD SECTION: %s\n", section.Title)
			section.buildChapters(c, e)
			book.Summary.Sections[sectionCount] = section
		})
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	c.Visit(fmt.Sprintf("%s%s", baseURL, "/gettingreal"))
}

func buildContentFromJson(book *Book) {
	bookBytes, _ := os.ReadFile("book.json")
	json.Unmarshal([]byte(bookBytes), &book)
}

func main() {
	book := &Book{Authors: []string{"Basecamp", "37Signals"}}

	buildContentFromCrawler(book)
	createEpub(book, "getting_real_crawler.epub")
	// bookJson, _ := json.Marshal(book)
	// fmt.Println(string(bookJson[:]))

	// buildContentFromJson(book)
	// book.Authors = []string{"Basecamp", "37Signals"}
	// createEpub(book, "getting_real_json.epub")
}
