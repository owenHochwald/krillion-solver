package game

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Exercise request construction and real JSON decoding without opening sockets.
// This also lets the tests run in environments that prohibit local listeners.
type handlerTransport struct{ handler http.Handler }

func (h handlerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if err := r.Context().Err(); err != nil {
		return nil, err
	}
	w := httptest.NewRecorder()
	h.handler.ServeHTTP(w, r)
	return w.Result(), nil
}

func testClient(handler http.HandlerFunc) *Client {
	return &Client{BaseURL: "https://krillion.test", HTTP: &http.Client{Transport: handlerTransport{handler}}}
}

func fixture() (today, reveal) {
	r := Demo()
	d := today{Date: r.Date, DayNumber: r.DayNumber, ResetsAt: json.RawMessage(`"2026-10-22T04:00:00Z"`)}
	v := reveal{Date: r.Date}
	for _, p := range r.Prompts {
		d.Prompts = append(d.Prompts, prompt{ID: p.ID, Text: p.Text})
		answers := append([]Answer{{Answer: "Common", Tier: "schooler", Score: 30}}, p.Winners...)
		v.Prompts = append(v.Prompts, prompt{ID: p.ID, Text: p.Text, Answers: answers})
	}
	return d, v
}

func TestSolveAnonymousAndMultipleWinners(t *testing.T) {
	d, v := fixture()
	// Round IDs, rather than response order, must determine the output order.
	v.Prompts[0], v.Prompts[6] = v.Prompts[6], v.Prompts[0]
	requests := 0
	c := testClient(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != "GET" || r.Header.Get("Cookie") != "" || r.Header.Get("Authorization") != "" || r.ContentLength > 0 {
			t.Errorf("request needs to be an anonymous GET: %v", r)
		}
		switch r.URL.Path {
		case "/api/today":
			json.NewEncoder(w).Encode(d)
		case "/api/reveal":
			if r.URL.Query().Get("date") != d.Date {
				t.Errorf("wrong date: %s", r.URL.RawQuery)
			}
			json.NewEncoder(w).Encode(v)
		default:
			t.Errorf("unexpected request: %s", r.URL)
		}
	})
	got, err := c.Solve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 || got.Score() != 700 || got.WinnerCount() != 8 || got.Prompts[0].Winners[0].Answer != "Dekalog" || len(got.Prompts[3].Winners) != 2 || got.ResetsAt == nil || got.Demo {
		t.Fatalf("unexpected result: %+v; requests=%d", got, requests)
	}
}

func TestSolveRejectsIncompleteOrChangedData(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*today, *reveal)
	}{
		{"missing winner", func(d *today, v *reveal) { v.Prompts[0].Answers = nil }},
		{"wrong score", func(d *today, v *reveal) { v.Prompts[0].Answers[1].Score = 90 }},
		{"empty winner", func(d *today, v *reveal) { v.Prompts[0].Answers[1].Answer = " " }},
		{"missing prompt", func(d *today, v *reveal) { v.Prompts = v.Prompts[:6] }},
		{"duplicate reveal ID", func(d *today, v *reveal) { v.Prompts[1].ID = v.Prompts[0].ID }},
		{"duplicate today ID", func(d *today, v *reveal) { d.Prompts[1].ID = d.Prompts[0].ID }},
		{"invalid date", func(d *today, v *reveal) { d.Date = "not-a-date" }},
		{"missing round", func(d *today, v *reveal) { d.Prompts = d.Prompts[:6] }},
	} {
		t.Run(test.name, func(t *testing.T) {
			d, v := fixture()
			test.mutate(&d, &v)
			c := testClient(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/today" {
					json.NewEncoder(w).Encode(d)
				} else {
					json.NewEncoder(w).Encode(v)
				}
			})
			if _, err := c.Solve(context.Background()); err == nil {
				t.Fatal("expected incomplete or changed data to fail")
			}
		})
	}
}

func TestDateMismatchRetriesOnce(t *testing.T) {
	d, v := fixture()
	calls := 0
	c := testClient(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path == "/api/today" {
			json.NewEncoder(w).Encode(d)
		} else {
			v.Date = "2026-10-20"
			json.NewEncoder(w).Encode(v)
		}
	})
	_, err := c.Solve(context.Background())
	if err == nil || calls != 4 {
		t.Fatalf("expected bounded retry, calls=%d err=%v", calls, err)
	}
}

func TestHTTPErrors(t *testing.T) {
	for _, status := range []int{403, 429, 500, 200} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			c := testClient(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(status)
				fmt.Fprint(w, "<html>not JSON</html>")
			})
			_, err := c.Solve(context.Background())
			if err == nil || calls != 1 {
				t.Fatalf("expected one failed request, calls=%d err=%v", calls, err)
			}
		})
	}
}

func TestCancelledRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewClient().Solve(ctx)
	if err == nil || !strings.Contains(err.Error(), "canceled") {
		t.Fatalf("unexpected cancellation: %v", err)
	}
}

func TestSafeOutput(t *testing.T) {
	if got := SafeText("name\x1b\n\a"); got != "name " {
		t.Fatalf("unsafe text: %q", got)
	}
	if Demo().Score() != 700 || !strings.Contains(Demo().Plain(), "DEMO") {
		t.Fatal("demo must be labeled and count each round once")
	}
}
