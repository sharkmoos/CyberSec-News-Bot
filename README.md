# ComSec News Bot

This is a Discord bot, written in Go, to retrieve news from interesting cybersecurity news outlets and send the article links
to the ComSec Discord server. Currently the bot polls the RSS feeds of the interesting websites. It is designed to be relatively 
low maintanence and simple to add new locations to pull information from.

## Adding a New RSS News Outlet

1. Create a new file for the RSS feed. Name it something like `feedName.go`, in the same directory as the main.go file.
2. Write the go structures of the XML data of the RSS feed. It does not need to be a comprehensive representation, as long as the fields
relevent to send to the Discord server are accessbile. 
3. Add the RSS endpoint to the map of endpoints in `main.go`. The string URL of the endpoint and the type of the top level structure should be added here.
4. Write a new interface method for finding new articles from the feed. An example of this would be `(hn *HackerNewsRssFeed) ParseNewRssContent(oldData RSSFeed, newData RSSFeed)`. 
Most of the code can just be copy & pasted, just changing the fields that get compared and the types to be casted.

