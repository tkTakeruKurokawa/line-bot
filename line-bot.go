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
	"reflect"

	// "github.com/kr/pretty"
	// "github.com/kr/pretty"

	"github.com/kr/pretty"
	"github.com/line/line-bot-sdk-go/linebot"
	"googlemaps.github.io/maps"
)

// ShopData 店の検索結果
type ShopData struct {
	NextShops []maps.PlacesSearchResult
	// ShopFrom10to19 []maps.PlacesSearchResult
	NextPageToken string
}

// ClickCounter 次の10件を検索するを何回押したかカウントする
type ClickCounter struct {
	counter int
}

func (cc *ClickCounter) nowCount() int {
	return cc.counter
}

func (cc *ClickCounter) inclement() {
	cc.counter++
}

func (cc *ClickCounter) reset() {
	cc.counter = 0
}

func main() {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	newClient()

	clickCounter := new(ClickCounter)
	shopData := &ShopData{}

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
			pretty.Println("Next Page Token: ", shopData.NextPageToken)

			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					// if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("alt txt", Get())).Do(); err != nil {
					// 	log.Print(err)
					// }

					location := getGeometryLocation(message.Text)

					if len(location) > 0 {
						shopData = buildAndSendFlexMessage(location, event.ReplyToken)
						// shopData.NextShops = shopData.ShopFrom10to19
						// shopData.ShopFrom10to19 = []maps.PlacesSearchResult{}
					} else {
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("入力された地名が見つかりません")).Do(); err != nil {
							log.Print(err)
						}
					}
				case *linebot.LocationMessage:
					location := []float64{message.Latitude, message.Longitude}
					shopData = buildAndSendFlexMessage(location, event.ReplyToken)
					// if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(detail.URL), linebot.NewTextMessage(detail.Website)).Do(); err != nil {
					// 	log.Print(err)
					// }
				}
			}
			if event.Type == linebot.EventTypePostback {
				if event.Postback.Data == "continue" {
					clickCounter.inclement()
					// fmt.Println(clickCounter.nowCount())

					if clickCounter.nowCount() == 1 {
						if _, err = bot.PushMessage(event.Source.UserID, linebot.NewTextMessage("次の10件を検索します")).Do(); err != nil {
							log.Print(err)
						}
						shopData = buildAndSendNextFlexMessage(shopData, event.ReplyToken)
						clickCounter.reset()
					}
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

func buildAndSendFlexMessage(location []float64, replyToken string) *ShopData {
	shopData, nextPageToken := getUsedClothingShop(location)

	return sendMessageAndBuildShopData(shopData, replyToken, nextPageToken)
}

func buildAndSendNextFlexMessage(shopData *ShopData, replyToken string) *ShopData {
	if reflect.ValueOf(shopData.NextShops).IsNil() {
		shopData, nextPageToken := getNextShops(shopData.NextPageToken)

		return sendMessageAndBuildShopData(shopData, replyToken, nextPageToken)
	}

	shops := [][]maps.PlacesSearchResult{shopData.NextShops, nil}

	return sendMessageAndBuildShopData(shops, replyToken, shopData.NextPageToken)
}

func sendMessageAndBuildShopData(shopData [][]maps.PlacesSearchResult, replyToken string, nextPageToken string) *ShopData {
	bubbles := getBubbles(shopData, nextPageToken)
	sendFlexMessage(bubbles, replyToken)

	return &ShopData{
		NextShops:     shopData[1],
		NextPageToken: nextPageToken,
	}
}

func getBubbles(shopData [][]maps.PlacesSearchResult, nextPageToken string) []*Bubble {
	var shopDetails []maps.PlaceDetailsResult
	var photos [][]string
	var bubbles []*Bubble

	for _, shop := range shopData[0] {
		shopDetail := getPlaceDetails(shop.PlaceID)
		shopDetails = append(shopDetails, shopDetail)

		photo := getPlacePhotos(shopDetail.Photos)
		photos = append(photos, photo)

		bubbles = append(bubbles, getShopBubble(shopDetail, photo))
	}
	if !(reflect.ValueOf(shopData[1]).IsNil() && len(nextPageToken) == 0) {
		bubbles = append(bubbles, getNextActionBubble())
	}

	return bubbles
}

func sendFlexMessage(bubbles []*Bubble, replyToken string) {
	req, err := buildRequest(bubbles, replyToken)
	if err != nil {
		fmt.Println(err.Error())
	}

	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer res.Body.Close()

	dumpResp, _ := httputil.DumpResponse(res, true)
	fmt.Printf("%s\n\n", dumpResp)
}

func buildRequest(bubbles []*Bubble, replyToken string) (*http.Request, error) {
	message, err := json.Marshal(getFlexMessage(bubbles, replyToken))
	if err != nil {
		fmt.Println(err.Error())
	}

	req, err := http.NewRequest("POST", "https://api.line.me/v2/bot/message/reply", bytes.NewReader(message))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer {"+os.Getenv("CHANNEL_TOKEN")+"}")

	return req, nil
}
