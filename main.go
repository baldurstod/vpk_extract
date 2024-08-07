package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hash/crc32"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/baldurstod/go-vpk"
	glob "github.com/ganbarodigital/go_glob"
)

const crcFilename = "vpk_extract.crc.json"

var sigchan = make(chan os.Signal, 1)

type fileCRC struct {
	crcs map[string]interface{}
}

func main() {
	var inputFile string
	var outputFolder string
	var command string
	var sleep int

	signal.Notify(sigchan, os.Interrupt)

	flag.StringVar(&inputFile, "i", "", "Input VPK")
	flag.StringVar(&outputFolder, "o", "", "Output folder")
	flag.StringVar(&command, "c", "extract", "Command: can be extract or crc")
	flag.IntVar(&sleep, "s", 0, "Sleep x milliseconds after each file written")
	flag.Parse()

	if (command == "extract") && (inputFile == "") {
		fmt.Println("No input file provided. Use the flag -i")
		os.Exit(1)
	}
	if outputFolder == "" {
		fmt.Println("No output folder provided. Use the flag -o")
		os.Exit(1)
	}

	globPatterns := flag.Args()
	if len(globPatterns) == 0 {
		globPatterns = []string{"*"}
	}

	switch command {
	case "extract":
		extractVPK(inputFile, outputFolder, globPatterns, time.Duration(sleep)*time.Millisecond)
	case "crc":
		generateCRCFile(outputFolder)
	}
}

func extractVPK(inputFile string, outputFolder string, globPatterns []string, sleep time.Duration) {
	var pak vpk.VPK
	var err error
	fileCRC := fileCRC{}
	fileCRC.init()

	crcPath := path.Join(outputFolder, crcFilename)

	crcFileContent, err := os.ReadFile(crcPath)
	if err == nil {
		_ = json.Unmarshal(crcFileContent, &fileCRC.crcs)
	}

	if strings.HasSuffix(inputFile, "_dir.vpk") {
		pak, err = vpk.OpenDir(inputFile)
	} else {
		pak, err = vpk.OpenSingle(inputFile)
	}
	if err != nil {
		panic(err)
	}
	defer pak.Close()

	// Prepare the globs
	globs := []*glob.Glob{}
	for _, globPattern := range globPatterns {
		globs = append(globs, glob.NewGlob(globPattern))
	}

	// Iterate through all files in the VPK
mainLoop:
	for _, entry := range pak.Entries() {
		fileName := entry.Filename()
		extractName := path.Join(outputFolder, fileName)
		for _, g := range globs {

			match, err := g.Match(fileName)
			if err != nil {
				panic(err)
			}

			if match {
				entryCRC := entry.CRC()

				if crc, exist := fileCRC.getCRC(fileName); !exist || crc != entryCRC {
					if fileReader, error := entry.Open(); error == nil {

						err := os.MkdirAll(path.Join(outputFolder, entry.Path()), 0755)
						if err != nil && !os.IsExist(err) {
							fmt.Println(err)
						}

						buf, _ := ioutil.ReadAll(fileReader)
						fileCRC.addFile(fileName, entryCRC)
						fmt.Println(extractName)
						error := os.WriteFile(extractName, buf, 0666)
						if error != nil {
							fmt.Println(error)
						}
						time.Sleep(sleep)
					}
				}
				break
			}
		}
		select {
		case <-sigchan:
			fmt.Println("Interrupted, exiting...")
			break mainLoop
		default:
		}
	}

	j, _ := json.Marshal(&fileCRC.crcs)
	os.WriteFile(crcPath, j, 0666)
}

func generateCRCFile(outputFolder string) {
	fileCRC := fileCRC{}
	fileCRC.init()

	crcPath := path.Join(outputFolder, crcFilename)

	e := filepath.WalkDir(outputFolder, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}
		if strings.HasSuffix(path, crcFilename) { // Skip vpk_extract.crc.json
			return nil
		}

		if !info.IsDir() {
			rel, _ := filepath.Rel(outputFolder, path)
			rel = filepath.ToSlash(rel)

			fileContent, e := os.ReadFile(path)
			if e != nil {
				fmt.Println(path, rel, e)
			} else {
				crc := crc32.ChecksumIEEE(fileContent)
				fileCRC.addFile(rel, crc)
				fmt.Println(path, rel, crc)
			}
		}
		return nil
	})

	if e != nil {
		log.Fatal(e)
	}

	j, _ := json.Marshal(&fileCRC.crcs)
	os.WriteFile(crcPath, j, 0666)
}

func (this *fileCRC) init() {
	this.crcs = make(map[string]interface{})
}

func (this *fileCRC) addFile(relativePath string, crc uint32) {
	path := strings.Split(relativePath, "/")

	current := this.crcs

	for index, p := range path {
		if index == len(path)-1 {
			current[p] = crc
		} else {
			//fmt.Println(index)
			next, exist := current[p]
			if !exist {
				next = make(map[string]interface{})
				current[p] = next
			}
			current = (next).(map[string]interface{})
		}
	}
}

func (this *fileCRC) getCRC(relativePath string) (uint32, bool) {
	path := strings.Split(relativePath, "/")

	current := this.crcs

	for index, p := range path {
		//fmt.Println(index)
		next, exist := current[p]
		if index == len(path)-1 {
			if exist {
				switch next.(type) {
				case uint32:
					return next.(uint32), true
				case float64:
					return uint32(next.(float64)), true
				default:
					fmt.Println(next)
					panic("Unknown type")
				}
			} else {
				return 0, false
			}
		}
		if !exist {
			next = make(map[string]interface{})
			current[p] = next
		}
		current = (next).(map[string]interface{})
	}
	return 0, false
}
