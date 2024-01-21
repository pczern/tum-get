package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Rule struct {
	Method                 string `json:"method"`
	File_pattern           string `json:"file_pattern"`
	Path_pattern           string `json:"path_pattern"`
	Neg_file_pattern       string `json:"neg_file_pattern"`
	Neg_path_pattern       string `json:"neg_path_pattern"`
	Regex_file_pattern     string `json:"regex_file_pattern"`
	Regex_path_pattern     string `json:"regex_path_pattern"`
	Neg_regex_file_pattern string `json:"neg_regex_file_pattern"`
	Neg_regex_path_pattern string `json:"neg_regex_path_pattern"`
	Dest                   string `json:"dest"`
}
type Data struct {
	CourseId int    `json:"courseId"`
	Course   string `json:"course"`
	Url      string `json:"url"`
	Rules    []Rule `json:"rules"`
}
type Input struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Root     string `json:"root"`
	DataArr  []Data `json:"data"`
}

func main() {
	skipFetch := flag.Bool("skip-fetch", false, "skips fetching")
	flag.Parse()
	jsonFile, err := os.Open("./input.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	var input Input
	jsonParser := json.NewDecoder(jsonFile)
	if err = jsonParser.Decode(&input); err != nil {
		fmt.Println(err)
		return
	}

	if strings.TrimSpace(input.Root) == "" {
		fmt.Println("root directory missing in input.json")
		return
	}
	if strings.TrimSpace(input.Username) == "" {
		fmt.Println("username missing in input.json")
		return
	}
	if strings.TrimSpace(input.Password) == "" {
		fmt.Println("password missing in input.json")
		return
	}

	dirs := filepath.Dir(input.Root)
	if err := os.MkdirAll(dirs, 0755); err != nil {
		fmt.Println("Error creating root directory:", err)
		return
	}

	if !*skipFetch {
		client, err := Auth(input.Username, input.Password)
		if err != nil {
			fmt.Println(err)
			return
		}

		var wg sync.WaitGroup
		fmt.Println("Requesting ZIPs")
		for _, data := range input.DataArr {
			wg.Add(1)
			go func(data Data) {
				defer wg.Done()

				req, err := http.NewRequest("GET", data.Url, nil)
				if err != nil {
					fmt.Println("Error creating request:", err)
					return
				}

				resp, err := client.Do(req)

				if err != nil {
					fmt.Println("Error making request:", err)
					return
				}
				defer resp.Body.Close()

				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Println("Error reading response body:", err)
					return
				}

				bodyString := string(bodyBytes)

				doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyString))
				if err != nil {
					fmt.Println("Error loading HTTP response body. ", err)
					return
				}

				actionURL, docData := findData(doc)

				docDataValues := url.Values{}
				for key, value := range docData {
					docDataValues.Set(key, value)
				}

				req, err = http.NewRequest("POST", actionURL, strings.NewReader(docDataValues.Encode()))
				if err != nil {
					fmt.Println("Error creating request:", err)
					return
				}
				req.Header.Set("Referer", data.Url)
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
				resp, err = client.Do(req)
				if err != nil {
					fmt.Println("Error making request:", err)
					return
				}
				defer resp.Body.Close()

				bodyBytes, err = io.ReadAll(resp.Body)
				if err != nil {
					fmt.Println("Error reading body:", err)
					return
				}

				bodyReader := bytes.NewReader(bodyBytes)

				zipReader, err := zip.NewReader(bodyReader, int64(bodyReader.Len()))
				if err != nil {
					fmt.Printf("%+v\n", resp)
					fmt.Printf("%+v\n", resp.Body)

					fmt.Println("Error creating zip reader:", err, resp.StatusCode, resp.Header)
					return
				}

				// iterate over zip archive files
				for _, f := range zipReader.File {

					rc, err := f.Open()
					if err != nil {
						fmt.Println("Error opening file:", err)
						return
					}

					filePath := filepath.Join(input.Root, data.Course, f.Name)

					// create destination directory if not exists
					dstDir := filepath.Dir(filePath)
					if err := os.MkdirAll(dstDir, 0755); err != nil {
						fmt.Println("Error creating destination directory:", err)
						return
					}

					if !f.FileInfo().IsDir() {
						_, err = os.Stat(filePath)
						if errors.Is(err, os.ErrNotExist) {
							fmt.Println("New:", filePath)
						} else if err != nil {
							fmt.Println("Error checking if file exists:", err)
							return
						}
					}
					dstFile, err := os.Create(filePath)
					if err != nil {
						fmt.Println("Error creating destination file:", err)
						return
					}
					_, err = io.Copy(dstFile, rc)
					if err != nil {
						fmt.Println("Error copying file:", err)
						return
					}

					rc.Close()
					dstFile.Close()
				}
			}(data)
		}
		wg.Wait()
	}
	fmt.Println("Processing Rules")
	err = filepath.Walk(input.Root, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(input.Root, path)
		if err != nil {
			fmt.Printf("error getting relative path: %v\n", err)
			return err
		}

		var firstDir string
		parts := strings.Split(relPath, string(os.PathSeparator))
		if len(parts) > 1 {
			firstDir = parts[0]
		}

		for _, data := range input.DataArr {
			if firstDir != data.Course {
				continue
			}
			for _, rule := range data.Rules {
				isMatch := true
				if rule.Regex_path_pattern != "" {
					pathRe, err := regexp.Compile(rule.Regex_path_pattern)
					if err != nil {
						fmt.Println("Error compiling regular expression:", err)
						return err
					}
					isMatch = isMatch && pathRe.MatchString(path)
				}
				if rule.Regex_file_pattern != "" {
					fileRe, err := regexp.Compile(rule.Regex_file_pattern)
					if err != nil {
						fmt.Println("Error compiling regular expression:", err)
						return err
					}
					isMatch = isMatch && fileRe.MatchString(filepath.Base(path))
				}
				if rule.Neg_regex_path_pattern != "" {
					negPathRe, err := regexp.Compile(rule.Neg_regex_path_pattern)
					if err != nil {
						fmt.Println("Error compiling regular expression:", err)
						return err
					}
					isMatch = isMatch && !negPathRe.MatchString(path)
				}
				if rule.Neg_regex_file_pattern != "" {
					negFileRe, err := regexp.Compile(rule.Neg_regex_file_pattern)
					if err != nil {
						fmt.Println("Error compiling regular expression:", err)
						return err
					}
					isMatch = isMatch && !negFileRe.MatchString(filepath.Base(path))
				}

				if rule.Path_pattern != "" {
					pathMatch, err := filepath.Match(rule.Path_pattern, path)
					if err != nil {
						fmt.Printf("error while trying to match pattern: %v\n", err)
						return err
					}
					isMatch = isMatch && pathMatch
				}
				if rule.File_pattern != "" {
					fileMatch, err := filepath.Match(rule.File_pattern, filepath.Base(path))
					if err != nil {
						fmt.Printf("error while trying to match pattern: %v\n", err)
						return err
					}
					isMatch = isMatch && fileMatch
				}
				if rule.Neg_path_pattern != "" {
					negPathMatch, err := filepath.Match(rule.Neg_path_pattern, path)
					if err != nil {
						fmt.Printf("error while trying to match pattern: %v\n", err)
						return err
					}
					isMatch = isMatch && !negPathMatch
				}

				if rule.Neg_file_pattern != "" {
					negFileMatch, err := filepath.Match(rule.Neg_file_pattern, filepath.Base(path))
					if err != nil {
						fmt.Printf("error while trying to match pattern: %v\n", err)
						return err
					}
					isMatch = isMatch && !negFileMatch
				}

				// make sure at least one pattern is set
				isMatch = isMatch && (rule.Regex_path_pattern != "" || rule.Regex_file_pattern != "" || rule.Neg_regex_path_pattern != "" || rule.Neg_regex_file_pattern != "" || rule.Path_pattern != "" || rule.File_pattern != "" || rule.Neg_path_pattern != "" || rule.Neg_file_pattern != "")

				if isMatch {

					// move the file to target folder
					newPath := filepath.Join(rule.Dest, filepath.Base(path))

					dirs := filepath.Dir(newPath)
					if err := os.MkdirAll(dirs, 0755); err != nil {
						fmt.Println("Error creating destination directory:", err)
						return err
					}

					if rule.Method != "rename" {
						src, err := os.Open(path)
						if err != nil {
							fmt.Println("Error opening source file:", err)
							return err
						}
						defer src.Close()
						dest, err := os.Create(newPath)
						if err != nil {
							fmt.Println("Error creating destination file:", err)
							return err
						}
						defer dest.Close()
						_, err = io.Copy(dest, src)
						if err != nil {
							fmt.Println("Error copying file:", err)
							return err
						}
						return nil
					}
					err = os.Rename(path, newPath)

					if err != nil {
						fmt.Printf("error while moving file: %v\n", err)
						return err
					}
					// create placeholder file, so that system detects if file had existed
					file, err := os.Create(path)
					if err != nil {
						fmt.Println("Error creating file:", err)
						return err
					}
					defer file.Close()
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %v: %v\n", input.Root, err)
	}

	fmt.Println("Done")
}

func findData(doc *goquery.Document) (string, map[string]string) {
	form := doc.Find("form").First()
	actionURL, _ := form.Attr("action")

	data := make(map[string]string)
	form.Find("input").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		value, _ := s.Attr("value")
		if name != "cancel" {
			data[name] = value
		}
	})

	return actionURL, data
}
