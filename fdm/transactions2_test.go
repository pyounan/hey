package fdm

import (
	"testing"
	"time"

	_ "pos-proxy/config"
	"pos-proxy/fdm"
	"pos-proxy/fdm/tickets"
)

func TestNormalSales(t *testing.T) {
	// connection to FDM
	f, err := fdm.New()
	if err != nil {
		t.Log(err)
	}

	items := make([]fdm.POSLineItem, 3)
	items[0] = fdm.POSLineItem{
		Quantity:    1,
		Description: "Drink 1",
		Price:       0.50,
		VAT:         'A',
	}
	items[1] = fdm.POSLineItem{
		Quantity:    1,
		Description: "food 1",
		Price:       5.00,
		VAT:         'B',
	}
	items[2] = fdm.POSLineItem{
		Quantity:    1,
		Description: "drink 1",
		Price:       0.50,
		VAT:         'A',
	}

	plu := fdm.GeneratePLUHash(items)
	t.Log(plu)

	vats := make([]tickets.VAT, 4)
	vats[0].Percentage = 30.00
	vats[0].FixedAmount = 40.00
	vats[1].Percentage = 15.00
	vats[1].FixedAmount = 10.00
	vats[2].Percentage = 15.00
	vats[2].FixedAmount = 15.00
	vats[3].Percentage = 10.00
	vats[3].FixedAmount = 10.00

	RCRS := "ACAS0001234567" // 14 letters
	user_id := "12345678910"
	ticket_number := 123
	total_amount := 100.10

	req := tickets.HashAndSignMsg("PS", plu, time.Now(), user_id,
		ticket_number, total_amount, vats, RCRS)
	res, err := f.Write(req, false, 109)
	if err != nil {
		t.Log(err)
	}
	t.Log(res)
}
