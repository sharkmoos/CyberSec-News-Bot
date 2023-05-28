package main 

import (
  "log"
	

  "github.com/bwmarrin/discordgo"
)

type discordMessageData struct {
  Title string
  Description string
  Link string
}

func submitNewRssContent(newRssContent []discordMessageData) {
  for _, item := range newRssContent {
    // content := fmt.Sprintf("New article from %s\n\n%s: %s\n", "The Hacker News", item.Title, item.Link)
    embed := &discordgo.MessageSend{
      Embeds: []*discordgo.MessageEmbed{{
        Type:        discordgo.EmbedTypeRich,
        Title:       item.Title,
        Description: item.Description,
        Fields: []*discordgo.MessageEmbedField{
         {
            Name:   "Read it here",
            Value:  item.Link,
            Inline: true,
          },
        },
      },},
    }

    log.Println("Sending message:", item.Title)
    sendDiscordMessage(dg_ptr, embed)
  }
}

func sendDiscordMessage(session *discordgo.Session, message *discordgo.MessageSend) {
  log.Println("Session:", session)
  _, err := session.ChannelMessageSendComplex(newsChannelId, message)
  if err != nil {
    log.Println("err: Message failed to send - ", err)
  }
}


