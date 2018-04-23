package db

import (
	"log"
	"pos-proxy/entity"

	"gopkg.in/mgo.v2/bson"
)

// GetCCVSettingsForTerminal returns an object for the configuration of a CCV pinpad
// attached to a certain terminal
func GetCCVSettingsForTerminal(terminalID int) (*entity.CCVSettings, error) {
	session := Session.Copy()
	cti := entity.CCVTerminalIntegration{}
	err := DB.C("ccv_terminal_integration_settings").With(session).Find(bson.M{"terminal_id": terminalID}).One(&cti)
	if err != nil {
		return nil, err
	}
	log.Println("found terminal integration with CCV", cti, cti.CCVSettingsID)
	ccv := entity.CCVSettings{}
	err = DB.C("ccv_settings").With(session).Find(bson.M{"id": cti.CCVSettingsID}).One(&ccv)
	if err != nil {
		return nil, err
	}
	return &ccv, nil
}
