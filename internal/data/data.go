// Package data holds the hard-coded demo orders.
//
// In production these will come from MYOB AccountRight Plus
// (`/Sale/Order/Item`). The shape of the returned `[]models.Order` is the
// boundary the rest of the app codes against, so swapping data sources is a
// drop-in change.
package data

import (
	"time"

	"github.com/wescoseeds/picknship/internal/models"
)

// demoOrders is the hard-coded data set for the demo.
var demoOrders []models.Order

func init() {
	day := time.Date(2026, 4, 17, 0, 0, 0, 0, time.UTC)
	due := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)

	order1 := models.Order{
		UID:     "uid-order-1",
		Number:  "S-6329",
		Date:    day,
		DueDate: due,
		Customer: models.Customer{
			UID:          "uid-cust-7471",
			DisplayID:    "7471",
			CompanyName:  "Wesco Farm Products",
			AddressLine1: "112 Catherwoods Road",
			Country:      "New Zealand",
			TaxCode:      "GST",
			Terms:        "Net 20 after EOM",
		},
		ShipVia:      "Mainfreight",
		Service:      "Freight",
		SalesRepName: "Jethro Larsen",
		Lines: []models.OrderLine{
			{
				LineID:      "l-1-1",
				Item:        models.Item{UID: "i-cocks", Number: "COCKS", Name: "Cocksfoot", UOM: "kg", SellPrice: 11.30},
				QtyOrdered:  200,
				UnitPrice:   11.30,
				Description: "Cocksfoot",
			},
			{
				LineID:      "l-1-2",
				Item:        models.Item{UID: "i-tim", Number: "TIM", Name: "Timothy", UOM: "kg", SellPrice: 14.75},
				QtyOrdered:  200,
				UnitPrice:   14.75,
				Description: "Timothy",
			},
			{
				LineID:      "l-1-3",
				Item:        models.Item{UID: "i-rad", Number: "DAIK", Name: "Daikon Radish", UOM: "kg", SellPrice: 15.00},
				QtyOrdered:  50,
				UnitPrice:   15.00,
				Description: "Daikon Radish",
			},
		},
	}

	order2 := models.Order{
		UID:     "uid-order-2",
		Number:  "S-6330",
		Date:    day,
		DueDate: due,
		Customer: models.Customer{
			UID:          "uid-cust-horti",
			DisplayID:    "319869",
			CompanyName:  "Horticentre Ltd",
			AddressLine1: "211 Manukau Road",
			Suburb:       "Pukekohe",
			PostCode:     "2120",
			Country:      "New Zealand",
			Phone:        "0800 855 255",
			Email:        "accounts@hortcentre.co.nz",
			Reference:    "319869",
			Attention:    "Chris",
			TaxCode:      "GST",
			Terms:        "Net 20 after EOM",
		},
		ShipVia:      "Mainfreight",
		Service:      "Freight",
		SalesRepName: "Jethro Larsen",
		Lines: []models.OrderLine{
			{
				LineID: "l-2-1",
				Item: models.Item{
					UID: "i-mix1", Number: "MIXED1", Name: "Mix 1",
					UOM: "kg", SellPrice: 0.00, IsMixParent: true,
				},
				QtyOrdered:  0,
				UnitPrice:   0.00,
				Description: "Mix 1 (Black Oats + Vetches blend)",
				MixGroupID:  "mix-1",
			},
			{
				LineID: "l-2-2",
				Item: models.Item{
					UID: "i-oats", Number: "OATS-BLK", Name: "Black Oats",
					UOM: "kg", SellPrice: 2.10, IsMixChild: true,
				},
				QtyOrdered:  140,
				UnitPrice:   2.10,
				Description: "Black Oats (component of Mix 1)",
				MixGroupID:  "mix-1",
			},
			{
				LineID: "l-2-3",
				Item: models.Item{
					UID: "i-vetch", Number: "VETCH", Name: "Vetches",
					UOM: "kg", SellPrice: 4.80, IsMixChild: true,
				},
				QtyOrdered:  60,
				UnitPrice:   4.80,
				Description: "Vetches (component of Mix 1)",
				MixGroupID:  "mix-1",
			},
			{
				LineID:      "l-2-4",
				Item:        models.Item{UID: "i-bare", Number: "BARE-SEP", Name: "Bare & Separate", UOM: "kg", SellPrice: 9.50},
				QtyOrdered:  1,
				UnitPrice:   9.50,
				Description: "Bare & Separate",
			},
			{
				LineID:      "l-2-5",
				Item:        models.Item{UID: "i-phac", Number: "PHAC", Name: "Phacelia", UOM: "kg", SellPrice: 18.40},
				QtyOrdered:  30,
				UnitPrice:   18.40,
				Description: "Phacelia",
			},
			{
				LineID:      "l-2-6",
				Item:        models.Item{UID: "i-frt", Number: "FREIGHT", Name: "Freight", UOM: "EA", SellPrice: 50.00},
				QtyOrdered:  1,
				UnitPrice:   50.00,
				Description: "Freight",
			},
		},
	}

	demoOrders = []models.Order{order1, order2}
}

// All returns the demo orders.
func All() []models.Order {
	return demoOrders
}

// Find returns an order by UID.
func Find(uid string) (models.Order, bool) {
	for _, o := range demoOrders {
		if o.UID == uid {
			return o, true
		}
	}
	return models.Order{}, false
}
