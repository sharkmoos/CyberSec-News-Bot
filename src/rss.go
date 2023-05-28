package main

import (
  "fmt"
	"errors"
	"log"
	"net/http"
	"io/ioutil"
)

// base interface for all of the RSS structs and routines
type RSSFeed interface {
	ParseNewRssContent(oldData RSSFeed, newData RSSFeed) ([]discordMessageData, error)
}

func parseNewRssContent(oldData RSSFeed, newData RSSFeed) ([]discordMessageData, error) {
  // call the relevent routine based on the type of RSSFeed interface
  return oldData.ParseNewRssContent(oldData, newData)
}

func queryRssFeed(feedUrl string) (pagePtr []byte, errorString error) {
  response, err := http.Get(feedUrl)
  if err != nil {
    log.Println("err:", err)
    errorString = fmt.Errorf("error when quering RSS feed: %v\n", err)
    return
  }
  
  if response.StatusCode == 429 {
    // rate limited sites like ZDI return this. 
    log.Printf("Got response 429 from server %v. Will try again on next iteration.\n", feedUrl)
    errorString = errors.New("Got response 429 from server")
    return
  } else if response.StatusCode != 200 {
    errorString = fmt.Errorf("err: rss status code for %v was '%d' not 200\n", feedUrl, response.StatusCode)
    return
  }

  body, err := ioutil.ReadAll(response.Body)
  if err != nil {
    errorString = err
    return
  }

  pagePtr = body
  return
}


