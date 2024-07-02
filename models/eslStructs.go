package Data

type ESLTransaction struct {
	Id             string `json:"@id"`
	Finished       string `json:"@finished"`
	Failed         string `json:"@failed"`
	TotalNumber    string `json:"TotalUpdates"`
	ErrorNumber    string `json:"ErrorUpdates"`
	FinishedNumber string `json:"FinishedUpdates"`
}

type ESLUpdateStatusPage struct {
	Records        string `json:"@records"`
	TotalRecords   string `json:"@totalRecords"`
	TotalPages     string `json:"@totalPages"`
	Page           string `json:"@page"`
	RecordsPerPage string `json:"@recordsPerPage"`
	UpdateStatus   ESLUpdateStatus
}

type ESLUpdateStatus struct {
	UpdateId      string `json:"@id"`
	Finished      string `json:"@finished"`
	LabelId       string
	TaskType      string
	Page          string
	TaskPriority  string
	TransactionId string
	Status        string
	AccessPointId string
	PowerStatus   string
}

type ESLLabelInfo struct {
	LabelId          string
	PowerStatus      string
	ConnectionStatus string
	TaskType         string
	Status           string
	SyncQuality      string
	Rssi             string
	CurrentPage      string
	Type             string
}

type ESLabelsPagedResult struct {
	Records        string `json:"@records"`
	TotalRecords   string `json:"@totalRecords"`
	TotalPages     string `json:"@totalPages"`
	Page           string `json:"@page"`
	RecordsPerPage string `json:"@recordsPerPage"`
	LabelInfo      []ESLLabelInfo
}

type ESLArticles []ESLArticle

type ESLArticle struct {
	ArticleNumber string            `json:"@articleNumber"`
	Gtin          []string          `json:"Gtin"`
	Field         []ESLArticleField `json:"Field"`
}

type CustomESLArticlesList struct {
	Article1 CustomESLArticle `json:"article1"`
	Article2 CustomESLArticle `json:"article2"`
}

type CustomESLArticle struct {
	Plu          string   `json:"articleNumber" xml:"articleNumber"`
	EAN          []string `json:"gtin" xml:"gtin"`
	ItemName     string   `json:"name" xml:"name"`
	Description  string   `json:"description" xml:"description"`
	Manufacturer string   `json:"manufacturer" xml:"manufacturer"`
	Stpr         string   `json:"stpr" xml:"stpr"`
	Unit         string   `json:"unit" xml:"unit"`
	Price        string   `json:"price" xml:"price"`
	PromoPrice   string   `json:"promoprice" xml:"promoprice"`
	OldPrice     string   `json:"oldprice" xml:"oldprice"`
	PromoQty     string   `json:"promoqty" xml:"promoqty"`
	PromoType    string   `json:"promotype" xml:"promotype"`
	Ctm          string   `json:"ctm" xml:"ctm"`
	TopType      string   `json:"toptype" xml:"toptype"`
	Ingredients  string   `json:"ingredients" xml:"ingredients"`
	Promo        string   `json:"promo" xml:"promo"`
	ScaleNumber  string   `json:"scalenumber" xml:"scalenumber"`
	ItemPlace    string   `json:"itemplace" xml:"itemplace"`
	StoreName    string   `json:"storename" xml:"storename"`
	InAssortment string   `json:"inassortment" xml:"inassortment"`
	ProdDate     string   `json:"proddate" xml:"proddate"`
	Barcode      string   `json:"barcode" xml:"barcode"`
	Style        string   `json:"style" xml:"style"`
	Location     string   `json:"location" xml:"location"`
	Alcohol      string   `json:"alcohol" xml:"alcohol"`
}

type ESLArticleField struct {
	Key   string `json:"@key"`
	Value string `json:"$"`
}

type ESLTemplateTask struct {
	Page             string `json:"@page"`
	Preload          string `json:"@preload"`
	SkipOnEqualImage string `json:"@skipOnEqualImage"`
	Template         string `json:"@template"`
	LabelId          string `json:"@labelId"`
	TaskPriority     string `json:"@taskPriority"`
	//ExternalId       string `json:"@externalId"`
	Article map[string]CustomESLArticle `json:"articles"`
}

type ESLSwitchPageTask struct {
	Page    string `json:"@page"`
	LabelId string `json:"@labelId"`
}

type ESLTaskList struct {
	TemplateTask   ESLTemplateTask
	SwitchPageTask ESLSwitchPageTask
}

type ESLLabelsList struct {
	Label []ESLLabel
}

type ESLLabel struct {
	Id string `json:"@id"`
}

type EslLabelsPagedResult struct {
	Records        string `json:"@records"`
	TotalRecords   string `json:"@totalRecords"`
	TotalPages     string `json:"@totalPages"`
	Page           string `json:"@page"`
	RecordsPerPage string `json:"@recordsPerPage"`
	LabelInfo      []ESLLabelInfo
}

type EslApPagedResult struct {
	Records        string `json:"@records"`
	TotalRecords   string `json:"@totalRecords"`
	TotalPages     string `json:"@totalPages"`
	Page           string `json:"@page"`
	RecordsPerPage string `json:"@recordsPerPage"`
	AccessPoint    []EslApInfo
}

type EslSingleApPagedResult struct {
	Records        string `json:"@records"`
	TotalRecords   string `json:"@totalRecords"`
	TotalPages     string `json:"@totalPages"`
	Page           string `json:"@page"`
	RecordsPerPage string `json:"@recordsPerPage"`
	AccessPoint    EslApInfo
}

type EslApInfo struct {
	Name             string `json:"Name"`
	AccessPointId    string `json:"AccessPointId"`
	Address          string `json:"Address"`
	Channel          string `json:"Channel"`
	ConnectionStatus string `json:"ConnectionStatus"`
	Version          string `json:"Version"`
	UpdateTime       string `json:"UpdateTime"`
	AutoConfig       string `json:"AutoConfig"`
	Protocol         string `json:"Protocol"`
}

type Property struct {
	Key   string `json:"@key"`
	Value string `json:"@value"`
}

type EslStatusData struct {
	ServiceName string     `json:"@serviceName"`
	StoreId     string     `json:"@storeId"`
	Properties  []Property `json:"Property"`
}
