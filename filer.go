package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	start string = "+++"
	end   string = "+++"
	title string = "[menu.wiki]"
)

var images = []string{".jpg", ".png", ".jpeg", ".bmp", ".jpg", ".svg"}

type fileInf struct {
	path string
	info os.FileInfo
}

var filesPath = "files"
var contentPath = "content"
var publicPath = "src/public"

func main() {
	var AuthCmd = &cobra.Command{
		Use:   "Filer",
		Short: "files helps organize User images and documents",
		Long: `
A command tool for copying files from the files directory
into the public directory, automatically builds the files directory
based so that there is a folder for every content article`,
		Run: func(cmd *cobra.Command, args []string) {
			walkContent()
		},
	}

	AuthCmd.PersistentFlags().StringVarP(&filesPath, "filesDir", "f", "files", "the path to the files directory")
	AuthCmd.PersistentFlags().StringVarP(&contentPath, "contentDir", "c", "content", "the path to the content directory")
	AuthCmd.PersistentFlags().StringVarP(&publicPath, "publicDir", "p", "public", "the path to the public directory")

	AuthCmd.Execute()

}

func walkContent() {
	contentList := make([]fileInf, 0)

	err := filepath.Walk(contentPath, func(path string, f os.FileInfo, err error) error {
		fileInfo := fileInf{
			path: path,
			info: f,
		}
		contentList = append(contentList, fileInfo)
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	for _, content := range contentList {
		extension := filepath.Ext(content.path)
		if !content.info.IsDir() {
			path := content.path[0 : len(content.path)-len(extension)]
			path = strings.Replace(path, contentPath, filesPath, 1)
			if err := os.MkdirAll(path, 0777); err != nil {
				log.Fatal(err)
			}

			fileList := make([]fileInf, 0)
			err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
				fileInfo := fileInf{
					path: path,
					info: f,
				}
				fileList = append(fileList, fileInfo)
				return nil
			})

			if err != nil {
				log.Fatal(err)
			}
			documents := make([]string, 0)
		outerloop:
			for _, file := range fileList {
				if file.info.IsDir() {
					continue
				}
				ext := filepath.Ext(file.path)
				for _, imageExt := range images {
					if ext == imageExt {
						break outerloop
					}
				}
				documents = append(documents, file.path)

			}
			writeFileToContent(content.path, documents)
		}
	}
}

func writeFileToContent(cPath string, documents []string) {
	documentline := strings.Join(documents, "\",\"")
	documentline = fmt.Sprintf("files=[\"%s\"]\n", documentline)
	lines, err := readLines(cPath)
	if err != nil {
		log.Fatal(err)
	}

	var withinToml bool = false
	var written bool = false
	// delete all entries within first
	for i, line := range lines {
		if strings.Contains(line, start) && !withinToml {
			withinToml = true
			continue
		}
		if strings.Contains(line, end) && withinToml {
			if !written {
				// then we need to insert a new line
				lines = append(lines[:i], append([]string{documentline}, lines[i:]...)...)
			}
			break
		}
		if strings.Contains(line, "files=") && withinToml {
			lines[i] = documentline
			written = true
		}
	}
	writeLines(cPath, lines)
}

// Utility Stuff

func readLines(file string) (lines []string, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		const delim = '\n'
		line, err := r.ReadString(delim)
		if err == nil || len(line) > 0 {
			if err != nil {
				line += string(delim)
			}
			lines = append(lines, line)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return lines, nil
}

func writeLines(file string, lines []string) (err error) {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	for _, line := range lines {
		_, err := w.WriteString(line)
		if err != nil {
			return err
		}
	}
	return nil
}
