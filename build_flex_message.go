package main

import (
	"googlemaps.github.io/maps"
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
	Contents   []ContentsContainer `json:"contents"`
}

// Button Flex Messageの要素
type Button struct {
	Type    componentType `json:"type"`
	Action  string        `json:"action,omitempty"`
	Flex    string        `json:"flex,omitempty"`
	Height  string        `json:"height,omitempty"`
	Style   string        `json:"style,omitempty"`
	Gravity string        `json:"gravity,omitempty"`
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
	Flex   string        `json:"flex,omitempty"`
	Margin componentSize `json:"margin,omitempty"`
	Size   componentSize `json:"size,omitempty"`
	Wrap   string        `json:"wrap,omitempty"`
	Weight string        `json:"weight,omitempty"`
	Color  string        `json:"color,omitempty"`
}

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

type componentLayout string

const (
	layoutVertical   componentLayout = "vertical"
	layoutHorizontal componentLayout = "horizontal"
)

type componentSize string

const (
	sizeXs componentSize = "xs"
	sizeSm componentSize = "sm"
	sizeMd componentSize = "md"
	sizeLg componentSize = "lg"
	sizeXl componentSize = "xl"
)

func getFlexMessage(searchResults []maps.PlaceDetailsResult) (Flex, Flex) {
	var bubblesFormer []*Bubble
	var bubblesLatter []*Bubble

	for index, searchResult := range searchResults {
		switch {
		case index < 10:
			bubblesFormer = append(bubblesFormer, buildResultBubble(searchResult))
		default:
			bubblesLatter = append(bubblesLatter, buildResultBubble(searchResult))
		}
	}
	// bubbles = append(bubbles, buildResultBubble(searchResults[0]))

	// message, err := json.Marshal(

	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }

	return Flex{
			Type:    typeFlex,
			AltText: "検索結果",
			Contents: Carousel{
				Type:     "carousel",
				Contents: bubblesFormer,
			},
		}, Flex{
			Type:    typeFlex,
			AltText: "検索結果",
			Contents: Carousel{
				Type:     "carousel",
				Contents: bubblesLatter,
			},
		}
}

func buildResultBubble(searchResult maps.PlaceDetailsResult) *Bubble {
	return &Bubble{
		Type:   typeBubble,
		Header: buildResultBubbleHeder(searchResult),
	}
}

func buildResultBubbleHeder(searchResult maps.PlaceDetailsResult) *Box {
	photoURLs := getPlacePhotos(searchResult.Photos)
	images := buildImageComponents(photoURLs)

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

func buildResultBubbleBody(searchResult maps.PlaceDetailsResult) *Box {
	return &Box{}
}

func buildResultBubbleFooter(searchResult maps.PlaceDetailsResult) *Box {
	return &Box{}
}

// func buildActionBubble() *linebot.BubbleContainer {

// }
