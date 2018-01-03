package syncer

var ConfApis map[string]string = make(map[string]string)
var SingleLoadApis map[string]string = make(map[string]string)

func init() {
	ConfApis["stores"] = "api/pos/store/"
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

	SingleLoadApis["tables"] = "api/pos/tables/"
	SingleLoadApis["posinvoices"] = "api/pos/posinvoices/?is_settled=false"
}
