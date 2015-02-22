package slacktotrello

import (
	"appengine"
	"appengine/urlfetch"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Config struct {
	Trello struct {
		Token string            `json:"token"`
		Key   string            `json:"key"`
		Lists map[string]string `json:"lists"`
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
	file, _ := os.Open("app.json")
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(configuration.Trello.Lists)
	return configuration
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
		"idList": {string(cfg.Trello.Lists[slack_values.Command])},
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
