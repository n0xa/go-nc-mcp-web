package nextcloud

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const deckBase = "/index.php/apps/deck/api/v1.0"

type Board struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Color  string `json:"color"`
	Labels []Label `json:"labels,omitempty"`
}

type Stack struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	BoardID int    `json:"boardId"`
	Order   int    `json:"order"`
	Cards   []Card `json:"cards,omitempty"`
}

type Card struct {
	ID          int              `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	StackID     int              `json:"stackId"`
	Type        string           `json:"type"`
	Owner       json.RawMessage  `json:"owner"`
	Order       int              `json:"order"`
	DueDate     string           `json:"duedate"`
	Done        *string          `json:"done"`
	Labels      []Label          `json:"labels,omitempty"`
}

// OwnerUID extracts the owner uid string from the owner field,
// which may be a JSON object {"uid":"..."} or a plain string.
func (c *Card) OwnerUID() string {
	if len(c.Owner) == 0 {
		return ""
	}
	// Try as string first
	var s string
	if json.Unmarshal(c.Owner, &s) == nil {
		return s
	}
	// Try as object with uid field
	var obj struct{ UID string `json:"uid"` }
	if json.Unmarshal(c.Owner, &obj) == nil {
		return obj.UID
	}
	return ""
}

type Label struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Color string `json:"color"`
}

func (c *Client) DeckGetBoards() ([]Board, error) {
	var boards []Board
	err := c.DoJSON("GET", deckBase+"/boards", nil, &boards)
	return boards, err
}

func (c *Client) DeckGetBoard(boardID int) (*Board, error) {
	var board Board
	err := c.DoJSON("GET", fmt.Sprintf("%s/boards/%d", deckBase, boardID), nil, &board)
	return &board, err
}

func (c *Client) DeckGetStacks(boardID int) ([]Stack, error) {
	var stacks []Stack
	err := c.DoJSON("GET", fmt.Sprintf("%s/boards/%d/stacks", deckBase, boardID), nil, &stacks)
	return stacks, err
}

func (c *Client) DeckGetCards(boardID, stackID int) ([]Card, error) {
	// Cards come embedded in the stack response
	var stack Stack
	err := c.DoJSON("GET", fmt.Sprintf("%s/boards/%d/stacks/%d", deckBase, boardID, stackID), nil, &stack)
	if err != nil {
		return nil, err
	}
	return stack.Cards, nil
}

func (c *Client) DeckGetCard(boardID, stackID, cardID int) (*Card, error) {
	var card Card
	err := c.DoJSON("GET", fmt.Sprintf("%s/boards/%d/stacks/%d/cards/%d", deckBase, boardID, stackID, cardID), nil, &card)
	return &card, err
}

func (c *Client) DeckCreateCard(boardID, stackID int, title, description string, dueDate string) (*Card, error) {
	body := map[string]interface{}{"title": title}
	if description != "" {
		body["description"] = description
	}
	if dueDate != "" {
		body["duedate"] = dueDate
	}
	data, _ := json.Marshal(body)
	var card Card
	err := c.DoJSON("POST", fmt.Sprintf("%s/boards/%d/stacks/%d/cards", deckBase, boardID, stackID), bytes.NewReader(data), &card)
	return &card, err
}

func (c *Client) DeckUpdateCard(boardID, stackID, cardID int, fields map[string]interface{}) (*Card, error) {
	// Fetch current card to fill in required fields the API demands
	current, err := c.DeckGetCard(boardID, stackID, cardID)
	if err != nil {
		return nil, fmt.Errorf("fetching card before update: %w", err)
	}

	// Start with all current card fields — the Deck API demands a full object on PUT
	cardType := current.Type
	if cardType == "" {
		cardType = "plain"
	}
	body := map[string]interface{}{
		"title":       current.Title,
		"type":        cardType,
		"owner":       current.OwnerUID(),
		"description": current.Description,
		"order":       current.Order,
	}
	if current.DueDate != "" {
		body["duedate"] = current.DueDate
	} else {
		body["duedate"] = nil
	}
	// Apply caller's fields on top (skip "done" — causes 500 on many NC Deck versions)
	for k, v := range fields {
		body[k] = v
	}

	data, _ := json.Marshal(body)
	var card Card
	err = c.DoJSON("PUT", fmt.Sprintf("%s/boards/%d/stacks/%d/cards/%d", deckBase, boardID, stackID, cardID), bytes.NewReader(data), &card)
	return &card, err
}

func (c *Client) DeckDeleteCard(boardID, stackID, cardID int) error {
	return c.DoJSON("DELETE", fmt.Sprintf("%s/boards/%d/stacks/%d/cards/%d", deckBase, boardID, stackID, cardID), nil, nil)
}

func (c *Client) DeckMoveCard(boardID, stackID, cardID, targetStackID int, order int) (*Card, error) {
	// The UI uses /apps/deck/cards/{id}/reorder (not the api/v1.0 path) for cross-stack moves.
	body := map[string]interface{}{
		"stackId": targetStackID,
		"order":   order,
	}
	data, _ := json.Marshal(body)
	var cards []Card
	err := c.DoJSON("PUT", fmt.Sprintf("/index.php/apps/deck/cards/%d/reorder", cardID), bytes.NewReader(data), &cards)
	if err != nil {
		return nil, err
	}
	for i := range cards {
		if cards[i].ID == cardID {
			return &cards[i], nil
		}
	}
	return nil, nil
}

func (c *Client) DeckCreateStack(boardID int, title string, order int) (*Stack, error) {
	body := map[string]interface{}{"title": title, "order": order}
	data, _ := json.Marshal(body)
	var stack Stack
	err := c.DoJSON("POST", fmt.Sprintf("%s/boards/%d/stacks", deckBase, boardID), bytes.NewReader(data), &stack)
	return &stack, err
}

func (c *Client) DeckAssignLabel(boardID, stackID, cardID, labelID int) error {
	body := map[string]interface{}{"labelId": labelID}
	data, _ := json.Marshal(body)
	return c.DoJSON("PUT", fmt.Sprintf("%s/boards/%d/stacks/%d/cards/%d/assignLabel", deckBase, boardID, stackID, cardID), bytes.NewReader(data), nil)
}

func (c *Client) DeckRemoveLabel(boardID, stackID, cardID, labelID int) error {
	body := map[string]interface{}{"labelId": labelID}
	data, _ := json.Marshal(body)
	return c.DoJSON("PUT", fmt.Sprintf("%s/boards/%d/stacks/%d/cards/%d/removeLabel", deckBase, boardID, stackID, cardID), bytes.NewReader(data), nil)
}
