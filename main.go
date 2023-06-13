package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/asg017/sqlite-vss/bindings/go"
	"github.com/bkono/vss-example/db"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	repo, err := db.New(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	ctx := context.Background()

	count, err := repo.CountArticles(ctx)
	if err != nil {
		log.Printf("listArticles: error: %+v\n", err)
	}
	log.Printf("listArticles: count: %d\n", count)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println()
		fmt.Printf("Enter a search term: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				os.Exit(0)
			}
			log.Printf("error reading input: %+v\n", err)
			os.Exit(1)
		}

		searched := repo.SearchHeadlines(ctx, text, 5)
		if len(searched) == 0 {
			fmt.Println("no results")
			continue
		}

		fmt.Println(">>> matches:")
		for _, article := range searched {
			fmt.Printf(">>>>>> Headline: %s - Distance %f\n", article.Headline, article.Distance)
		}
	}
}

/*
This is the alternative approach, instead of using the Makefile. I've opted for the Makefile to allow the homebrew prefix envvar to be used for the libomp library.

// #cgo linux,amd64 LDFLAGS: -L./extensions -Wl,-undefined,dynamic_lookup -lstdc++
// #cgo darwin,amd64 LDFLAGS: -L./extensions -Wl,-undefined,dynamic_lookup -lomp
// #cgo darwin,arm64 LDFLAGS: -L/opt/homebrew/opt/libomp/lib -L./extensions -Wl,-undefined,dynamic_lookup -lomp
import "C"
*/
