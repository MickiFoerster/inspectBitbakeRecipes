package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func fileSearch(dir string) chan struct{} {
	ch := make(chan string)
	done := make(chan struct{})

	go func() {
		semaphore := make(chan struct{}, 32)
		for fn := range ch {
			semaphore <- struct{}{}
			readRecipe(fn, semaphore)
		}
		fmt.Println("Done with consuming file paths")
		done <- struct{}{}
	}()

	go func() {
		err := filepath.Walk(dir,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					fmt.Println("error while accessing", path, err)
					return err
				}
				if !info.IsDir() {
					if strings.HasSuffix(path, ".bb") || strings.HasSuffix(path, ".bbappend") || strings.HasSuffix(path, ".bbclass") {
						ch <- path
					}
				}
				return nil
			})
		if err != nil {
			log.Fatal("error while recursive search: ", err)
		}
		fmt.Println("End of recursive file search")
		close(ch)
	}()

	return done
}
