package slacktotrello

import (
	"appengine"
	"appengine/urlfetch"
	"encoding/json"
	"io/ioutil"
	//"log"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type Config struct {
	Trello struct {
		Token         string            `json:"token"`
		Key           string            `json:"key"`
		Lists         map[string]string `json:"lists"`
		TrelloSupport struct {
			Lists map[string]string `json:"lists"`
		} `json:"trello_support"`
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

type UserVoiceMessage struct {
	Data      UserVoiceTicket
	Message   string
	Signature string
	Event     string
}

type UserVoiceTicket struct {
	Ticket struct {
		Subject string
		Url     string
	}
}

func GetConfigs() Config {
	file, _ := os.Open("app.json")
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	return configuration
}

func DecodeSlackMessage(r *http.Request) SlackMessage {
	c := appengine.NewContext(r)
	err := r.ParseForm()
	if err != nil {
		c.Errorf("%v", err)
	}
	slack_values := SlackMessage{Token: r.FormValue("token"), TeamId: r.FormValue("team_id"), ChannelId: r.FormValue("channel_id"), ChannelName: r.FormValue("channel_name"), UserId: r.FormValue("user_id"), UserName: r.FormValue("user_name"), Command: r.FormValue("command"), Text: r.FormValue("text")}
	return slack_values
}

func DecodeUserVoiceMessage(r *http.Request) UserVoiceMessage {
	c := appengine.NewContext(r)
	err := r.ParseForm()
	if err != nil {
		c.Errorf("%v", err)
	}
	var value UserVoiceTicket
	json_structure := r.FormValue("data")
	err = json.Unmarshal([]byte(json_structure), &value)
	c.Infof("User Voice Message %v", value)
	if err != nil {
		c.Criticalf("%v", err)
	}
	user_voice_values := UserVoiceMessage{Data: value, Message: r.FormValue("message"), Signature: r.FormValue("signature"), Event: r.FormValue("event")}
	return user_voice_values
}

func PostToTrello(r *http.Request, list_id string, name string, description string, cfg Config) (response []byte) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	resp, err := client.PostForm("https://api.trello.com/1/cards", url.Values{
		"token":  {string(cfg.Trello.Token)},
		"key":    {string(cfg.Trello.Key)},
		"due":    {"null"},
		"idList": {list_id},
		"name":   {name},
		"desc":   {description},
		"pos":    {"top"}})
	if err != nil {
		c.Criticalf("%v", err)
	}
	defer resp.Body.Close()
	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Criticalf("%v", err)
	}
	return response
}

func UserVoiceToTrello(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	user_voice_values := DecodeUserVoiceMessage(r)
	cfg := GetConfigs()
	list_id := cfg.Trello.TrelloSupport.Lists["tickets_to_respond_to"]
	name := user_voice_values.Data.Ticket.Subject
	url := user_voice_values.Data.Ticket.Url
	description := "URL: " + url
	if user_voice_values.Event == "new_ticket" {
		response := PostToTrello(r, list_id, name, description, cfg)
		var value map[string]interface{}
		if err := json.Unmarshal(response, &value); err != nil {
			c.Errorf("%v", err)
		}
	}
}

func SlackToTrello(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	slack_values := DecodeSlackMessage(r)
	cfg := GetConfigs()
	list_id := string(cfg.Trello.Lists[slack_values.Command])
	name := string(slack_values.Text)
	description := ""
	response := PostToTrello(r, list_id, name, description, cfg)
	var value map[string]interface{}
	if err := json.Unmarshal(response, &value); err != nil {
		c.Errorf("%v", err)
	}
	enc := json.NewEncoder(w)
	enc.Encode(map[string]interface{}{"text": "Added To Trello"})
}

func init() {
	http.HandleFunc("/send_to_trello", SlackToTrello)
	http.HandleFunc("/uservoice_to_trello", UserVoiceToTrello)
}
