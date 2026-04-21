package api

type User struct {
	Name string `json:"name"`
}

type UserSummary struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
