/*
Code for handling the RSS feed provided by The Hacker News
*/
package main

import (
	"encoding/xml"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type HackerNewsRssFeed struct {
	XMLName xml.Name             `xml:"rss"`
	Version string               `xml:"version,attr"`
	Channel hackerNewsRssChannel `xml:"channel"`
}

type hackerNewsRssChannel struct {
	Title       string              `xml:"title"`
	Link        string              `xml:"link"`
	Description string              `xml:"description"`
	Items       []hackerNewsRssItem `xml:"item"` // TODO: Refactor as a map, to make comparisons much faster
}

type hackerNewsRssItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Author      string `xml:"author"`
	Enclosure   struct {
		URL  string `xml:"url,attr"`
		Type string `xml:"type,attr"`
	} `xml:"enclosure"`
}

type hackerNewsHtml struct {
	XMLName xml.Name `xml:"html"`
	Body    struct {
		Div struct {
			Span struct {
				Text string `xml:",chardata"`
			} `xml:"span"`
		} `xml:"div"`
	} `xml:"body"`
}

var interestingList = []string{
	"Vulnerability",
	"Zero-Day",
	"Espionage",
	"Ransomware",
	"Web Security",
	"Cyber Threat",
	"AppSec",
	"Mobile Security",
	"Malware",
	"Cyber Espionage",
	"APT",
	"National Security",
	"Cloud Security",
	"Linux",
}

func (hn *HackerNewsRssFeed) getPageCategories(pageUrl string) (string, error) {
	/*
		Scrapes the page for the categories of the article
	*/
	var (
		resp *http.Response
		body []byte
		err  error
	)

	if resp, err = http.Get(pageUrl); err != nil {
		log.Fatalln("err: ", err)
	}

	defer resp.Body.Close()
	if body, err = io.ReadAll(resp.Body); err != nil {
		log.Fatalln("err: ", err)
	}

	re := regexp.MustCompile(`<span class='p-tags'(.*)</span>`)
	match := re.FindStringSubmatch(string(body))
	log.Printf("%v", match)

	if len(match) > 1 {
		log.Printf("Not posting article '%v' due to lack of interesting tags\n", pageUrl)
		return match[1], nil
	}

	return "", nil

}

func (hn *HackerNewsRssFeed) filterNewsCats(category string) bool {
	/*
		Filter out articles that are not interesting, using the tags provided by the website
	*/

	//log.Println(category)
	for _, str := range interestingList {
		if strings.Contains(category, str) {
			return true
		}
	}
	return false
}

func (hn *HackerNewsRssFeed) ParseNewRssContent(oldData RSSFeed, newData RSSFeed) ([]discordMessageData, error) {
	var (
		oldHNData  *HackerNewsRssFeed
		newHNData  *HackerNewsRssFeed
		newContent []discordMessageData

		ok bool
		//err error
	)
	if oldHNData, ok = oldData.(*HackerNewsRssFeed); !ok {
		return nil, errors.New("error: oldData is not of type HackerNewsRssFeed")
	}

	if newHNData, ok = newData.(*HackerNewsRssFeed); !ok {
		return nil, errors.New("error: newData is not of type HackerNewsRssFeed")
	}

	for _, newFeedItem := range newHNData.Channel.Items {
		itemExists := false
		for _, oldFeedItem := range oldHNData.Channel.Items {
			if newFeedItem.Title == oldFeedItem.Title {
				log.Printf("Article titled '%v' already exists in old data. Stopping iteration.\n", newFeedItem.Title)
				itemExists = true
				break
			}
		}
		if !itemExists {
			log.Printf("Article '%v' is new", newFeedItem.Title)

			category, err := hn.getPageCategories(newFeedItem.Link)
			if err != nil {
				log.Fatalf("err: %v\n", err)
			}

			interestingCategory := hn.filterNewsCats(category)
			if !interestingCategory {
				log.Printf("'%v' is not an interesting item, skipping", interestingCategory)
				continue
			}

			messageContent := discordMessageData{
				Title:       newFeedItem.Title,
				Description: newFeedItem.Description,
				Link:        newFeedItem.Link,
			}
			newContent = append(newContent, messageContent)
		}
	}
	if len(newContent) > 20 {
		return nil, errors.New("err: more than 20 new items. Logic bug likely. Exiting")
	}
	return newContent, nil
}
