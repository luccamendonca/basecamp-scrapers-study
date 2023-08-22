package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/repr"
	"github.com/bmaupin/go-epub"
	"github.com/gocolly/colly"
)

var sel = map[string]string{
	"mainContent":             "main#main > div.content",
	"section":                 "div.toc__part",
	"sectionTitle":            "h2.toc__part-title",
	"sectionChapters":         "ul.toc__chapters",
	"chapter":                 "li.toc__chapter",
	"chapterListNumber":       "p.toc__chapter-number",
	"chapterListAnchor":       "h3.toc__chapter-title > a",
	"chapterPageHeader":       "div.intro__content",
	"chapterPageHeaderNumber": "p.intro__masthead",
	"chapterPageHeaderTitle":  "h1.intro__title",
	"chapterPageContent":      "main#main > div.content",
	"chapterPageRemoveNav":    "nav.pagination",
	"chapterPageRemoveFooter": "footer.footer",
}

const baseURL = "https://basecamp.com/"

type Book struct {
	Summary     Summary
	Title       string
	Authors     []string
	Description string
}

type Summary struct {
	Sections map[int]Section
}

type Section struct {
	Number   int
	Title    string
	Chapters map[int]Chapter
}

type Chapter struct {
	Number  int
	Title   string
	Content string
	URL     string
}

func chapterTitleToNumber(title string) int {
	splitTitle := strings.Split(title, " ")
	converted, _ := strconv.Atoi(splitTitle[1])
	return converted
}

func BuildChapter(chapter *Chapter, c *colly.Collector, e *colly.HTMLElement) {
	chapter.Title = e.ChildText(sel["chapterListAnchor"])
	chapter.Number = chapterTitleToNumber(e.ChildText(sel["chapterListNumber"]))
	chapter.URL = fmt.Sprintf("%s%s", baseURL, e.ChildAttr(sel["chapterListAnchor"], "href"))

	c.OnHTML(sel["chapterPageContent"], func(e *colly.HTMLElement) {
		e.ForEach(sel["chapterPageRemoveNav"], func(_ int, e *colly.HTMLElement) {
			e.DOM.Remove()
		})
		e.ForEach(sel["chapterPageRemoveFooter"], func(_ int, e *colly.HTMLElement) {
			e.DOM.Remove()
		})

		chapter.Content, _ = e.DOM.Html()
	})

	c.Visit(chapter.URL)
}

func (s *Section) buildChapters(c *colly.Collector, e *colly.HTMLElement) {
	s.Chapters = make(map[int]Chapter)
	colChapter := c.Clone()
	e.ForEach(sel["chapter"], func(_ int, e *colly.HTMLElement) {
		chapter := Chapter{}
		BuildChapter(&chapter, colChapter, e)
		s.Chapters[chapter.Number] = chapter
	})

	repr.Println(s)
}

func convertToFilename(s string) string {
	return strings.ReplaceAll(s, " ", "_")
}

func createEpub(b *Book) *epub.Epub {
	e := epub.NewEpub(b.Title)
	e.SetAuthor(strings.Join(b.Authors[:], ", "))
	e.SetDescription(b.Description)

	for _, section := range b.Summary.Sections {
		sFilename := convertToFilename(section.Title)
		e.AddSection("", section.Title, sFilename, "")
		for _, chapter := range section.Chapters {
			cFilename := convertToFilename(chapter.Title)
			e.AddSubSection(sFilename, chapter.Content, chapter.Title, cFilename, "")
		}
	}

	return e
}

func main() {
	var c = colly.NewCollector(colly.AllowedDomains("basecamp.com"))

	var book = Book{}
	book.Summary = Summary{}
	book.Summary.Sections = make(map[int]Section)

	i := 1
	// c.OnHTML(sel["section"], func(e *colly.HTMLElement) {
	c.OnHTML(sel["section"], func(e *colly.HTMLElement) {
		section := Section{}
		section.Title = e.ChildText(sel["sectionTitle"])
		section.Number = i

		section.buildChapters(c, e)
		book.Summary.Sections[i] = section

		if i > 1 {
			epub := createEpub(&book)
			err := epub.Write("getting_real.epub")

			repr.Println(err)

			return
		}

		i++
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// repr.Println(summary)

	c.Visit(fmt.Sprintf("%s%s", baseURL, "/gettingreal"))
}
