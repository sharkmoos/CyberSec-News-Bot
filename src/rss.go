/*
Handles the querying of RSS feeds and the logic for handing off to the relevant RSSFeed interface
*/
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

// RSSFeed base interface for all the RSS structs and routines
type RSSFeed interface {
	ParseNewRssContent(oldData RSSFeed, newData RSSFeed) ([]discordMessageData, error)
}

func parseNewRssContent(oldData RSSFeed, newData RSSFeed) ([]discordMessageData, error) {
	/*
	   call the relevant routine based on the type of RSSFeed interface
	*/
	return oldData.ParseNewRssContent(oldData, newData)
}

func queryRssFeed(feedUrl string) (pageData []byte, err error) {
	/*
	   Queries the RSS feed and returns the response body as a byte array
	*/
	var (
		response *http.Response
	)

	if response, err = http.Get(feedUrl); err != nil {
		log.Println("err:", err)
		err = fmt.Errorf("error when quering RSS feed: %v\n", err)
		return
	}

	if response.StatusCode == 429 {
		// rate limited sites like ZDI return this.
		log.Printf("Got response 429 from server %v. Will try again on next iteration.\n", feedUrl)
		err = errors.New("err: got response 429 from server")
		return
	} else if response.StatusCode != 200 {
		err = fmt.Errorf("err: rss status code for %v was '%d' not 200\n", feedUrl, response.StatusCode)
		return
	}

	pageData, err = io.ReadAll(response.Body)
	return
}
