package vkAPI

type Wall struct {
    Count    int `json:"count"`
    Items WallItems `json:"items"`
}

type WallResponse struct {
    Response Wall `json:"response"`
}

type WallItems []WallItem

type Attachments []Attachment

type WallItem struct {
    Id            int `json:"id"`
    OwnerId       int `json:"owner_id"`
    FromId        int `json:"from_id"`
    Date          int `json:"date"`
    Text          string `json:"text"`
    ReplyOwnerId  int `json:"reply_owner_id"`
    ReplyPostId   int `json:"reply_post_id"`
    FriendsOnly   int `json:"friends_only"`
    CommentsObj   Comments `json:"comments"`
    LikesObj      Likes `json:"likes"`
    RepostsObj    Reposts `json:"reposts"`
    ViewsObj      Views `json:"views"`
    PostType      string `json:"post_type"`
    PostSourceObj PostSource `json:"post_source"`
    AttachmentsArr   Attachments `json:"attachments"`
    GeoObj        Geo `json:"geo"`
    SignerId      int `json:"signer_id"`
    CanPin        int `json:"can_pin"`
    CanDelete     int `json:"can_delete"`
    CanEdit       int `json:"can_edit"`
    IsPinned      int `json:"is_pinned"`
    MarkedAsAds   int `json:"marked_as_ads"`
}

type Comments struct {
    Count   int `json:"count"`
    CanPost int `json:"can_post"`
}
type Likes struct {
    Count      int `json:"count"`
    UserLikes  int `json:"user_likes"`
    CanLike    int `json:"can_like"`
    CanPublish int `json:"can_publish"`
}
type Reposts struct {
    Count        int `json:"count"`
    UserReposted int `json:"user_reposted"`
}
type Views struct {
    Count int `json:"count"`
}
type PostSource struct {
    Type     string `json:"type"`
    Platform string `json:"platform"`
    Data     string `json:"data"`
    Url      string `json:"url"`
}
type Attachment struct {
    Type     string `json:"type"`
    PhotoObj Photo `json:"photo"`
}

type Photo struct {
    Id        int `json:"id"`
    AlbumId   int `json:"album_id"`
    OwnerId   int `json:"owner_id"`
    Photo75   string `json:"photo_75"`
    Photo130  string `json:"photo_130"`
    Photo604  string `json:"photo_604"`
    Photo807  string `json:"photo_807"`
    Photo1200 string `json:"photo_1200"`
    Width     int `json:"width"`
    Height    int `json:"height"`
    Text      string `json:"text"`
    Date      int `json:"date"`
    AccessKey string `json:"access_key"`
}

type Geo struct {
    Type        string `json:"type"`
    Coordinates string `json:"coordinates"`
    PlaceObj    Place `json:"place"`
}

type Place struct {
    Id        int `json:"id"`
    Title     string `json:"title"`
    Latitude  int `json:"latitude"`
    Longitude int `json:"longitude"`
    Created   int `json:"created"`
    Icon      string `json:"icon"`
    Country   string `json:"country"`
    City      string `json:"city"`
}
