package main

import (
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strings"
	"time"

	"googlemaps.github.io/maps"
)

const (
	starGold     = "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gold_star_28.png"
	starGray     = "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gray_star_28.png"
	searchGoogle = "http://www.google.co.jp/search?hl=ja&lr=lang_ja&q="
)

// ContentsContainer インタフェース
type ContentsContainer interface {
	ContentsContainer()
}

// ContentsContainer 構造体 Box で Contents Container を実装
func (*Box) ContentsContainer() {}

// ContentsContainer 構造体 Button で Contents Container を実装
func (*Button) ContentsContainer() {}

// ContentsContainer 構造体 Icon で Contents Container を実装
func (*Icon) ContentsContainer() {}

// ContentsContainer 構造体 Image で Contents Container を実装
func (*Image) ContentsContainer() {}

// ContentsContainer 構造体 Spacer で Contents Container を実装
func (*Spacer) ContentsContainer() {}

// ContentsContainer 構造体 Text で Contents Container を実装
func (*Text) ContentsContainer() {}

// Action アクションメッセージ
type Action interface {
	Action()
}

// Action 構造体 URIAction で Action を実装
func (*URIAction) Action() {}

// Action 構造体 PostbackAction で Action を実装
func (*PostbackAction) Action() {}

// FlexMessage Flex Message
type FlexMessage struct {
	ReplyToken string `json:"replyToken"`
	Messages   []Flex `json:"messages"`
}

// Flex Flex Messageの要素
type Flex struct {
	Type     componentType `json:"type"`
	AltText  string        `json:"altText"`
	Contents Carousel      `json:"contents"`
}

// Carousel Flex Messageの要素
type Carousel struct {
	Type     componentType `json:"type"`
	Contents []*Bubble     `json:"contents"`
}

// Bubble Flex Messageの要素
type Bubble struct {
	Type   componentType `json:"type"`
	Header *Box          `json:"header,omitempty"`
	Body   *Box          `json:"body,omitempty"`
	Footer *Box          `json:"footer,omitempty"`
}

// Box Flex Messageの要素
type Box struct {
	Type       componentType       `json:"type"`
	Layout     componentLayout     `json:"layout"`
	PaddingAll string              `json:"paddingAll,omitempty"`
	Margin     componentSize       `json:"margin,omitempty"`
	Spacing    componentSize       `json:"spacing,omitempty"`
	Contents   []ContentsContainer `json:"contents"`
}

// Button Flex Messageの要素
type Button struct {
	Type    componentType `json:"type"`
	Action  Action        `json:"action,omitempty"`
	Flex    int           `json:"flex,omitempty"`
	Height  componentSize `json:"height,omitempty"`
	Style   string        `json:"style,omitempty"`
	Gravity string        `json:"gravity,omitempty"`
}

// URIAction Flex Messageの要素
type URIAction struct {
	Type   string           `json:"type"`
	Label  string           `json:"label,omitempty"`
	URI    string           `json:"uri"`
	AltURI *URIActionAltURI `json:"altUri,omitempty"`
}

// PostbackAction Flex Messageの要素
type PostbackAction struct {
	Type        string `json:"type"`
	Label       string `json:"label,omitempty"`
	Data        string `json:"data"`
	DisplayText string `json:"displayText,omitempty"`
}

// URIActionAltURI デスクトップ用
type URIActionAltURI struct {
	Desktop string `json:"desktop"`
}

// Icon Flex Messageの要素
type Icon struct {
	Type componentType `json:"type"`
	URL  string        `json:"url"`
	Size componentSize `json:"size,omitempty"`
}

// Image Flex Messageの要素
type Image struct {
	Type        componentType `json:"type"`
	URL         string        `json:"url"`
	Gravity     string        `json:"gravity,omitempty"`
	Size        componentSize `json:"size,omitempty"`
	AspectRatio string        `json:"aspectRatio,omitempty"`
	AspectMode  string        `json:"aspectMode,omitempty"`
}

// Spacer Flex Messageの要素
type Spacer struct {
	Type componentType `json:"type"`
	Size componentSize `json:"size,omitempty"`
}

// Text Flex Messageの要素
type Text struct {
	Type   componentType `json:"type"`
	Text   string        `json:"text,omitempty"`
	Flex   int           `json:"flex,omitempty"`
	Margin componentSize `json:"margin,omitempty"`
	Size   componentSize `json:"size,omitempty"`
	Wrap   bool          `json:"wrap,omitempty"`
	Weight string        `json:"weight,omitempty"`
	Color  string        `json:"color,omitempty"`
}

// componentType コンポーネントのタイプ
type componentType string

const (
	typeFlex     componentType = "flex"
	typeCarousel componentType = "carousel"
	typeBubble   componentType = "bubble"
	typeBox      componentType = "box"
	typeButton   componentType = "button"
	typeIcon     componentType = "icon"
	typeImage    componentType = "image"
	typeSpacer   componentType = "spacer"
	typeText     componentType = "text"
)

// componentLayout コンポーネントのレイアウト
type componentLayout string

const (
	layoutVertical   componentLayout = "vertical"
	layoutHorizontal componentLayout = "horizontal"
	layoutBaseline   componentLayout = "baseline"
)

// componentSize コンポーネントのサイズ
type componentSize string

const (
	sizeXs componentSize = "xs"
	sizeSm componentSize = "sm"
	sizeMd componentSize = "md"
	sizeLg componentSize = "lg"
	sizeXl componentSize = "xl"
)

// getFlexMessage Flex Message を構築し，返す
func getFlexMessage(bubbles []*Bubble, replyToken string) FlexMessage {
	return FlexMessage{
		ReplyToken: replyToken,
		Messages: []Flex{
			buildFlexComponent(bubbles),
		},
	}
}

// buildFlexComponent Flex Component を構築
func buildFlexComponent(bubbles []*Bubble) Flex {
	return Flex{
		Type:    typeFlex,
		AltText: "検索結果",
		Contents: Carousel{
			Type:     "carousel",
			Contents: bubbles,
		},
	}
}

// buildResultBubble バブルを構築し，返す
func getBubble(shopDetail maps.PlaceDetailsResult, photo []string) *Bubble {
	return &Bubble{
		Type:   typeBubble,
		Header: buildResultBubbleHeder(photo),
		Body:   buildResultBubbleBody(shopDetail),
		Footer: buildResultBubbleFooter(shopDetail),
	}
}

// buildResultBubbleHeader バブルのヘッダーを構築
func buildResultBubbleHeder(photo []string) *Box {
	images := buildImageComponents(photo)

	return &Box{
		Type:       typeBox,
		Layout:     layoutVertical,
		PaddingAll: "0px",
		Contents: []ContentsContainer{
			&Box{
				Type:   typeBox,
				Layout: layoutHorizontal,
				Contents: []ContentsContainer{
					images[0],
					&Box{
						Type:   typeBox,
						Layout: layoutVertical,
						Contents: []ContentsContainer{
							images[1],
							images[2],
						},
					},
				},
			},
		},
	}
}

// buildImageComponents Imageを構築
func buildImageComponents(photoURLs []string) []*Image {
	var images []*Image

	for index, photoURL := range photoURLs {
		switch index {
		case 0:
			images = append(images, buildImageComponent(photoURL, "150:196"))
		default:
			images = append(images, buildImageComponent(photoURL, "150:98"))
		}
	}
	return images
}

// buildImageComponent 各Imageを構築
func buildImageComponent(photoURL string, aspectRatio string) *Image {
	return &Image{
		Type:        typeImage,
		URL:         photoURL,
		Gravity:     "center",
		Size:        "full",
		AspectRatio: aspectRatio,
		AspectMode:  "cover",
	}
}

// buildResultBubbleBody ボディを構築
func buildResultBubbleBody(shopDetail maps.PlaceDetailsResult) *Box {
	return &Box{
		Type:   typeBox,
		Layout: layoutVertical,
		Contents: []ContentsContainer{
			&Text{
				Type:   typeText,
				Text:   shopDetail.Name,
				Size:   sizeLg,
				Wrap:   true,
				Weight: "bold",
			},
			buildEvaluation(shopDetail.Rating, shopDetail.UserRatingsTotal),
			buildStoreInformation(shopDetail.Vicinity, shopDetail.OpeningHours),
		},
	}
}

// buildEvaluation 店の評価（星の数と評価件数）を構築
func buildEvaluation(rating float32, ratingCount int) *Box {
	icons := buildIconComponents(rating)
	evaluation := buildEvaluationText(rating, ratingCount)

	return &Box{
		Type:   typeBox,
		Layout: layoutBaseline,
		Margin: sizeMd,
		Contents: []ContentsContainer{
			icons[0],
			icons[1],
			icons[2],
			icons[3],
			icons[4],
			evaluation,
		},
	}
}

// buildIconComponents 星集合を構築
func buildIconComponents(rating float32) []*Icon {
	var icons []*Icon
	r := int(math.Round(float64(rating)))

	for i := 0; i < 5; i++ {
		switch {
		case r >= (i + 1):
			icons = append(icons, buildIconComponent(starGold))
		default:
			icons = append(icons, buildIconComponent(starGray))
		}
	}

	return icons
}

// buildIconComponent 星のコンポーネントを構築
func buildIconComponent(url string) *Icon {
	return &Icon{
		Type: typeIcon,
		URL:  url,
		Size: sizeSm,
	}
}

// buildEvaluationText 評価値と評価件数を構築
func buildEvaluationText(rating float32, ratingCount int) *Text {
	value := math.Round(float64(rating))

	return &Text{
		Type:   typeText,
		Text:   fmt.Sprint(value) + "(" + fmt.Sprint(ratingCount) + ")",
		Margin: sizeMd,
		Size:   sizeSm,
		Color:  "#999999",
	}
}

// buildStoreInformation 店の情報（住所，営業時間）を構築
func buildStoreInformation(address string, openingHours *maps.OpeningHours) *Box {
	return &Box{
		Type:    typeBox,
		Layout:  layoutVertical,
		Margin:  sizeMd,
		Spacing: sizeSm,
		Contents: []ContentsContainer{
			buildStoreAddress(address),
			buildStoreOpeningHours(openingHours),
		},
	}
}

// buildStoreAddress 店の住所を構築
func buildStoreAddress(address string) *Box {
	return &Box{
		Type:    typeBox,
		Layout:  layoutBaseline,
		Spacing: sizeSm,
		Contents: []ContentsContainer{
			&Text{
				Type:  typeText,
				Text:  "場所",
				Flex:  1,
				Size:  sizeSm,
				Color: "#aaaaaa",
			},
			&Text{
				Type:  typeText,
				Text:  address,
				Flex:  5,
				Wrap:  true,
				Size:  sizeSm,
				Color: "#666666",
			},
		},
	}
}

// buildStoreOpeningHours 店の営業時間を構築
func buildStoreOpeningHours(openingHours *maps.OpeningHours) *Box {
	businessHours, status, color := buildStoreOpeningHoursPeriod(openingHours)

	return &Box{
		Type:    typeBox,
		Layout:  layoutBaseline,
		Spacing: sizeSm,
		Contents: []ContentsContainer{
			buildStoreInformationText("営業時間", 1, "#aaaaaa"),
			buildStoreInformationText(businessHours+" "+status, 3, color),
		},
	}
}

// buildStoreOpeningHoursPeriod 営業時間によって営業ステータスと色を変化させて構築
func buildStoreOpeningHoursPeriod(openingHours *maps.OpeningHours) (string, string, string) {
	if reflect.ValueOf(openingHours).IsNil() {
		return "", "営業時間未記載", "#666666"
	}

	for _, period := range (*openingHours).Periods {
		if period.Open.Day == time.Now().Weekday() {
			open := period.Open.Time
			close := period.Close.Time
			businessHours := applyTimeFormat(open, close)

			if *openingHours.OpenNow {
				return businessHours, "(営業中)", "#32cd32"
			}
			return businessHours, "(準備中)", "#ff0000"
		}
	}

	return "", "定休日", "#ff0000"
}

// applyTimeFormat 時間の表記に変更
func applyTimeFormat(open string, close string) string {
	return open[:2] + ":" + open[2:] + " ~ " + close[:2] + ":" + close[2:]
}

// buildStoreInformationText 店の営業時間を表す Text Component を構築
func buildStoreInformationText(text string, flex int, color string) *Text {
	return &Text{
		Type:  typeText,
		Text:  text,
		Flex:  flex,
		Size:  sizeSm,
		Color: color,
	}
}

// buildResultBubbleFooter フッターを構築
func buildResultBubbleFooter(shopDetail maps.PlaceDetailsResult) *Box {
	buttonURI := buildURIActionButtonComponent(shopDetail.URL, "GoogleMapを開く")
	var buttonWebSite *Button
	webSite := shopDetail.Website

	if len(webSite) > 0 {
		buttonWebSite = buildURIActionButtonComponent(webSite, "お店のURLを開く")
	} else {
		webSite = searchGoogle + url.QueryEscape(shopDetail.Vicinity+" "+shopDetail.Name)
		buttonWebSite = buildURIActionButtonComponent(webSite, "お店をGoogleで検索する")
	}

	return &Box{
		Type:    typeBox,
		Layout:  layoutVertical,
		Spacing: sizeSm,
		Contents: []ContentsContainer{
			buttonURI,
			buttonWebSite,
			&Spacer{
				Type: typeSpacer,
				Size: sizeSm,
			},
		},
	}
}

func removeSpace(term string) string {
	words := strings.Split(term, " ")
	return strings.Join(words, "")
}

// buildURIActionButtonComponent URIActionメッセージを構築
func buildURIActionButtonComponent(uri string, label string) *Button {
	return &Button{
		Type:   typeButton,
		Height: sizeSm,
		Style:  "link",
		Action: &URIAction{
			Type:  "uri",
			Label: label,
			URI:   uri,
			AltURI: &URIActionAltURI{
				Desktop: uri,
			},
		},
	}
}

// getNextActionBubble ユーザが次に行う操作一覧を返す
func getNextActionBubble() *Bubble {
	return &Bubble{
		Type: typeBubble,
		Body: &Box{
			Type:   typeBox,
			Layout: layoutVertical,
			Contents: []ContentsContainer{
				buildPostbackActionButtonComponent("次の10件を検索", "next"),
			},
		},
	}
}

// buildPostbackActionButtonComponent postbackメッセージを構築
func buildPostbackActionButtonComponent(label string, data string) *Button {
	return &Button{
		Type:    typeButton,
		Height:  sizeMd,
		Style:   "link",
		Flex:    5,
		Gravity: "center",
		Action: &PostbackAction{
			Type:  "postback",
			Label: label,
			Data:  data,
		},
	}
}
