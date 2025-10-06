package forms

//easyjson:json
type MessageAttachmentForm struct {
	MediaURLs   []string `json:"media"`
	AudioURLs   []string `json:"audio"`
	FileURLs    []string `json:"files"`
	StickerURLs []string `json:"stickers,omitempty"`
}
