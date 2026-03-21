package domain

type Address struct {
	Street      string `json:"street"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	Country     string `json:"country"`
	Coordinates string `json:"coordinates,omitempty"` // latitude,longitude
}
