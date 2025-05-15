package models

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type Receipt struct {
	Retailer     string  `json:"retailer"`
	PurchaseDate string  `json:"purchaseDate"`
	PurchaseTime string  `json:"purchaseTime"`
	Items        []Item  `json:"items"`
	Total        string  `json:"total"`
}

type ReceiptResponse struct {
	ID string `json:"id"`
}

type PointsResponse struct {
	Points int64 `json:"points"`
}