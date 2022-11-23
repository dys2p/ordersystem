package html

import (
	"embed"
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"strings"
)

//go:embed *
var files embed.FS

func centsToFloat(cents int) float64 {
	return math.Round(float64(cents)) / 100.0
}

func FmtHuman(cents int) string {
	return strings.Replace(fmt.Sprintf("%.2f Euro", centsToFloat(cents)), ".", ",", 1)
}

func FmtMachine(cents int) string {
	return fmt.Sprintf("%.2f", centsToFloat(cents)) // for some APIs and HTML <input> tags
}

var funcs = template.FuncMap{
	"FmtHuman":   FmtHuman,
	"FmtMachine": FmtMachine,
}

func parse(fn ...string) *template.Template {
	fn = append([]string{"layout"}, fn...)
	for i := range fn {
		fn[i] = fn[i] + ".html"
	}
	return template.Must(template.Must(template.New("layout.html").Funcs(funcs).ParseFS(files, fn...)).ParseGlob(filepath.Join(os.Getenv("CONFIGURATION_DIRECTORY"), "custom.html")))
}

var (
	ClientError         = parse("client", "client/error")
	ClientHello         = parse("client", "client/hello")
	ClientCreate        = parse("client", "client/collection-create")
	ClientCollCancel    = parse("client", "client/collection-cancel")
	ClientCollDelete    = parse("client", "client/collection-delete")
	ClientCollEdit      = parse("client", "client/collection-edit")
	ClientCollLogin     = parse("client", "client/collection-login")
	ClientCollMessage   = parse("client", "client/collection-message")
	ClientCollPayBTCPay = parse("client", "client/collection-pay-btcpay")
	ClientCollSubmit    = parse("client", "client/collection-submit")
	ClientCollView      = parse("client", "client/collection-view")
	ClientStateGet      = parse("client", "client/state-get")
	ClientStatePost     = parse("client", "client/state-post")

	StoreError                = parse("store", "store/error")
	StoreIndex                = parse("store", "store/index")
	StoreLogin                = parse("store", "store/login")
	StoreCollAccept           = parse("store", "store/collection-accept")
	StoreCollConfirmPayment   = parse("store", "store/collection-confirm-payment")
	StoreCollConfirmPickup    = parse("store", "store/collection-confirm-pickup")
	StoreCollConfirmReshipped = parse("store", "store/collection-confirm-reshipped")
	StoreCollDelete           = parse("store", "store/collection-delete")
	StoreCollEdit             = parse("store", "store/collection-edit")
	StoreCollMarkSpam         = parse("store", "store/collection-mark-spam")
	StoreCollMessage          = parse("store", "store/collection-message")
	StoreCollPriceRised       = parse("store", "store/collection-price-rised")
	StoreCollReturn           = parse("store", "store/collection-return")
	StoreCollReject           = parse("store", "store/collection-reject")
	StoreCollView             = parse("store", "store/collection-view")
	StoreTaskConfirmArrived   = parse("store", "store/task-confirm-arrived")
	StoreTaskConfirmOrdered   = parse("store", "store/task-confirm-ordered")
	StoreTaskConfirmPickup    = parse("store", "store/task-confirm-pickup")
	StoreTaskConfirmReshipped = parse("store", "store/task-confirm-reshipped")
	StoreTaskMarkFailed       = parse("store", "store/task-mark-failed")
)
