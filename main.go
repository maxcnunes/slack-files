package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

const (
	baseURLSearch = "https://slack.com/api/search.files"
	baseURLList   = "https://slack.com/api/files.list"
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
	PrettyType         string `json:"pretty_type"`
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

// FilesSearchAPIResponse ...
type FilesSearchAPIResponse struct {
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

// FilesAPIResponse ...
type FilesAPIResponse struct {
	OK     bool
	Files  []File
	Paging struct {
		Count int
		Total int
		Page  int
		Pages int
	}
}

type files []File

func (f files) Len() int           { return len(f) }
func (f files) Less(i, j int) bool { return f[i].Size < f[j].Size }
func (f files) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func getFilesPerPage(token *string, query string, page int) (*FilesAPIResponse, error) {
	var url string
	if query != "" {
		url = baseURLSearch + "?token=" + *token + "&query=" + query + "&page=" + strconv.Itoa(page)
	} else {
		url = baseURLList + "?token=" + *token + "&page=" + strconv.Itoa(page)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if query == "" {
		data := &FilesAPIResponse{}
		dec := json.NewDecoder(resp.Body)
		dec.Decode(data)
		return data, nil
	}

	data := &FilesSearchAPIResponse{}
	dec := json.NewDecoder(resp.Body)
	dec.Decode(data)

	return &FilesAPIResponse{OK: data.OK, Paging: data.Files.Paging, Files: data.Files.Matches}, nil
}

func getFiles(token *string, query string) (files, error) {
	page := 1
	paging := true

	var files files

	if query == "" {
		fmt.Print("Fetching files.")
	} else {
		fmt.Printf("Fetching files by searching them with query %s.", query)
	}

	for paging {
		data, err := getFilesPerPage(token, query, page)
		if err != nil {
			return nil, err
		}

		files = append(files, data.Files...)

		paging = data.Paging.Pages > data.Paging.Page
		page = data.Paging.Page + 1

		if data.Paging.Pages > 1 {
			if data.Paging.Page == 1 {
				fmt.Printf(" Total pages: %d ", data.Paging.Pages)
			}

			fmt.Printf(".")
		}
	}

	fmt.Printf("\n")

	return files, nil
}

func floatToString(num float64) string {
	return strconv.FormatFloat(num, 'f', 2, 64)
}

func getHumanSize(fileSize int) string {
	size := float64(fileSize)

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
	query := flag.String("query", "", "Search query. Accept multiple values separated by \",\".")
	types := flag.String("types", "", "Filter files by type. Accept multiple values separated by \",\".")
	flag.Parse()

	if *token == "" {
		color.Red("Missing token")
		os.Exit(1)
	}

	if *types == "" && *query == "" {
		*query = ".rar;.tar;.zip;.mp3;.mp4;.pdf;.ppt;.csv;.jpeg;.json"
	}

	fileIds := make(map[string]bool)
	totalSize := 0
	sizeByTypes := make(map[string]int)
	var uniqFiles files
	var allFiles files

	if *types != "" {
		result, err := getFiles(token, "")
		fmt.Println(len(result))
		if err != nil {
			panic(err)
		}
		allFiles = append(allFiles, result...)
	}

	queries := strings.Split(*query, ",")
	for _, q := range queries {
		if q == "" {
			continue
		}

		result, err := getFiles(token, q)
		if err != nil {
			panic(err)
		}

		allFiles = append(allFiles, result...)
	}

	for _, file := range allFiles {
		if _, ok := fileIds[file.ID]; !ok {
			uniqFiles = append(uniqFiles, file)
			fileIds[file.ID] = true
			sizeByTypes[file.PrettyType] += file.Size
			totalSize += file.Size
		}
	}

	sort.Sort(sort.Reverse(uniqFiles))

	for _, file := range uniqFiles {
		fmt.Printf("  %s - %s\n", cyan(getHumanSize(file.Size)), white(file.Name))
	}

	fmt.Printf("  %s => %s\n", cyan(getHumanSize(totalSize)), white("Total"))

	for name, size := range sizeByTypes {
		fmt.Printf("  %s => %s\n", cyan(getHumanSize(size)), white("Total "+name))
	}
}
