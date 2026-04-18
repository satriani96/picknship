// Package handlers wires HTTP routes to data + templates.
package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/wescoseeds/picknship/internal/data"
	"github.com/wescoseeds/picknship/internal/models"
	"github.com/wescoseeds/picknship/internal/store"
)

// Server bundles deps for the HTTP handlers.
type Server struct {
	Store *store.Store
	// pages maps page name (e.g. "orders.html") to a fully-parsed template
	// set rooted at the "layout" template. Each page provides its own
	// "content" definition.
	pages map[string]*template.Template
	// fragments holds standalone templates (modal, row) parsed once.
	fragments *template.Template
}

var funcMap = template.FuncMap{
	"sub": func(a, b float64) float64 { return a - b },
	"mul": func(a, b float64) float64 { return a * b },
	"add": func(a, b float64) float64 { return a + b },
}

// shared templates that every page needs (layout, header, partials used by
// both pages and fragments). Keep this list small and explicit.
var sharedTemplates = []string{
	"templates/layout.html",
	"templates/header.html",
	"templates/pickrow.html",
	"templates/consignment.html",
	"templates/packing.html",
	"templates/invoice.html",
}

// pageFiles maps a page key to its file. Each page must define "content".
var pageFiles = map[string]string{
	"orders.html":    "templates/orders.html",
	"pick.html":      "templates/pick.html",
	"documents.html": "templates/documents.html",
}

// New parses templates and returns a Server.
func New(s *store.Store, tplFS fs.FS) (*Server, error) {
	pages := map[string]*template.Template{}
	for name, file := range pageFiles {
		files := append([]string{}, sharedTemplates...)
		files = append(files, file)
		t, err := template.New("").Funcs(funcMap).ParseFS(tplFS, files...)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		pages[name] = t
	}
	frag, err := template.New("").Funcs(funcMap).ParseFS(tplFS,
		"templates/pickrow.html",
		"templates/confirm.html",
	)
	if err != nil {
		return nil, fmt.Errorf("parse fragments: %w", err)
	}
	return &Server{Store: s, pages: pages, fragments: frag}, nil
}

// MustNew panics on error (used at program start).
func MustNew(s *store.Store, tplFS embed.FS) *Server {
	srv, err := New(s, tplFS)
	if err != nil {
		panic(err)
	}
	return srv
}

// Routes registers handlers on a mux.
func (s *Server) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/orders/", s.handleOrders)
}

// ---- view-models ----------------------------------------------------------

type pickRow struct {
	OrderUID   string
	Line       models.OrderLine
	Picked     float64
	MixGroupID string
	MixPos     string // "first" | "mid" | "last" | ""
}

func buildRows(o models.Order, picks map[string]float64) []pickRow {
	rows := make([]pickRow, len(o.Lines))
	for i, l := range o.Lines {
		rows[i] = pickRow{
			OrderUID:   o.UID,
			Line:       l,
			Picked:     picks[l.LineID],
			MixGroupID: l.MixGroupID,
		}
	}
	for i := range rows {
		if rows[i].MixGroupID == "" {
			continue
		}
		prev := i > 0 && rows[i-1].MixGroupID == rows[i].MixGroupID
		next := i < len(rows)-1 && rows[i+1].MixGroupID == rows[i].MixGroupID
		switch {
		case !prev && next:
			rows[i].MixPos = "first"
		case prev && next:
			rows[i].MixPos = "mid"
		case prev && !next:
			rows[i].MixPos = "last"
		}
	}
	return rows
}

// ---- routes ---------------------------------------------------------------

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.render(w, "orders.html", "layout", map[string]any{
		"Title":     "Orders",
		"Subtitle":  "Pick & Ship",
		"Summaries": s.Store.Summaries(),
	})
}

// handleOrders dispatches /orders/{uid}/...
func (s *Server) handleOrders(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/orders/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	uid := parts[0]
	o, ok := data.Find(uid)
	if !ok {
		http.NotFound(w, r)
		return
	}
	switch {
	case len(parts) == 1:
		s.showPick(w, r, o)
	case len(parts) == 2 && parts[1] == "confirm":
		s.showConfirm(w, r, o)
	case len(parts) == 2 && parts[1] == "complete" && r.Method == http.MethodPost:
		s.completeOrder(w, r, o)
	case len(parts) == 2 && parts[1] == "documents":
		s.showDocuments(w, r, o)
	case len(parts) == 4 && parts[1] == "lines" && parts[3] == "pick" && r.Method == http.MethodPost:
		s.setQty(w, r, o, parts[2])
	case len(parts) == 4 && parts[1] == "lines" && parts[3] == "tick" && r.Method == http.MethodPost:
		s.tickQty(w, r, o, parts[2])
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) showPick(w http.ResponseWriter, r *http.Request, o models.Order) {
	if s.Store.Status(o.UID) == "completed" {
		http.Redirect(w, r, "/orders/"+o.UID+"/documents", http.StatusSeeOther)
		return
	}
	picks := s.Store.GetAll(o.UID)
	s.render(w, "pick.html", "layout", map[string]any{
		"Title":    o.Number,
		"Subtitle": o.Customer.CompanyName,
		"Back":     "/",
		"Order":    o,
		"Rows":     buildRows(o, picks),
	})
}

func (s *Server) setQty(w http.ResponseWriter, r *http.Request, o models.Order, lineID string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	raw := strings.TrimSpace(r.FormValue("qty"))
	q, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		q = 0
	}
	s.Store.Set(o.UID, lineID, q)
	s.renderRow(w, o, lineID)
}

func (s *Server) tickQty(w http.ResponseWriter, r *http.Request, o models.Order, lineID string) {
	for _, l := range o.Lines {
		if l.LineID == lineID {
			s.Store.Set(o.UID, lineID, l.QtyOrdered)
			break
		}
	}
	s.renderRow(w, o, lineID)
}

func (s *Server) renderRow(w http.ResponseWriter, o models.Order, lineID string) {
	picks := s.Store.GetAll(o.UID)
	rows := buildRows(o, picks)
	for _, row := range rows {
		if row.Line.LineID == lineID {
			if err := s.fragments.ExecuteTemplate(w, "pickrow", row); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}
	http.NotFound(w, nil)
}

func (s *Server) showConfirm(w http.ResponseWriter, r *http.Request, o models.Order) {
	picks := s.Store.GetAll(o.UID)
	if err := s.fragments.ExecuteTemplate(w, "confirm", map[string]any{
		"Order": o,
		"Rows":  buildRows(o, picks),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) completeOrder(w http.ResponseWriter, r *http.Request, o models.Order) {
	s.Store.Complete(o.UID)
	s.renderDocuments(w, r, o)
}

func (s *Server) showDocuments(w http.ResponseWriter, r *http.Request, o models.Order) {
	s.renderDocuments(w, r, o)
}

func (s *Server) renderDocuments(w http.ResponseWriter, r *http.Request, o models.Order) {
	picks := s.Store.GetAll(o.UID)
	rows := buildRows(o, picks)
	picked := s.Store.PickedOrder(o)

	pickedRows := []models.OrderLine{}
	for _, l := range picked.Lines {
		if l.QtyOrdered > 0 || l.Item.IsMixParent {
			pickedRows = append(pickedRows, l)
		}
	}

	invoiceRows := []models.OrderLine{}
	subTotal := 0.0
	for _, l := range picked.Lines {
		if l.QtyOrdered <= 0 && !l.Item.IsMixParent {
			continue
		}
		invoiceRows = append(invoiceRows, l)
		subTotal += l.QtyOrdered * l.UnitPrice
	}
	subTotal += o.FreightExc
	gst := subTotal * 0.15
	total := subTotal + gst

	totalWeight := 0.0
	for _, l := range pickedRows {
		if strings.EqualFold(l.Item.UOM, "kg") {
			totalWeight += l.QtyOrdered
		}
	}

	connote := fmt.Sprintf("MFL%s", strings.TrimPrefix(o.Number, "SO-"))
	invoiceNo := fmt.Sprintf("INV-%s", strings.TrimPrefix(o.Number, "SO-"))

	s.render(w, "documents.html", "layout", map[string]any{
		"Title":         "Documents · " + o.Number,
		"Order":         o,
		"Rows":          rows,
		"PickedRows":    pickedRows,
		"InvoiceRows":   invoiceRows,
		"SubTotal":      subTotal,
		"GST":           gst,
		"Total":         total,
		"TotalWeightKg": totalWeight,
		"ConnoteNo":     connote,
		"InvoiceNo":     invoiceNo,
		"PrintDate":     time.Now().Format("2 Jan 2006"),
	})
}

// ---- helpers --------------------------------------------------------------

func (s *Server) render(w http.ResponseWriter, page, layout string, data map[string]any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, ok := s.pages[page]
	if !ok {
		http.Error(w, "unknown page: "+page, http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, layout, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
