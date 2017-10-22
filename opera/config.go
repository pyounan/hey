package opera

import (
	"gopkg.in/mgo.v2/bson"
	"pos-proxy/db"
	"strconv"
)

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

func GetPaymentMethod(department int) (int, error) {
	paymentConfig := []RevenuePaymentServiceConfig{}
	_ = db.DB.C("operasettings").Find(bson.M{"config_name": "payment_method"}).All(&paymentConfig)
	paymentMethod := ""
	for _, p := range paymentConfig {
		for _, dept := range p.Value.Departments {
			if dept == department {
				paymentMethod = p.Value.Code
				break
			}
		}
	}
	return strconv.Atoi(paymentMethod)
}

func GetRoomDepartmentID() (int, error) {
	var roomDepartment RoomDepartmentConfig
	err := db.DB.C("operasettings").Find(bson.M{"config_name": "room_department"}).One(&roomDepartment)
	deptID := -1
	if err != nil {
		return deptID, err
	}
	return roomDepartment.Value.DepartmentID, err
}
