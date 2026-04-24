package entities

import (
	"fmt"
	"strings"
	"time"

	"finance_tracker/internal/storage/database/types"
)

type Currency string

const (
	CurrencyAED Currency = "AED"
	CurrencyAFN Currency = "AFN"
	CurrencyALL Currency = "ALL"
	CurrencyAMD Currency = "AMD"
	CurrencyANG Currency = "ANG"
	CurrencyAOA Currency = "AOA"
	CurrencyARS Currency = "ARS"
	CurrencyAUD Currency = "AUD"
	CurrencyAWG Currency = "AWG"
	CurrencyAZN Currency = "AZN"
	CurrencyBAM Currency = "BAM"
	CurrencyBBD Currency = "BBD"
	CurrencyBDT Currency = "BDT"
	CurrencyBGN Currency = "BGN"
	CurrencyBHD Currency = "BHD"
	CurrencyBIF Currency = "BIF"
	CurrencyBMD Currency = "BMD"
	CurrencyBND Currency = "BND"
	CurrencyBOB Currency = "BOB"
	CurrencyBRL Currency = "BRL"
	CurrencyBSD Currency = "BSD"
	CurrencyBTC Currency = "BTC"
	CurrencyBTN Currency = "BTN"
	CurrencyBWP Currency = "BWP"
	CurrencyBYN Currency = "BYN"
	CurrencyBZD Currency = "BZD"
	CurrencyCAD Currency = "CAD"
	CurrencyCDF Currency = "CDF"
	CurrencyCHF Currency = "CHF"
	CurrencyCLF Currency = "CLF"
	CurrencyCLP Currency = "CLP"
	CurrencyCNH Currency = "CNH"
	CurrencyCNY Currency = "CNY"
	CurrencyCOP Currency = "COP"
	CurrencyCRC Currency = "CRC"
	CurrencyCUC Currency = "CUC"
	CurrencyCUP Currency = "CUP"
	CurrencyCVE Currency = "CVE"
	CurrencyCZK Currency = "CZK"
	CurrencyDJF Currency = "DJF"
	CurrencyDKK Currency = "DKK"
	CurrencyDOP Currency = "DOP"
	CurrencyDZD Currency = "DZD"
	CurrencyEGP Currency = "EGP"
	CurrencyERN Currency = "ERN"
	CurrencyETB Currency = "ETB"
	CurrencyEUR Currency = "EUR"
	CurrencyFJD Currency = "FJD"
	CurrencyFKP Currency = "FKP"
	CurrencyGBP Currency = "GBP"
	CurrencyGEL Currency = "GEL"
	CurrencyGGP Currency = "GGP"
	CurrencyGHS Currency = "GHS"
	CurrencyGIP Currency = "GIP"
	CurrencyGMD Currency = "GMD"
	CurrencyGNF Currency = "GNF"
	CurrencyGTQ Currency = "GTQ"
	CurrencyGYD Currency = "GYD"
	CurrencyHKD Currency = "HKD"
	CurrencyHNL Currency = "HNL"
	CurrencyHRK Currency = "HRK"
	CurrencyHTG Currency = "HTG"
	CurrencyHUF Currency = "HUF"
	CurrencyIDR Currency = "IDR"
	CurrencyILS Currency = "ILS"
	CurrencyIMP Currency = "IMP"
	CurrencyINR Currency = "INR"
	CurrencyIQD Currency = "IQD"
	CurrencyIRR Currency = "IRR"
	CurrencyISK Currency = "ISK"
	CurrencyJEP Currency = "JEP"
	CurrencyJMD Currency = "JMD"
	CurrencyJOD Currency = "JOD"
	CurrencyJPY Currency = "JPY"
	CurrencyKES Currency = "KES"
	CurrencyKGS Currency = "KGS"
	CurrencyKHR Currency = "KHR"
	CurrencyKMF Currency = "KMF"
	CurrencyKPW Currency = "KPW"
	CurrencyKRW Currency = "KRW"
	CurrencyKWD Currency = "KWD"
	CurrencyKYD Currency = "KYD"
	CurrencyKZT Currency = "KZT"
	CurrencyLAK Currency = "LAK"
	CurrencyLBP Currency = "LBP"
	CurrencyLKR Currency = "LKR"
	CurrencyLRD Currency = "LRD"
	CurrencyLSL Currency = "LSL"
	CurrencyLYD Currency = "LYD"
	CurrencyMAD Currency = "MAD"
	CurrencyMDL Currency = "MDL"
	CurrencyMGA Currency = "MGA"
	CurrencyMKD Currency = "MKD"
	CurrencyMMK Currency = "MMK"
	CurrencyMNT Currency = "MNT"
	CurrencyMOP Currency = "MOP"
	CurrencyMRU Currency = "MRU"
	CurrencyMUR Currency = "MUR"
	CurrencyMVR Currency = "MVR"
	CurrencyMWK Currency = "MWK"
	CurrencyMXN Currency = "MXN"
	CurrencyMYR Currency = "MYR"
	CurrencyMZN Currency = "MZN"
	CurrencyNAD Currency = "NAD"
	CurrencyNGN Currency = "NGN"
	CurrencyNIO Currency = "NIO"
	CurrencyNOK Currency = "NOK"
	CurrencyNPR Currency = "NPR"
	CurrencyNZD Currency = "NZD"
	CurrencyOMR Currency = "OMR"
	CurrencyPAB Currency = "PAB"
	CurrencyPEN Currency = "PEN"
	CurrencyPGK Currency = "PGK"
	CurrencyPHP Currency = "PHP"
	CurrencyPKR Currency = "PKR"
	CurrencyPLN Currency = "PLN"
	CurrencyPYG Currency = "PYG"
	CurrencyQAR Currency = "QAR"
	CurrencyRON Currency = "RON"
	CurrencyRSD Currency = "RSD"
	CurrencyRUB Currency = "RUB"
	CurrencyRWF Currency = "RWF"
	CurrencySAR Currency = "SAR"
	CurrencySBD Currency = "SBD"
	CurrencySCR Currency = "SCR"
	CurrencySDG Currency = "SDG"
	CurrencySEK Currency = "SEK"
	CurrencySGD Currency = "SGD"
	CurrencySHP Currency = "SHP"
	CurrencySLE Currency = "SLE"
	CurrencySLL Currency = "SLL"
	CurrencySOS Currency = "SOS"
	CurrencySRD Currency = "SRD"
	CurrencySSP Currency = "SSP"
	CurrencySTD Currency = "STD"
	CurrencySTN Currency = "STN"
	CurrencySVC Currency = "SVC"
	CurrencySYP Currency = "SYP"
	CurrencySZL Currency = "SZL"
	CurrencyTHB Currency = "THB"
	CurrencyTJS Currency = "TJS"
	CurrencyTMT Currency = "TMT"
	CurrencyTND Currency = "TND"
	CurrencyTOP Currency = "TOP"
	CurrencyTRY Currency = "TRY"
	CurrencyTTD Currency = "TTD"
	CurrencyTWD Currency = "TWD"
	CurrencyTZS Currency = "TZS"
	CurrencyUAH Currency = "UAH"
	CurrencyUGX Currency = "UGX"
	CurrencyUSD Currency = "USD"
	CurrencyUYU Currency = "UYU"
	CurrencyUZS Currency = "UZS"
	CurrencyVES Currency = "VES"
	CurrencyVND Currency = "VND"
	CurrencyVUV Currency = "VUV"
	CurrencyWST Currency = "WST"
	CurrencyXAF Currency = "XAF"
	CurrencyXAG Currency = "XAG"
	CurrencyXAU Currency = "XAU"
	CurrencyXCD Currency = "XCD"
	CurrencyXCG Currency = "XCG"
	CurrencyXDR Currency = "XDR"
	CurrencyXOF Currency = "XOF"
	CurrencyXPD Currency = "XPD"
	CurrencyXPF Currency = "XPF"
	CurrencyXPT Currency = "XPT"
	CurrencyYER Currency = "YER"
	CurrencyZAR Currency = "ZAR"
	CurrencyZMW Currency = "ZMW"
	CurrencyZWG Currency = "ZWG"
	CurrencyZWL Currency = "ZWL"
)

var knownCurrency = map[string]Currency{
	"AED": CurrencyAED,
	"AFN": CurrencyAFN,
	"ALL": CurrencyALL,
	"AMD": CurrencyAMD,
	"ANG": CurrencyANG,
	"AOA": CurrencyAOA,
	"ARS": CurrencyARS,
	"AUD": CurrencyAUD,
	"AWG": CurrencyAWG,
	"AZN": CurrencyAZN,
	"BAM": CurrencyBAM,
	"BBD": CurrencyBBD,
	"BDT": CurrencyBDT,
	"BGN": CurrencyBGN,
	"BHD": CurrencyBHD,
	"BIF": CurrencyBIF,
	"BMD": CurrencyBMD,
	"BND": CurrencyBND,
	"BOB": CurrencyBOB,
	"BRL": CurrencyBRL,
	"BSD": CurrencyBSD,
	"BTC": CurrencyBTC,
	"BTN": CurrencyBTN,
	"BWP": CurrencyBWP,
	"BYN": CurrencyBYN,
	"BZD": CurrencyBZD,
	"CAD": CurrencyCAD,
	"CDF": CurrencyCDF,
	"CHF": CurrencyCHF,
	"CLF": CurrencyCLF,
	"CLP": CurrencyCLP,
	"CNH": CurrencyCNH,
	"CNY": CurrencyCNY,
	"COP": CurrencyCOP,
	"CRC": CurrencyCRC,
	"CUC": CurrencyCUC,
	"CUP": CurrencyCUP,
	"CVE": CurrencyCVE,
	"CZK": CurrencyCZK,
	"DJF": CurrencyDJF,
	"DKK": CurrencyDKK,
	"DOP": CurrencyDOP,
	"DZD": CurrencyDZD,
	"EGP": CurrencyEGP,
	"ERN": CurrencyERN,
	"ETB": CurrencyETB,
	"EUR": CurrencyEUR,
	"FJD": CurrencyFJD,
	"FKP": CurrencyFKP,
	"GBP": CurrencyGBP,
	"GEL": CurrencyGEL,
	"GGP": CurrencyGGP,
	"GHS": CurrencyGHS,
	"GIP": CurrencyGIP,
	"GMD": CurrencyGMD,
	"GNF": CurrencyGNF,
	"GTQ": CurrencyGTQ,
	"GYD": CurrencyGYD,
	"HKD": CurrencyHKD,
	"HNL": CurrencyHNL,
	"HRK": CurrencyHRK,
	"HTG": CurrencyHTG,
	"HUF": CurrencyHUF,
	"IDR": CurrencyIDR,
	"ILS": CurrencyILS,
	"IMP": CurrencyIMP,
	"INR": CurrencyINR,
	"IQD": CurrencyIQD,
	"IRR": CurrencyIRR,
	"ISK": CurrencyISK,
	"JEP": CurrencyJEP,
	"JMD": CurrencyJMD,
	"JOD": CurrencyJOD,
	"JPY": CurrencyJPY,
	"KES": CurrencyKES,
	"KGS": CurrencyKGS,
	"KHR": CurrencyKHR,
	"KMF": CurrencyKMF,
	"KPW": CurrencyKPW,
	"KRW": CurrencyKRW,
	"KWD": CurrencyKWD,
	"KYD": CurrencyKYD,
	"KZT": CurrencyKZT,
	"LAK": CurrencyLAK,
	"LBP": CurrencyLBP,
	"LKR": CurrencyLKR,
	"LRD": CurrencyLRD,
	"LSL": CurrencyLSL,
	"LYD": CurrencyLYD,
	"MAD": CurrencyMAD,
	"MDL": CurrencyMDL,
	"MGA": CurrencyMGA,
	"MKD": CurrencyMKD,
	"MMK": CurrencyMMK,
	"MNT": CurrencyMNT,
	"MOP": CurrencyMOP,
	"MRU": CurrencyMRU,
	"MUR": CurrencyMUR,
	"MVR": CurrencyMVR,
	"MWK": CurrencyMWK,
	"MXN": CurrencyMXN,
	"MYR": CurrencyMYR,
	"MZN": CurrencyMZN,
	"NAD": CurrencyNAD,
	"NGN": CurrencyNGN,
	"NIO": CurrencyNIO,
	"NOK": CurrencyNOK,
	"NPR": CurrencyNPR,
	"NZD": CurrencyNZD,
	"OMR": CurrencyOMR,
	"PAB": CurrencyPAB,
	"PEN": CurrencyPEN,
	"PGK": CurrencyPGK,
	"PHP": CurrencyPHP,
	"PKR": CurrencyPKR,
	"PLN": CurrencyPLN,
	"PYG": CurrencyPYG,
	"QAR": CurrencyQAR,
	"RON": CurrencyRON,
	"RSD": CurrencyRSD,
	"RUB": CurrencyRUB,
	"RWF": CurrencyRWF,
	"SAR": CurrencySAR,
	"SBD": CurrencySBD,
	"SCR": CurrencySCR,
	"SDG": CurrencySDG,
	"SEK": CurrencySEK,
	"SGD": CurrencySGD,
	"SHP": CurrencySHP,
	"SLE": CurrencySLE,
	"SLL": CurrencySLL,
	"SOS": CurrencySOS,
	"SRD": CurrencySRD,
	"SSP": CurrencySSP,
	"STD": CurrencySTD,
	"STN": CurrencySTN,
	"SVC": CurrencySVC,
	"SYP": CurrencySYP,
	"SZL": CurrencySZL,
	"THB": CurrencyTHB,
	"TJS": CurrencyTJS,
	"TMT": CurrencyTMT,
	"TND": CurrencyTND,
	"TOP": CurrencyTOP,
	"TRY": CurrencyTRY,
	"TTD": CurrencyTTD,
	"TWD": CurrencyTWD,
	"TZS": CurrencyTZS,
	"UAH": CurrencyUAH,
	"UGX": CurrencyUGX,
	"USD": CurrencyUSD,
	"UYU": CurrencyUYU,
	"UZS": CurrencyUZS,
	"VES": CurrencyVES,
	"VND": CurrencyVND,
	"VUV": CurrencyVUV,
	"WST": CurrencyWST,
	"XAF": CurrencyXAF,
	"XAG": CurrencyXAG,
	"XAU": CurrencyXAU,
	"XCD": CurrencyXCD,
	"XCG": CurrencyXCG,
	"XDR": CurrencyXDR,
	"XOF": CurrencyXOF,
	"XPD": CurrencyXPD,
	"XPF": CurrencyXPF,
	"XPT": CurrencyXPT,
	"YER": CurrencyYER,
	"ZAR": CurrencyZAR,
	"ZMW": CurrencyZMW,
	"ZWG": CurrencyZWG,
	"ZWL": CurrencyZWL,
}

func (c Currency) String() string {
	return string(c)
}

func CurrencyFromString(s string) (Currency, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if c, ok := knownCurrency[s]; ok {
		return c, nil
	}
	return "", fmt.Errorf("currencies: unknown currency: %s", s)
}

type CurrencyRate struct {
	RateAgainstUSD map[Currency]int64
}

type CurrencyDB struct {
	Timestamp    time.Time      `db:"timestamp" json:"timestamp"`
	TimestampNum int64          `db:"timestamp_num" json:"timestamp_num"`
	CurrencyData types.Int64Map `db:"rates" json:"rates"`
}
