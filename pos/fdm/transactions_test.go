package fdm

import (
	"testing"
	"time"

	_ "pos-proxy/config"
	"pos-proxy/integrations/fdm/models"
)

func TestNormalSales(t *testing.T) {
	// connection to FDM
	f, err := fdm.New()
	if err != nil {
		t.Log(err)
	}

	items := make([]models.POSLineItem, 8)
	items[0] = models.POSLineItem{
		Quantity:    3,
		Description: "soda LIGHT 33 CL",
		Price:       6.60,
		VAT:         'A',
	}
	items[1] = models.POSLineItem{
		Quantity:    2,
		Description: "Spaghetti Bolognaise (KLEIN)",
		Price:       10.00,
		VAT:         'B',
	}
	items[2] = models.POSLineItem{
		Quantity:    0.527,
		Description: "Salad Bar",
		Price:       8.53,
		VAT:         'B',
	}
	items[3] = models.POSLineItem{
		Quantity:    1,
		Description: "Steak Hach",
		Price:       14.50,
		VAT:         'B',
	}
	items[4] = models.POSLineItem{
		Quantity:    2,
		Description: "Koffie verkeerd medium",
		Price:       6,
		VAT:         'A',
	}
	items[5] = models.POSLineItem{
		Quantity:    1,
		Description: "Dame Blanche",
		Price:       7.00,
		VAT:         'B',
	}
	items[6] = models.POSLineItem{
		Quantity:    -1,
		Description: "Soda LIGHT 33",
		Price:       -2.20,
		VAT:         'A',
	}
	items[7] = models.POSLineItem{
		Quantity:    1.25,
		Description: "Huiswijnliter",
		Price:       12.50,
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
	vats[2].Percentage = 15.00
	vats[3].FixedAmount = 10.00
	vats[3].FixedAmount = 10.00

	RCRS := "ACAS0001234567" // 14 letters
	user_id := "12345678910"
	ticket_number := 123
	total_amount := 100.10

	req := hashAndSignMsg("PS", plu, time.Now(), user_id,
		ticket_number, total_amount, vats, RCRS)
	res, err := f.Write(req, false, 109)
	if err != nil {
		t.Log(err)
	}
	t.Log(res)
}
