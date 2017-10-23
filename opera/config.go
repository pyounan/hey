package opera

import (
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/syncer"
	"strconv"
)

type OperaConfigValue struct {
	Code        string `json:"code" bson:"code"`
	Departments []int  `json:"departments" bson:"departments"`
}

type OperaConfig struct {
	ConfigName string           `json:"config_name" bson:"config_name"`
	Value      OperaConfigValue `json:"value,omitempty" bson:"value,omitempty"`
}

type RoomDepartmentConfigValue struct {
	DepartmentID int `json:"department_id" bson:"department_id"`
}

type RoomDepartmentConfig struct {
	Value RoomDepartmentConfigValue `json:"value" bson:"value"`
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

func GetPaymentMethod(department int) (int, error) {
	paymentConfig := []OperaConfig{}
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

func DeleteConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configID, _ := strconv.Atoi(vars["id"])
	db.DB.C("operasettings").Remove(bson.M{"id": configID})
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, nil)
	helpers.ReturnSuccessMessage(w, true)
}

func CheckInArray(number int, arr []int) bool {
	found := false
	for _, val := range arr {
		if val == number {
			found = true
		}
	}
	return found
}