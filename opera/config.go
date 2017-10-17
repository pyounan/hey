package opera

type OperaConfigValue struct {
	Code        string `json:"code" bson:"code"`
	Departments []int  `json:"departments" bson:"departments"`
}

type OperaConfig struct {
	ConfigName string           `json:"config_name" bson:"config_name"`
	Value      OperaConfigValue `json:"value" bson:"value"`
}

func FlattenToMap(operaConfigs []OperaConfig) map[int]string {
	flattenedMap := make(map[int]string)
	for _, obj := range operaConfigs {
		for _, department := range obj.Value.Departments {
			flattenedMap[department] = obj.Value.Code
		}
	}
	return flattenedMap
}
