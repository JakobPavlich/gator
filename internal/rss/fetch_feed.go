// Package rss - tu bom verjetno dajal kodo v zvezi s RSS prejemanjen in urejanjem
package rss

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "gator")
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var feed RSSFeed
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		return nil, err
	}
	// title := html.UnescapeString(feed.Channel.Title)
	// description := html.UnescapeString(feed.Channel.Description)
	// fmt.Println(title)
	// fmt.Println(feed.Channel.Link)
	// fmt.Println(description)
	// for _, item := range feed.Channel.Item {
	// 	title := html.UnescapeString(item.Title)
	// 	description := html.UnescapeString(item.Description)
	// 	fmt.Println(title)
	// 	fmt.Println(item.Link)
	// 	fmt.Println(description)
	// 	fmt.Println(item.PubDate)
	// }
	return &feed, nil
}
