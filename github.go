package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GitHubCommit struct {
	Commit struct {
		Author struct {
			Date string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}
type SearchResult struct {
	TotalCount int            `json:"total_count"`
	Items      []GitHubCommit `json:"items"`
}

func GetTodayCommitsCount(username string) (int, error) {

	today := time.Now().Format("2006-01-02")
	url := fmt.Sprintf("https://api.github.com/search/commits?q=author:%s+committer-date:%s",
		username, today)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/vnd.github.cloak-preview")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("GitHub API вернул статус: %s", resp.Status)
	}
	var result SearchResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, err

	}
	return result.TotalCount, nil
}
