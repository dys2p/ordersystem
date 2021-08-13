package html

import (
	"html/template"

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
	}, // /store/collection-delete.html, /store/collection-reject.html, /store/collection-mark-spam.html
	"coll-cancel-invalid": []TagStr{
		TagStr{language.AmericanEnglish, "Please confirm that you would like to cancel the order."},
		TagStr{language.German, "Bitte bestätige, dass du den Auftrag abbrechen möchtest."},
	},
	"go-back": []TagStr{
		TagStr{language.AmericanEnglish, "Cancel and return"},
		TagStr{language.German, "Abbrechen und zurück"},
	}, // /client/collection-pay-btcpay.html, /client/collection-delete.html, /client/collection-edit.html, /client/collection-submit.html, /client/collection-message.html, /store/task-confirm-ordered.html, /store/collection-delete.html, /store/collection-accept.html, /store/collection-return.html, /store/collection-reject.html, /store/collection-price-rised.html, /store/task-mark-failed.html, /store/collection-edit.html, /store/task-confirm-reshipped.html, /store/collection-message.html, /store/collection-confirm-reshipped.html, /store/task-confirm-pickup.html, /store/task-confirm-arrived.html, /store/collection-confirm-pickup.html, /store/collection-mark-spam.html, /store/collection-confirm-payment.html

	// collection-create.html
	"coll-new": []TagStr{
		TagStr{language.AmericanEnglish, "Create new order"},
		TagStr{language.German, "Neuen Bestellauftrag beginnen"},
	},
	"coll-view-edit": []TagStr{
		TagStr{language.AmericanEnglish, "View or edit order"},
		TagStr{language.German, "Auftrag ansehen oder bearbeiten"},
	},
	"coll-new-intro": []TagStr{
		TagStr{language.AmericanEnglish, "We have generated a random order ID and passphrase for your new order:"},
		TagStr{language.German, "Für deinen neuen Bestellauftrag haben wir dir eine Auftragsnummer und eine Passphrase ausgewürfelt:"},
	},
	"coll-id": []TagStr{
		TagStr{language.AmericanEnglish, "Order ID"},
		TagStr{language.German, "Auftragsnummer"},
	}, // /client/state-post.html, /client/state-get.html, /store/index.html
	"coll-id-invalid": []TagStr{
		TagStr{language.AmericanEnglish, "There should be a ten-digit ID here. Please reload the page."},
		TagStr{language.German, "Hier sollte eine zehnstellige Buchstabenkombination stehen. Bitte lade die Seite neu."},
	},
	"coll-pass": []TagStr{
		TagStr{language.AmericanEnglish, "Passphrase"},
		TagStr{language.German, "Passphrase"},
	}, // /client/state-post.html, /client/collection-login.html
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
	}, // /client/collection-cancel.html, /client/state-post.html, /client/state-get.html, /client/collection-login.html

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
	}, // /client/state-post.html, /store/collection-view.html
	"coll-edit-status": []TagStr{
		TagStr{language.AmericanEnglish, "Status:"},
		TagStr{language.German, "Status:"},
	}, // /client/state-post.html, /store/collection-view.html, layout.html
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
	}, // /client/state-post.html, /client/state-get.html, /store/login.html
	"coll-login-invalid-id": []TagStr{
		TagStr{language.AmericanEnglish, "Please enter your ten-digit order ID."},
		TagStr{language.German, "Bitte gib deine zehnstellige Auftragsnummer ein."},
	}, // /state-get.html
	"coll-login-invalid-pass": []TagStr{
		TagStr{language.AmericanEnglish, "Please enter the correct passphrase."},
		TagStr{language.German, "Bitte gib die korrekte Passphrase ein."},
	},

	// collection-message.html
	"coll-msg": []TagStr{
		TagStr{language.AmericanEnglish, "Leave a message"},
		TagStr{language.German, "Nachricht hinterlassen"},
	}, // /client/collection-view.html, /store/collection-message.html, /store/collection-view.html
	"coll-md": []TagStr{
		TagStr{language.AmericanEnglish, "You can type Markdown (CommonMark)."},
		TagStr{language.German, "Du kannst Markdown (CommonMark) eingeben."},
	}, // /client/collection-submit.html, /store/collection-accept.html, /store/collection-return.html, /store/collection-reject.html, /store/collection-price-rised.html, /store/task-mark-failed.html, /store/collection-message.html, /store/collection-confirm-payment.html

	// collection-pay-btcpay.html
	"coll-btcpay": []TagStr{
		TagStr{language.AmericanEnglish, "Pay with Bitcoin or Monero"},
		TagStr{language.German, "Mit Bitcoin oder Monero bezahlen"},
	},
	"coll-btcpay-text": []TagStr{
		TagStr{language.AmericanEnglish, `<p>If you continue, an invoice of <strong>%s</strong> will be generated on our BTCPayServer. You can pay it with Bitcoin (BTC) or Monero (XMR). The invoice is valid for <strong>60&nbsp;minutes</strong>. Your transaction must become visible in the blockchain before this time expires.</p>
<p>Please make sure that you pay the bill <strong>in time, completely and with a single transaction</strong>. Only then we can accept the conversion rate. If your exchange or client deducts transaction fees from the amount, you must add them before. <strong>If your payment is late or partial, the coins will still be sold automatically on an exchange.</strong> After that, we will manually enter the achieved sale value here.</p>`},
		TagStr{language.German, `<p>Wenn du fortfährst, wird auf unserem BTCPayServer eine Rechnung über <strong>%s</strong> erzeugt, du mit Bitcoin (BTC) oder Monero (XMR) bezahlen kannst. Die Rechnung ist <strong>60&nbsp;Minuten</strong> lang gültig. Bis zum Ablauf der Zeit muss deine Transaktion in der Blockchain sichtbar sein.</p>
<p>Bitte achte darauf, dass du die Rechnung <strong>rechtzeitig, vollständig und mit einer einzelnen Transaktion</strong> bezahlst. Nur dann können wir den Umrechnungskurs akzeptieren. <strong>Falls deine Börse oder dein Client die Transaktionsgebühren von dem Betrag abzieht, musst du sie vorher hinzuaddieren.</strong> Falls deine Zahlung verspätet oder nur teilweise eintrifft, werden die Coins trotzdem automatisch an einer Börse verkauft. Danach werden wir den erzielten Verkaufswert manuell hier eintragen.</p>`},
	},
	"coll-btcpay-button": []TagStr{
		TagStr{language.AmericanEnglish, "Create invoice"},
		TagStr{language.German, "Rechnung erzeugen"},
	},

	// collection-submit.html
	"coll-submit": []TagStr{
		TagStr{language.AmericanEnglish, "Submit order"},
		TagStr{language.German, "Auftrag einreichen"},
	},
	"coll-submit-confirm": []TagStr{
		TagStr{language.AmericanEnglish, `Please confirm your order by clicking on "Place binding order" below.`},
		TagStr{language.German, `Bitte bestätige deinen Auftrag, indem du unten auf "Verbindlichen Bestellauftrag erteilen" klickst.`},
	},
	"coll-submit-payment": []TagStr{
		TagStr{language.AmericanEnglish, "Once we have accepted your order, you can pay it <strong>in advance only</strong> in one of the following ways:"},
		TagStr{language.German, "Sobald wir deine Bestellung akzeptiert haben, kannst du sie <strong>nur per Vorkasse</strong> auf folgende Weise bezahlen:"},
	},
	"coll-submit-cash": []TagStr{
		TagStr{language.AmericanEnglish, `cash (on site in our store or e.&nbsp;g. up to 100 Euro with Deutsche Post "Einschreiben Wert")`},
		TagStr{language.German, `bar (vor Ort in unserem Laden oder z.&nbsp;B. bis 100 Euro mit Deutsche Post "Einschreiben Wert")`},
	},
	"coll-submit-bank": []TagStr{
		TagStr{language.AmericanEnglish, "by bank transfer"},
		TagStr{language.German, "per Überweisung"},
	},
	"coll-submit-coins": []TagStr{
		TagStr{language.AmericanEnglish, "with Bitcoin or Monero"},
		TagStr{language.German, "mit Bitcoin oder Monero"},
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
		TagStr{language.AmericanEnglish, "Space for remarks"},
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
	"coll-view-cancel": []TagStr{
		TagStr{language.AmericanEnglish, "Cancel order"},
		TagStr{language.German, "Bestellauftrag abbrechen"},
	},
	"coll-view-delete": []TagStr{
		TagStr{language.AmericanEnglish, "Delete order"},
		TagStr{language.German, "Bestellauftrag löschen"},
	},
	"coll-view-course": []TagStr{
		TagStr{language.AmericanEnglish, "Log"},
		TagStr{language.German, "Verlauf"},
	}, // /store/collection-view.html

	// error.html
	"error1": []TagStr{
		TagStr{language.AmericanEnglish, "Error"},
		TagStr{language.German, "Fehler"},
	}, // /store/error.html
	"error2": []TagStr{
		TagStr{language.AmericanEnglish, "Back to main page"},
		TagStr{language.German, "Zurück zur Startseite"},
	}, // /store/error.html

	// state-get.html
	"state-get": []TagStr{
		TagStr{language.AmericanEnglish, "Order status"},
		TagStr{language.German, "Auftragsstatus"},
	},
	"state-get-numbers": []TagStr{
		TagStr{language.AmericanEnglish, "Please type the digits:"},
		TagStr{language.German, "Bitte tippe die Ziffern ab:"},
	},
	"state-get-captcha": []TagStr{
		TagStr{language.AmericanEnglish, "Please type the captcha correctly."},
		TagStr{language.German, "Bitte tippe das Captcha korrekt ab."},
	}, // layout.html
	"state-get-reload": []TagStr{
		TagStr{language.AmericanEnglish, "Load other image (requires JavaScript)"},
		TagStr{language.German, "Anderes Bild laden (erfordert JavaScript)"},
	},
	"state-get-status": []TagStr{
		TagStr{language.AmericanEnglish, "View order status"},
		TagStr{language.German, "Auftragsstatus anzeigen"},
	},

	// hello.html
	"welcome": []TagStr{
		TagStr{language.AmericanEnglish, "Welcome!"},
		TagStr{language.German, "Willkommen!"},
	},
	"coll-display-state": []TagStr{
		TagStr{language.AmericanEnglish, "Show order state only"},
		TagStr{language.German, "Nur Auftragsstatus anzeigen"},
	},

	// state-post.html
	"state-post-login": []TagStr{
		TagStr{language.AmericanEnglish, "Log in to view or edit the order:"},
		TagStr{language.German, "Logge dich ein, um den Auftrag anzusehen oder zu bearbeiten:"},
	},

	// client.html
	"client-login": []TagStr{
		TagStr{language.AmericanEnglish, "Logged in as"},
		TagStr{language.German, "Angemeldet als"},
	},
	"logout": []TagStr{
		TagStr{language.AmericanEnglish, "Logout"},
		TagStr{language.German, "Abmelden"},
	},

	// store.html
	"store-overview": []TagStr{
		TagStr{language.AmericanEnglish, "What's new?"},
		TagStr{language.German, "Übersicht"},
	},

	// layout.html
	"layout-numbers": []TagStr{
		TagStr{language.AmericanEnglish, "Please type the numbers to prove you mean it:"},
		TagStr{language.German, "Bitte tippe die Ziffern ab, um zu beweisen, dass du es ernst meinst:"},
	},
	"log-date": []TagStr{
		TagStr{language.AmericanEnglish, "Date"},
		TagStr{language.German, "Datum"},
	},
	"log-remark": []TagStr{
		TagStr{language.AmericanEnglish, "Remark"},
		TagStr{language.German, "Vermerk"},
	},
	"log-paid": []TagStr{
		TagStr{language.AmericanEnglish, "Amount paid"},
		TagStr{language.German, "Bezahlter Betrag"},
	},
	"log-sum": []TagStr{
		TagStr{language.AmericanEnglish, "Sum"},
		TagStr{language.German, "Summe"},
	},
	"layout-client-name": []TagStr{
		TagStr{language.AmericanEnglish, "The name under which we represent you. This is the name we will give to the actual merchant or provider. Keep in mind that involved banks and payment services may also know the name."},
		TagStr{language.German, "Der Name, unter dem wir dich vertreten. Diesen Namen werden wir bei dem tatsächlichen Versandhändler bzw. Anbieter angeben. Bedenke dass möglicherweise auch beteiligte Banken und Bezahldienste den Namen erfahren."},
	},
	"layout-client-name-help": []TagStr{
		TagStr{language.AmericanEnglish, "Keep in mind that a false statement may lead to problems in case of warranty or guarantee claims."},
		TagStr{language.German, "Bedenke dass eine falsche Angabe im Gewährleistungs- oder Garantiefall unter Umständen zu Problemen führen kann."},
	},
	"layout-client-contact": []TagStr{
		TagStr{language.AmericanEnglish, "Contact option for possible further inquiries (optional, will not be passed on, will be deleted 14 days after completion of the order)"},
		TagStr{language.German, "Kontaktmöglichkeit für mögliche Rückfragen (freiwillig, wird nicht weitergegeben und 14 Tage nach Abschluss des Auftrags gelöscht)"},
	},
	"layout-client-contact-help": []TagStr{
		TagStr{language.AmericanEnglish, "please do not mention the order number when contacting us"},
		TagStr{language.German, "bei Kontaktaufnahme bitte keine Auftragsnummer nennen"},
	},
	"layout-client-age": []TagStr{
		TagStr{language.AmericanEnglish, "If proof of age is required, it can be provided in person at our store. Goods requiring other proof cannot be ordered."},
		TagStr{language.German, "Falls ein Altersnachweis nötig ist, kann er persönlich in unserem Ladenlokal erbracht werden. Anderweitig nachweispflichtige Waren können nicht bestellt werden."},
	},
	"layout-add": []TagStr{
		TagStr{language.AmericanEnglish, "Add another order"},
		TagStr{language.German, "Weiteren Auftrag hinzufügen"},
	},
	"layout-delivery": []TagStr{
		TagStr{language.AmericanEnglish, "Pickup or forwarding"},
		TagStr{language.German, "Abholung oder Weiterversand"},
	},
	"layout-delivery-text": []TagStr{
		TagStr{language.AmericanEnglish, "You can pick up the goods in our store (Bernhard-Goering-Strasse 162, 04277 Leipzig, note the opening hours). We can also forward them to you by mail or parcel service."},
		TagStr{language.German, "Du kannst die Ware im Ladenlokal (Bernhard-Göring-Straße 162, 04277 Leipzig, beachte die Öffnungszeiten) abholen. Wir können sie auch per Post oder Paketdienst an dich weiterleiten."},
	},
	"layout-pickup": []TagStr{
		TagStr{language.AmericanEnglish, "I pick up the goods in the store, <strong>mentioning the order ID</strong>."},
		TagStr{language.German, "Ich hole die Ware <strong>unter Nennung der Auftragsnummer</strong> im Ladenlokal ab."},
	},
	"layout-locker": []TagStr{
		TagStr{language.AmericanEnglish, "I pick up the goods <strong>from a locker</strong> in the store. I'm bringing you an opened padlock before."},
		TagStr{language.German, "Ich hole die Ware <strong>aus einem Schließfach</strong> im Ladenlokal ab. Ein geöffnetes Vorhängeschloss bringe ich vorher vorbei."},
	},
	"layout-stamp": []TagStr{
		TagStr{language.AmericanEnglish, "Please send me the goods. I will bring a paid and labeled parcel stamp."},
		TagStr{language.German, "Bitte schickt mir die Ware. Eine bezahlte und beschriftete Paketmarke bringe ich vorbei."},
	},
	"layout-shipping": []TagStr{
		TagStr{language.AmericanEnglish, "Please send me the goods to the following address."},
		TagStr{language.German, "Bitte schickt mir die Ware an folgende Adresse."},
	},
	"layout-name1": []TagStr{
		TagStr{language.AmericanEnglish, "First name"},
		TagStr{language.German, "Vorname"},
	},
	"layout-name2": []TagStr{
		TagStr{language.AmericanEnglish, "Last name"},
		TagStr{language.German, "Nachname"},
	},
	"layout-address-supplement": []TagStr{
		TagStr{language.AmericanEnglish, "Address supplement"},
		TagStr{language.German, "Adresszusatz"},
	},
	"layout-street": []TagStr{
		TagStr{language.AmericanEnglish, "Street"},
		TagStr{language.German, "Straße"},
	},
	"layout-street-number": []TagStr{
		TagStr{language.AmericanEnglish, "House number"},
		TagStr{language.German, "Hausnummer"},
	},
	"layout-postcode": []TagStr{
		TagStr{language.AmericanEnglish, "Postal Code"},
		TagStr{language.German, "Postleitzahl"},
	},
	"layout-town": []TagStr{
		TagStr{language.AmericanEnglish, "Town"},
		TagStr{language.German, "Stadt"},
	},
	"layout-parcel": []TagStr{
		TagStr{language.AmericanEnglish, `You can find out how to enter the address of a DHL Packstation at <a href="https://dont.re/?https://parcelshopfinder.dhlparcel.com/html/note_direct_access_germany.html?setLng=en" target="_blank">dhlparcel.com</a>. You will need a DHL customer account with a PostNumber.`},
		TagStr{language.German, `Wie du die Adresse einer DHL-Packstation eingeben kannst, erfährst du auf <a href="https://dont.re/?https://parcelshopfinder.dhlparcel.com/html/note_direct_access_germany.html?setLng=de" target="_blank">dhlparcel.com</a>. Dazu brauchst du ein DHL-Kundenkonto mit Postnummer.`},
	},
	"layout-reshipment": []TagStr{
		TagStr{language.AmericanEnglish, "Cost of reshipment (if zero, the minimum value is taken)."},
		TagStr{language.German, "Kosten für Weiterversand (bei null wird der Mindestwert genommen)."},
	},
	"layout-summary": []TagStr{
		TagStr{language.AmericanEnglish, "Summary"},
		TagStr{language.German, "Zusammenfassung"},
	},
	"layout-name": []TagStr{
		TagStr{language.AmericanEnglish, "Name"},
		TagStr{language.German, "Bezeichnung"},
	},
	"layout-value": []TagStr{
		TagStr{language.AmericanEnglish, "Order value"},
		TagStr{language.German, "Bestellwert"},
	},
	"layout-fee": []TagStr{
		TagStr{language.AmericanEnglish, "Order fee"},
		TagStr{language.German, "Auftragsgebühr"},
	},
	"layout-sum": []TagStr{
		TagStr{language.AmericanEnglish, "= Sum"},
		TagStr{language.German, "= Summe"},
	},
	"layout-status-arrived": []TagStr{
		TagStr{language.AmericanEnglish, "Has arrived"},
		TagStr{language.German, "Ist angekommen"},
	},
	"layout-status-ordered": []TagStr{
		TagStr{language.AmericanEnglish, "Mark as ordered"},
		TagStr{language.German, "Als bestellt markieren"},
	},
	"layout-status-pickup": []TagStr{
		TagStr{language.AmericanEnglish, "Mark as picked up"},
		TagStr{language.German, "Als abgeholt markieren"},
	},
	"layout-status-reshipped": []TagStr{
		TagStr{language.AmericanEnglish, "Mark as reshipped"},
		TagStr{language.German, "Als weiterverschickt markieren"},
	},
	"layout-status-failed": []TagStr{
		TagStr{language.AmericanEnglish, "Is failed"},
		TagStr{language.German, "Ist gescheitert"},
	},

	// store/collection-accept.html
	"store-accept": []TagStr{
		TagStr{language.AmericanEnglish, "Accept order"},
		TagStr{language.German, "Auftrag akzeptieren"},
	}, // /store/collection-view.html
	"store-msg": []TagStr{
		TagStr{language.AmericanEnglish, "Message"},
		TagStr{language.German, "Nachricht"},
	}, // /store/collection-return.html, /store/collection-reject.html, /store/task-mark-failed.html
	"store-accept-text": []TagStr{
		TagStr{language.AmericanEnglish, "The order has been accepted."},
		TagStr{language.German, "Der Auftrag wurde akzeptiert."},
	},
	// store/collection-confirm-payment.html
	"store-payment": []TagStr{
		TagStr{language.AmericanEnglish, "Confirm payment"},
		TagStr{language.German, "Zahlung bestätigen"},
	},
	"store-paid-amount": []TagStr{
		TagStr{language.AmericanEnglish, "Amount received (positive) or spent (negative)"},
		TagStr{language.German, "Eingenommener (positiver) oder ausgegebener (negativer) Betrag"},
	},
	"store-value": []TagStr{
		TagStr{language.AmericanEnglish, "Please enter a value."},
		TagStr{language.German, "Bitte gib einen Wert ein."},
	},
	"store-payment-message": []TagStr{
		TagStr{language.AmericanEnglish, "Message. Please specify the payment method here, e.g. cash or bank transfer."},
		TagStr{language.German, "Nachricht. Bitte gib hier die Zahlungsmethode an, z. B. bar oder Überweisung."},
	},

	// store/collection-confirm-pickup.html
	"store-pickup": []TagStr{
		TagStr{language.AmericanEnglish, "Confirm pickup"},
		TagStr{language.German, "Abholung bestätigen"},
	},
	"store-pickup-text": []TagStr{
		TagStr{language.AmericanEnglish, "If you continue, these individual orders will be marked as picked up:"},
		TagStr{language.German, "Wenn du fortfährst, werden diese Einzelaufträge als abgeholt markiert:"},
	},

	// store/collection-confirm-reshipped.html
	"store-reshipment": []TagStr{
		TagStr{language.AmericanEnglish, "Confirm reshipment"},
		TagStr{language.German, "Weiterversand bestätigen"},
	},
	"store-reshipment-text": []TagStr{
		TagStr{language.AmericanEnglish, "If you continue, these individual orders will be marked as reshipped:"},
		TagStr{language.German, "Wenn du fortfährst, werden diese Einzelaufträge als weiterverschickt markiert:"},
	},

	// store/collection-delete.html
	"store-delete-title": []TagStr{
		TagStr{language.AmericanEnglish, "Delete already accepted order"},
		TagStr{language.German, "Bereits akzeptierten Auftrag löschen"},
	},
	"store-delete-confirm": []TagStr{
		TagStr{language.AmericanEnglish, "Please confirm that you want to delete the already accepted order."},
		TagStr{language.German, "Bitte bestätige, dass den bereits akzeptierten Auftrag löschen möchtest."},
	},
	"store-delete": []TagStr{
		TagStr{language.AmericanEnglish, "Delete order"},
		TagStr{language.German, "Auftrag löschen"},
	},

	// store/collection-edit.html
	"store-edit": []TagStr{
		TagStr{language.AmericanEnglish, "Edit order"},
		TagStr{language.German, "Auftrag bearbeiten"},
	},
	"store-save": []TagStr{
		TagStr{language.AmericanEnglish, "Save"},
		TagStr{language.German, "Speichern"},
	},

	// store/collection-mark-spam.html
	"store-spam": []TagStr{
		TagStr{language.AmericanEnglish, "Mark order as spam"},
		TagStr{language.German, "Auftrag als Spam markieren"},
	}, // /store/collection-view.html
	"store-spam-text": []TagStr{
		TagStr{language.AmericanEnglish, "Please confirm that you want to mark the order as spam."},
		TagStr{language.German, "Bitte bestätige, dass du den Auftrag als Spam markieren möchtest."},
	},

	// store/collection-price-rised.html
	"store-rised": []TagStr{
		TagStr{language.AmericanEnglish, "Price has risen"},
		TagStr{language.German, "Preis ist gestiegen"},
	},
	"store-rised-text": []TagStr{
		TagStr{language.AmericanEnglish, "The price of your order has risen."},
		TagStr{language.German, "Der Preis deines Auftrags hat sich erhöht."},
	},

	// store/collection-reject.html
	"store-reject": []TagStr{
		TagStr{language.AmericanEnglish, "Reject order"},
		TagStr{language.German, "Auftrag ablehnen"},
	}, // /store/collection-view.html
	"store-reject-confirm": []TagStr{
		TagStr{language.AmericanEnglish, "Please confirm that you wish to reject the order."},
		TagStr{language.German, "Bitte bestätige, dass du den Auftrag ablehnen möchtest."},
	},
	"store-reject-msg": []TagStr{
		TagStr{language.AmericanEnglish, "Unfortunately, your order had to be rejected."},
		TagStr{language.German, "Dein Auftrag musste leider abgelehnt werden."},
	},

	// store/collection-return.html
	"store-return": []TagStr{
		TagStr{language.AmericanEnglish, "Return order for revision"},
		TagStr{language.German, "Auftrag zur Überarbeitung zurückgeben"},
	}, // /store/collection-view.html
	"store-return-text": []TagStr{
		TagStr{language.AmericanEnglish, "Your order requires revision."},
		TagStr{language.German, "Dein Auftrag erfordert Überarbeitung."},
	},

	// store/collection-view.html
	"store-view-amount": []TagStr{
		TagStr{language.AmericanEnglish, "Amount:"},
		TagStr{language.German, "Summe:"},
	},
	"store-view-payment": []TagStr{
		TagStr{language.AmericanEnglish, "Enter payment"},
		TagStr{language.German, "Zahlung vermerken"},
	},
	"store-view-pickup": []TagStr{
		TagStr{language.AmericanEnglish, "Has been picked up"},
		TagStr{language.German, "Wurde abgeholt"},
	},
	"store-view-reshipped": []TagStr{
		TagStr{language.AmericanEnglish, "Has been reshipped"},
		TagStr{language.German, "Wurde weiterverschickt"},
	},
	"store-view-rised": []TagStr{
		TagStr{language.AmericanEnglish, "Price has risen, additional payment required"},
		TagStr{language.German, "Preis ist gestiegen, Nachzahlung erforderlich"},
	},
	"store-view-delete": []TagStr{
		TagStr{language.AmericanEnglish, "Delete"},
		TagStr{language.German, "Löschen"},
	},

	// store/index.html
	"store-submitted": []TagStr{
		TagStr{language.AmericanEnglish, "Submitted"},
		TagStr{language.German, "Eingereicht"},
	},
	"store-event": []TagStr{
		TagStr{language.AmericanEnglish, "Last event"},
		TagStr{language.German, "Letztes Event"},
	},
	"store-noorder": []TagStr{
		TagStr{language.AmericanEnglish, "No orders"},
		TagStr{language.German, "Keine Bestellungen"},
	},
	"store-accepted": []TagStr{
		TagStr{language.AmericanEnglish, "Accepted"},
		TagStr{language.German, "Angenommen"},
	},
	"store-underpaid": []TagStr{
		TagStr{language.AmericanEnglish, "Underpaid"},
		TagStr{language.German, "Unterbezahlt"},
	},
	"store-paid": []TagStr{
		TagStr{language.AmericanEnglish, "Paid"},
		TagStr{language.German, "Bezahlt"},
	}, // /store/collection-confirm-payment.html,
	"store-iorders": []TagStr{
		TagStr{language.AmericanEnglish, "Individual orders"},
		TagStr{language.German, "Einzelaufträge"},
	},
	"store-open-order": []TagStr{
		TagStr{language.AmericanEnglish, "Not ordered yet thereof"},
		TagStr{language.German, "Davon noch nicht bestellt"},
	},
	"store-completed": []TagStr{
		TagStr{language.AmericanEnglish, "Completed"},
		TagStr{language.German, "Abgeschlossen"},
	},

	// store/login.html
	"login-user": []TagStr{
		TagStr{language.AmericanEnglish, "Username"},
		TagStr{language.German, "Nutzername"},
	},
	"login-pass": []TagStr{
		TagStr{language.AmericanEnglish, "Password"},
		TagStr{language.German, "Passwort"},
	},

	// store/task-confirm-arrived.html
	"store-arrived": []TagStr{
		TagStr{language.AmericanEnglish, "Mark individual order as arrived"},
		TagStr{language.German, "Einzelauftrag als eingetroffen markieren"},
	},
	"store-arrived-button": []TagStr{
		TagStr{language.AmericanEnglish, "Goods have arrived"},
		TagStr{language.German, "Ware ist eingetroffen"},
	},

	// store/task-confirm-ordered.html
	"store-ordered": []TagStr{
		TagStr{language.AmericanEnglish, "Mark individual order as executed"},
		TagStr{language.German, "Einzelauftrag als ausgeführt markieren"},
	},

	// store/task-confirm-pickup.html
	"store-mark-pickup": []TagStr{
		TagStr{language.AmericanEnglish, "Mark individual order as picked up"},
		TagStr{language.German, "Einzelauftrag als abgeholt markieren"},
	},
	"store-confirm-pickup": []TagStr{
		TagStr{language.AmericanEnglish, "Individual order was picked up"},
		TagStr{language.German, "Einzelauftrag wurde abgeholt"},
	},

	// store/task-confirm-reshipped.html
	"store-mark-reshipped": []TagStr{
		TagStr{language.AmericanEnglish, "Mark individual order as reshipped"},
		TagStr{language.German, "Einzelauftrag als weiterverschickt markieren"},
	},
	"store-confirm-reshipped": []TagStr{
		TagStr{language.AmericanEnglish, "Individual order was reshipped"},
		TagStr{language.German, "Einzelauftrag wurde weiterverschickt"},
	},

	// store/task-mark-failed.html
	"store-mark-failed": []TagStr{
		TagStr{language.AmericanEnglish, "Mark individual order as failed"},
		TagStr{language.German, "Einzelauftrag als gescheitert markieren"},
	},
	"store-failed": []TagStr{
		TagStr{language.AmericanEnglish, `Unfortunately, the individual order at "%s" failed.`},
		TagStr{language.German, `Die Einzelbestellung bei "%s" ist leider gescheitert.`},
	},
}

// Language is any string. It will be matched by golang.org/x/text/language.Make and golang.org/x/text/language.NewMatcher.
type Language string

func (lang Language) Translate(key string, args ...interface{}) template.HTML {
	item, ok := translations[key]
	if !ok {
		// no translation available, create language tag and print
		return template.HTML(message.NewPrinter(language.Make(string(lang))).Sprintf(key, args...))
	}
	// choose language tag from list of translations
	langs := make([]language.Tag, len(item))
	for i := range item {
		langs[i] = item[i].Tag
	}
	tag, i := language.MatchStrings(language.NewMatcher(langs), string(lang))
	return template.HTML(message.NewPrinter(tag).Sprintf(item[i].Str, args...))
}
