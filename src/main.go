package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/xml"
	"log"
	"time"
  "sync"
  "os"

	"github.com/bwmarrin/discordgo"
)

// generators for feed structures. Code can generate new RSS structures
// based on the feeds URL
type RSSFeedFactory func() RSSFeed
var rssFeeds = map[string]RSSFeedFactory{
   "https://googleprojectzero.blogspot.com/feeds/posts/default":     func() RSSFeed { return &ProjectZeroRssFeed{} },
  "https://feeds.feedburner.com/TheHackersNews":                     func() RSSFeed { return &HackerNewsRssFeed{} },
  "https://www.zerodayinitiative.com/blog?format=rss":               func() RSSFeed { return &ZDIRssFeed{} },
  "https://portswigger.net/research/rss":                            func() RSSFeed { return &PortSwiggerRSSFeed{} },
  // "http://127.0.0.1:8081/rss_tests/hackernews/xmlfeed.xml":             func() RSSFeed { return &HackerNewsRssFeed{} },
  // "http://127.0.0.1:8081/rss_tests/project_zero/newgooglefeed.xml":     func() RSSFeed { return &ProjectZeroRssFeed{} },
  // "http://127.0.0.1:8081/rss_tests/zdi/feed.xml":                       func() RSSFeed { return &ZDIRssFeed{} },
  // "http://127.0.0.1:8081/rss_tests/portswigger/feed.xml":               func() RSSFeed { return &PortSwiggerRSSFeed{} },
}


const ( 
  pollFreq time.Duration = 10 
)

var (
  dg_ptr *discordgo.Session
  discordToken string 
  newsChannelId string 
)

func rssPollLoop(feedUrl string) {
  // TODO: Pointless to store the entire RSS feed. After unmarshalling we could just keep the most recent 20 results or something
  var (
    pageHash []byte
    pageContents []byte
    oldPageXmlData RSSFeed 
    pageXmlData RSSFeed
  )
  
  // get the starting data
  oldPageHash, err := pollOnce(feedUrl)
  oldPageContents, err := queryRssFeed(feedUrl) 
  createFeedType := rssFeeds[feedUrl] // is a ptr to the relevent struct generator func

  if createFeedType == nil {
    // this will occur if the URL has no feed structure type
    log.Printf("Unsupported RSS feed: %s. Will not create monitor\n", feedUrl)
    return
  }

	oldPageXmlData = createFeedType() 
  pageXmlData = createFeedType()

  if err := xml.Unmarshal(oldPageContents, &oldPageXmlData); err != nil {
    // mostly occurs when the page struct does not represent the XML data closely enough
    log.Panicln("err unmarshaling XML:", err)
    return
  }
  for {
    time.Sleep(pollFreq * time.Minute)
    // time.Sleep(5 * time.Second)
    pageHash, err = pollOnce(feedUrl)
    if err != nil {
      log.Printf("err: %v. Continuing\n", err)
      continue
    }

    if bytes.Equal(oldPageHash, pageHash) {
      log.Printf("hash for '%v' is the same. Sleeping", feedUrl)
      continue
    }

    // Doing another request so soon doesn't player super nice with some feeds with rate limits (looking at you ZDI)
    // TODO: Switch pollOnce to returning the page contents, call page hash from this function, this way after the comparison
    //  we still have the contents to parse
    pageContents, err = queryRssFeed(feedUrl)
    if err != nil {
      log.Printf("err: querying page - %v\tSkipping iteration\n", err)
      continue
    }
    if err := xml.Unmarshal(pageContents, &pageXmlData); err != nil {
      log.Printf("err: unmarshaling XML - %v\tStopping monitor\n", err)
      break
    } 
 
    newRssContent, err := parseNewRssContent(oldPageXmlData, pageXmlData)
    if err != nil {
      log.Printf("err: parsing RSS feed '%v'\tStopping monitor\n", newRssContent)
      break
    }

    submitNewRssContent(newRssContent)
    // the new data replaces the old for future iterations
    oldPageHash = pageHash
    oldPageContents = pageContents
    oldPageXmlData = pageXmlData
  }
}


func pollOnce(feedUrl string) (pageHash []byte, err error) {
  pageContents, err := queryRssFeed(feedUrl) 
  
  if err != nil {
    log.Panicln("err:", err)
  }

  pageHash, err = getPageHash(pageContents)
  if err != nil {
    log.Panicln("err:", err.Error())
  }
  return 
}

func getPageHash(pageBody []byte) (pageHash []byte, errorString error) {
  // sha256 the byte slice of a page. Return the hash as a byte slice.
  hasher := sha256.New()
  hasher.Write(pageBody)
  pageHash = hasher.Sum(nil)
  log.Printf("Hash for webpage was: %x\n", pageHash)
  return
}

func startPollingRss() {
  // hacky way to stop the program from exiting after creating all of the goroutines
  // will idle until all of the goroutines returns
  var wg sync.WaitGroup
  wg.Add(len(rssFeeds))
  for rssFeed := range rssFeeds {
    go func(url string) {
      defer wg.Done()
      rssPollLoop(url)
    }(rssFeed)
  }
  wg.Wait()
}

func main() {
  discordToken = os.Getenv("DISCORD_BOT_TOKEN ")
  newsChannelId = os.Getenv("DISCORD_CHANNEL_ID ")

  if len(discordToken) < 1 || len(newsChannelId) < 1 {
    log.Fatalln("err: reading env vars")
  }

  dg, err := discordgo.New("Bot " + discordToken)
  dg_ptr = dg
  if err != nil {
    log.Println("err: creating Discord session")
  }

  defer dg.Close()

  dg.AddHandlerOnce(func(session *discordgo.Session, event *discordgo.Ready) {
    log.Println("Bot is connected and ready.")
  })

  dg.Open()
  
  log.Println("News polling started")
  startPollingRss()
}
