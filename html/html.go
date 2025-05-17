package html

import (
	"embed"
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/dys2p/eco/captcha"
	"github.com/dys2p/eco/ssg"
	"gitlab.com/golang-commonmark/markdown"
)

//go:embed *
var Files embed.FS

var md = markdown.New(markdown.HTML(true), markdown.Linkify(false))

type TemplateData struct {
	ssg.TemplateData
	AuthorizedCollID string
}

func centsToFloat(cents int) float64 {
	return math.Round(float64(cents)) / 100.0
}

func FmtEuro(cents int) string {
	return strings.Replace(fmt.Sprintf("%.2f Euro", centsToFloat(cents)), ".", ",", 1)
}

func FmtMachine(cents int) string {
	return fmt.Sprintf("%.2f", centsToFloat(cents)) // for some APIs and HTML <input> tags
}

func parse(fn ...string) *template.Template {
	t := template.New("html").Funcs(template.FuncMap{
		"Cut":        func(s string, maxlen int) string {
			if len(s) > maxlen+3 {
				return s[:maxlen]+"..."
			}
			return s
		},
		"FmtEuro":    FmtEuro,
		"FmtMachine": FmtMachine,
		"Markdown": func(input string) template.HTML {
			return template.HTML(md.RenderToString([]byte(input)))
		},
	})
	t = template.Must(t.Parse(captcha.TemplateString))
	t = template.Must(t.ParseFS(Files, fn...))
	t = template.Must(t.ParseGlob(filepath.Join(os.Getenv("CONFIGURATION_DIRECTORY"), "*.html")))
	return t
}

var (
	ClientError         = parse("order.proxysto.re/*.html", "common.html", "client.html", "client/error.html")
	ClientHello         = parse("order.proxysto.re/*.html", "common.html", "client.html", "client/hello.html")
	ClientCreate        = parse("order.proxysto.re/*.html", "common.html", "client.html", "client/collection-create.html")
	ClientCollCancel    = parse("order.proxysto.re/*.html", "common.html", "client.html", "client/collection-cancel.html")
	ClientCollDelete    = parse("order.proxysto.re/*.html", "common.html", "client.html", "client/collection-delete.html")
	ClientCollEdit      = parse("order.proxysto.re/*.html", "common.html", "client.html", "client/collection-edit.html")
	ClientCollLogin     = parse("order.proxysto.re/*.html", "common.html", "client.html", "client/collection-login.html")
	ClientCollMessage   = parse("order.proxysto.re/*.html", "common.html", "client.html", "client/collection-message.html")
	ClientCollPayBTCPay = parse("order.proxysto.re/*.html", "common.html", "client.html", "client/collection-pay-btcpay.html")
	ClientCollSubmit    = parse("order.proxysto.re/*.html", "common.html", "client.html", "client/collection-submit.html")
	ClientCollView      = parse("order.proxysto.re/*.html", "common.html", "client.html", "client/collection-view.html")
	ClientSite          = parse("order.proxysto.re/*.html", "common.html", "client.html")
	ClientStateGet      = parse("order.proxysto.re/*.html", "common.html", "client.html", "client/state-get.html")
	ClientStatePost     = parse("order.proxysto.re/*.html", "common.html", "client.html", "client/state-post.html")

	StoreError                = parse("common.html", "store.html", "store/error.html")
	StoreIndex                = parse("common.html", "store.html", "store/index.html")
	StoreLogin                = parse("common.html", "store.html", "store/login.html")
	StoreCollAccept           = parse("common.html", "store.html", "store/collection-accept.html")
	StoreCollConfirmPayment   = parse("common.html", "store.html", "store/collection-confirm-payment.html")
	StoreCollConfirmPickup    = parse("common.html", "store.html", "store/collection-confirm-pickup.html")
	StoreCollConfirmReshipped = parse("common.html", "store.html", "store/collection-confirm-reshipped.html")
	StoreCollDelete           = parse("common.html", "store.html", "store/collection-delete.html")
	StoreCollEdit             = parse("common.html", "store.html", "store/collection-edit.html")
	StoreCollMarkSpam         = parse("common.html", "store.html", "store/collection-mark-spam.html")
	StoreCollMessage          = parse("common.html", "store.html", "store/collection-message.html")
	StoreCollReturn           = parse("common.html", "store.html", "store/collection-return.html")
	StoreCollReject           = parse("common.html", "store.html", "store/collection-reject.html")
	StoreCollSubmit           = parse("common.html", "store.html", "store/collection-submit.html")
	StoreCollView             = parse("common.html", "store.html", "store/collection-view.html")
	StoreTaskConfirmArrived   = parse("common.html", "store.html", "store/task-confirm-arrived.html")
	StoreTaskConfirmOrdered   = parse("common.html", "store.html", "store/task-confirm-ordered.html")
	StoreTaskConfirmPickup    = parse("common.html", "store.html", "store/task-confirm-pickup.html")
	StoreTaskConfirmReshipped = parse("common.html", "store.html", "store/task-confirm-reshipped.html")
	StoreTaskMarkFailed       = parse("common.html", "store.html", "store/task-mark-failed.html")
)
