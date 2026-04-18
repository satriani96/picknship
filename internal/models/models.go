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
	UID            string  // MYOB UID
	DisplayID      string  // e.g. "7471"
	CompanyName    string  // e.g. "Wesco Farm Products"
	AddressLine1   string
	AddressLine2   string
	Suburb         string
	PostCode       string
	Country        string
	FreightTaxCode string
	TaxCode        string  // GST code
	Terms          string  // e.g. "20th of following month"
	BalanceDue     float64 // unused in demo
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
	UID        string
	Number     string // e.g. "SO-1001"
	Date       time.Time
	Customer   Customer
	Lines      []OrderLine
	Comment    string
	ShipVia    string
	FreightExc float64 // freight charge ex GST
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
