package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/fatih/color"
)

const (
	baseURL = "https://slack.com/api/search.files"
)

// Files ...
type Files struct {
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
		Matches []Files
	}
}

func getFiles(token *string, query *string, page int) (*FilesAPIResponse, error) {
	resp, err := http.Get(baseURL + "?token=" + *token + "&query=" + *query + "&page=" + strconv.Itoa(page))
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

func main() {
	cyan := color.New(color.FgCyan).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	token := flag.String("token", "", "Slack Authentication token (Requires scope: search:read).")
	query := flag.String("query", "", "Search query. May contain booleans, etc.")
	flag.Parse()

	if *token == "" {
		color.Red("Missing token")
		os.Exit(1)
	}

	if *query == "" {
		*query = "zip"
	}

	page := 1
	paging := true
	total := 0

	var files []Files

	for paging {
		fmt.Printf("Fetching page %d of %d with query %s\n", page, total, *query)
		data, err := getFiles(token, query, page)
		if err != nil {
			panic(err)
		}

		total = data.Files.Paging.Pages
		paging = data.Files.Paging.Pages > data.Files.Paging.Page
		page = data.Files.Paging.Page + 1

		files = append(files, data.Files.Matches...)
	}

	for _, file := range files {
		sizeMB := float64(file.Size) / 1024 / 1024
		fmt.Printf("%s - %s\n", cyan(floatToString(sizeMB)+" MB"), white(file.Name))
	}
}
