package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"sync"
)

var (
	lut                         = make(map[string]string)
	muxLut                      sync.Mutex
	setOfRecipes                = make(map[string]bool)
	mapFilenameWithoutExtToPath = make(map[string][]string)
	mapFilenameWithExtToPath    = make(map[string][]string)
	sortedListOfRecipes         []string
)

func readRecipe(path string, semaphore chan struct{}) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("could not read file", path, err)
	} else {
		muxLut.Lock()
		lut[path] = string(b)
		muxLut.Unlock()
		log.Println("successfully read file", path)
	}
	<-semaphore
}

func createSetOfRecipes() chan struct{} {
	ch := make(chan struct{})
	go func() {
		for k := range lut {
			setOfRecipes[k] = true

			base := filepath.Base(k)
			ext := filepath.Ext(k)
			baseWithoutExt := base[:len(base)-len(ext)]

			paths := mapFilenameWithoutExtToPath[baseWithoutExt]
			paths = append(paths, k)
			mapFilenameWithoutExtToPath[baseWithoutExt] = paths

			paths = mapFilenameWithExtToPath[base]
			paths = append(paths, k)
			mapFilenameWithExtToPath[base] = paths
		}
		ch <- struct{}{}
	}()
	return ch
}

func sortRecipes() chan struct{} {
	ch := make(chan struct{})

	go func() {
		for k := range lut {
			sortedListOfRecipes = append(sortedListOfRecipes, k)
		}
		sort.Strings(sortedListOfRecipes)
		ch <- struct{}{}
	}()

	return ch
}
