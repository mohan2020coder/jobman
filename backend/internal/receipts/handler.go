package receipts

import (
	"fmt"
	"html"
	"net/http"
	"strconv"

	"github.com/jobman/backend/internal/identity"
	"github.com/jobman/backend/pkg/httpapi"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetByJob(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	rec, err := h.svc.GetByJob(r.Context(), p.BusinessID, r.PathValue("id"))
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{
		"receipt":      rec,
		"public_url":   "/r/" + rec.PublicToken,
	})
}

// PublicJSON serves the public receipt as JSON (no auth).
func (h *Handler) PublicJSON(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.PublicReceipt(r.Context(), r.PathValue("token"))
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, res)
}

// PublicPage renders a simple HTML receipt at /r/{token}.
func (h *Handler) PublicPage(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.PublicReceipt(r.Context(), r.PathValue("token"))
	if err != nil {
		httpapi.WriteError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Receipt not found.")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page := renderReceiptHTML(res)
	_, _ = w.Write([]byte(page))
}

func renderReceiptHTML(res *PublicReceipt) string {
	bizName := html.EscapeString(res.Business.Name)
	if bizName == "" {
		bizName = "Receipt"
	}
	cust := html.EscapeString(res.CustomerName)
	service := html.EscapeString(res.Service)
	currency := "₹"

	var itemsHTML string
	for _, it := range res.Items {
		itemsHTML += fmt.Sprintf(
			"<tr><td>%s</td><td>x%s</td><td style='text-align:right'>%s</td></tr>",
			html.EscapeString(it.Description), trimZero(it.Quantity), currency+fmtMoney(it.Amount))
	}
	if itemsHTML == "" {
		itemsHTML = "<tr><td colspan='3' style='text-align:center;color:#888'>No items listed</td></tr>"
	}

	phone := html.EscapeString(res.Business.Phone)
	statusColor := map[string]string{
		"PAID":    "#1f9d55",
		"PARTIAL": "#b7791f",
		"UNPAID":  "#c53030",
	}[res.PaymentStatus]

	completed := ""
	if res.CompletedAt != nil {
		completed = html.EscapeString(res.CompletedAt.Format("02 Jan 2006, 03:04 PM"))
	}

	args := []any{
		bizName,                 // %[1]s title + h1
		phone,                   // %[2]s phone under h1
		receiptNumber(res),      // %[3]s receipt number
		cust,                    // %[4]s customer
		service,                 // %[5]s service
		itemsHTML,               // %[6]s items table body
		currency + fmtMoney(res.Total), // %[7]s
		currency + fmtMoney(res.Paid),  // %[8]s
		joinMethods(res.PaymentMethods), // %[9]s
		statusColor,             // %[10]s badge background
		res.PaymentStatus,       // %[11]s badge text
		completed,               // %[12]s completed date
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%[1]s</title>
<style>
  body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:#f4f6fb;margin:0;padding:24px;color:#1a202c}
  .card{max-width:520px;margin:0 auto;background:#fff;border-radius:14px;box-shadow:0 4px 16px rgba(0,0,0,.08);padding:32px}
  h1{font-size:20px;margin:0 0 4px}
  .muted{color:#718096;font-size:13px}
  table{width:100%;border-collapse:collapse;margin-top:16px;font-size:14px}
  th,td{padding:8px 6px;border-bottom:1px solid #edf2f7;text-align:left}
  th{color:#718096;font-weight:500;font-size:12px}
  .total-row td{font-weight:700;font-size:16px;border-top:2px solid #e2e8f0}
  .badge{display:inline-block;padding:3px 10px;border-radius:999px;color:#fff;font-size:12px;font-weight:600;margin-top:8px}
  .hr{border:none;border-top:1px dashed #cbd5e0;margin:20px 0}
  footer{margin-top:24px;font-size:12px;color:#a0aec0;text-align:center}
</style>
</head>
<body>
<div class="card">
  <h1>%[1]s</h1>
  <div class="muted">%[2]s</div>
  <div class="muted">Receipt %[3]s</div>
  <div style="margin-top:16px">
    <div class="muted">Bill to</div>
    <div style="font-weight:600">%[4]s</div>
  </div>
  <div style="margin-top:12px">
    <div class="muted">Service</div>
    <div style="font-weight:600">%[5]s</div>
  </div>
  <table>
    <tr><th>Description</th><th>Qty</th><th style="text-align:right">Amount</th></tr>
    %[6]s
    <tr class="total-row"><td colspan="2">Total</td><td style="text-align:right">%[7]s</td></tr>
  </table>
  <div style="margin-top:12px;font-size:14px">
    <span class="muted">Paid:</span> <strong>%[8]s</strong>
    <span style="margin-left:16px">Methods: <strong>%[9]s</strong></span>
  </div>
  <div class="badge" style="background:%[10]s">%[11]s</div>
  <hr class="hr">
  <div class="muted">Completed: %[12]s</div>
  <footer>Thank you for your business!</footer>
</div>
</body>
</html>`, args...)
}

func receiptNumber(res *PublicReceipt) string { return html.EscapeString(res.ReceiptNumber) }
func trimZero(v float64) string               { return strconv.FormatFloat(v, 'f', -1, 64) }
func joinMethods(m []string) string {
	if len(m) == 0 {
		return "—"
	}
	out := ""
	for i, s := range m {
		if i > 0 {
			out += ", "
		}
		out += html.EscapeString(s)
	}
	return out
}

func fmtMoney(v float64) string {
	raw := strconv.FormatFloat(v, 'f', 2, 64)
	neg := ""
	if raw[0] == '-' {
		neg = "-"
		raw = raw[1:]
	}
	intPart := raw
	frac := ""
	if i := indexByte(raw, '.'); i >= 0 {
		intPart = raw[:i]
		frac = raw[i:]
	}
	// insert thousands separators
	var b []byte
	n := len(intPart)
	for i, c := range []byte(intPart) {
		if i > 0 && (n-i)%3 == 0 {
			b = append(b, ',')
		}
		b = append(b, c)
	}
	return neg + string(b) + frac
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}