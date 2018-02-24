package app

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/pmmp/CrashArchive/app/database"
)

// App ...
type App struct {
	Config   *Config
	Database *database.DB

	mux       sync.Mutex
	slackTime time.Time
}

func (a *App) ReportToSlack(name string, id int64, msg string) {
	if a.Config.SlackURL == "" {
		return
	}

	if !a.slackTime.IsZero() && time.Now().Sub(a.slackTime).Minutes() < 5.0 {
		return
	}

	data := &slackMessage{
		Attachments: []slackAttachment{
			{
				AuthorName: fmt.Sprintf("New report from %s", name),
				Title:      fmt.Sprintf("Report #%d: %s", id, msg),
				TitleLink:  fmt.Sprintf("https://crash.pmmp.io/view/%d", id),
				Color:      "#36a64f",
				Text:       fmt.Sprintf("<https://crash.pmmp.io/download/%d|Download>", id),
			},
		},
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.Encode(data)

	req, err := http.NewRequest("POST", a.Config.SlackURL, buf)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error happened when posting to webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("error happened posting update to webhook")
		log.Println(hex.Dump(buf.Bytes()))
		log.Println("response Status:", resp.Status)
		log.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("response Body:", string(body))
	} else {
		log.Println("posted update to webhook successfully")
	}

	a.mux.Lock()
	a.slackTime = time.Now()
	a.mux.Unlock()
}

type slackMessage struct {
	Attachments []slackAttachment `json:"attachments"`
}
type slackAttachment struct {
	AuthorName string `json:"author_name"`
	Title      string `json:"title"`
	TitleLink  string `json:"title_link"`
	Color      string `json:"color"`
	Text       string `json:"text"`
}
