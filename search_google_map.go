package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kr/pretty"
	"googlemaps.github.io/maps"
)

// Client クライアントを生成
var Client *maps.Client

const baseURL = "https://maps.googleapis.com/maps/api/place/photo?maxwidth=300&photoreference="
const noImage = "https://via.placeholder.com/150x150?text=NO%20IMAGE"

// NewClient GoogleMapAPIのクライアントを生成
func newClient() {
	var err error
	Client, err = maps.NewClient(maps.WithAPIKey(os.Getenv("GCP_API")))

	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
}

// GetGeometryLocation 検索用語を受け取り，地名を検索し，緯度・経度を返す
func getGeometryLocation(term string) []float64 {
	request := &maps.GeocodingRequest{
		Address:  term,
		Language: "ja",
	}

	respons, err := Client.Geocode(context.Background(), request)
	if len(respons) == 0 {
		return []float64{}
	}
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	return []float64{respons[0].Geometry.Location.Lat, respons[0].Geometry.Location.Lng}
}

// GetUsedClothingShop 緯度・経度を受け取り，付近の古着屋を検索し，検索結果を返す
func getUsedClothingShop(location []float64) ([][]maps.PlacesSearchResult, string) {
	request := &maps.TextSearchRequest{
		Location: &maps.LatLng{Lat: location[0], Lng: location[1]},
		Radius:   1000,
		Query:    "男性古着屋",
		Type:     maps.PlaceTypeClothingStore,
	}

	// for _, result := range response.Results {
	// 	result := GetPlaceDetails(result.PlaceID)
	// 	pretty.Println(GetPlacePhotos(result.Photos))
	// }

	// for {
	// 	for _, result := range response.Results {
	// 		fmt.Println(result.Name)
	// 	}
	// 	if len(response.NextPageToken) == 0 {
	// 		break
	// 	}

	// 	// time.Sleep(time.Second * 5)
	// 	request = &maps.TextSearchRequest{
	// 		PageToken: response.NextPageToken,
	// 	}
	// 	response, err = Client.TextSearch(context.Background(), request)
	// 	if err != nil {
	// 		// if response.HTMLAttribution {

	// 		// }
	// 		log.Fatalf("fatal error: %s", err)
	// 	}
	// }
	// pretty.Println(response)

	return searchShops(request)
}

// GetSelectClothingShop 緯度・経度を受け取り，付近のセレクトショップを検索し，検索結果を返す
func getSelectClothingShop(location []float64) {
	request := &maps.TextSearchRequest{
		Location: &maps.LatLng{Lat: location[0], Lng: location[1]},
		Radius:   1000,
		Query:    "男性セレクトショップ",
		Type:     maps.PlaceTypeClothingStore,
	}

	response, err := Client.TextSearch(context.Background(), request)
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
		response, err = Client.TextSearch(context.Background(), request)
		if err != nil {
			// if response.HTMLAttribution {

			// }
			log.Fatalf("fatal error: %s", err)
		}
	}
	pretty.Println(response)
}

// GetOtherClothingShop 緯度・経度を受け取り，付近の衣料品店を検索し，検索結果を返す
func getOtherClothingShop(location []float64) {
	request := &maps.TextSearchRequest{
		Location: &maps.LatLng{Lat: location[0], Lng: location[1]},
		Radius:   1000,
		Query:    "(男性衣料品ブランド) || (Clothing Brand && Men)",
		Type:     maps.PlaceTypeClothingStore,
	}

	response, err := Client.TextSearch(context.Background(), request)
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
		response, err = Client.TextSearch(context.Background(), request)
		if err != nil {
			// if response.HTMLAttribution {

			// }
			log.Fatalf("fatal error: %s", err)
		}
	}
	pretty.Println(response)
}

// GetPlaceDetails 位置情報を受け取り，その位置の詳細情報を取得し，返す
func getPlaceDetails(placeID string) maps.PlaceDetailsResult {
	detailRequest := &maps.PlaceDetailsRequest{
		PlaceID:  placeID,
		Language: "ja",
		Fields: []maps.PlaceDetailsFieldMask{
			maps.PlaceDetailsFieldMaskPlaceID,
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

	detailResult, err := Client.PlaceDetails(context.Background(), detailRequest)
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	// pretty.Println(detailResult)

	return detailResult
}

// GetPlacePhotos 写真参照コードを受け取り，写真を最大3枚取得し，返す
func getPlacePhotos(photos []maps.Photo) []string {
	var photoResponses []string
	for i := 0; i < 3; i++ {
		photoResponses = append(photoResponses, noImage)
	}
	// photoResponses = make([]maps.PlacePhotoResponse, 3, 3)

	for index, photo := range photos {
		if index > 2 {
			break
		}

		// pretty.Println(photo)

		photoResponses[index] = getPlacePhotoURL(photo.PhotoReference)
	}

	return photoResponses
}

// getPlacePhotoURL 写真参照コードを受け取り，写真のURLを返す
func getPlacePhotoURL(photoReference string) string {
	photoURL := baseURL + photoReference + "&key=" + os.Getenv("GCP_API")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(photoURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if len(resp.Header["Location"]) > 0 {
		return resp.Header["Location"][0]
	}

	return noImage
}

func getNextShops(nextPageToken string) ([][]maps.PlacesSearchResult, string) {
	request := &maps.TextSearchRequest{
		PageToken: nextPageToken,
	}

	return searchShops(request)
}

func searchShops(request *maps.TextSearchRequest) ([][]maps.PlacesSearchResult, string) {
	response, err := Client.TextSearch(context.Background(), request)
	if err != nil {
		log.Fatalf("fatal error: %s", err)

	}

	var shops [][]maps.PlacesSearchResult = make([][]maps.PlacesSearchResult, 2)

	// pretty.Println(shops)

	for index, shopResult := range response.Results {
		if index < 10 {
			shops[0] = append(shops[0], shopResult)
		} else {
			shops[1] = append(shops[1], shopResult)
		}
	}

	return shops, response.NextPageToken
}
