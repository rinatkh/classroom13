// Package note содержит модель заметки и бизнес-сценарии ownership.
package note

type Note struct {
	ID      int64
	OwnerID int64
	Title   string
	Text    string
}

type CreateInput struct {
	Title string
	Text  string
}

type UpdateInput struct {
	Title string
	Text  string
}
