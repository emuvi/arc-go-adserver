package motor

import (
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"golang.org/x/text/language"
)

type FormatType string

const (
	FormatString = "String"
	FormatDate = "Date"
	FormatCurrency = "Currency"
)

type Style struct {
	Type FormatType
	Model string
}

func (format *Style) Format(transit *Convey, value interface{}) string {
	switch format.Type {
	case FormatDate:
		return transit.FormatDate(value)
	case FormatCurrency:
		return transit.FormatCurrency(value)
	default:
		return transit.FormatString(value)
	}
}

var languageMatcher = language.NewMatcher([]language.Tag{
	language.Make("pt"),
	language.Make("en"),
})

func (transit *Convey) getLanguage() language.Tag {
	langFormString := transit.request.FormValue("adserver-lang")
	langForm := language.Make(langFormString)
	langCookieString := ""
	langCookieData, err := transit.request.Cookie("adserver-lang")
	if err == nil {
		langCookieString = langCookieData.String()
	}
	langCookie := language.Make(langCookieString)
	langHeaderString := transit.request.Header.Get("Accept-Language")
	langHeader := language.Make(langHeaderString)
	result, _, _ := languageMatcher.Match(langForm, langCookie, langHeader)
	return result
}

func (transit *Convey) GetDateFormat() string {
	if transit.getLanguage() == language.Portuguese {
		return "01/02/2006"
	}
	return "02/01/2006"
}

func (transit *Convey) GetDateActual() string {
	return time.Now().Format(transit.GetDateFormat())
}

func (transit *Convey) FormatString(value interface{}) string {
	return fmt.Sprintf("%s", value)
}

func (transit *Convey) PutFormatStringAs(column, as string) bool {
	value, err := transit.Take(column)
	if err != nil {
		transit.PutError(err)
		transit.PutError("can't put the formatted of", column, "as", as)
		return false
	}
	transit.Set(as, transit.FormatString(value))
	return true
}

func (transit *Convey) PutFormatString(column string) bool {
	return transit.PutFormatStringAs(column, column)
}

func (transit *Convey) FormatDate(date interface{}) string {
	if value, ok := date.(time.Time); ok {
		return value.Format(transit.GetDateFormat())
	}
	return transit.FormatString(date)
}

func (transit *Convey) PutFormatDateAs(column, as string) bool {
	value, err := transit.Take(column)
	if err != nil {
		transit.PutError(err)
		transit.PutError("can't put the formatted date of", column, "as", as)
		return false
	}
	transit.Set(as, transit.FormatDate(value))
	return true
}

func (transit *Convey) PutFormatDate(column string) bool {
	return transit.PutFormatDateAs(column, column)
}

func (transit *Convey) FormatCurrency(currency interface{}) string {
	if value, ok := currency.(pgtype.Numeric); ok {
		var converted float64
		err := value.AssignTo(&converted)
		if err == nil {
			return fmt.Sprintf("%.2f", converted)
		}
	}
	return transit.FormatString(currency)
}

func (transit *Convey) PutFormatCurrencyAs(column, as string) bool {
	value, err := transit.Take(column)
	if err != nil {
		transit.PutError(err)
		transit.PutError("can't put the formatted currency of", column, "as", as)
		return false
	}
	transit.Set(as, transit.FormatCurrency(value))
	return true
}

func (transit *Convey) PutFormatCurrency(column string) bool {
	return transit.PutFormatCurrencyAs(column, column)
}
