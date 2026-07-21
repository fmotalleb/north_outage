package message

import (
	"strings"

	"github.com/go-telegram/bot/models"

	"github.com/fmotalleb/north_outage/internal/template"
)

// EvaluateMessageTemplate evaluates a template string with the provided data
// and Telegram update context. It extracts chat name and message info from the
// update and makes them available as "name" and "msg" in the template data.
func EvaluateMessageTemplate(tmplt string, data map[string]any, update *models.Update) (string, error) {
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
	return template.EvaluateTemplate(tmplt, data)
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
