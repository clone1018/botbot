package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"github.com/tjarratt/babble"
)

// Variables used for command line parameters
var (
	Token string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

var db *sql.DB

//https://discordapp.com/oauth2/authorize?client_id=682524025544900608&scope=bot
func main() {
	conn, err := sql.Open("sqlite3", "./chat.db")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	db = conn

	//"id": m.ID,
	//	"guild": m.GuildID,
	//	"channel": m.ChannelID,
	//	"author": m.Author.Username,

	sqlStmt := `
	create table IF NOT EXISTS chat (id integer not null primary key, guild_id integer, channel_id integer, author_id integer, author_name text, content text);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)
	//makeTestMessages(dg)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into chat(id, guild_id, channel_id, author_id, author_name, content) " +
		"values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(m.ID, m.GuildID, m.ChannelID, m.Author.ID, m.Author.Username, m.Content)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()

	log.WithFields(log.Fields{
		"id": m.ID,
		"guild": m.GuildID,
		"channel": m.ChannelID,
		"author": m.Author.Username,
	}).Info(m.Content)

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
	// If the message is "pong" reply with "Ping!"
	if strings.HasPrefix(m.Content, "!m") {
		s.ChannelMessageSend(m.ChannelID, "You're doing good work!")
	}
}

func makeTestMessages(s *discordgo.Session) {
	babbler := babble.NewBabbler()
	babbler.Separator = " "
	babbler.Count = 10

	ticker := time.NewTicker(1 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <- ticker.C:
				s.ChannelMessageSend("682527019379982351", babbler.Babble())
				// do stuff
			case <- quit:
				ticker.Stop()
				return
			}
		}
	}()
}
