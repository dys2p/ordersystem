// Code generated by running "go generate" in golang.org/x/text. DO NOT EDIT.

package ordersystem

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

type dictionary struct {
	index []uint32
	data  string
}

func (d *dictionary) Lookup(key string) (data string, ok bool) {
	p, ok := messageKeyToIndex[key]
	if !ok {
		return "", false
	}
	start, end := d.index[p], d.index[p+1]
	if start == end {
		return "", false
	}
	return d.data[start:end], true
}

func init() {
	dict := map[string]catalog.Dictionary{
		"de_DE": &dictionary{index: de_DEIndex, data: de_DEData},
	}
	fallback := language.MustParse("en-US")
	cat, err := catalog.NewFromMap(dict, catalog.Fallback(fallback))
	if err != nil {
		panic(err)
	}
	message.DefaultCatalog = cat
}

var messageKeyToIndex = map[string]int{
	"All Services and Projects":       39,
	"Australian dollars":              4,
	"Battery Disposal":                33,
	"Bulgarian lev":                   5,
	"Canadian dollars":                6,
	"Cancellation Policy":             32,
	"Cash":                            2,
	"Cash by mail":                    36,
	"Cash in Foreign Currency":        1,
	"Cash payment in local store":     35,
	"Chinese renminbi":                8,
	"Contact & News":                  46,
	"Contact us":                      47,
	"Czech koruna":                    9,
	"DHL parcel, franked digitally":   27,
	"DHL parcel, franked handwritten": 26,
	"Danish krone":                    10,
	"Delivery":                        23,
	"Digital Goods":                   42,
	"Germany":                         48,
	"Got an idea or found an error? Drop us a note!": 52,
	"Icelandic króna":                       12,
	"Imprint":                               31,
	"Japanese yen":                          13,
	"Legal":                                 28,
	"Local Store":                           41,
	"Mon+Thu 2pm-6pm":                       50,
	"Monero and Bitcoin":                    37,
	"Monero or Bitcoin":                     0,
	"New Israeli shekel (NIS)":              14,
	"New Taiwan dollars":                    21,
	"New Zealand dollars":                   16,
	"New opening hours!":                    49,
	"Norwegian krone":                       15,
	"Online printing":                       45,
	"Online shop":                           43,
	"Order Service":                         44,
	"Payment":                               34,
	"Pickup from locker in our local store": 25,
	"Pickup in our local store":             24,
	"Polish złoty":                          17,
	"Pound sterling":                        11,
	"Privacy policy":                        30,
	"Romanian leu":                          18,
	"SEPA Bank Transfer":                    3,
	"SEPA bank transfer":                    38,
	"Serbian dinar":                         19,
	"Swedish krona":                         20,
	"Swiss francs":                          7,
	"Terms and Conditions":                  29,
	"Tue+Wed+Fri+Sat 10am-2pm":              51,
	"United States dollars":                 22,
	"Why?":                                  40,
}

var de_DEIndex = []uint32{ // 54 elements
	// Entry 0 - 1F
	0x00000000, 0x00000014, 0x0000002d, 0x00000035,
	0x0000004b, 0x0000005f, 0x0000006f, 0x00000081,
	0x00000093, 0x000000a8, 0x000000bc, 0x000000cd,
	0x000000dd, 0x000000f1, 0x00000100, 0x0000011f,
	0x00000132, 0x00000144, 0x00000155, 0x00000165,
	0x00000175, 0x00000188, 0x00000196, 0x000001a0,
	0x000001a8, 0x000001c3, 0x000001ef, 0x00000213,
	0x0000022f, 0x0000023b, 0x0000023f, 0x0000024b,
	// Entry 20 - 3F
	0x00000255, 0x00000268, 0x00000288, 0x00000292,
	0x000002af, 0x000002c0, 0x000002d3, 0x000002e5,
	0x00000300, 0x00000307, 0x00000316, 0x00000326,
	0x00000331, 0x00000340, 0x00000350, 0x0000035f,
	0x00000367, 0x00000373, 0x00000389, 0x00000399,
	0x000003af, 0x000003d2,
} // Size: 240 bytes

const de_DEData string = "" + // Size: 978 bytes
	"\x02Monero oder Bitcoin\x02Bargeld in Fremdwährung\x02Bargeld\x02SEPA-Ba" +
	"nküberweisung\x02Australische Dollar\x02Bulgarische Lew\x02Kanadische Do" +
	"llar\x02Schweizer Franken\x02Chinesische Renminbi\x02Tschechische Kronen" +
	"\x02Dänische Kronen\x02Britische Pfund\x02Isländische Kronen\x02Japanisc" +
	"he Yen\x02Neue israelische Schekel (NIS)\x02Norwegische Kronen\x02Neusee" +
	"land-Dollar\x02Polnische Złoty\x02Rumänische Leu\x02Serbische Dinar\x02S" +
	"chwedische Kronen\x02Taiwan-Dollar\x02US-Dollar\x02Versand\x02Abholung i" +
	"m Ladengeschäft\x02Abholung aus Schließfach im Ladengeschäft\x02DHL-Pake" +
	"t handschriftlich frankiert\x02DHL-Paket digital frankiert\x02Rechtliche" +
	"s\x02AGB\x02Datenschutz\x02Impressum\x02Widerrufsbelehrung\x02Hinweise z" +
	"ur Batterieentsorgung\x02Bezahlung\x02Barzahlung im Ladengeschäft\x02Bar" +
	"geld per Post\x02Monero und Bitcoin\x02SEPA-Überweisung\x02Alle Angebote" +
	" und Projekte\x02Warum?\x02Ladengeschäft\x02Digitale Güter\x02Onlineshop" +
	"\x02Bestellservice\x02Onlinedruckerei\x02Kontakt & News\x02Kontakt\x02De" +
	"utschland\x02Neue Öffnungszeiten!\x02Mo+Do 14-18 Uhr\x02Di+Mi+Fr+Sa 10-1" +
	"4 Uhr\x02Fehler oder Hinweise? Schreib uns!"

	// Total table size 1218 bytes (1KiB); checksum: E7146949
