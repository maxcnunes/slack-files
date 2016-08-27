package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

const (
	baseURL = "https://slack.com/api/search.files"
)

// File ...
type File struct {
	ID                 string
	Created            int
	Timestamp          int
	Name               string
	Title              string
	Mimetype           string
	Filetype           string
	PrettyType         string
	User               string
	Editable           bool
	Size               int
	Mode               string
	IsExternal         bool
	ExternalType       string
	IsPublic           bool
	PublicURLShared    bool
	DisplayAsBot       bool
	Username           string
	URLPrivate         string
	URLPrivateDownload string
	Permalink          string
	PermalinkPublic    string
	CommentsCount      int
}

// FilesAPIResponse ...
type FilesAPIResponse struct {
	OK           bool
	Query        string
	AllStopWords bool
	Files        struct {
		Total      int
		Pagination struct {
			TotalCount int
			Page       int
			PerPage    int
			PageCount  int
			First      int
			Last       int
		}
		Paging struct {
			Count int
			Total int
			Page  int
			Pages int
		}
		Matches []File
	}
}

func getFiles(token *string, query string, page int) (*FilesAPIResponse, error) {
	resp, err := http.Get(baseURL + "?token=" + *token + "&query=" + query + "&page=" + strconv.Itoa(page))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data := &FilesAPIResponse{}
	dec := json.NewDecoder(resp.Body)
	dec.Decode(data)

	return data, nil
}

func floatToString(num float64) string {
	return strconv.FormatFloat(num, 'f', 2, 64)
}

func getHumanSize(file File) string {
	size := float64(file.Size)

	humanSize := floatToString(size) + " B"
	size = size / 1024
	if size < 1 {
		return humanSize
	}

	humanSize = floatToString(size) + " KB"
	size = size / 1024
	if size < 1 {
		return humanSize
	}

	humanSize = floatToString(size) + " MB"
	size = size / 1024
	if size < 1 {
		return humanSize
	}

	return floatToString(size) + " GB"
}

func main() {
	cyan := color.New(color.FgCyan).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	token := flag.String("token", "", "Slack Authentication token.")
	query := flag.String("query", "", "Search query. Accept multiple values separated by \";\".")
	flag.Parse()

	if *token == "" {
		color.Red("Missing token")
		os.Exit(1)
	}

	if *query == "" {
		*query = ".rar;.tar;.zip;.mp3;.mp4;.pdf;.ppt;.csv;.jpg;.jpeg;.json"
	}

	var files []File

	queries := strings.Split(*query, ";")
	for _, q := range queries {
		page := 1
		paging := true

		fmt.Printf("Fetching files with query %s", q)
		for paging {
			data, err := getFiles(token, q, page)
			if err != nil {
				panic(err)
			}

			paging = data.Files.Paging.Pages > data.Files.Paging.Page
			page = data.Files.Paging.Page + 1

			files = append(files, data.Files.Matches...)

			if data.Files.Paging.Pages > 1 {
				if data.Files.Paging.Page == 1 {
					fmt.Printf(" Total pages: %d ", data.Files.Paging.Pages)
				}

				fmt.Printf(".")
			}
		}

		fmt.Printf("\n")
	}

	for _, file := range files {
		fmt.Printf("  %s - %s\n", cyan(getHumanSize(file)), white(file.Name))
	}
}
