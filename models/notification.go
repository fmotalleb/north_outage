package models

type Notification struct {
	Listener *Listener
	Event    *Event
	Message  string
}
