package main

import (
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"time"
)

func init() {

	//flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

var db *sql.DB

func main() {
	conn, err := sql.Open("sqlite3", "../chat.db")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	db = conn


	http.HandleFunc("/logs", logHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getMessages() ChatMessages {
	var messages ChatMessages

	rows, err := db.Query("select id, guild_id, channel_id, author_id, author_name, content from chat order by id desc")
	if err != nil {
		log.Error(err)
	}
	defer rows.Close()
	for rows.Next() {
		var message ChatMessage
		err = rows.Scan(
			&message.Id,
			&message.GuildId,
			&message.ChannelId,
			&message.AuthorId,
			&message.AuthorName,
			&message.Content,
		)
		if err != nil {
			log.Error(err)
		}

		message.Time = snowflakeToUnix(message.Id)

		messages = append(messages, message)
	}
	err = rows.Err()
	if err != nil {
		log.Error(err)
	}

	return messages
}

type ChatMessage struct {
	Id int64
	GuildId int64
	ChannelId int64
	AuthorId int64
	AuthorName string
	Content string

	Time time.Time
}
type ChatMessages []ChatMessage

func logHandler(w http.ResponseWriter, r *http.Request) {

	messages := getMessages()
	t, _ := template.ParseFiles("log.html")
	t.Execute(w, messages)
}

func snowflakeToUnix(snowflake int64) time.Time {
	return time.Unix(0, ((snowflake >> 22) + 1420070400000) * int64(1000000))
}