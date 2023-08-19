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

var (
	discordCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "send",
			Description: "Send a link to the news channel",

			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "The link to send",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title",
					Description: "The title of the article",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "description",
					Description: "The description of the article",
					Required:    false,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"send": slashCommandHandler,
	}
)

func slashCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ID == s.State.User.ID {
		return
	}
	if i.ChannelID != adminChannelId {
		log.Println("Not the right channel")
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Please only use this command in <#%s>", adminChannelId),
			},
		})
		if err != nil {
			log.Printf("Interaction response failed: %v", err)
		}
		return
	}

	roles, err := s.GuildRoles(i.GuildID)
	if err != nil {
		fmt.Println("Error retrieving roles:", err)
		return
	}

	var hasRole bool
	for _, role := range roles {
		if !(role.ID == committeeRoleID) && !(role.ID == priorCommitteeRoleID) {
			continue
		}
		for _, memberRole := range i.Member.Roles {
			if memberRole == role.ID {
				hasRole = true
				break
			}
		}
	}
	if !hasRole {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("You do not have the required role to use this command"),
			},
		})
		if err != nil {
			log.Printf("Interaction response failed: %v", err)
		}
	}

	var messageData discordMessageData
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))

	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	if _, ok := optionMap["title"]; ok {
		messageData.Title = "Admin submitted article: " + optionMap["title"].StringValue()
	} else {
		messageData.Title = "Admin submitted article"
	}

	if _, ok := optionMap["description"]; ok {
		messageData.Description = optionMap["description"].StringValue()
	} else {
		messageData.Description = ""
	}

	messageData.Link = optionMap["link"].StringValue()

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Recieved link: %s, title: %s, description: %s", messageData.Link, messageData.Title, messageData.Description),
		},
	})
	if err != nil {
		log.Printf("Interaction response failed: %v", err)
	}

	submitNewRssContent([]discordMessageData{{Title: messageData.Title, Description: messageData.Description, Link: messageData.Link}})

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
