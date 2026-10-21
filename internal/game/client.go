package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const Origin = "https://krillion.io"
const maxResponse = 8 << 20

var errDateMismatch = errors.New("the daily puzzle changed while loading")

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

type prompt struct {
	ID      string   `json:"id"`
	Text    string   `json:"text"`
	Answers []Answer `json:"answers"`
}

type today struct {
	Date      string          `json:"date"`
	DayNumber int             `json:"dayNumber"`
	ResetsAt  json.RawMessage `json:"resetsAt"`
	Prompts   []prompt        `json:"prompts"`
}

type reveal struct {
	Date    string   `json:"date"`
	Prompts []prompt `json:"prompts"`
}

func NewClient() *Client {
	return &Client{BaseURL: Origin, HTTP: &http.Client{Timeout: 12 * time.Second}}
}

// Solve makes two anonymous GETs. A date mismatch gets one fresh attempt.
func (c *Client) Solve(ctx context.Context) (Result, error) {
	result, err := c.solve(ctx)
	if errors.Is(err, errDateMismatch) {
		return c.solve(ctx)
	}
	return result, err
}

func (c *Client) solve(ctx context.Context) (Result, error) {
	var day today
	if err := c.get(ctx, "/api/today", &day); err != nil {
		return Result{}, err
	}
	if _, err := time.Parse("2006-01-02", day.Date); err != nil {
		return Result{}, errors.New("today returned an invalid puzzle date")
	}
	if len(day.Prompts) != 7 {
		return Result{}, fmt.Errorf("expected 7 daily prompts, got %d; the API may have changed", len(day.Prompts))
	}
	var data reveal
	if err := c.get(ctx, "/api/reveal?date="+url.QueryEscape(day.Date), &data); err != nil {
		return Result{}, err
	}
	if data.Date != day.Date {
		return Result{}, errDateMismatch
	}
	byID := make(map[string]prompt, len(data.Prompts))
	for _, p := range data.Prompts {
		if _, exists := byID[p.ID]; exists || p.ID == "" {
			return Result{}, errors.New("reveal returned duplicate or empty prompt IDs")
		}
		byID[p.ID] = p
	}
	result := Result{Date: day.Date, DayNumber: day.DayNumber}
	// Reset metadata is optional. An unfamiliar format must not prevent solving.
	var reset string
	if json.Unmarshal(day.ResetsAt, &reset) == nil {
		if at, err := time.Parse(time.RFC3339, reset); err == nil {
			result.ResetsAt = &at
		}
	}
	seen := make(map[string]bool)
	for _, p := range day.Prompts {
		if p.ID == "" || seen[p.ID] || strings.TrimSpace(p.Text) == "" {
			return Result{}, errors.New("today returned invalid prompt metadata")
		}
		seen[p.ID] = true
		answers, ok := byID[p.ID]
		if !ok {
			return Result{}, fmt.Errorf("reveal is missing round %s", SafeText(p.ID))
		}
		round := Round{ID: p.ID, Text: p.Text, Accepted: len(answers.Answers)}
		for _, a := range answers.Answers {
			if a.Tier == "krillion" {
				if strings.TrimSpace(a.Answer) == "" || a.Score != 100 {
					return Result{}, fmt.Errorf("round %s has an unexpected Krillion answer or score; the API may have changed", SafeText(p.ID))
				}
				round.Winners = append(round.Winners, a)
			}
		}
		if len(round.Winners) == 0 {
			return Result{}, fmt.Errorf("round %s has no Krillion answer; reveal may no longer be available", SafeText(p.ID))
		}
		result.Prompts = append(result.Prompts, round)
	}
	return result, nil
}

func (c *Client) get(ctx context.Context, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.BaseURL, "/")+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "krillion-solver/0.1")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("request stopped: %w", ctx.Err())
		}
		return fmt.Errorf("could not reach krillion.io: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			return errors.New("Krillion rate limited this request; wait before refreshing")
		case http.StatusUnauthorized, http.StatusForbidden:
			return errors.New("Krillion denied access to the daily data; anonymous reveal may no longer be available")
		default:
			return fmt.Errorf("Krillion returned HTTP %d for %s", resp.StatusCode, path)
		}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse+1))
	if err != nil {
		return fmt.Errorf("could not read Krillion's response: %w", err)
	}
	if len(body) > maxResponse {
		return errors.New("Krillion's response exceeded 8 MiB")
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("Krillion returned unexpected JSON for %s: %w", path, err)
	}
	return nil
}
