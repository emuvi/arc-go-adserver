package motor

import (
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"golang.org/x/text/language"
)

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

func (transit *Convey) Format(value interface{}) string {
	return fmt.Sprintf("%s", value)
}

func (transit *Convey) PutFormatAs(column, as string) bool {
	value, err := transit.Take(column)
	if err != nil {
		transit.PutError(err)
		transit.PutError("can't put the formatted of", column, "as", as)
		return false
	}
	transit.Set(as, transit.Format(value))
	return true
}

func (transit *Convey) PutFormat(column string) bool {
	return transit.PutFormatAs(column, column)
}

func (transit *Convey) FormatDate(date interface{}) string {
	if value, ok := date.(time.Time); ok {
		return value.Format(transit.GetDateFormat())
	}
	return transit.Format(date)
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
	return transit.Format(currency)
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
