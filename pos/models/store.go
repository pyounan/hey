package models

// CondimentGroup swagger:model CondimentGroup
// defines attributes of CondimentGroup entity combinaation
type CondimentGroups struct {
	ID   int    `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
}

// Store swagger:model store
// defines attributes of a Store entity
type Store struct {
	ID             int                    `json:"id" bson:"id"`
	Code           string                 `json:"code" bson:"code"`
	Description    string                 `json:"description" bson:"description"`
	InvoiceFooter  string                 `json:"invoice_footer" bson:"invoice_footer"`
	InvoiceHeader  string                 `json:"invoice_header" bson:"invoice_header"`
	KitchenSubmit  string                 `json:"kitchen_submit" bson:"kitchen_submit"`
	Logo           string                 `json:"logo" bson:"logo"`
	NumberOfTables int                    `json:"number_of_tables" bson:"number_of_tables"`
	ShowPaymaster  bool                   `json:"show_paymaster" bson:"show_paymaster"`
	AllowedRooms   string                 `json:"allowed_rooms" bson:"allowed_rooms"`
	LayoutJson     map[string]interface{} `json:"layout_json" bson:"layout_json"`
}

// StoreDetails swagger:model storeDetails
// defines attributes of a StoreDetails entity
// basically it's a Store with it's menus attahced to it
type StoreDetails struct {
	Store `bson:",inline"`
	Menus []Menu `json:"menus" bson:"menus"`
}

// Menu swagger:model menu
// defines attributes of Menu entity
type Menu struct {
	ID       int     `json:"id" bson:"id"`
	Name     string  `json:"name" bson:"name"`
	FromTime *string `json:"from_time" bson:"from_time"`
	ToTime   *string `json:"to_time" bson:"to_time"`
	Groups   []Group `json:"groups" bson:"groups"`
}

// Group swagger:model group
// defines attributes of Group entity
type Group struct {
	ID              int               `json:"id" bson:"id"`
	Name            string            `json:"name" bson:"name"`
	Items           []Item            `json:"items" bson:"items"`
	CondimentGroups []CondimentGroups `json:"condimentgroups" bson:"condimentgroups"`
}
