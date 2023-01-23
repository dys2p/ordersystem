package html

import (
	"embed"
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/golang-commonmark/markdown"
	"golang.org/x/text/language"
)

//go:embed *
var files embed.FS

var md = markdown.New(markdown.HTML(true), markdown.Linkify(false))

type TemplateData struct {
	AuthorizedCollID string
}

func (TemplateData) Website() string {
	return "order.proxysto.re"
}

func (TemplateData) Locale() string {
	return "de"
}

func (TemplateData) Locales() []string {
	return []string{"de"}
}

// copied
func (td TemplateData) Map(args ...string) (template.HTML, error) {
	var tags = []language.Tag{}
	var strs = []string{}
	for i := 0; i+1 < len(args); i += 2 {
		tag, err := language.Parse(args[i])
		if err != nil {
			return template.HTML(""), fmt.Errorf("%w (%s)", err, args[i])
		}
		tags = append(tags, tag)
		strs = append(strs, args[i+1])
	}

	locale, _ := language.Parse(td.Locale())

	_, index, _ := language.NewMatcher(tags).Match(locale)
	return template.HTML(strs[index]), nil
}

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
		"Caps":       strings.ToUpper,
		"FmtHuman":   FmtHuman,
		"FmtMachine": FmtMachine,
		"Markdown": func(input string) template.HTML {
			return template.HTML(md.RenderToString([]byte(input)))
		},
	}).ParseFS(files, fn...)).ParseGlob(filepath.Join(os.Getenv("CONFIGURATION_DIRECTORY"), "*.html")))
}

var (
	ClientError         = parse("template.html", "layout.html", "client.html", "client/error.html")
	ClientHello         = parse("template.html", "layout.html", "client.html", "client/hello.html")
	ClientCreate        = parse("template.html", "layout.html", "client.html", "client/collection-create.html")
	ClientCollCancel    = parse("template.html", "layout.html", "client.html", "client/collection-cancel.html")
	ClientCollDelete    = parse("template.html", "layout.html", "client.html", "client/collection-delete.html")
	ClientCollEdit      = parse("template.html", "layout.html", "client.html", "client/collection-edit.html")
	ClientCollLogin     = parse("template.html", "layout.html", "client.html", "client/collection-login.html")
	ClientCollMessage   = parse("template.html", "layout.html", "client.html", "client/collection-message.html")
	ClientCollPayBTCPay = parse("template.html", "layout.html", "client.html", "client/collection-pay-btcpay.html")
	ClientCollSubmit    = parse("template.html", "layout.html", "client.html", "client/collection-submit.html")
	ClientCollView      = parse("template.html", "layout.html", "client.html", "client/collection-view.html")
	ClientSite          = parse("template.html", "layout.html", "client.html", "client/site.html")
	ClientStateGet      = parse("template.html", "layout.html", "client.html", "client/state-get.html")
	ClientStatePost     = parse("template.html", "layout.html", "client.html", "client/state-post.html")

	StoreError                = parse("template.html", "layout.html", "store.html", "store/error.html")
	StoreIndex                = parse("template.html", "layout.html", "store.html", "store/index.html")
	StoreLogin                = parse("template.html", "layout.html", "store.html", "store/login.html")
	StoreCollAccept           = parse("template.html", "layout.html", "store.html", "store/collection-accept.html")
	StoreCollConfirmPayment   = parse("template.html", "layout.html", "store.html", "store/collection-confirm-payment.html")
	StoreCollConfirmPickup    = parse("template.html", "layout.html", "store.html", "store/collection-confirm-pickup.html")
	StoreCollConfirmReshipped = parse("template.html", "layout.html", "store.html", "store/collection-confirm-reshipped.html")
	StoreCollDelete           = parse("template.html", "layout.html", "store.html", "store/collection-delete.html")
	StoreCollEdit             = parse("template.html", "layout.html", "store.html", "store/collection-edit.html")
	StoreCollMarkSpam         = parse("template.html", "layout.html", "store.html", "store/collection-mark-spam.html")
	StoreCollMessage          = parse("template.html", "layout.html", "store.html", "store/collection-message.html")
	StoreCollPriceRised       = parse("template.html", "layout.html", "store.html", "store/collection-price-rised.html")
	StoreCollReturn           = parse("template.html", "layout.html", "store.html", "store/collection-return.html")
	StoreCollReject           = parse("template.html", "layout.html", "store.html", "store/collection-reject.html")
	StoreCollView             = parse("template.html", "layout.html", "store.html", "store/collection-view.html")
	StoreTaskConfirmArrived   = parse("template.html", "layout.html", "store.html", "store/task-confirm-arrived.html")
	StoreTaskConfirmOrdered   = parse("template.html", "layout.html", "store.html", "store/task-confirm-ordered.html")
	StoreTaskConfirmPickup    = parse("template.html", "layout.html", "store.html", "store/task-confirm-pickup.html")
	StoreTaskConfirmReshipped = parse("template.html", "layout.html", "store.html", "store/task-confirm-reshipped.html")
	StoreTaskMarkFailed       = parse("template.html", "layout.html", "store.html", "store/task-mark-failed.html")
)
