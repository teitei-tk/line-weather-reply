package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
)

const (
	apiBase = "http://weather.livedoor.com/forecast/webservice/json/v1"

	// from http://weather.livedoor.com/forecast/rss/primary_area.xml
	cityID = 130010
)

// from http://weather.livedoor.com/weather_hacks/webservice
type response struct {
	PinpointLocations []struct {
		Link string `json:"link"`
		Name string `json:"name"`
	} `json:"pinpointLocations"`
	Link      string `json:"link"`
	Forecasts []struct {
		DateLabel   string `json:"dateLabel"`
		Telop       string `json:"telop"`
		Date        string `json:"date"`
		Temperature struct {
			Min interface{} `json:"min"`
			Max interface{} `json:"max"`
		} `json:"temperature"`
		Image struct {
			Width  int    `json:"width"`
			URL    string `json:"url"`
			Title  string `json:"title"`
			Height int    `json:"height"`
		} `json:"image"`
	} `json:"forecasts"`
	Location struct {
		City       string `json:"city"`
		Area       string `json:"area"`
		Prefecture string `json:"prefecture"`
	} `json:"location"`
	PublicTime string `json:"publicTime"`
	Copyright  struct {
		Provider []struct {
			Link string `json:"link"`
			Name string `json:"name"`
		} `json:"provider"`
		Link  string `json:"link"`
		Title string `json:"title"`
		Image struct {
			Width  int    `json:"width"`
			Link   string `json:"link"`
			URL    string `json:"url"`
			Title  string `json:"title"`
			Height int    `json:"height"`
		} `json:"image"`
	} `json:"copyright"`
	Title       string `json:"title"`
	Description struct {
		Text       string `json:"text"`
		PublicTime string `json:"publicTime"`
	} `json:"description"`
}

func requestAPI() (*response, error) {
	reqURL := fmt.Sprintf("%s?city=%d", apiBase, cityID)
	_, err := url.ParseRequestURI(reqURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	result := &response{}
	json.NewDecoder(res.Body).Decode(result)

	return result, nil
}

func replyText(r *response) string {
	return fmt.Sprintf("%s, %s", r.Title, r.Description)
}

func main() {
	fmt.Println(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"))

	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}

		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				_, ok := event.Message.(*linebot.TextMessage)
				if !ok {
					return
				}

				res, err := requestAPI()
				if err != nil {
					log.Fatal(err)
				}

				msg := linebot.NewTextMessage(replyText(res))
				if _, err = bot.ReplyMessage(event.ReplyToken, msg).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
