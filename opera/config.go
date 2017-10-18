package opera

type RevenuePaymentServiceConfigValue struct {
	Code        string `json:"code" bson:"code"`
	Departments []int  `json:"departments" bson:"departments"`
}

type RevenuePaymentServiceConfig struct {
	ConfigName string                           `json:"config_name" bson:"config_name"`
	Value      RevenuePaymentServiceConfigValue `json:"value,omitempty" bson:"value,omitempty"`
}

type RoomDepartmentConfigValue struct {
	DepartmentID int `json:"department_id" bson:"department_id"`
}

type RoomDepartmentConfig struct {
	Value RoomDepartmentConfigValue `json:"value" bson:"value"`
}

func FlattenToMap(operaConfigs []RevenuePaymentServiceConfig) map[int]string {
	flattenedMap := make(map[int]string)
	for _, obj := range operaConfigs {
		for _, department := range obj.Value.Departments {
			flattenedMap[department] = obj.Value.Code
		}
	}
	return flattenedMap
}
