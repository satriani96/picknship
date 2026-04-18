# PicknShip

Pick and Ship demo application for **Wesco Seeds**, designed to sit on top of
MYOB AccountRight Plus.

This is a mobile-first demo; sales-order data is currently hard-coded but the
internal types mirror the MYOB AccountRight `SalesOrder` / `Item` / `Customer`
shapes so the wiring can be swapped to the real API later.

## Tech stack

- **Backend:** Go 1.22 (`net/http`, `html/template`)
- **Frontend:** Tailwind CSS (CDN), HTMX (CDN)
- **Storage:** in-memory store (demo only)

## Run

```bash
go run .
```

Then open <http://localhost:8080> on a phone (or a phone-sized browser window).

## Flow

1. Pick a sales order from the list.
2. Enter picked qty for each line, or hit the green tick to auto-fill the
   ordered qty (and edit down for partial picks).
3. Press **Complete**, confirm in the modal.
4. The Consignment Note, Packing Slip and Tax Invoice are generated and the
   browser opens the print dialog.

## Project layout

```
main.go                       routing + bootstrap
internal/models               Order / Item / Customer types (MYOB-shaped)
internal/data                 hard-coded demo orders
internal/store                in-memory pick-state store
internal/handlers             HTTP + HTMX handlers
templates/                    html/template files
static/                       CSS + Wesco logo
```

The Wesco logo (`Wesco_Primary_RGB.webp`) lives in `static/` and is served at
`/static/Wesco_Primary_RGB.webp`.
