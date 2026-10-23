package dto

type CreateNoteRequest struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type UpdateNoteRequest struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type NoteResponse struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}
