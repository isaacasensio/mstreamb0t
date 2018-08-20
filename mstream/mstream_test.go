package mstream

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func responseStub() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `<?xml version="1.0" encoding="iso-8859-1"?>
			<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
				<channel>
					<atom:link href="http://readms.net/rss" rel="self" type="application/rss+xml" />
					<title>MangaStream Releases</title>
					<link>http://readms.net/rss</link>
					<description>The latest MangaStream.com Releases</description>
					<language>en</language>
					<ttl>30</ttl>
					<item>
						<title>Fairy Tail 100 Years Quest 004</title>
						<link>http://readms.net/r/fairy_tail_100_years_quest/004/5275/1</link>
						<pubDate>Tue, 07 Aug 2018 12:09:46 -0700</pubDate>
						<description>Amazing Elmina</description>
						<guid isPermaLink="true">http://readms.net/read/fairy_tail_100_years_quest/004/5275/1</guid>
					</item>
					<item>
						<title>Dragon Ball Super 038</title>
						<link>http://readms.net/r/dragon_ball_super/038/5279/1</link>
						<pubDate>Tue, 07 Aug 2018 12:09:46 -0700</pubDate>
						<description>Dragon Ball Super</description>
						<guid isPermaLink="true">http://readms.net/r/dragon_ball_super/038/5279/1</guid>
					</item>
				</channel>
			</rss>`
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(resp))
	}))
}

func errorStub() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp string
		switch r.RequestURI {
		case "/not-found":
			http.Error(w, "Not found", http.StatusNotFound)
			return
		case "/internal-server-error":
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		default:
			resp = `{ incorrect-json }`
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
		w.WriteHeader(404)
	}))
}
func TestFindUpdatesSince(t *testing.T) {

	getDate := func(date string) time.Time {
		parsedDate, _ := time.Parse(time.RFC1123Z, date)
		return parsedDate
	}

	var flagtests = []struct {
		name           string
		mangaNames     []string
		since          time.Time
		expectedMangas []string
	}{
		{
			name:       "returns empty list when no manga found for today's date",
			mangaNames: []string{"unknown-manga"},
			since:      time.Now(),
		},
		{
			name:       "returns empty list when no manga found since an old date",
			mangaNames: []string{"unknown-manga"},
			since:      getDate("Tue, 07 Aug 1970 12:09:46 -0700"),
		},
		{
			name:           "returns a list when an update is found since the last execution",
			mangaNames:     []string{"Fairy"},
			since:          getDate("Mon, 06 Aug 2018 10:00:00 -0700"),
			expectedMangas: []string{"Fairy Tail 100 Years Quest 004"},
		},
		{
			name:           "returns multiple manga when multiple updates are found since the last execution",
			mangaNames:     []string{"Fairy", "Dragon"},
			since:          getDate("Mon, 06 Aug 2018 10:00:00 -0700"),
			expectedMangas: []string{"Fairy Tail 100 Years Quest 004", "Dragon Ball Super 038"},
		},
		{
			name:       "returns empty list when manga found but pubdate is older than last execution",
			mangaNames: []string{"Fairy"},
			since:      getDate("Thu, 09 Aug 2018 10:00:00 -0700"),
		},
	}

	for _, tt := range flagtests {
		t.Run(tt.name, func(t *testing.T) {
			server := responseStub()
			defer server.Close()

			c := NewClient(server.URL)
			updates, err := c.FindNewReleasesSince(tt.mangaNames, tt.since)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedMangas, updates)
		})
	}

}

func TestFindUpdatesSince_ApiCallErrors(t *testing.T) {
	server := errorStub()
	defer server.Close()

	var flagtests = []struct {
		name     string
		endpoint string
	}{
		{
			name:     "returns empty list when no manga found for today's date",
			endpoint: "/not-found",
		},
		{
			name:     "returns empty list when no manga found since an old date",
			endpoint: "/internal-server-error",
		},
		{
			name:     "returns a list when an update is found since the last execution",
			endpoint: "/invalid",
		},
	}

	for _, tt := range flagtests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(server.URL + tt.endpoint)

			updates, err := c.FindNewReleasesSince([]string{"manga"}, time.Now())
			assert.Error(t, err)
			var expectedUpdates []string
			assert.Equal(t, expectedUpdates, updates)
		})
	}
}

func TestFindUpdatesSince_WhenInvalidDate(t *testing.T) {

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `<?xml version="1.0" encoding="iso-8859-1"?>
			<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
				<channel>
					<atom:link href="http://readms.net/rss" rel="self" type="application/rss+xml" />
					<title>MangaStream Releases</title>
					<link>http://readms.net/rss</link>
					<description>The latest MangaStream.com Releases</description>
					<language>en</language>
					<ttl>30</ttl>
					<item>
						<title>Fairy Tail 100 Years Quest 004</title>
						<link>http://readms.net/r/fairy_tail_100_years_quest/004/5275/1</link>
						<pubDate>this-is-not-a-date</pubDate>
						<description>Amazing Elmina</description>
						<guid isPermaLink="true">http://readms.net/read/fairy_tail_100_years_quest/004/5275/1</guid>
					</item>
				</channel>
			</rss>`
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(resp))

	}))

	defer s.Close()

	c := NewClient(s.URL)
	_, err := c.FindNewReleasesSince([]string{"something"}, time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parsing time")
}
