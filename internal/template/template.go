package template

import (
	"strings"
	"time"

	"github.com/fmotalleb/go-tools/template"
	"github.com/mshafiee/jalali"
	"github.com/spf13/cast"
)

var funcs = map[string]any{
	"toJalali": toJalali,
	"jFormat":  jFormat,
	"fanum":    faNum,
	"relDate":  relativeDate,
}

// EvaluateTemplate evaluates a template string with the provided data using the
// built-in custom functions (toJalali, jFormat, fanum, relDate).
func EvaluateTemplate(tmplt string, data map[string]any) (string, error) {
	if data == nil {
		data = make(map[string]any)
	}
	out, err := template.EvaluateTemplateWithFuncs(tmplt, data, funcs)
	return out, err
}

func toJalali(t any) jalali.JalaliTime {
	realValue := cast.ToTime(t)
	return jalali.JalaliFromTime(realValue)
}

func jFormat(format string, t time.Time) string {
	return jalali.JalaliFromTime(t).Format(format)
}

var faNumMap = map[rune]rune{
	'1': '۱',
	'2': '۲',
	'3': '۳',
	'4': '۴',
	'5': '۵',
	'6': '۶',
	'7': '۷',
	'8': '۸',
	'9': '۹',
	'0': '۰',
}

func faNum(in any) string {
	sin := cast.ToString(in)
	return strings.Map(
		func(r rune) rune {
			if fa, ok := faNumMap[r]; ok {
				return fa
			}
			return r
		},
		sin,
	)
}

func relativeDate(t time.Time) string {
	now := time.Now()

	// Normalize both times to midnight in local timezone
	y1, m1, d1 := now.Date()
	y2, m2, d2 := t.Date()
	n1 := time.Date(y1, m1, d1, 0, 0, 0, 0, now.Location())
	n2 := time.Date(y2, m2, d2, 0, 0, 0, 0, t.Location())

	diff := int(n2.Sub(n1).Hours() / 24)

	switch diff {
	case 0:
		return "امروز"
	case -1:
		return "دیروز"
	case 1:
		return "فردا"
	default:
		if diff < 0 {
			return faNum(-diff) + " روز پیش"
		}
		return faNum(-diff) + " روز دیگه"
	}
}
