package main

import (
	"C"
	"bytes"
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// main function
func main() {
	dg, e := discordgo.New("Bot " + cfg.BotToken)
	if e != nil {
		log.Println("Error:", e)
		return
	}
	dg.AddHandlerOnce(ready)
	dg.AddHandler(presenceUpdate)
	if e := dg.Open(); e != nil {
		log.Println("Error:", e)
		return
	}
	<-make(chan struct{})
	return
}

// vars and declarations here
var (
	adminRoleID string
	botID       string
	configText  []byte = []byte("{\n\t\"BotToken\": \"\",\n\t\"GuildID\": \"\"\n}")
)

type Configuration struct {
	BotToken string
	GuildID  string
}

var cfg Configuration

// functions after here

func init() {
	if _, e := os.Stat("config.json"); os.IsNotExist(e) {
		os.Create("config.json")
		ioutil.WriteFile("config.json", configText, os.ModePerm)
		log.Println("No config file found, creating one. Please configure and restart")
		return
	}
	configFile, _ := os.Open("config.json")
	configFileContents, e := ioutil.ReadFile("config.json")
	if e != nil {
		log.Println(e)
		return
	}
	if bytes.Compare(configFileContents, configText) == 0 {
		log.Println("Not configured, aborting. Please configure and restart")
		return
	} else {
		decoder := json.NewDecoder(configFile)
		cfg = Configuration{}
		e := decoder.Decode(&cfg)
		if e != nil {
			log.Println("Error:", e)
			return
		}
	}
}
func ready(s *discordgo.Session, event *discordgo.Event) {
	go func() {
		time.Sleep(time.Second * 2)
		guild, e := s.Guild(cfg.GuildID)
		if e != nil {
			log.Println("Error:", e)
		}
		for _, p := range guild.Presences {
			correctRoles(s, p)
		}
	}()
}
func presenceUpdate(s *discordgo.Session, p *discordgo.PresenceUpdate) {
	correctRoles(s, &p.Presence)
}
func correctRoles(s *discordgo.Session, p *discordgo.Presence) {
	if authenticate(s, p.User) {
		log.Println("Admin or trusted user, cannot modify role")
		return
	}
	var updatedRoles []string
	var role string

	guildRoles, e := s.GuildRoles(cfg.GuildID)
	if e != nil {
		log.Println("Error:", e)
		return
	}
	for _, gr := range guildRoles {
		if p.User.Bot {
			if gr.Name == "Bots" {
				role = "Bot"
				updatedRoles = append(updatedRoles, gr.ID)
			}
		} else {
			if p.Game != nil {
				if gr.Name == p.Game.Name {
					role = gr.Name
					updatedRoles = append(updatedRoles, gr.ID)
				}
			} else {
				guild, e := s.Guild(cfg.GuildID)
				if e != nil {
					log.Println("Error:", e)
				}
				role = guild.Name
				updatedRoles = append(updatedRoles, guild.ID)
			}
			if role == "" {
				for _, gr := range guildRoles {
					if gr.Name == "Other Games" {
						updatedRoles = append(updatedRoles, gr.ID)
						role = "Other Games"
					}
				}
			}
			if role == "" {
				guild, e := s.Guild(cfg.GuildID)
				if e != nil {
					log.Println("Error:", e)
				}
				role = guild.Name
				updatedRoles = append(updatedRoles, guild.ID)
				log.Println("No role by name \"Other Games\", putting user in default role")
			}
		}
	}
	log.Println("Changing user role to:", role)
	if e := s.GuildMemberEdit(cfg.GuildID, p.User.ID, updatedRoles); e != nil {
		log.Println("Error:", e)
	}
}
func authenticate(s *discordgo.Session, u *discordgo.User) bool {
	user, e := s.GuildMember(cfg.GuildID, u.ID)
	if e != nil {
		log.Println("Error:", e)
		return false
	}
	roles, e := s.GuildRoles(cfg.GuildID)
	if e != nil {
		log.Println("Error:", e)
		return false
	}
	for _, ar := range roles {
		if ar.Name == "Troll Patrol" || ar.Name == "The Crew" {
			adminRoleID = ar.ID
		}
	}
	if adminRoleID != "" {
		for _, r := range user.Roles {
			if r == adminRoleID {
				return true
			}
		}
	} else {
		log.Println("No role by name of \"Admin\", Things might not go so well :/")
		return false
	}
	return false
}
