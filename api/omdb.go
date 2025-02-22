package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type OMDBResponse struct {
	Search       []Movie `json:"Search"`
	TotalResults string  `json:"totalResults"`
	Response     string  `json:"Response"`
	Error        string  `json:"Error,omitempty"`
}

type Movie struct {
	Title      string `json:"Title"`
	Year       string `json:"Year"`
	ImdbID     string `json:"imdbID"`
	Type       string `json:"Type"`
	Poster     string `json:"Poster"`
	Plot       string `json:"Plot"`
	ImdbRating string `json:"imdbRating"`
}

func SearchMovies(apiKey, query string) (*OMDBResponse, error) {
	baseURL := "http://www.omdbapi.com/"
	params := url.Values{}
	params.Add("apikey", apiKey)
	params.Add("s", query)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(fullURL)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from OMDB: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OMDB returned non-OK status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var omdbResp OMDBResponse

	if err := json.Unmarshal(body, &omdbResp); err != nil {
		return nil, fmt.Errorf("failed to decode OMDB response: %w", err)
	}

	if omdbResp.Response != "True" {
		return &omdbResp, fmt.Errorf("OMDB error: %s", omdbResp.Error)
	}

	return &omdbResp, nil
}

func GetMovieDetails(apiKey, imdbID string) (*Movie, error) {
	baseURL := "http://www.omdbapi.com/"
	params := url.Values{}
	params.Add("apikey", apiKey)
	params.Add("i", imdbID)
	params.Add("plot", "full")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(fullURL)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie details: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OMDB returned non-OK status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read details response: %w", err)
	}

	var movie Movie

	if err := json.Unmarshal(body, &movie); err != nil {
		return nil, fmt.Errorf("failed to decode movie details: %w", err)
	}

	if movie.Title == "" {
		return nil, fmt.Errorf("movie details not found")
	}

	return &movie, nil
}
