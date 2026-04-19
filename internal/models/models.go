// Package models contains the domain types for PicknShip.
//
// The shapes intentionally mirror a subset of the MYOB AccountRight Plus REST
// API so the demo can later be wired to the real endpoints with minimal
// refactoring:
//
//   - Customer  -> /Contact/Customer
//   - Item      -> /Inventory/Item
//   - Order     -> /Sale/Order/Item
//   - Invoice   -> /Sale/Invoice/Item
package models

import "time"

// Customer mirrors MYOB Contact/Customer (subset).
type Customer struct {
	UID          string // MYOB UID
	DisplayID    string // e.g. "7471"
	CompanyName  string // e.g. "Wesco Farm Products"
	AddressLine1 string
	AddressLine2 string
	Suburb       string
	PostCode     string
	Country      string
	Phone        string
	Email        string
	// Reference / project the customer wants printed on docs.
	Reference string
	// Attention name on the picking list.
	Attention      string
	FreightTaxCode string
	TaxCode        string // GST code
	Terms          string // e.g. "Net 20 after EOM"
	BalanceDue     float64
}

// WescoCompany is the seller's own details, printed on every document.
// Pulled out of MYOB CompanyFile in production.
type WescoCompany struct {
	Name      string
	POBox     string
	Town      string
	PostCode  string
	Phone     string
	Freephone string
	Email     string
	Web       string
	GSTNumber string
	BankName  string
	BankAcct  string
	// Sender address used on the consignment note (despatch warehouse).
	SenderStreet string
	SenderTown   string
	SenderPost   string
}

// Wesco is the live company record for the demo (matches the example invoice).
var Wesco = WescoCompany{
	Name:         "Wesco Seeds Ltd",
	POBox:        "PO Box 22",
	Town:         "Rangiora",
	PostCode:     "7440",
	Phone:        "03 312 5860",
	Freephone:    "0800 643 643",
	Email:        "sales@wesco.co.nz",
	Web:          "www.wesco.co.nz",
	GSTNumber:    "110-255-624",
	BankName:     "Wesco Seeds Ltd",
	BankAcct:     "01-0822-0182632-00",
	SenderStreet: "PO Box 22",
	SenderTown:   "Rangiora",
	SenderPost:   "7440",
}

// Item mirrors MYOB Inventory/Item (subset).
type Item struct {
	UID         string
	Number      string  // item code, e.g. "COCKS"
	Name        string  // display name, e.g. "Cocksfoot"
	UOM         string  // selling unit, e.g. "kg", "EA"
	SellPrice   float64 // ex-GST unit price
	IsMixParent bool    // true for Mix1 (BOM parent)
	IsMixChild  bool    // true for the BOM components highlighted with parent
}

// OrderLine is one row of an order (mirrors MYOB Sale/Order/Item line).
type OrderLine struct {
	LineID      string
	Item        Item
	QtyOrdered  float64
	UnitPrice   float64 // ex GST
	Description string
	// MixGroupID lets us visually group a Mix parent with its components on
	// the picking screen. Lines sharing a non-empty MixGroupID get the same
	// background colour. Empty string means "not in a mix group".
	MixGroupID string
}

// Order mirrors MYOB Sale/Order/Item (subset).
type Order struct {
	UID          string
	Number       string // e.g. "S-6330"
	Date         time.Time
	DueDate      time.Time
	Customer     Customer
	Lines        []OrderLine
	Comment      string
	ShipVia      string // courier company, e.g. "Mainfreight"
	Service      string // courier service, e.g. "Freight"
	SalesRepName string
	// FreightExc is freight ex-GST when it isn't already a line item.
	FreightExc float64
}

// Totals computed for invoice display.
func (o Order) SubTotal() float64 {
	t := 0.0
	for _, l := range o.Lines {
		t += l.QtyOrdered * l.UnitPrice
	}
	return t + o.FreightExc
}

func (o Order) GST() float64 {
	return o.SubTotal() * 0.15
}

func (o Order) Total() float64 {
	return o.SubTotal() + o.GST()
}
