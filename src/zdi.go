package main 

import (
  "errors"
  "log"
)

type ZDIRssFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Language    string    `xml:"language"`
		PubDate     string    `xml:"pubDate"`
		LastBuild   string    `xml:"lastBuildDate"`
		Items       []ZDIItem `xml:"item"`
	} `xml:"channel"`
}

type ZDIItem struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"`
	GUID        string    `xml:"guid"`
}


func (zdi *ZDIRssFeed) ParseNewRssContent(oldData RSSFeed, newData RSSFeed) ([]discordMessageData, error) {
  oldZdiData, ok := oldData.(*ZDIRssFeed); if !ok {
    return nil, errors.New("oldData is not of type ZDIRssFeed")
  }

  newZdiData, ok := newData.(*ZDIRssFeed); if !ok {
    return nil, errors.New("newData is not of type ZDIRssFeed")
  }

  var newContent []discordMessageData
  for _, newFeedItem := range newZdiData.Channel.Items {
    itemExists := false
    for _, oldFeedItem := range oldZdiData.Channel.Items {
      if oldFeedItem.GUID == newFeedItem.GUID {
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

