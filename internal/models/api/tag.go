package api

type Tag struct {
	Name string `json:"tag"`
}

type TagResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"tag"`
}
