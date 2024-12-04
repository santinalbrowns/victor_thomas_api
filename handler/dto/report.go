package dto

type Report struct {
	Sales      int64 `json:"sales"`
	Purchases  int64 `json:"purchases"`
	Products   int64 `json:"products"`
	Categories int64 `json:"categories"`
}
