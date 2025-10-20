package forms

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"

	time2 "quickflow/config/time"
	"quickflow/shared/models"
)

//easyjson:json
type File struct {
	Name string
	Data []byte
}

//easyjson:json
type PostForm struct {
	Text        string    `json:"text,omitempty"`
	Media       []string  `form:"media" json:"media,omitempty"`
	Audio       []string  `form:"audio" json:"audio,omitempty"`
	File        []string  `form:"files" json:"files,omitempty"`
	Stickers    []string  `form:"stickers" json:"stickers,omitempty"`
	IsRepost    bool      `json:"is_repost,omitempty"`
	CreatorId   uuid.UUID `json:"author_id,omitempty"`
	CreatorType string    `json:"author_type,omitempty"`
}

func ParseCreatorType(creatorType string) (models.PostCreatorType, error) {
	switch creatorType {
	case string(models.PostUser):
		return models.PostUser, nil
	case string(models.PostCommunity):
		return models.PostCommunity, nil
	default:
		return "", errors.New("invalid creator type")
	}
}

func (p *PostForm) ToPostModel(userId uuid.UUID) (models.Post, error) {
	var attachments []*models.File
	for _, file := range p.Media {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeMedia,
		})
	}

	for _, file := range p.Audio {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeAudio,
		})
	}

	for _, file := range p.File {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeFile,
		})
	}

	for _, sticker := range p.Stickers {
		attachments = append(attachments, &models.File{
			URL:         sticker,
			DisplayType: models.DisplayTypeSticker,
		})
	}
	var postModel models.Post

	if len(p.CreatorType) == 0 {
		postModel.CreatorType = models.PostUser
		postModel.CreatorId = userId
	} else {
		var err error
		postModel.CreatorType, err = ParseCreatorType(p.CreatorType)
		if err != nil {
			return models.Post{}, errors.New("invalid creator type")
		}

		postModel.CreatorId = userId
	}

	postModel.Desc = p.Text
	postModel.CreatedAt = time.Now()
	postModel.UpdatedAt = time.Now()
	postModel.Files = attachments
	postModel.IsRepost = p.IsRepost

	return postModel, nil
}

//easyjson:json
type FeedForm struct {
	Posts int    `json:"count"`
	Ts    string `json:"ts"`
}

// GetParams gets parameters from the map
func (f *FeedForm) GetParams(values url.Values) error {
	var (
		err      error
		numPosts int64
	)

	if !values.Has("count") {
		return errors.New("count parameter missing")
	}

	numPosts, err = strconv.ParseInt(values.Get("count"), 10, 64)
	if err != nil {
		return errors.New("failed to parse count")
	}

	f.Posts = int(numPosts)
	f.Ts = values.Get("ts")
	return nil
}

//easyjson:json
type PublicUserInfoOut struct {
	ID        string              `json:"id"`
	Username  string              `json:"username"`
	AvatarURL string              `json:"avatar_url,omitempty"`
	FirstName string              `json:"firstname"`
	LastName  string              `json:"lastname"`
	IsOnline  *bool               `json:"online,omitempty"`
	Relation  models.UserRelation `json:"relation,omitempty"`
}

func PublicUserInfoToOut(info models.PublicUserInfo, relation models.UserRelation) PublicUserInfoOut {
	return PublicUserInfoOut{
		ID:        info.Id.String(),
		Username:  info.Username,
		FirstName: info.Firstname,
		LastName:  info.Lastname,
		AvatarURL: info.AvatarURL,
		Relation:  relation,
	}
}

//easyjson:json
type PostOut struct {
	Id           string      `json:"id"`
	Creator      interface{} `json:"author,omitempty"`
	CreatorType  string      `json:"author_type"`
	Desc         string      `json:"text"`
	MediaURLs    []FileOut   `json:"media,omitempty"`
	AudioURLs    []FileOut   `json:"audio,omitempty"`
	FileURLs     []FileOut   `json:"files,omitempty"`
	StickerURLs  []FileOut   `json:"stickers,omitempty"`
	CreatedAt    string      `json:"created_at"`
	UpdatedAt    string      `json:"updated_at"`
	LikeCount    int         `json:"like_count"`
	RepostCount  int         `json:"repost_count"`
	CommentCount int         `json:"comment_count"`
	IsRepost     bool        `json:"is_repost"`
	IsLiked      bool        `json:"is_liked"`
	LastComment  *CommentOut `json:"last_comment,omitempty"`
}

//easyjson:json
type PostsOut []PostOut

func (p *PostOut) FromPost(post models.Post) {
	mediaURLs := make([]FileOut, 0)
	audioURLs := make([]FileOut, 0)
	fileURLs := make([]FileOut, 0)
	stickerUrls := make([]FileOut, 0)

	for _, file := range post.Files {
		if file.DisplayType == models.DisplayTypeMedia {
			mediaURLs = append(mediaURLs, ToFileOut(*file))
		} else if file.DisplayType == models.DisplayTypeAudio {
			audioURLs = append(audioURLs, ToFileOut(*file))
		} else if file.DisplayType == models.DisplayTypeSticker {
			stickerUrls = append(stickerUrls, ToFileOut(*file))
		} else {
			fileURLs = append(fileURLs, ToFileOut(*file))
		}
	}

	p.Id = post.Id.String()
	p.Desc = post.Desc
	p.MediaURLs = mediaURLs
	p.AudioURLs = audioURLs
	p.FileURLs = fileURLs
	p.StickerURLs = stickerUrls
	p.CreatedAt = post.CreatedAt.Format(time2.TimeStampLayout)
	p.UpdatedAt = post.UpdatedAt.Format(time2.TimeStampLayout)
	p.Creator = &PublicUserInfoOut{
		ID: post.CreatorId.String(),
	}
	p.CreatorType = string(post.CreatorType)
	p.LikeCount = post.LikeCount
	p.RepostCount = post.RepostCount
	p.CommentCount = post.CommentCount
	p.IsRepost = post.IsRepost
	p.IsLiked = post.IsLiked
}

//easyjson:json
type UpdatePostForm struct {
	Id    string   `json:"-"`
	Text  string   `json:"text"`
	Media []string `form:"media" json:"media,omitempty"`
	Audio []string `form:"audio" json:"audio,omitempty"`
	File  []string `form:"files" json:"files,omitempty"`
}

func (p *UpdatePostForm) ToPostUpdateModel(postId uuid.UUID) (models.PostUpdate, error) {
	var attachments []*models.File
	for _, file := range p.Media {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeMedia,
		})
	}

	for _, file := range p.Audio {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeAudio,
		})
	}

	for _, file := range p.File {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeFile,
		})
	}

	return models.PostUpdate{
		Id:    postId,
		Desc:  p.Text,
		Files: attachments,
	}, nil
}

type LikeOut struct {
	TargetType string            `json:"target_type"`
	TargetId   uuid.UUID         `json:"target_id"`
	User       PublicUserInfoOut `json:"user"`
	CreatedAt  string            `json:"created_at"`
}

func LikeFromModel(likeOut models.Like, user PublicUserInfoOut) LikeOut {
	return LikeOut{
		TargetType: string(likeOut.TargetType),
		TargetId:   likeOut.UserId,
		User:       user,
		CreatedAt:  likeOut.CreatedAt.Format(time2.TimeStampLayout),
	}
}
