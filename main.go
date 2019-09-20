package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	header = fmt.Sprintf(`<html><body><a href="/">home</a><br>`)
	footer = fmt.Sprintf("</body></html>")
)

func main() {
	dir := checkArgs()
	doneFileSearch := fileSearch(dir)
	<-doneFileSearch

	doneCreatingSetOfRecipes := createSetOfRecipes()
	doneSortingListOfRecipes := sortRecipes()

	<-doneCreatingSetOfRecipes
	<-doneSortingListOfRecipes

	// setup web server
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":1234", nil))
}

func checkArgs() string {
	if len(os.Args) < 2 {
		log.Fatalf("syntax error: %s <path where recursive search should start>\n", os.Args[0])
	}
	dir := os.Args[1]

	if info, err := os.Stat(dir); os.IsNotExist(err) || !info.IsDir() {
		log.Fatal("error: `", dir, "` is not a valid path to a directory.")
	}

	d, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return dir
	}
	return d
}
