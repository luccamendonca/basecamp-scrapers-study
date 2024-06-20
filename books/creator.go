package books

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bmaupin/go-epub"
	"golang.org/x/exp/slices"
)

func convertToFilename(s string) string {
	return strings.ReplaceAll(s, " ", "_")
}

func CreateEpubFromBook(b BuildableBook, filename string, coverImagePath string) error {
	e := epub.NewEpub(b.GetTitle())
	e.SetAuthor(strings.Join(b.Authors[:], ", "))
	e.SetDescription(b.Description)

	coverImageLocalPath, _ := e.AddImage(coverImagePath, "cover.png")
	e.SetCover(coverImageLocalPath, "")

	var sKeys []int
	for k := range b.Summary.Sections {
		sKeys = append(sKeys, k)
	}
	sort.Ints(sKeys)

	sections = b.GetSortedSections()

	for _, sk := range sKeys {
		section := b.Summary.Sections[sk]
		sFilename := convertToFilename(section.Title)
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
