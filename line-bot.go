// Copyright 2016 LINE Corporation
//
// LINE Corporation licenses this file to you under the Apache License,
// version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at:
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	// "fmt"
	// "fmt"

	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	// "github.com/kr/pretty"
	// "github.com/kr/pretty"

	"github.com/line/line-bot-sdk-go/linebot"
	"googlemaps.github.io/maps"
)

func main() {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	newClient()

	// Setup HTTP Server for receiving requests from LINE platform
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
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					// if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("alt txt", Get())).Do(); err != nil {
					// 	log.Print(err)
					// }

					location := getGeometryLocation(message.Text)

					if len(location) > 0 {
						searchResults := getUsedClothingShop(location)
						detailResults := getPlaceDetailResults(searchResults)
						sendFlexMessages(detailResults, event.ReplyToken)
					} else {
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("入力された地名が見つかりません")).Do(); err != nil {
							log.Print(err)
						}
					}
				case *linebot.LocationMessage:
					location := []float64{message.Latitude, message.Longitude}
					searchResults := getUsedClothingShop(location)
					detailResults := getPlaceDetailResults(searchResults)
					sendFlexMessages(detailResults, event.ReplyToken)
					// if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(detail.URL), linebot.NewTextMessage(detail.Website)).Do(); err != nil {
					// 	log.Print(err)
					// }
				}
			}
		}
	})
	// This is just sample code.
	// For actual use, you must support HTTPS by using `ListenAndServeTLS`, a reverse proxy or something else.
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}

func newRequest(flex Flex, replyToken string) (*http.Request, error) {
	message, err := json.Marshal(&struct {
		ReplyToken string `json:"replyToken"`
		Messages   []Flex `json:"messages"`
	}{
		ReplyToken: replyToken,
		Messages:   []Flex{flex},
	})

	// pretty.Println(string(message))

	// pretty.Println(flex)
	req, err := http.NewRequest("POST", "https://api.line.me/v2/bot/message/reply", bytes.NewReader(message))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer {"+os.Getenv("CHANNEL_TOKEN")+"}")

	return req, nil
}

func sendFlexMessages(detailResults []maps.PlaceDetailsResult, replyToken string) {
	flexFormer, _ := getFlexMessage(detailResults)
	// for _, flex := range []Flex{flexFormer, flexLatter} {

	// }
	req, err := newRequest(flexFormer, replyToken)
	if err != nil {
		fmt.Println(err.Error())
	}

	client := new(http.Client)
	res, err := client.Do(req)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	dumpResp, _ := httputil.DumpResponse(res, true)
	fmt.Printf("%s", dumpResp)
	// res, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	defer res.Body.Close()
	// pretty.Println(res.)
}
