package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type discordMessageData struct {
	Title       string
	Description string
	Link        string
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
			}},
		}

		log.Println("Sending message:", item.Title)
		sendDiscordMessage(dg_ptr, embed)
	}
}

// DiscordMessageHandler monitor #disord-updates channel for commands
func discordMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.ChannelID != adminChannelId {
		log.Println("Not the right channel")
		return
	}

	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		fmt.Println("Error retrieving roles:", err)
		return
	}

	var hasRole bool
	for _, role := range roles {
		if !(role.ID == committeeRoleID) && !(role.ID == priorCommitteeRoleID) {
			continue
		}
		for _, memberRole := range m.Member.Roles {
			if memberRole == role.ID {
				hasRole = true
				break
			}
		}
	}
	if !hasRole {
		return
	}

	// Check if the message content starts with "!send"
	content := strings.TrimSpace(m.Content)
	if !strings.Contains(content, "!send") {
		return
	}

	// extract the link from the message by splitting starting at https:// and ending at a space
	link := "https://" + strings.Split(content, "https://")[1]

	// Respond to the message with the link
	response := fmt.Sprintf("Received link: %s", link)
	_, err = s.ChannelMessageSend(m.ChannelID, response)
	if err != nil {
		fmt.Println("Error sending message:", err)
	}

	_, err = s.ChannelMessageSend(newsChannelId, "Admin submitted article\n\n"+link)
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}

func sendDiscordMessage(session *discordgo.Session, message *discordgo.MessageSend) {
	log.Println("Session:", session)
	_, err := session.ChannelMessageSendComplex(newsChannelId, message)
	if err != nil {
		log.Println("err: Message failed to send - ", err)
	}
}
