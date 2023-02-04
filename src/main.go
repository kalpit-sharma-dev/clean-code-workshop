package main

import (
	"clean-code-workshop/src/constants"
	"crypto/sha1"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync/atomic"
)

func traverseDir(hashes, duplicates map[string]string, dupeSize *int64, entries []os.FileInfo, directory string) {
	for _, entry := range entries {
		fullpath := (path.Join(directory, entry.Name()))

		if !entry.Mode().IsDir() && !entry.Mode().IsRegular() {
			continue
		}

		if entry.IsDir() {
			dirFiles, err := ioutil.ReadDir(fullpath)
			if err != nil {
				panic(err)
			}
			traverseDir(hashes, duplicates, dupeSize, dirFiles, fullpath)
			continue
		}
		file, err := ioutil.ReadFile(fullpath)
		if err != nil {
			panic(err)
		}
		saveHash(file, fullpath, hashes, duplicates, dupeSize, entry)

	}
}

func saveHash(file []byte, fullpath string, hashes, duplicates map[string]string, dupeSize *int64, entry fs.FileInfo) {
	hash := sha1.New()
	if _, err := hash.Write(file); err != nil {
		panic(err)
	}
	hashSum := hash.Sum(nil)
	hashString := fmt.Sprintf("%x", hashSum)
	if hashEntry, ok := hashes[hashString]; ok {
		duplicates[hashEntry] = fullpath
		atomic.AddInt64(dupeSize, entry.Size())
	} else {
		hashes[hashString] = fullpath
	}
}

func toReadableSize(nbytes int64) string {

	if nbytes > constants.TB {
		return strconv.FormatInt(nbytes/(1000*1000*1000*1000), 10) + constants.TBString
	}
	if nbytes > constants.GB {
		return strconv.FormatInt(nbytes/(1000*1000*1000), 10) + constants.GBString
	}
	if nbytes > constants.MB {
		return strconv.FormatInt(nbytes/(1000*1000), 10) + constants.MBString
	}
	if nbytes >= constants.KB {
		return strconv.FormatInt(nbytes/1000, 10) + constants.KBString
	}

	return strconv.FormatInt(nbytes, 10) + constants.ByteString
}

func main() {
	var err error
	dir := flag.String("path", "", "the path to traverse searching for duplicates")
	flag.Parse()

	if *dir == "" {
		*dir, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}

	hashes := map[string]string{}
	duplicates := map[string]string{}
	var dupeSize int64

	entries, err := ioutil.ReadDir(*dir)
	if err != nil {
		panic(err)
	}

	traverseDir(hashes, duplicates, &dupeSize, entries, *dir)

	fmt.Println("DUPLICATES")

	fmt.Println("TOTAL FILES:", len(hashes))
	fmt.Println("DUPLICATES:", len(duplicates))
	fmt.Println("TOTAL DUPLICATE SIZE:", toReadableSize(dupeSize))
}

// running into problems of not being able to open directories inside .app folders
