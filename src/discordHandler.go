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

func hasAdminRole(userRoles []string) bool {
	/*
		Checks if the user has the committee or prior committee role
	*/
	for _, role := range userRoles {
		if role == committeeRoleID || role == priorCommitteeRoleID {
			return true
		}
	}
	return false
}

func slashCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	/*
		Allow an admin to send links to the news channel. This is a lot better than the previous method !send
		feature, because slash commands move the splitting of data to Discord, so we don't need to create logic for
		parsing link, title etc. from the message.
	*/
	var (
		messageData = discordMessageData{
			Title:       "Admin submitted article",
			Description: "",
			Link:        "",
		}
		err error
	)

	if i.ID == s.State.User.ID {
		return
	}
	if i.ChannelID != adminChannelId {
		log.Println("Not the right channel")
		if s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Please only use this command in <#%s>", adminChannelId),
			},
		}) != nil {
			log.Printf("Interaction response failed: %v", err)
		}
		return
	}

	// I guess this is actually redundant because the channel defined above is currently an admin-only channel,
	// but I'll leave it in, in case we want to change the channel in the future.
	if !hasAdminRole(i.Member.Roles) {
		if s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("You do not have the required role to use this command"),
			},
		}) != nil {
			log.Printf("Interaction response failed: %v", err)
		}
	}

	// retrieve the options from the slash command, thankfully Discord does the parsing for us
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	// this is a required option, so we can assume it exists
	messageData.Link = optionMap["link"].StringValue()

	// not going to get too clever about trying to parse the title and description from the web page or anything
	// just leave it up to the admin
	if _, ok := optionMap["title"]; ok {
		messageData.Title = "Admin submitted article: " + optionMap["title"].StringValue()
	}
	if _, ok := optionMap["description"]; ok {
		messageData.Description = optionMap["description"].StringValue()
	}

	if s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Recieved link: %s, title: %s, description: %s", messageData.Link, messageData.Title, messageData.Description),
		},
	}) != nil {
		log.Printf("Interaction response failed: %v", err)
	}

	// use the same function as the RSS feed to send the message, making a nice rich text embed
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
		sendDiscordMessage(discordSession, embed)
	}
}

// DiscordMessageHandler monitor #disord-updates channel for commands
func discordMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.ChannelID != adminChannelId || !hasAdminRole(m.Member.Roles) {
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
	if _, err := s.ChannelMessageSend(m.ChannelID, response); err != nil {
		fmt.Println("Error sending message:", err)
	}

	if _, err := s.ChannelMessageSend(newsChannelId, "Admin submitted article\n\n"+link); err != nil {
		fmt.Println("Error sending message:", err)
	}
}

func sendDiscordMessage(session *discordgo.Session, message *discordgo.MessageSend) {
	log.Println("Session:", session)
	if _, err := session.ChannelMessageSendComplex(newsChannelId, message); err != nil {
		log.Println("err: Message failed to send - ", err)
	}
}
