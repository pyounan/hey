package syncer

var ConfApis map[string]string = make(map[string]string)
var SingleLoadApis map[string]string = make(map[string]string)

func init() {
	ConfApis["stores"] = "api/pos/store/?outlet=true"
	ConfApis["fixeddiscounts"] = "api/pos/fixeddiscount/"
	ConfApis["storedetails"] = "api/pos/storedetails/"
	ConfApis["terminals"] = "api/pos/terminal/"
	ConfApis["condiments"] = "api/pos/condiment/"
	ConfApis["courses"] = "api/pos/course/"
	ConfApis["printers"] = "api/pos/printer/"
	ConfApis["printersettings"] = "api/pos/printersettings/"

	ConfApis["company"] = "shadowinn/api/company/"
	ConfApis["audit_date"] = "shadowinn/api/auditdate/"

	ConfApis["departments"] = "income/api/department/"
	ConfApis["currencies"] = "income/api/currency/"
	ConfApis["permissions"] = "income/api/poscashierpermissions/"
	ConfApis["cashiers"] = "income/api/cashier/sync/"
	ConfApis["attendance"] = "income/api/attendance/"
	ConfApis["usergroups"] = "core/getallusergroups/"
	ConfApis["operasettings"] = "api/pos/opera/"

	ConfApis["sunexportdate"] = "api/inventory/sunexportdate/"

	ConfApis["tables"] = "api/pos/tables/"

	ConfApis["ccv_settings"] = "payment/ccv/settings/"
	ConfApis["ccv_terminal_integration_settings"] = "api/pos/settings/ccv/terminals/"

	ConfApis["storemenuitemconfig"] = "api/pos/storemenuitemconfig/"

	SingleLoadApis["posinvoices"] = "api/pos/posinvoices/?is_settled=false"
}
