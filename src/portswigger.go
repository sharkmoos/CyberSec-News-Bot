/*
Code for handling the RSS feed provided by PortSwigger
*/
package main

import (
	"encoding/xml"
	"errors"
	"log"
)

type PortSwiggerRSSFeed struct {
	XMLName xml.Name           `xml:"rss"`
	Version string             `xml:"version,attr"`
	Channel PortSwiggerChannel `xml:"channel"`
}

type PortSwiggerChannel struct {
	Title       string            `xml:"title"`
	Link        string            `xml:"link"`
	Description string            `xml:"description"`
	Items       []PortSwiggerItem `xml:"item"`
}

type PortSwiggerItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Author      string `xml:"author"`
	Category    string `xml:"category"`
	Comments    string `xml:"comments"`
}

func (pz *PortSwiggerRSSFeed) ParseNewRssContent(oldData RSSFeed, newData RSSFeed) ([]discordMessageData, error) {
	var (
		oldHNData  *PortSwiggerRSSFeed
		newHNData  *PortSwiggerRSSFeed
		newContent []discordMessageData

		ok bool
	)

	if oldHNData, ok = oldData.(*PortSwiggerRSSFeed); !ok {
		return nil, errors.New("error: oldData is not of type HackerNewsRssFeed")
	}
	if newHNData, ok = newData.(*PortSwiggerRSSFeed); !ok {
		return nil, errors.New("error: newData is not of type HackerNewsRssFeed")
	}

	for _, newFeedItem := range newHNData.Channel.Items {
		itemExists := false
		for _, oldFeedItem := range oldHNData.Channel.Items {
			if newFeedItem.Title == oldFeedItem.Title {
				// log.Printf("Article titled '%v' already exists in old data. Stopping iteration.\n", newFeedItem.Title)
				itemExists = true
				break
			}
		}
		if !itemExists {
			log.Printf("Article '%v' is new", newFeedItem.Title)
			log.Println(newFeedItem)

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
