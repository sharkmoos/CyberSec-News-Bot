/*
Code for handling the RSS feed provided by Google Project Zero
*/
package main

import (
	"encoding/xml"
	"errors"
	"log"
	"regexp"
)

type ProjectZeroRssFeed struct {
	XMLName xml.Name             `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string               `xml:"title"`
	Link    string               `xml:"link"`
	Updated string               `xml:"updated"`
	Items   []ProjectZeroRssItem `xml:"entry"`
}

type ProjectZeroRssItem struct {
	Title     string                `xml:"title"`
	Link      []ProjectZeroRssLink  `xml:"link"`
	Published string                `xml:"published"`
	Updated   string                `xml:"updated"`
	Summary   ProjectZeroRssSummary `xml:"summary"`
	Content   ProjectZeroRssContent `xml:"content"`
	Id        string                `xml:"id"`
}

type ProjectZeroRssLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type ProjectZeroRssSummary struct {
	Type    string `xml:"type,attr"`
	Summary string `xml:",chardata"`
}

type ProjectZeroRssContent struct {
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

func (pz *ProjectZeroRssFeed) ParseNewRssContent(oldData RSSFeed, newData RSSFeed) ([]discordMessageData, error) {
	var (
		oldHNData  *ProjectZeroRssFeed
		newHNData  *ProjectZeroRssFeed
		newContent []discordMessageData

		ok bool
	)
	if oldHNData, ok = oldData.(*ProjectZeroRssFeed); !ok {
		return nil, errors.New("error: oldData is not of type HackerNewsRssFeed")
	}
	if newHNData, ok = newData.(*ProjectZeroRssFeed); !ok {
		return nil, errors.New("error: newData is not of type HackerNewsRssFeed")
	}

	for _, newFeedItem := range newHNData.Items {
		itemExists := false
		for _, oldFeedItem := range oldHNData.Items {
			if newFeedItem.Title == oldFeedItem.Title {
				// log.Printf("Article titled '%v' already exists in old data. Stopping iteration.\n", newFeedItem.Title)
				itemExists = true
				break
			}
		}
		if !itemExists {
			log.Printf("Article '%v' is new", newFeedItem.Title)
			log.Println(newFeedItem)

			pattern := regexp.MustCompile("https://googleprojectzero\\.blogspot\\.com/\\d\\d\\d\\d")
			var newsLink string
			for _, str := range newFeedItem.Link {
				matches := pattern.FindString(str.Href)
				if matches != "" {
					newsLink = str.Href
					break
				}
			}
			if newsLink == "" {
				newsLink = "Unable to resolve link. Scream at @sharkmoos to fix."
			}

			messageContent := discordMessageData{
				Title:       newFeedItem.Title,
				Description: newFeedItem.Summary.Summary,
				Link:        newsLink,
			}
			newContent = append(newContent, messageContent)
		}
	}

	if len(newContent) > 20 {
		return nil, errors.New("err: more than 20 new items. Logic bug likely. Exiting")
	}
	return newContent, nil
}
