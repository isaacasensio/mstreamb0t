package mstream

import (
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
)

// Client is a mangaStream client
type Client struct {
	feedURL string
}

// NewClient creates a new Client instance
func NewClient(feedURL string) Client {
	return Client{
		feedURL: feedURL,
	}
}

// FindNewReleasesSince finds all manga updates since the provided time
func (c Client) FindNewReleasesSince(mangaNames []string, since time.Time) ([]string, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(c.feedURL)
	if err != nil {
		logrus.Errorf("Error fetching data from %s : %v", c.feedURL, err)
		return nil, err
	}

	var updates []string
	for _, item := range feed.Items {

		t, err := time.Parse(time.RFC1123Z, item.Published)
		if err != nil {
			logrus.Errorf("Error parsing publication date %s : %v", item.Published, err)
			return nil, err
		}

		for _, mangaName := range mangaNames {
			if caseInsensitiveContains(item.Title, strings.TrimSpace(mangaName)) && t.After(since) {
				updates = append(updates, item.Title)
			}
		}
	}
	return updates, nil
}

func caseInsensitiveContains(str string, substr string) bool {
	return strings.Contains(strings.ToUpper(str), strings.ToUpper(substr))
}
