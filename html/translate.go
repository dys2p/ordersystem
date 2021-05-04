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
	"welcome": []TagStr{
		TagStr{language.AmericanEnglish, "Welcome!"},
		TagStr{language.German, "Willkommen!"},
	},
	"coll-view-edit": []TagStr{
		TagStr{language.AmericanEnglish, "View or edit order"},
		TagStr{language.German, "Auftrag ansehen oder bearbeiten"},
	},
	"coll-display-state": []TagStr{
		TagStr{language.AmericanEnglish, "Display order state only"},
		TagStr{language.German, "Nur Auftragsstatus anzeigen"},
	},
	"cancel": []TagStr{
		TagStr{language.AmericanEnglish, "Cancel"},
		TagStr{language.German, "Abbrechen"},
	},
	"go-back": []TagStr{
		TagStr{language.AmericanEnglish, "Go back"},
		TagStr{language.German, "Zurück"},
	},
	"confirm-yes": []TagStr{
		TagStr{language.AmericanEnglish, "Yes, I am sure."},
		TagStr{language.German, "Ja, ich bin mir sicher."},
	},
	"coll-cancel": []TagStr{
		TagStr{language.AmericanEnglish, "Cancel order"},
		TagStr{language.German, "Auftrag abbrechen"},
	},
	"coll-cancel-invalid": []TagStr{
		TagStr{language.AmericanEnglish, "Please confirm that you would like to cancel the order."},
		TagStr{language.German, "Bitte bestätige, dass den Auftrag abbrechen möchtest."},
	},
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
	"coll-new-check-invalid": []TagStr{
		TagStr{language.AmericanEnglish, "Please confirm that you have written down, memorized or saved the order ID and passphrase."},
		TagStr{language.German, "Bitte bestätige, dass du dir die Auftragsnummer und die Passphrase aufgeschrieben, gemerkt oder gespeichert hast."},
	},
	"coll-delete": []TagStr{
		TagStr{language.AmericanEnglish, "Delete draft"},
		TagStr{language.German, "Auftragsentwurf löschen"},
	},
	"coll-delete-invalid": []TagStr{
		TagStr{language.AmericanEnglish, "Please confirm that you would like to delete the order."},
		TagStr{language.German, "Bitte bestätige, dass den Auftragsentwurf löschen möchtest."},
	},
	"": []TagStr{
		TagStr{language.AmericanEnglish, ""},
		TagStr{language.German, ""},
	},
	"": []TagStr{
		TagStr{language.AmericanEnglish, ""},
		TagStr{language.German, ""},
	},
	"": []TagStr{
		TagStr{language.AmericanEnglish, ""},
		TagStr{language.German, ""},
	},
	"": []TagStr{
		TagStr{language.AmericanEnglish, ""},
		TagStr{language.German, ""},
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
