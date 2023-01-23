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

func parse(fn ...string) *template.Template {
	return template.Must(template.Must(template.New(fn[0]).Funcs(template.FuncMap{
		"FmtHuman":   FmtHuman,
		"FmtMachine": FmtMachine,
	}).ParseFS(files, fn...)).ParseGlob(filepath.Join(os.Getenv("CONFIGURATION_DIRECTORY"), "*.html")))
}

var (
	ClientError         = parse("layout.html", "client.html", "client/error.html")
	ClientHello         = parse("layout.html", "client.html", "client/hello.html")
	ClientCreate        = parse("layout.html", "client.html", "client/collection-create.html")
	ClientCollCancel    = parse("layout.html", "client.html", "client/collection-cancel.html")
	ClientCollDelete    = parse("layout.html", "client.html", "client/collection-delete.html")
	ClientCollEdit      = parse("layout.html", "client.html", "client/collection-edit.html")
	ClientCollLogin     = parse("layout.html", "client.html", "client/collection-login.html")
	ClientCollMessage   = parse("layout.html", "client.html", "client/collection-message.html")
	ClientCollPayBTCPay = parse("layout.html", "client.html", "client/collection-pay-btcpay.html")
	ClientCollSubmit    = parse("layout.html", "client.html", "client/collection-submit.html")
	ClientCollView      = parse("layout.html", "client.html", "client/collection-view.html")
	ClientStateGet      = parse("layout.html", "client.html", "client/state-get.html")
	ClientStatePost     = parse("layout.html", "client.html", "client/state-post.html")

	StoreError                = parse("layout.html", "store.html", "store/error.html")
	StoreIndex                = parse("layout.html", "store.html", "store/index.html")
	StoreLogin                = parse("layout.html", "store.html", "store/login.html")
	StoreCollAccept           = parse("layout.html", "store.html", "store/collection-accept.html")
	StoreCollConfirmPayment   = parse("layout.html", "store.html", "store/collection-confirm-payment.html")
	StoreCollConfirmPickup    = parse("layout.html", "store.html", "store/collection-confirm-pickup.html")
	StoreCollConfirmReshipped = parse("layout.html", "store.html", "store/collection-confirm-reshipped.html")
	StoreCollDelete           = parse("layout.html", "store.html", "store/collection-delete.html")
	StoreCollEdit             = parse("layout.html", "store.html", "store/collection-edit.html")
	StoreCollMarkSpam         = parse("layout.html", "store.html", "store/collection-mark-spam.html")
	StoreCollMessage          = parse("layout.html", "store.html", "store/collection-message.html")
	StoreCollPriceRised       = parse("layout.html", "store.html", "store/collection-price-rised.html")
	StoreCollReturn           = parse("layout.html", "store.html", "store/collection-return.html")
	StoreCollReject           = parse("layout.html", "store.html", "store/collection-reject.html")
	StoreCollView             = parse("layout.html", "store.html", "store/collection-view.html")
	StoreTaskConfirmArrived   = parse("layout.html", "store.html", "store/task-confirm-arrived.html")
	StoreTaskConfirmOrdered   = parse("layout.html", "store.html", "store/task-confirm-ordered.html")
	StoreTaskConfirmPickup    = parse("layout.html", "store.html", "store/task-confirm-pickup.html")
	StoreTaskConfirmReshipped = parse("layout.html", "store.html", "store/task-confirm-reshipped.html")
	StoreTaskMarkFailed       = parse("layout.html", "store.html", "store/task-mark-failed.html")
)
