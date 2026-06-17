package nextcloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
)

const notesBase = "/index.php/apps/notes/api/v1/notes"

type Note struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Modified int64  `json:"modified"`
	Favorite bool   `json:"favorite"`
}

func (c *Client) NotesSearch(query string) ([]Note, error) {
	path := notesBase
	if query != "" {
		path += "?exclude=content&search=" + url.QueryEscape(query)
	} else {
		path += "?exclude=content"
	}
	var notes []Note
	err := c.DoJSON("GET", path, nil, &notes)
	return notes, err
}

func (c *Client) NotesGet(id int) (*Note, error) {
	var note Note
	err := c.DoJSON("GET", fmt.Sprintf("%s/%d", notesBase, id), nil, &note)
	return &note, err
}

func (c *Client) NotesCreate(title, content, category string) (*Note, error) {
	body := map[string]interface{}{"title": title, "content": content}
	if category != "" {
		body["category"] = category
	}
	data, _ := json.Marshal(body)
	var note Note
	err := c.DoJSON("POST", notesBase, bytes.NewReader(data), &note)
	return &note, err
}

func (c *Client) NotesUpdate(id int, fields map[string]interface{}) (*Note, error) {
	data, _ := json.Marshal(fields)
	var note Note
	err := c.DoJSON("PUT", fmt.Sprintf("%s/%d", notesBase, id), bytes.NewReader(data), &note)
	return &note, err
}

func (c *Client) NotesDelete(id int) error {
	return c.DoJSON("DELETE", fmt.Sprintf("%s/%d", notesBase, id), nil, nil)
}
