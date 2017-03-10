package fdm

import (
	"testing"
	"time"

	_ "pos-proxy/config"
	"pos-proxy/fdm"
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

	ticket := fdm.Ticket{}
	ticket.TicketNumber = "999000"
	ticket.CreatedAt = time.Now()
	ticket.Items = items
	ticket.RCRS = "ACAS0001234567" // 14 letters
	ticket.UserID = "12345678910"  // 11 letters
	ticket.InvoiceNumber = "1-10"
	ticket.TotalAmount = 3.72
	ticket.PLUHash = fdm.GeneratePLUHash(items)

	vats := make([]fdm.VAT, 4)
	vats[0].Percentage = 21.00
	vats[0].FixedAmount = 10.00
	vats[1].Percentage = 12.00
	vats[1].FixedAmount = 7.00
	vats[2].Percentage = 6.00
	vats[2].FixedAmount = 0.00
	vats[3].Percentage = 0.00
	vats[3].FixedAmount = 7.50

	ticket.VATs = vats

	req := fdm.HashAndSignMsg("PS", ticket)
	res, err := f.Write(req, false, 64)
	if err != nil {
		t.Log(err)
	}
	t.Log(res)
}
