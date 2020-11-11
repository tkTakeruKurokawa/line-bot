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
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/kr/pretty"
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

	client, err := maps.NewClient(maps.WithAPIKey("GCP_API"))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	request := &maps.TextSearchRequest{
		Query: "男性古着屋",
		// Query: "(男性衣料品ブランド) || (Clothing Brand && Men)",
		// Query:    "男性セレクトショップ",
		// Location: &maps.LatLng{Lat: 35.665251, Lng: 139.712092}, // 表参道駅
		Location: &maps.LatLng{Lat: 35.006949, Lng: 135.766404}, //御幸町通
		Radius:   1000,
		// Keyword:  "河原町 && (Men's Clothing Shop ||  Men's Used Shop || Men's Select Shop)",
		Type: maps.PlaceTypeClothingStore,
	}

	res, err := client.TextSearch(context.Background(), request)
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	// for {
	// 	for _, result := range res.Results {
	// 		fmt.Println(result.Name)
	// 	}
	// 	if len(res.NextPageToken) == 0 {
	// 		break
	// 	}

	// 	time.Sleep(time.Second * 5)
	// 	request = &maps.TextSearchRequest{
	// 		PageToken: res.NextPageToken,
	// 	}
	// 	res, err = client.TextSearch(context.Background(), request)
	// 	if err != nil {
	// 		log.Fatalf("fatal error: %s", err)
	// 	}
	// }

	for _, result := range res.Results {
		placeDetail := getPlaceDetails(result.PlaceID, client)
		pretty.Println(placeDetail)
	}

	// os.Exit(0)

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
					location := getGeometryLocation(message.Text, client)
					fmt.Println(location)
					if len(location) > 0 {
						longitude, latitude := strconv.FormatFloat(location[0], 'f', -1, 64), strconv.FormatFloat(location[1], 'f', -1, 64)
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(longitude+", "+latitude)).Do(); err != nil {
							log.Print(err)
						}
					} else {
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("入力された地名が見つかりません")).Do(); err != nil {
							log.Print(err)
						}
					}
				case *linebot.LocationMessage:
					log.Println(message.Title)
					log.Println(message.Address)
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

func getGeometryLocation(address string, client *maps.Client) []float64 {
	request := &maps.GeocodingRequest{
		Address: address,
	}

	respons, err := client.Geocode(context.Background(), request)
	fmt.Printf("%+v\n", respons)
	if len(respons) == 0 {
		return []float64{}
	}
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	return []float64{respons[0].Geometry.Location.Lat, respons[0].Geometry.Location.Lng}
}

func getUsedClothingShop(location []float64, client *maps.Client) {
	request := &maps.TextSearchRequest{
		Location: &maps.LatLng{Lat: location[0], Lng: location[1]},
		Radius:   1000,
		Query:    "男性古着屋",
		Type:     maps.PlaceTypeClothingStore,
	}

	response, err := client.TextSearch(context.Background(), request)
	if err != nil {
		log.Fatalf("fatal error: %s", err)

	}

	for {
		for _, result := range response.Results {
			fmt.Println(result.Name)
		}
		if len(response.NextPageToken) == 0 {
			break
		}

		// time.Sleep(time.Second * 5)
		request = &maps.TextSearchRequest{
			PageToken: response.NextPageToken,
		}
		response, err = client.TextSearch(context.Background(), request)
		if err != nil {
			// if response.HTMLAttribution {

			// }
			log.Fatalf("fatal error: %s", err)
		}
	}
	pretty.Println(response)

	// showPlaceResponse(response)
}

// func showPlaceResponse(response maps.PlacesSearchResponse){
// 	for {
// 		for _, result := range response.Results {
// 			fmt.Println(result.Name)
// 		}
// 		if len(response.NextPageToken) == 0 {
// 			break
// 		}

// 		time.Sleep(time.Second * 5)
// 		request = &maps.TextSearchRequest{
// 			PageToken: response.NextPageToken,
// 		}
// 		response, err = client.TextSearch(context.Background(), request)
// 		if err != nil {
// 			log.Fatalf("fatal error: %s", err)
// 		}
// 	}

// 	for _, result := range response.Results {
// 		fmt.Println(result.Name)
// 	}
// 	// pretty.Println(res)
// }

func getPlaceDetails(placeID string, client *maps.Client) maps.PlaceDetailsResult {
	detailRequest := &maps.PlaceDetailsRequest{
		PlaceID:  placeID,
		Language: "ja",
		Fields: []maps.PlaceDetailsFieldMask{
			maps.PlaceDetailsFieldMaskVicinity,
			maps.PlaceDetailsFieldMaskName,
			maps.PlaceDetailsFieldMaskRatings,
			maps.PlaceDetailsFieldMaskUserRatingsTotal,
			maps.PlaceDetailsFieldMaskOpeningHours,
			maps.PlaceDetailsFieldMaskPhotos,
			maps.PlaceDetailsFieldMaskURL,
			maps.PlaceDetailsFieldMaskWebsite,
		},
	}

	detailResult, err := client.PlaceDetails(context.Background(), detailRequest)
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	return detailResult
}
