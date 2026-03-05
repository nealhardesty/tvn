package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const tvdbBaseURL = "https://api4.thetvdb.com/v4"

// SeriesSearchResult represents a series from TVDB search.
type SeriesSearchResult struct {
	TVDBID          int
	Name            string
	PrimaryLanguage string
	Country         string
	Year            string
	Overview        string
}

// Episode represents a single TV episode from TVDB.
type Episode struct {
	ID           int
	Name         string
	SeasonNumber int
	Number       int
	Aired        string
}

// TVDBClient handles communication with TheTVDB API v4.
type TVDBClient struct {
	apiKey   string
	token    string
	baseURL  string
	client   *http.Client
	testMode bool

	// Caches
	mu             sync.Mutex
	episodeCache   map[int][]Episode // series ID -> episodes
}

// NewTVDBClient creates a new TVDB API client.
func NewTVDBClient(cfg *Config) *TVDBClient {
	apiKey := ""
	if cfg.TVDBApiKey != nil {
		apiKey = *cfg.TVDBApiKey
	}

	return &TVDBClient{
		apiKey:       apiKey,
		baseURL:      tvdbBaseURL,
		client:       &http.Client{},
		testMode:     os.Getenv("TVNAMER_TEST_MODE") == "1",
		episodeCache: make(map[int][]Episode),
	}
}

// Login authenticates with TVDB and obtains a bearer token.
func (c *TVDBClient) Login() error {
	if c.testMode {
		c.token = "test-token"
		return nil
	}
	if c.apiKey == "" {
		return fmt.Errorf("TVDB API key not configured. Set tvdb_api_key in config or use --config")
	}

	body, _ := json.Marshal(map[string]string{"apikey": c.apiKey})
	resp, err := c.client.Post(c.baseURL+"/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("TVDB login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("TVDB login failed (HTTP %d): %s", resp.StatusCode, string(b))
	}

	var result struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decoding login response: %w", err)
	}
	c.token = result.Data.Token
	return nil
}

// SearchSeries searches for a TV series by name.
func (c *TVDBClient) SearchSeries(name string) ([]SeriesSearchResult, error) {
	if c.testMode {
		return c.searchFixture(name)
	}

	u := fmt.Sprintf("%s/search?query=%s&type=series", c.baseURL, url.QueryEscape(name))
	data, err := c.doGet(u)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []struct {
			TVDBID          string `json:"tvdb_id"`
			Name            string `json:"name"`
			PrimaryLanguage string `json:"primary_language"`
			Country         string `json:"country"`
			Year            string `json:"year"`
			Overview        string `json:"overview"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding search response: %w", err)
	}

	results := make([]SeriesSearchResult, 0, len(resp.Data))
	for _, d := range resp.Data {
		id, _ := strconv.Atoi(d.TVDBID)
		results = append(results, SeriesSearchResult{
			TVDBID:          id,
			Name:            d.Name,
			PrimaryLanguage: d.PrimaryLanguage,
			Country:         d.Country,
			Year:            d.Year,
			Overview:        d.Overview,
		})
	}
	return results, nil
}

// GetEpisodes retrieves all episodes for a series, using cache.
func (c *TVDBClient) GetEpisodes(seriesID int, order string) ([]Episode, error) {
	c.mu.Lock()
	if cached, ok := c.episodeCache[seriesID]; ok {
		c.mu.Unlock()
		return cached, nil
	}
	c.mu.Unlock()

	if c.testMode {
		return c.episodesFixture(seriesID)
	}

	seasonType := "default"
	if order == "dvd" {
		seasonType = "dvd"
	}

	var allEpisodes []Episode
	page := 0
	for {
		u := fmt.Sprintf("%s/series/%d/episodes/%s?page=%d", c.baseURL, seriesID, seasonType, page)
		data, err := c.doGet(u)
		if err != nil {
			return nil, err
		}

		var resp struct {
			Data struct {
				Episodes []struct {
					ID           int    `json:"id"`
					Name         string `json:"name"`
					SeasonNumber int    `json:"seasonNumber"`
					Number       int    `json:"number"`
					Aired        string `json:"aired"`
				} `json:"episodes"`
			} `json:"data"`
			Links struct {
				Next *int `json:"next"`
			} `json:"links"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("decoding episodes response: %w", err)
		}

		for _, e := range resp.Data.Episodes {
			allEpisodes = append(allEpisodes, Episode{
				ID:           e.ID,
				Name:         e.Name,
				SeasonNumber: e.SeasonNumber,
				Number:       e.Number,
				Aired:        e.Aired,
			})
		}

		if resp.Links.Next == nil {
			break
		}
		page = *resp.Links.Next
	}

	c.mu.Lock()
	c.episodeCache[seriesID] = allEpisodes
	c.mu.Unlock()

	return allEpisodes, nil
}

// GetSeriesByID retrieves series info by TVDB ID.
func (c *TVDBClient) GetSeriesByID(seriesID int) (*SeriesSearchResult, error) {
	if c.testMode {
		return &SeriesSearchResult{TVDBID: seriesID, Name: fmt.Sprintf("Series %d", seriesID)}, nil
	}

	u := fmt.Sprintf("%s/series/%d", c.baseURL, seriesID)
	data, err := c.doGet(u)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data struct {
			ID              int    `json:"id"`
			Name            string `json:"name"`
			PrimaryLanguage string `json:"primaryLanguage"`
			Country         string `json:"country"`
			Year            string `json:"year"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding series response: %w", err)
	}

	return &SeriesSearchResult{
		TVDBID:          resp.Data.ID,
		Name:            resp.Data.Name,
		PrimaryLanguage: resp.Data.PrimaryLanguage,
		Country:         resp.Data.Country,
		Year:            resp.Data.Year,
	}, nil
}

// FindEpisode finds an episode by season and episode number.
func FindEpisode(episodes []Episode, season, number int) *Episode {
	for i := range episodes {
		if episodes[i].SeasonNumber == season && episodes[i].Number == number {
			return &episodes[i]
		}
	}
	return nil
}

// FindEpisodeByDate finds an episode by air date.
func FindEpisodeByDate(episodes []Episode, year, month, day int) *Episode {
	dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	for i := range episodes {
		if episodes[i].Aired == dateStr {
			return &episodes[i]
		}
	}
	return nil
}

func (c *TVDBClient) doGet(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("TVDB request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TVDB API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Test mode fixture support

func (c *TVDBClient) searchFixture(name string) ([]SeriesSearchResult, error) {
	safe := strings.ReplaceAll(strings.ToLower(name), " ", "_")
	path := filepath.Join("testdata", "tvdb", "search_"+safe+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		// Return a default result for test mode
		return []SeriesSearchResult{{
			TVDBID: 1,
			Name:   name,
			Year:   "2020",
		}}, nil
	}

	var results []SeriesSearchResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (c *TVDBClient) episodesFixture(seriesID int) ([]Episode, error) {
	path := filepath.Join("testdata", "tvdb", fmt.Sprintf("episodes_%d.json", seriesID))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("test fixture not found: %s", path)
	}

	var episodes []Episode
	if err := json.Unmarshal(data, &episodes); err != nil {
		return nil, err
	}
	return episodes, nil
}
