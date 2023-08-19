package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/xml"
	"log"
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// generators for feed structures. Code can generate new RSS structures
// based on the feeds URL
type RSSFeedFactory func() RSSFeed

var rssFeeds = map[string]RSSFeedFactory{
	"https://googleprojectzero.blogspot.com/feeds/posts/default": func() RSSFeed { return &ProjectZeroRssFeed{} },
	"https://feeds.feedburner.com/TheHackersNews":                func() RSSFeed { return &HackerNewsRssFeed{} },
	"https://www.zerodayinitiative.com/blog?format=rss":          func() RSSFeed { return &ZDIRssFeed{} },
	"https://portswigger.net/research/rss":                       func() RSSFeed { return &PortSwiggerRSSFeed{} },
	// "http://127.0.0.1:8081/rss_tests/hackernews/xmlfeed.xml":             func() RSSFeed { return &HackerNewsRssFeed{} },
	// "http://127.0.0.1:8081/rss_tests/project_zero/newgooglefeed.xml":     func() RSSFeed { return &ProjectZeroRssFeed{} },
	// "http://127.0.0.1:8081/rss_tests/zdi/feed.xml":                       func() RSSFeed { return &ZDIRssFeed{} },
	// "http://127.0.0.1:8081/rss_tests/portswigger/feed.xml":               func() RSSFeed { return &PortSwiggerRSSFeed{} },
}

const (
	pollFreq time.Duration = 10
)

var (
	discordSession       *discordgo.Session
	discordToken         string
	newsChannelId        string
	serverId             string
	committeeRoleID      string
	priorCommitteeRoleID string
	adminChannelId       string
)

func rssPollLoop(feedUrl string) {
	// TODO: Pointless to store the entire RSS feed. After unmarshalling we could just keep the most recent 20 results or something
	var (
		pageHash        []byte
		oldPageHash     []byte
		pageContents    []byte
		oldPageContents []byte
		oldPageXmlData  RSSFeed
		pageXmlData     RSSFeed
	)

	oldPageContents, err := queryRssFeed(feedUrl)
	oldPageHash, err = getPageHash(oldPageContents)
	if err != nil {
		log.Panicln("err:", err.Error())
	}
	log.Printf("starting with page hash %x for site %s", oldPageHash, feedUrl)

	oldPageXmlData = rssFeeds[feedUrl]()
	pageXmlData = rssFeeds[feedUrl]()

	if err := xml.Unmarshal(oldPageContents, &oldPageXmlData); err != nil {
		// mostly occurs when the page struct does not represent the XML data closely enough
		log.Panicln("err unmarshaling XML:", err)
		return
	}
	for {
		time.Sleep(pollFreq * time.Minute)
		// time.Sleep(5 * time.Second)
		pageContents, err = queryRssFeed(feedUrl)
		if err != nil {
			log.Printf("err: %v. Continuing\n", err)
			continue
		}

		pageHash, err = getPageHash(pageContents)
		if err != nil {
			log.Panicln("err:", err.Error())
		}

		if bytes.Equal(oldPageHash, pageHash) {
			log.Printf("hash for '%v' is the same. Sleeping", feedUrl)
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
		pageXmlData = rssFeeds[feedUrl]()
	}
}

func getPageHash(pageBody []byte) (pageHash []byte, errorString error) {
	/*
		sha256 the byte slice of a page. Return the hash as a byte slice.
	*/
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

func setCommitteeRoles(roles []*discordgo.Role) bool {
	/*
		Find the role IDs for the committee and prior committee roles.
	*/
	for _, role := range roles {
		if role.Name == "Committee" {
			committeeRoleID = role.ID
		} else if role.Name == "Prior Committee" {
			priorCommitteeRoleID = role.ID
		}
		if committeeRoleID != "" && priorCommitteeRoleID != "" {
			return true
		}
	}
	return false
}

func main() {
	var (
		err   error
		roles []*discordgo.Role
	)

	discordToken = os.Getenv("DISCORD_BOT_TOKEN")
	newsChannelId = os.Getenv("DISCORD_CHANNEL_ID")
	adminChannelId = os.Getenv("ADMIN_CHANNEL_ID")
	serverId = os.Getenv("DISCORD_SERVER_ID")

	if len(discordToken) < 1 || len(newsChannelId) < 1 {
		log.Fatalln("err: reading env vars")
	}

	if discordSession, err = discordgo.New("Bot " + discordToken); err != nil {
		log.Fatalln("err: creating Discord session")
	}

	discordSession.AddHandlerOnce(func(session *discordgo.Session, event *discordgo.Ready) {
		log.Println("Bot is connected and ready.")
	})
	discordSession.AddHandler(discordMessageHandler)
	discordSession.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	if err = discordSession.Open(); err != nil {
		log.Fatalln("err: opening connection to Discord")
	}

	if roles, err = discordSession.GuildRoles(serverId); err != nil {
		log.Fatalln("err: retrieving roles:")
	}

	if !setCommitteeRoles(roles) {
		log.Fatalln("err: retrieving roles:")
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(discordCommands))
	for i, v := range discordCommands {
		cmd, err := discordSession.ApplicationCommandCreate(discordSession.State.User.ID, serverId, v)
		if err != nil {
			log.Panicln("err: creating application command")
		}
		registeredCommands[i] = cmd
	}

	defer discordSession.Close()
	log.Println("News polling started")
	startPollingRss()
}
