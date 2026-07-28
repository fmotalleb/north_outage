package message

import (
	"embed"
	"fmt"
)

//go:embed files/*
var templates embed.FS

func readOrPanic(key string) string {
	fPath := fmt.Sprintf("files/%s.tmpl", key)
	o, err := templates.ReadFile(fPath)
	if err != nil {
		panic(fmt.Errorf("failed to read template: %w", err))
	}
	return string(o)
}

var (
	Help             = readOrPanic("help")
	Search           = readOrPanic("search")
	List             = readOrPanic("list")
	ClearConfirm     = readOrPanic("clear_confirm")
	ClearDone        = readOrPanic("clear_done")
	Notification     = readOrPanic("notification")
	MMNotification   = readOrPanic("mm_notification")
	RemoveBtn        = readOrPanic("remove_btn")
	WeatherLine      = readOrPanic("weather")
)
