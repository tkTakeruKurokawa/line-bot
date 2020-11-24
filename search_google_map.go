package main

import (
	"context"
	"log"
	"net/http"
	"os"

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

// getGeometryLocation 検索用語を受け取り，地名を検索し，緯度・経度を返す
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

// getShopData 緯度経度，検索対象を受け取り，検索し，結果一覧を返す
func getShopData(location []float64, shopType string) ([][]maps.PlacesSearchResult, string) {
	keyword, category := getQuery(shopType)
	if len(keyword) == 0 {
		log.Fatalln("Invalid shop type")
	}

	request := &maps.NearbySearchRequest{
		Location: &maps.LatLng{Lat: location[0], Lng: location[1]},
		RankBy:   maps.RankByDistance,
		Keyword:  keyword,
		Type:     category,
	}

	return searchShops(request)
}

// getQuery 検索対象を受け取り，それに応じた検索用語と検索場所の種類を返す
func getQuery(shopType string) (string, maps.PlaceType) {
	switch shopType {
	case "used":
		return "男性古着屋", maps.PlaceTypeClothingStore
	case "select":
		return "男性セレクトショップ", maps.PlaceTypeClothingStore
	case "other":
		return "(男性衣料品ブランド) || (Clothing Brand && Men)", maps.PlaceTypeClothingStore
	case "cafe":
		return "カフェ || Cafe", maps.PlaceTypeCafe
	}

	return "", ""
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

	return detailResult
}

// GetPlacePhotos 写真参照コードを受け取り，写真を最大3枚取得し，返す
func getPlacePhotos(photos []maps.Photo) []string {
	var photoResponses []string
	for i := 0; i < 3; i++ {
		photoResponses = append(photoResponses, noImage)
	}

	for index, photo := range photos {
		if index > 2 {
			break
		}

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

// getNextShops 次の20件の検索結果一覧を返す
func getNextShops(nextPageToken string) ([][]maps.PlacesSearchResult, string) {
	request := &maps.NearbySearchRequest{
		PageToken: nextPageToken,
	}

	return searchShops(request)
}

// searchShops リクエスト内容を受け取り，NearbySearchRequestを行い，検索結果一覧と次の20件の検索結果一覧にアクセスするトークンを返す
func searchShops(request *maps.NearbySearchRequest) ([][]maps.PlacesSearchResult, string) {
	response, err := Client.NearbySearch(context.Background(), request)
	if err != nil {
		log.Fatalf("fatal error: %s", err)

	}

	var shops [][]maps.PlacesSearchResult = make([][]maps.PlacesSearchResult, 2)

	for index, shopResult := range response.Results {
		if index < 10 {
			shops[0] = append(shops[0], shopResult)
		} else {
			shops[1] = append(shops[1], shopResult)
		}
	}

	return shops, response.NextPageToken
}
