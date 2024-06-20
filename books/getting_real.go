package books

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

type GettingRealBook Book

const baseURL = "https://basecamp.com"

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

func chapterTitleToNumber(title string) int {
	splitTitle := strings.Split(title, " ")
	converted, _ := strconv.Atoi(splitTitle[1])
	return converted
}

func buildChapters(s *Section, c *colly.Collector, e *colly.HTMLElement) {
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

func (book *GettingRealBook) BuildContent() {
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
			buildChapters(&section, c, e)
			book.Summary.Sections[sectionCount] = section
		})
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	c.Visit(fmt.Sprintf("%s%s", baseURL, "/gettingreal"))
}
