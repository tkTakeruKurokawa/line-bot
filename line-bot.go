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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"reflect"

	"github.com/line/line-bot-sdk-go/linebot"
	"googlemaps.github.io/maps"
)

// ShopData 店の検索結果
type ShopData struct {
	NextShops     []maps.PlacesSearchResult
	NextPageToken string
}

// SearchData 検索に使うデータ
type SearchData struct {
	Type          string
	TypeName      string
	Location      []float64
	LocationName  string
	UserID        string
	ReplyToken    string
	SelectMessage *linebot.ButtonsTemplate
}

// ClickLocker 次の10件を検索するを何回押したかカウントする
type ClickLocker struct {
	counter int
}

func (cc *ClickLocker) nowCount() int {
	return cc.counter
}

func (cc *ClickLocker) inclement() {
	cc.counter++
}

func (cc *ClickLocker) reset() {
	cc.counter = 0
}

func initializeSearchData() *SearchData {
	return &SearchData{
		SelectMessage: &linebot.ButtonsTemplate{
			Text: "検索する店の種類を選んで下さい",
			Actions: []linebot.TemplateAction{
				&linebot.PostbackAction{
					Label: "古着屋",
					Data:  "used",
				},
				&linebot.PostbackAction{
					Label: "セレクトショップ",
					Data:  "select",
				},
				&linebot.PostbackAction{
					Label: "その他の衣料品店",
					Data:  "other",
				},
				&linebot.PostbackAction{
					Label: "カフェ",
					Data:  "cafe",
				},
			},
		},
	}
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

	shopData := &ShopData{}
	searchData := initializeSearchData()
	clickLocker := new(ClickLocker)

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
			searchData.UserID = event.Source.UserID
			searchData.ReplyToken = event.ReplyToken

			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					searchData.Location = getGeometryLocation(message.Text)
					searchData.LocationName = message.Text

					if len(searchData.Location) > 0 {
					} else {
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("入力された地名が見つかりません")).Do(); err != nil {
							log.Print(err)
						}
						return
					}

				case *linebot.LocationMessage:
					searchData.Location = []float64{message.Latitude, message.Longitude}
					searchData.LocationName = message.Address
				}
			}
			if event.Type == linebot.EventTypePostback {
				switch data := event.Postback.Data; data {
				case "next":
					searchData.Type = data

				case "used":
					searchData.Type = data
					searchData.TypeName = "古着屋"

				case "select":
					searchData.Type = data
					searchData.TypeName = "セレクトショップ"
				case "other":
					searchData.Type = data
					searchData.TypeName = "その他の衣料品店"
				case "cafe":
					searchData.Type = data
					searchData.TypeName = "カフェ"
				}
			}
		}

		clickLocker.inclement()
		if clickLocker.nowCount() == 1 {
			shopData, searchData = startSearchOrSendMessage(bot, shopData, searchData)
			clickLocker.reset()
		}
	})
	// This is just sample code.
	// For actual use, you must support HTTPS by using `ListenAndServeTLS`, a reverse proxy or something else.
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}

// startSearchOrSendMessage 検索を開始するもしくは次の行動を促すメッセージを送信する
func startSearchOrSendMessage(bot *linebot.Client, shopData *ShopData, searchData *SearchData) (*ShopData, *SearchData) {
	var situationMessage, nextActionMessage linebot.SendingMessage

	switch {
	case searchData.Type == "next":
		shopData = executeNextAction(bot, shopData, searchData)
		searchData = initializeSearchData()
		return shopData, searchData

	case len(searchData.Location) > 0 && len(searchData.Type) > 0:
		if _, err := bot.PushMessage(searchData.UserID, linebot.NewTextMessage("種類： "+searchData.TypeName+"\n場所: "+searchData.LocationName), linebot.NewTextMessage("上記内容で検索します")).Do(); err != nil {
			log.Print(err)
		}

		shopData = buildAndSendFlexMessage(searchData.Location, searchData.Type, searchData.ReplyToken)
		searchData = initializeSearchData()
		return shopData, searchData

	case len(searchData.Location) > 0 && len(searchData.Type) == 0:
		situationMessage = linebot.NewTextMessage("種類： " + searchData.TypeName + "\n場所: " + searchData.LocationName)
		nextActionMessage = linebot.NewTemplateMessage("検索する店の種類を選んで下さい", searchData.SelectMessage)

	case len(searchData.Location) == 0 && len(searchData.Type) > 0:
		situationMessage = linebot.NewTextMessage("種類： " + searchData.TypeName + "\n場所: " + searchData.LocationName)
		nextActionMessage = linebot.NewTextMessage("位置情報を送るか検索したい場所の名称を送ってください\n(例：東京駅)")
	}

	shopData = &ShopData{}
	if _, err := bot.PushMessage(searchData.UserID, situationMessage, nextActionMessage).Do(); err != nil {
		log.Print(err)
	}

	return shopData, searchData
}

// executeNextAction 次の10件を検索する
func executeNextAction(bot *linebot.Client, shopData *ShopData, searchData *SearchData) *ShopData {
	if reflect.ValueOf(shopData.NextShops).IsNil() && len(shopData.NextPageToken) == 0 {
		if _, err := bot.PushMessage(searchData.UserID, linebot.NewTextMessage("検索できません．検索場所，検索対象を入力して下さい")).Do(); err != nil {
			log.Print(err)
		}

	} else {
		if _, err := bot.PushMessage(searchData.UserID, linebot.NewTextMessage("次の10件を検索します")).Do(); err != nil {
			log.Print(err)
		}

		shopData = buildAndSendNextFlexMessage(shopData, searchData.ReplyToken)

		if reflect.ValueOf(shopData.NextShops).IsNil() && len(shopData.NextPageToken) == 0 {
			if _, err := bot.PushMessage(searchData.UserID, linebot.NewTextMessage("最大検索数に達したため，検索を終了します")).Do(); err != nil {
				log.Print(err)
			}
		}
	}

	return shopData
}

// buildAndSendFlexMessage FlexMessageを構築し，送信する
func buildAndSendFlexMessage(location []float64, shopType string, replyToken string) *ShopData {
	shopData, nextPageToken := getShopData(location, shopType)

	return sendMessageAndBuildShopData(shopData, replyToken, nextPageToken)
}

// buildAndSendNextFlexMessage 次の10件のFlexMessageを構築し，送信する
func buildAndSendNextFlexMessage(shopData *ShopData, replyToken string) *ShopData {
	if reflect.ValueOf(shopData.NextShops).IsNil() {
		shopData, nextPageToken := getNextShops(shopData.NextPageToken)

		return sendMessageAndBuildShopData(shopData, replyToken, nextPageToken)
	}

	shops := [][]maps.PlacesSearchResult{shopData.NextShops, nil}

	return sendMessageAndBuildShopData(shops, replyToken, shopData.NextPageToken)
}

// sendMessageAndBuildShopData FlexMessageを構築し，送信する
func sendMessageAndBuildShopData(shopData [][]maps.PlacesSearchResult, replyToken string, nextPageToken string) *ShopData {
	bubbles := getBubbles(shopData, nextPageToken)
	sendFlexMessage(bubbles, replyToken)

	return &ShopData{
		NextShops:     shopData[1],
		NextPageToken: nextPageToken,
	}
}

// getBubbles FlexMessageを構成するバブルを構築する
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

// sendFlexMessage http.Clientを利用してFlexMessageを送る
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

// buildRequest リクエストを構築する
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
