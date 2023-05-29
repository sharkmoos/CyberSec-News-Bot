package main

import (
  "log"
  "errors"
  "net/http"
  "encoding/xml"
  "strings"
  "regexp"
  "io/ioutil"
)

type HackerNewsRssFeed struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel hackerNewsRssChannel `xml:"channel"`
}

type hackerNewsRssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
  Items       []hackerNewsRssItem    `xml:"item"` // TODO: Refactor as a map, to make comparisons much faster
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



func (hn *HackerNewsRssFeed) getPageCategories(pageUrl string) ( string, error) {
  resp, err := http.Get(pageUrl)
  
  if err != nil {
    log.Panicln(err)
  }

  defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("err:", err)
	}

  content := string(body)
  re := regexp.MustCompile(`<span class='p-tags'(.*)</span>`)
  match := re.FindStringSubmatch(content)
	log.Printf("%v", match)
  if len(match) > 1 {
    log.Printf("Not posting article '%v' due to lack of interesting\n", pageUrl)
    return match[1], nil
	}

	return "", nil

}

func (hn *HackerNewsRssFeed) filterNewsCats(category string) (bool) {
  interestingList := []string {
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

  log.Println(category)
  for _, str := range interestingList {
    if strings.Contains(category, str) {
      return true
    }
  }
  return false
} 

func (hn *HackerNewsRssFeed) ParseNewRssContent(oldData RSSFeed, newData RSSFeed) ([]discordMessageData, error) {
	oldHNData, ok := oldData.(*HackerNewsRssFeed)

	if !ok {
		return nil, errors.New("error: oldData is not of type HackerNewsRssFeed")
	}

	newHNData, ok := newData.(*HackerNewsRssFeed)
	if !ok {
		return nil, errors.New("error: newData is not of type HackerNewsRssFeed")
	}

	var newContent []discordMessageData
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

      interestingCategory := hn.filterNewsCats(category); if !interestingCategory {
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

