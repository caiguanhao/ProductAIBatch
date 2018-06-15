package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	netUrl "net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	serviceId   string
	accessKeyId string

	stdout = log.New(os.Stdout, "", 0)
	stderr = log.New(os.Stderr, "", 0)
)

type (
	searchResponse struct {
		DetectedObjs []struct {
			Loc []float64 `json:"loc"`
		} `json:"detected_objs"`
		Results []struct {
			MetaData string  `json:"metadata"`
			Score    float64 `json:"score"`
			URL      string  `json:"url"`
		} `json:"results"`
		Type []string `json:"type"`
	}

	processedSearchResponseResult struct {
		ImageUrl string `json:"image_url"`
		Id       string `json:"id"`
	}

	processedSearchResponse struct {
		Coordinates [][]float64                     `json:"coordinates"`
		Results     []processedSearchResponseResult `json:"results"`
	}

	searchResult struct {
		url    string
		coords []string
		result string
	}
)

func search(url string, coords []string) (results []searchResult, err error) {
	v := netUrl.Values{}
	v.Set("ret_detected_objs", "1")
	v.Set("url", url)
	if len(coords) == 4 {
		v.Set("loc", strings.Join(coords, "-"))
	}
	var req *http.Request
	req, err = http.NewRequest("POST", "https://api.productai.cn/search/"+serviceId, strings.NewReader(v.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-CA-Version", "1.0")
	req.Header.Set("X-CA-AccessKeyId", accessKeyId)
	client := http.Client{
		Timeout: time.Duration(3 * time.Second),
	}
	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		return
	}

	var response searchResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	resp.Body.Close()
	if err != nil {
		return
	}

	var processed processedSearchResponse
	for _, result := range response.DetectedObjs {
		processed.Coordinates = append(processed.Coordinates, result.Loc)
	}
	for _, result := range response.Results {
		processed.Results = append(processed.Results, processedSearchResponseResult{
			ImageUrl: result.URL,
			Id:       result.MetaData,
		})
	}

	result, _ := json.Marshal(processed)
	results = append(results, searchResult{url, coords, string(result)})

	if coords == nil {
		for _, result := range response.DetectedObjs {
			var locs []string
			for _, loc := range result.Loc {
				locs = append(locs, strconv.FormatFloat(loc, 'f', -1, 32))
			}
			if len(locs) == 0 {
				continue
			}
			_results, err := search(url, locs)
			if err == nil {
				results = append(results, _results...)
			}
		}
	}

	time.Sleep(2 * time.Second)
	return
}

func init() {
	flag.StringVar(&serviceId, "service-id", "", "Service ID")
	flag.StringVar(&accessKeyId, "access-key-id", "", "Access Key ID")
}

func main() {
	flag.Parse()

	concurrency := 3
	jobs := make(chan interface{})

	go func() {
		defer close(jobs)
		file, err := os.Open(flag.Arg(0))
		if err == nil {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				jobs <- scanner.Text()
			}
		}
		file.Close()
	}()

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for job := range jobs {
				url := (job).(string)
				if url == "" {
					continue
				}
				results, err := search(url, nil)
				if err != nil {
					stderr.Println("Error", url, err)
					continue
				}
				for _, result := range results {
					coords, _ := json.Marshal(result.coords)
					stdout.Println(result.url, "\t", string(coords), "\t", string(result.result))
				}
				stderr.Println("Finished", url)
			}
		}()
	}
	wg.Wait()
}
