package main

import (
	"luccamendonca/basecamp-scraper/books"

	"github.com/alecthomas/repr"
)

// func buildContentFromJson(book *Book) {
// 	bookBytes, _ := os.ReadFile("book.json")
// 	json.Unmarshal([]byte(bookBytes), &book)
// }

func main() {
	book := books.NewBook("https://basecamp.com/gettingreal")

	repr.Println(book)

	book.BuildContent()
	books.CreateEpubFromBook(book, "getting_real_crawler.epub", "getting_real_cover.png")

	// bookJson, _ := json.Marshal(book)
	// fmt.Println(string(bookJson[:]))

	// buildContentFromJson(book)
	// book.Authors = []string{"Basecamp", "37Signals"}
	// createEpub(book, "getting_real_json.epub")
}
