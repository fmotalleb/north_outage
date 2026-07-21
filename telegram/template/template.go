package template

import (
	"strings"
	"time"

	"github.com/fmotalleb/go-jalali"
	"github.com/fmotalleb/go-tools/template"
	"github.com/go-telegram/bot/models"
	"github.com/spf13/cast"
)

var funcs = map[string]any{
	"toJalali": toJalali,
	"jFormat":  jFormat,
	"fanum":    faNum,
	"relDate":  relativeDate,
}

func EvaluateTemplate(tmplt string, data map[string]any, update *models.Update) (string, error) {
	if data == nil {
		data = make(map[string]any)
	}
	var msg *models.Message
	if update != nil {
		msg = update.Message
		if msg == nil && update.CallbackQuery != nil {
			msg = update.CallbackQuery.Message.Message
		}
	}
	if msg == nil {
		return "", nil
	}
	data["msg"] = msg
	data["name"] = getName(&msg.Chat)
	out, err := template.EvaluateTemplateWithFuncs(tmplt, data, funcs)
	return out, err
}

func getName(c *models.Chat) string {
	sb := new(strings.Builder)
	if c.FirstName != "" {
		sb.WriteString(c.FirstName)
		if c.LastName != "" {
			sb.WriteRune(' ')
			sb.WriteString(c.LastName)
		}
	} else if c.LastName != "" {
		sb.WriteString(c.LastName)
	}
	if sb.String() == "" {
		return c.Title
	}
	return sb.String()
}

func toJalali(t any) jalali.Time {
	realValue := cast.ToTime(t)
	return jalali.FromGregorian(realValue)
}

func jFormat(format string, t time.Time) string {
	return jalali.FromGregorian(t).FormatPersian(format)
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
