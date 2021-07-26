package html

import (
	"net/http"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type TagStr struct {
	Tag language.Tag
	Str string
}

// global variable used by type Language
var translations = map[string][]TagStr{
	// collection-cancel.html
	"coll-cancel": []TagStr{
		TagStr{language.AmericanEnglish, "Cancel order"},
		TagStr{language.German, "Auftrag abbrechen"},
	},
	"confirm-yes": []TagStr{
		TagStr{language.AmericanEnglish, "Yes, I am sure."},
		TagStr{language.German, "Ja, ich bin mir sicher."},
	},	
	"coll-cancel-invalid": []TagStr{
		TagStr{language.AmericanEnglish, "Please confirm that you would like to cancel the order."},
		TagStr{language.German, "Bitte bestätige, dass du den Auftrag abbrechen möchtest."},
	},	
	"go-back": []TagStr{
		TagStr{language.AmericanEnglish, "Cancel and return"},
		TagStr{language.German, "Abbrechen und zurück"},
	},
	
// collection-create.html	
	"coll-new": []TagStr{
		TagStr{language.AmericanEnglish, "Create new order"},
		TagStr{language.German, "Neuen Bestellauftrag beginnen"},
	},	
	"coll-new-intro": []TagStr{
		TagStr{language.AmericanEnglish, "We have generated a random order ID and passphrase for your new order:"},
		TagStr{language.German, "Für deinen neuen Bestellauftrag haben wir dir eine Auftragsnummer und eine Passphrase ausgewürfelt:"},
	},	
	"coll-id": []TagStr{
		TagStr{language.AmericanEnglish, "Order ID"},
		TagStr{language.German, "Auftragsnummer"},
	},	
	"coll-id-invalid": []TagStr{
		TagStr{language.AmericanEnglish, "There should be a ten-digit ID here. Please reload the page."},
		TagStr{language.German, "Hier sollte eine zehnstellige Buchstabenkombination stehen. Bitte lade die Seite neu."},
	},
	"coll-pass": []TagStr{
		TagStr{language.AmericanEnglish, "Passphrase"},
		TagStr{language.German, "Passphrase"},
	},
"coll-pass-invalid": []TagStr{
		TagStr{language.AmericanEnglish, "Please enter a passphrase."},
		TagStr{language.German, "Bitte gib eine Passphrase ein."},
	},
	"coll-new-check": []TagStr{
		TagStr{language.AmericanEnglish, "Yes, I have written down, memorized or saved the order ID and passphrase."},
		TagStr{language.German, "Ja, ich habe mir die Auftragsnummer und die Passphrase aufgeschrieben, gemerkt oder gespeichert."},
	},
	"coll-new-invalid": []TagStr{
		TagStr{language.AmericanEnglish, "Please confirm that you have written down, memorized or saved the order ID and passphrase."},
		TagStr{language.German, "Bitte bestätige, dass du dir die Auftragsnummer und die Passphrase aufgeschrieben, gemerkt oder gespeichert hast."},
	},
	"cancel": []TagStr{
		TagStr{language.AmericanEnglish, "Cancel"},
		TagStr{language.German, "Abbrechen"},
	},

// collection-delete.html
	"coll-delete": []TagStr{
		TagStr{language.AmericanEnglish, "Delete draft"},
		TagStr{language.German, "Auftragsentwurf löschen"},
	},
	"coll-delete-invalid": []TagStr{
		TagStr{language.AmericanEnglish, "Please confirm that you want to delete the draft."},
		TagStr{language.German, "Bitte bestätige, dass den Auftragsentwurf löschen möchtest."},
	},

// collection-edit.html
	"coll-edit-order": []TagStr{
		TagStr{language.AmericanEnglish, "Order"},
		TagStr{language.German, "Auftrag"},
	},
	"coll-edit-status": []TagStr{
		TagStr{language.AmericanEnglish, "Status:"},
		TagStr{language.German, "Status:"},
	},
	"coll-edit-js": []TagStr{
		TagStr{language.AmericanEnglish, "Sorry, this page requires JavaScript."},
		TagStr{language.German, "Dieses Seite erfordert leider JavaScript."},
	},
	"coll-edit-button": []TagStr{
		TagStr{language.AmericanEnglish, "Save order draft"},
		TagStr{language.German, "Auftragsentwurf speichern"},
	},

// collection-login.html
	"coll-login": []TagStr{
		TagStr{language.AmericanEnglish, "Login"},
		TagStr{language.German, "Anmelden"},
	},
	"coll-login-invalid-id": []TagStr{
		TagStr{language.AmericanEnglish, "Please enter your ten-digit order ID."},
		TagStr{language.German, "Bitte gib deine zehnstellige Auftragsnummer ein.},
	},
	"coll-login-invalid-pass": []TagStr{
		TagStr{language.AmericanEnglish, "Please enter the correct passphrase."},
		TagStr{language.German, "Bitte gib die korrekte Passphrase ein.},
	},	
	
// collection-message.html
	"coll-msg": []TagStr{
		TagStr{language.AmericanEnglish, "Leave a message"},
		TagStr{language.German, "Nachricht hinterlassen"},
	},	
	"coll-md": []TagStr{
		TagStr{language.AmericanEnglish, "You can type Markdown (CommonMark)."},
		TagStr{language.German, "Du kannst Markdown (CommonMark) eingeben."},
	},	

// collection-pay-btcpay.html
	"coll-btcpay": []TagStr{
		TagStr{language.AmericanEnglish, "Pay with Bitcoin or Monero"},
		TagStr{language.German, "Mit Bitcoin oder Monero bezahlen"},
	},
	"coll-btcpay-text": []TagStr{
		TagStr{language.AmericanEnglish, "<p>If you continue, an invoice about <strong>{{FmtHuman .Due}}</strong> will be generated on our BTCPayServer, you can pay with Bitcoin (BTC) or Monero (XMR). The invoice is valid for <strong>60&nbsp;minutes</strong>. Until the time expires, your transaction must be visible in the blockchain.</p>
<p>Please make sure that you pay the bill <strong>on time, complete and in a single transaction</strong>. Only then we can accept the conversion rate. If your exchange or client deducts transaction fees from the amount, you must add them before. <strong>If your payment is late or partial, the coins will still be sold automatically on an exchange.</strong> After that, we will manually enter the achieved sale value here.</p>"},
		TagStr{language.German, "<p>Wenn du fortfährst, wird auf unserem BTCPayServer eine Rechnung über <strong>{{FmtHuman .Due}}</strong> erzeugt, du mit Bitcoin (BTC) oder Monero (XMR) bezahlen kannst. Die Rechnung ist <strong>60&nbsp;Minuten</strong> lang gültig. Bis zum Ablauf der Zeit muss deine Transaktion in der Blockchain sichtbar sein.</p>
<p>Bitte achte darauf, dass du die Rechnung <strong>rechtzeitig, vollständig und mit einer einzelnen Transaktion</strong> bezahlst. Nur dann können wir den Umrechnungskurs akzeptieren. <strong>Falls deine Börse oder dein Client die Transaktionsgebühren von dem Betrag abzieht, musst du sie vorher hinzuaddieren.</strong> Falls deine Zahlung verspätet oder nur teilweise eintrifft, werden die Coins trotzdem automatisch an einer Börse verkauft. Danach werden wir den erzielten Verkaufswert manuell hier eintragen.</p>"},
	},	
	"coll-btcpay-button": []TagStr{
		TagStr{language.AmericanEnglish, "Generate invoice"},
		TagStr{language.German, "Rechnung erzeugen"},
	},

// collection-submit.html
	"coll-submit": []TagStr{
		TagStr{language.AmericanEnglish, "Submit order"},
		TagStr{language.German, "Auftrag einreichen"},
	},
	"coll-submit-confirm": []TagStr{
		TagStr{language.AmericanEnglish, "Please confirm your order by clicking on "Place binding order" below."},
		TagStr{language.German, "Bitte bestätige deinen Auftrag, indem du unten auf "Verbindlichen Bestellauftrag erteilen" klickst."},
	},
	"coll-submit-payment": []TagStr{
		TagStr{language.AmericanEnglish, "Once we have accepted your order, you can pay it <strong>only by prepayment</strong> in the following way:"},
		TagStr{language.German, "Sobald wir deine Bestellung akzeptiert haben, kannst du sie <strong>nur per Vorkasse</strong> auf folgende Weise bezahlen:"},
	},
	"coll-submit-cash": []TagStr{
		TagStr{language.AmericanEnglish, "cash (on site in our store or e.&nbsp;g. up to 100 Euro with Deutsche Post "Einschreiben Wert")"},
		TagStr{language.German, "bar (vor Ort in unserem Laden oder z.&nbsp;B. bis 100 Euro mit Deutsche Post "Einschreiben Wert")"},
	},
	"coll-submit-bank": []TagStr{
		TagStr{language.AmericanEnglish, "by bank transfer"},
		TagStr{language.German, "per Überweisung"},
	},
	"coll-submit-coins": []TagStr{
		TagStr{language.AmericanEnglish, "with Bitcoin and Monero"},
		TagStr{language.German, "mit Bitcoin und Monero"},
	},
	"coll-submit-gwg": []TagStr{
		TagStr{language.AmericanEnglish, "Please note the respective maximum limits according to the Money Laundering Act."},
		TagStr{language.German, "Beachte bitte die jeweiligen Höchstgrenzen gemäß Geldwäschegesetz."},
	},
	"coll-submit-tos": []TagStr{
		TagStr{language.AmericanEnglish, "Please note our terms and conditions and our cancellation policy."},
		TagStr{language.German, "Bitte beachte unsere AGB und unsere Widerrufsbelehrung."},
	},
	"coll-submit-note": []TagStr{
		TagStr{language.AmericanEnglish, "Space for notes"},
		TagStr{language.German, "Raum für Anmerkungen"},
	},
	"coll-submit-button": []TagStr{
		TagStr{language.AmericanEnglish, "Place binding order"},
		TagStr{language.German, "Verbindlichen Bestellauftrag erteilen"},
	},

// collection-view.html
	"coll-view-reminder": []TagStr{
		TagStr{language.AmericanEnglish, "Don't forget to submit your order when you're done."},
		TagStr{language.German, "Vergiss nicht, deinen Bestellauftrag einzureichen, wenn du fertig bist."},
	},
	"coll-view-coins": []TagStr{
		TagStr{language.AmericanEnglish, "Pay with Bitcoin or Monero"},
		TagStr{language.German, "Mit Bitcoin oder Monero bezahlen"},
	},
	"coll-view-submit": []TagStr{
		TagStr{language.AmericanEnglish, "Submit order"},
		TagStr{language.German, "Bestellauftrag einreichen"},
	},
	"coll-view-edit": []TagStr{
		TagStr{language.AmericanEnglish, "Edit order"},
		TagStr{language.German, "Bestellauftrag bearbeiten"},
	},
	"coll-view-cancel": []TagStr{
		TagStr{language.AmericanEnglish, "Cancel order"},
		TagStr{language.German, "Bestellauftrag abbrechen"},
	},
	"coll-view-delete": []TagStr{
		TagStr{language.AmericanEnglish, "Delete order"},
		TagStr{language.German, "Bestellauftrag löschen"},
	},
	"coll-view-course": []TagStr{
		TagStr{language.AmericanEnglish, "Course"},
		TagStr{language.German, "Verlauf"},
	},

// error.html
	"error1": []TagStr{
		TagStr{language.AmericanEnglish, "Error"},
		TagStr{language.German, "Fehler"},
	},
	"error2": []TagStr{
		TagStr{language.AmericanEnglish, "Back to home page"},
		TagStr{language.German, "Zurück zur Startseite"},
	},

// state-get.html
	"state-get": []TagStr{
		TagStr{language.AmericanEnglish, "Order status"},
		TagStr{language.German, "Auftragsstatus"},
	},
	"state-get-numbers": []TagStr{
		TagStr{language.AmericanEnglish, "Please type in the numbers:"},
		TagStr{language.German, "Bitte tippe die Ziffern ab:"},
	},
	"state-get-captcha": []TagStr{
		TagStr{language.AmericanEnglish, "Please type the captcha correctly."},
		TagStr{language.German, "Bitte tippe das Captcha korrekt ab."},
	},
	"state-get-reload": []TagStr{
		TagStr{language.AmericanEnglish, "Load other image (requires JavaScript)"},
		TagStr{language.German, "Anderes Bild laden (erfordert JavaScript)"},
	},
	"state-get-status": []TagStr{
		TagStr{language.AmericanEnglish, "View order status"},
		TagStr{language.German, "Auftragsstatus anzeigen"},
	},

// hello.html
	"coll-display-state": []TagStr{
		TagStr{language.AmericanEnglish, "Display order state only"},
		TagStr{language.German, "Nur Auftragsstatus anzeigen"},
	},

// state-post.html
	"state-post-login": []TagStr{
		TagStr{language.AmericanEnglish, "Log in to view or edit the order:"},
		TagStr{language.German, "Logge dich ein, um den Auftrag anzusehen oder zu bearbeiten:"},
	},
}

// Language is any string. It will be matched by golang.org/x/text/language.Make and golang.org/x/text/language.NewMatcher.
type Language string

// GetLanguage returns the "lang" GET parameter or, if not present, the Accept-Language header value.
// No matching is performed.
func GetLanguage(r *http.Request) Language {
	if lang := r.URL.Query().Get("lang"); lang != "" {
		if len(lang) > 35 {
			lang = lang[:35] // max length of language tag
		}
		return Language(lang)
	}
	return Language(r.Header.Get("Accept-Language"))
}

func (lang Language) Translate(key string, args ...interface{}) string {
	item, ok := translations[key]
	if !ok {
		// no translation available, create language tag and print
		return message.NewPrinter(language.Make(string(lang))).Sprintf(key, args...)
	}
	// choose language tag from list of translations
	langs := make([]language.Tag, len(item))
	for i := range item {
		langs[i] = item[i].Tag
	}
	tag, i := language.MatchStrings(language.NewMatcher(langs), string(lang))
	return message.NewPrinter(tag).Sprintf(item[i].Str, args...)
}
