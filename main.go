package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	baseURLSearch = "https://slack.com/api/search.files"
	baseURLList   = "https://slack.com/api/files.list"
	baseURLDelete = "https://slack.com/api/files.delete"
)

// File ...
type File struct {
	ID                 string
	Name               string
	PrettyType         string `json:"pretty_type"`
	Size               int
	URLPrivateDownload string `json:"url_private_download"`
	Permalink          string
}

// Paging ...
type Paging struct {
	Count int
	Total int
	Page  int
	Pages int
}

// FilesSearchAPIResponse ...
type FilesSearchAPIResponse struct {
	OK           bool
	Query        string
	AllStopWords bool
	Files        struct {
		Total   int
		Paging  Paging
		Matches []File
	}
}

// FilesAPIResponse ...
type FilesAPIResponse struct {
	OK     bool
	Error  string
	Files  []File
	Paging Paging
}

type files []File

func (f files) Len() int           { return len(f) }
func (f files) Less(i, j int) bool { return f[i].Size < f[j].Size }
func (f files) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func getFilesPerPage(token *string, types string, tsTo string, query string, page int) (*FilesAPIResponse, error) {
	var url string
	if query != "" {
		url = baseURLSearch + "?token=" + *token + "&query=" + query + "&page=" + strconv.Itoa(page)
	} else {
		url = baseURLList + "?token=" + *token + "&page=" + strconv.Itoa(page)
		if types != "" {
			url += "&types=" + types
		}
		if tsTo != "" {
			url += "&ts_to=" + tsTo
		}
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

func getFiles(token *string, types string, tsTo string, query string) (files, error) {
	cyan := color.New(color.FgCyan).SprintFunc()
	page := 1
	paging := true

	var files files

	if query == "" {
		fmt.Print("Fetching files.")
	} else {
		fmt.Printf("Fetching files by searching them with query %s.", query)
	}

	for paging {
		data, err := getFilesPerPage(token, types, tsTo, query, page)
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

			fmt.Print(cyan("."))
		}
	}

	fmt.Printf("\n")

	return files, nil
}

func deleteFile(token *string, file File) (*FilesAPIResponse, error) {
	url := baseURLDelete + "?token=" + *token + "&file=" + file.ID
	resp, err := http.Get(url)
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

func downloadFile(token *string, backup string, file File) error {
	if file.URLPrivateDownload == "" {
		return errors.New("URL for download not available " + file.Permalink)
	}

	out, err := os.Create(backup + "/" + file.ID + "___" + file.Name)
	if err != nil {
		return err
	}
	defer out.Close()

	color.Cyan("Downloading file \"%s\"", file.Name)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", file.URLPrivateDownload, nil)
	req.Header.Set("Authorization", "Bearer "+*token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func main() {
	cyan := color.New(color.FgCyan).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	token := flag.String("token", "", "Slack Authentication token.")
	query := flag.String("query", "", "Search query. Accept multiple values separated by \",\".")
	types := flag.String("types", "", "Filter files by type. Accept multiple values separated by \",\".")
	daysTo := flag.Int("days-to", 0, "Filter files created before this timestamp (inclusive).")
	backup := flag.String("backup", "", "Path to backup files before delete.")
	version := flag.Bool("version", false, "Version.")
	flag.Parse()

	if *version {
		fmt.Println("v1.2.0")
		os.Exit(0)
	}

	if *token == "" {
		color.Red("Missing token")
		os.Exit(1)
	}

	if *query == "" && *types == "" && *daysTo == 0 {
		*types = "all"
	}

	var tsTo string
	if *daysTo != 0 {
		tsTo = strconv.FormatInt(time.Now().AddDate(0, 0, -*daysTo).UTC().Unix(), 10)
	}

	fileIds := make(map[string]bool)
	totalSize := 0
	sizeByTypes := make(map[string]int)
	var uniqFiles files
	var allFiles files

	if *types != "" {
		result, err := getFiles(token, *types, tsTo, "")
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

		result, err := getFiles(token, "", "", q)
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

	// Sort files by size and print result
	sort.Sort(sort.Reverse(uniqFiles))

	color.Magenta("Found %d files", len(uniqFiles))
	for _, file := range uniqFiles {
		fmt.Printf("  %s - %s (%s)\n", cyan(getHumanSize(file.Size)), file.Name, file.Permalink)
	}

	color.Magenta("Summary: Total")
	fmt.Printf("  %s - %d Files", cyan(getHumanSize(totalSize)), len(uniqFiles))

	// Sort types by size and print result
	var listSizeByTypes files
	for name, size := range sizeByTypes {
		listSizeByTypes = append(listSizeByTypes, File{Size: size, Name: name})
	}

	sort.Sort(sort.Reverse(listSizeByTypes))

	color.Magenta("\nSummary: Total size by types")
	for _, file := range listSizeByTypes {
		fmt.Printf("  %s - %s\n", cyan(getHumanSize(file.Size)), white(file.Name))
	}

	if len(uniqFiles) == 0 {
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\nWould you like to delete those files?")
	fmt.Printf("  %s No. Stop here.\n", cyan("1)"))
	fmt.Printf("  %s Yes. Delete all them.\n", cyan("2)"))
	fmt.Printf("  %s Yes. But ask me to confirm each one of them.\n", cyan("3)"))
	fmt.Print("=> ")
	answer, _ := reader.ReadString('\n')

	if answer != "2\n" && answer != "3\n" {
		color.Red("Stopping without delete any file.")
		return
	}

	shouldBackup := *backup != ""

	if !shouldBackup {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("\nBackup is not defined. Are you sure you will delete the files without backup?")
		fmt.Printf("%s => ", cyan("(y/n)"))
		answer, _ := reader.ReadString('\n')

		if answer != "y\n" {
			color.Red("Stopping without delete any file.")
			return
		}
	}

	shouldAsk := answer == "3\n"
	totalDeleted := 0
	totalSizeDeleted := 0

	for _, file := range uniqFiles {
		if shouldAsk {
			fmt.Printf("Delete file \"%s\" %s (%s)?\n", file.Name, cyan(getHumanSize(file.Size)), file.Permalink)
			fmt.Printf("%s => ", cyan("(y/n)"))
			answer, _ := reader.ReadString('\n')

			if answer == "n\n" {
				color.Blue("Skipped.")
				continue
			}
		}

		if shouldBackup {
			err := downloadFile(token, *backup, file)
			if err != nil {
				color.Red("Error: %v", err)
				continue
			}
		}

		data, err := deleteFile(token, file)
		if err != nil {
			panic(err)
		}

		if !data.OK {
			if shouldAsk {
				color.Red("Error: %s", data.Error)
			} else {
				color.Red("Error deleting file \"%s\": %s", file.Name, data.Error)
			}
			continue
		}
		totalDeleted++
		totalSizeDeleted += file.Size
		color.Green("Deleted.")
	}

	color.Magenta("\nSummary: Deleted")
	fmt.Printf("  %s - %d Files\n", cyan(getHumanSize(totalSizeDeleted)), totalDeleted)
}
