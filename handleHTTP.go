package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

func handler(w http.ResponseWriter, req *http.Request) {
	uri := req.URL.RequestURI()
	uri, err := url.QueryUnescape(uri)
	if err != nil {
		log.Println("could not unescape URI ", req.URL.RequestURI(), err)
		return
	}

	if uri == "/" {
		listRecipes(w)
		return
	}

	if setOfRecipes[uri] {
		printRecipeContent(w, uri)
		return
	}

	bbn := uri[1:]
	done := listRecipeGivenByFilename(w, bbn, mapFilenameWithExtToPath)
	if done {
		return
	}

	done = listRecipeGivenByFilename(w, bbn, mapFilenameWithoutExtToPath)
	if done {
		return
	}

	io.WriteString(w, fmt.Sprint("Could not find ", uri))
}

func listRecipeGivenByFilename(w io.Writer, recipename string, lookupTable map[string][]string) bool {
	if p, ok := lookupTable[recipename]; ok {
		if len(p) == 1 {
			printRecipeContent(w, p[0])
			return true
		}
		io.WriteString(w, "<ul>")
		for _, e := range p {
			io.WriteString(w, `<li><a href="`)
			io.WriteString(w, e)
			io.WriteString(w, `">`)
			io.WriteString(w, e)
			io.WriteString(w, `</a></li>`)
		}
		io.WriteString(w, "</ul>")
		return true
	}
	return false
}

func listRecipes(w io.Writer) {
	io.WriteString(w, header)
	io.WriteString(w, "<ul>")
	for _, e := range sortedListOfRecipes {
		io.WriteString(w, `<li><a href="`+url.QueryEscape(e)+`">`+e+`</a></li>`)
	}
	io.WriteString(w, "</ul>")
	io.WriteString(w, footer)
}

func printRecipeContent(w io.Writer, recipename string) {
	io.WriteString(w, header)
	io.WriteString(w, `<pre>`)

	s := lut[recipename]

	// replace recipes occurring in DEPENDS, require, inherit statements
	reader := strings.NewReader(s)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		origline := scanner.Text() + "\n"
		line := strings.Trim(origline, "\t \n")
		if done := checkForPrefix(w, line, "inherit "); done {
			io.WriteString(w, "\n")
			continue
		}
		if done := checkForPrefix(w, line, "require "); done {
			io.WriteString(w, "\n")
			continue
		}
		if done := checkForPrefix(w, line, "DEPENDS "); done {
			io.WriteString(w, "\n")
			continue
		}
		io.WriteString(w, origline)
	}
	if err := scanner.Err(); err != nil {
		log.Println("could not read content of recipe ", recipename, err)
	}
	io.WriteString(w, `</pre>`)
	io.WriteString(w, footer)
}

func checkForPrefix(w io.Writer, line string, prefix string) bool {
	if strings.HasPrefix(line, prefix) {
		io.WriteString(w, prefix)
		req := line[len(prefix):]
		linereader := strings.NewReader(req)
		wordscanner := bufio.NewScanner(linereader)
		wordscanner.Split(bufio.ScanWords)
		for wordscanner.Scan() {
			recipename := wordscanner.Text()
			trimmed := strings.Trim(recipename, `"`)
			basename := filepath.Base(trimmed)
			withoutExt := basename[:len(basename)-len(filepath.Ext(basename))]
			log.Println("basename: ", basename)
			log.Println("withoutExt: ", withoutExt)
			if _, ok := mapFilenameWithoutExtToPath[withoutExt]; ok {
				replaceRecipeNameByLink(w, recipename)
			} else if _, ok := mapFilenameWithExtToPath[basename]; ok {
				replaceRecipeNameByLink(w, recipename)
			} else {
				io.WriteString(w, recipename+" ")
			}
		}
		if err := wordscanner.Err(); err != nil {
			log.Printf("line %q could not split into words: %s", line, err)
		}
		return true
	}
	return false
}

func replaceRecipeNameByLink(w io.Writer, recipename string) {
	io.WriteString(w, `<span style="text-decoration: underline;cursor:pointer;color:blue;" `)
	io.WriteString(w, `onclick="window.location.href = '/`+recipename+`';" >`+recipename+`</span> `)
}
