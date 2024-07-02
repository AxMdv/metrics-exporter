package Data

type BOUpdateMessage []BOArticle

type BOArticle struct {
	Plu            string       `json:"plu"`
	HasUpdErrors   bool         `json:"hasUpdErrors,omitempty"`
	EAN            []string     `json:"ean"`
	ItemName       string       `json:"itemName"`
	Description    string       `json:"description"`
	Manufacturer   string       `json:"manufacturer"`
	Stpr           string       `json:"stpr"`
	Unit           string       `json:"unit"`
	Price          []float32    `json:"price"`
	OldPrice       float32      `json:"oldPrice"`
	PromoQty       []int        `json:"promoQty"`
	PromoType      string       `json:"promoType"`
	Ctm            string       `json:"ctm"`
	TopType        string       `json:"topType"`
	Ingredients    string       `json:"ingredients,omitempty"`
	Promo          bool         `json:"promo,omitempty"`
	ScaleNumber    int          `json:"scaleNumber,omitempty"`
	ItemPlace      string       `json:"itemPlace,omitempty"`
	StoreName      string       `json:"storeName,omitempty"`
	InAssortment   bool         `json:"inAssortment,omitempty"`
	ProdDate       string       `json:"prodDate,omitempty"`
	Barcode        string       `json:"barcode,omitempty"`
	Style          string       `json:"style,omitempty"`
	Location       string       `json:"location,omitempty"`
	Alcohol        string       `json:"alcoholDescription,omitempty"`
	BOPluOrderList []BOPluOrder `json:"eslIdPluOrder"`
}

type BOPluOrder struct {
	EslId    string `json:"eslID"`
	PluOrder int    `json:"pluOrder"`
}

type PluInfo struct {
	Plu      string `json:"plu"`
	ItemName string `json:"itemName"`
}

type BOBind struct {
	EslId string   `json:"eslID"`
	Plu   []string `json:"goodsID"`
}

type BOUpdateStatus struct {
	Plu       string   `json:"plu"`
	UpdStatus string   `json:"updStatus"`
	EslId     []string `json:"eslID,omitempty"`
}

type BOEslStatus struct {
	EslId         string `json:"eslID"`
	EslStatus     string `json:"eslStatus"`
	BatteryStatus string `json:"batteryStatus"`
	ScreenSize    string `json:"screenSize"`
}

type BOErrorDescription struct {
	EslId     string   `json:"eslID"`
	Plu       []string `json:"plu"`
	ErrorDesc string   `json:"errorDesc"`
}

var LabelTypesRelations = map[string]string{
	"1.6 BWR": "152x152",
	"1.6 BWY": "152x152",
	"2.6 BWR": "296x152",
	"2.6 BWY": "296x152",
	"2.6 BW":  "296x152",
	"4.2 BWR": "400x300",
	"4.2 BWY": "400x300",
	"4.2 BW":  "400x300",
	"7.4 BWR": "480x800",
	"7.4 BWY": "480x800",
	"7.4 BW":  "480x800",
	"4.5 BWR": "480x176",
	"4.5 BWY": "480x176",
}
