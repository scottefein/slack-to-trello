package slacktotrello

import (
	"appengine"
	"appengine/urlfetch"
	"code.google.com/p/gcfg"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Config struct {
	Trello struct {
		Token  string
		Key    string
		IdList string
	}
}

type SlackMessage struct {
	Token       string
	TeamId      string
	ChannelId   string
	ChannelName string
	UserId      string
	UserName    string
	Command     string
	Text        string
}

func GetConfigs() Config {
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, "app.gcfg")
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

func DecodeSlackMessage(r *http.Request) SlackMessage {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	slack_values := SlackMessage{Token: r.FormValue("token"), TeamId: r.FormValue("team_id"), ChannelId: r.FormValue("channel_id"), ChannelName: r.FormValue("Channel_name"), UserId: r.FormValue("user_id"), UserName: r.FormValue("user_name"), Command: r.FormValue("command"), Text: r.FormValue("text")}
	return slack_values
}

func PostToTrello(r *http.Request, slack_values SlackMessage, cfg Config) (response []byte) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	resp, err := client.PostForm("https://api.trello.com/1/cards", url.Values{
		"token":  {string(cfg.Trello.Token)},
		"key":    {string(cfg.Trello.Key)},
		"due":    {"null"},
		"idList": {string.cfg.Trello.IdList},
		"name":   {string(slack_values.Text)},
		"pos":    {"top"}})
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return response
}

func SendToTrello(w http.ResponseWriter, r *http.Request) {
	slack_values := DecodeSlackMessage(r)
	cfg := GetConfigs()
	response := PostToTrello(r, slack_values, cfg)
	var value map[string]interface{}
	if err := json.Unmarshal(response, &value); err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(w)
	enc.Encode(map[string]interface{}{"text": "Added To Trello"})
}

func init() {
	http.HandleFunc("/send_to_trello", SendToTrello)
}
