package tools

import (
	"context"
	"time"

	"go-nc-mcp-web/internal/mcp"
)

func registerDeckTools(srv *mcp.Server) {
	srv.RegisterTool(mcp.Tool{
		Name:        "deck_get_boards",
		Description: "List all Deck boards",
		InputSchema: mcp.InputSchema{Type: "object"},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		boards, err := client(ctx).DeckGetBoards()
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(boards)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "deck_get_board",
		Description: "Get a Deck board by ID",
		InputSchema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"boardId": prop("boardId", "integer", "Board ID")},
			Required:   []string{"boardId"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		id, err := requireInt(args, "boardId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		board, err := client(ctx).DeckGetBoard(id)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(board)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "deck_get_stacks",
		Description: "List stacks in a Deck board",
		InputSchema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"boardId": prop("boardId", "integer", "Board ID")},
			Required:   []string{"boardId"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		id, err := requireInt(args, "boardId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		stacks, err := client(ctx).DeckGetStacks(id)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(stacks)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "deck_get_cards",
		Description: "List cards in a Deck stack",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"boardId": prop("boardId", "integer", "Board ID"),
				"stackId": prop("stackId", "integer", "Stack ID"),
			},
			Required: []string{"boardId", "stackId"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		boardID, err := requireInt(args, "boardId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		stackID, err := requireInt(args, "stackId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		cards, err := client(ctx).DeckGetCards(boardID, stackID)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(cards)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "deck_get_card",
		Description: "Get a specific Deck card",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"boardId": prop("boardId", "integer", "Board ID"),
				"stackId": prop("stackId", "integer", "Stack ID"),
				"cardId":  prop("cardId", "integer", "Card ID"),
			},
			Required: []string{"boardId", "stackId", "cardId"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		boardID, err := requireInt(args, "boardId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		stackID, err := requireInt(args, "stackId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		cardID, err := requireInt(args, "cardId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		card, err := client(ctx).DeckGetCard(boardID, stackID, cardID)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(card)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "deck_create_card",
		Description: "Create a new card in a Deck stack",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"boardId":     prop("boardId", "integer", "Board ID"),
				"stackId":     prop("stackId", "integer", "Stack ID"),
				"title":       prop("title", "string", "Card title"),
				"description": prop("description", "string", "Card description"),
				"duedate":     prop("duedate", "string", "Due date (ISO 8601)"),
			},
			Required: []string{"boardId", "stackId", "title"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		boardID, err := requireInt(args, "boardId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		stackID, err := requireInt(args, "stackId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		title, err := requireString(args, "title")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		card, err := client(ctx).DeckCreateCard(boardID, stackID, title, optString(args, "description"), optString(args, "duedate"))
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(card)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "deck_update_card",
		Description: "Update a Deck card (title, description, duedate, done)",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"boardId":     prop("boardId", "integer", "Board ID"),
				"stackId":     prop("stackId", "integer", "Stack ID"),
				"cardId":      prop("cardId", "integer", "Card ID"),
				"title":       prop("title", "string", "New title"),
				"description": prop("description", "string", "New description"),
				"duedate":     prop("duedate", "string", "New due date (ISO 8601)"),
				"done":        prop("done", "boolean", "Set to true to mark as done (uses current time), or false/omit to leave unchanged"),
			},
			Required: []string{"boardId", "stackId", "cardId"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		boardID, err := requireInt(args, "boardId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		stackID, err := requireInt(args, "stackId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		cardID, err := requireInt(args, "cardId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		fields := make(map[string]interface{})
		if v, ok := args["title"]; ok {
			fields["title"] = v
		}
		if v, ok := args["description"]; ok {
			fields["description"] = v
		}
		if v, ok := args["duedate"]; ok {
			fields["duedate"] = v
		}
		if _, ok := args["done"]; ok {
			fields["done"] = time.Now().Format(time.RFC3339)
		}
		card, err := client(ctx).DeckUpdateCard(boardID, stackID, cardID, fields)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(card)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "deck_delete_card",
		Description: "Delete a Deck card",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"boardId": prop("boardId", "integer", "Board ID"),
				"stackId": prop("stackId", "integer", "Stack ID"),
				"cardId":  prop("cardId", "integer", "Card ID"),
			},
			Required: []string{"boardId", "stackId", "cardId"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		boardID, _ := requireInt(args, "boardId")
		stackID, _ := requireInt(args, "stackId")
		cardID, err := requireInt(args, "cardId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		if err := client(ctx).DeckDeleteCard(boardID, stackID, cardID); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("card deleted")
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "deck_move_card",
		Description: "Move a card to a different stack",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"boardId":       prop("boardId", "integer", "Board ID"),
				"stackId":       prop("stackId", "integer", "Current stack ID"),
				"cardId":        prop("cardId", "integer", "Card ID"),
				"targetStackId": prop("targetStackId", "integer", "Target stack ID"),
				"order":         prop("order", "integer", "Position in target stack (default 0)"),
			},
			Required: []string{"boardId", "stackId", "cardId", "targetStackId"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		boardID, _ := requireInt(args, "boardId")
		stackID, _ := requireInt(args, "stackId")
		cardID, _ := requireInt(args, "cardId")
		targetStackID, err := requireInt(args, "targetStackId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		order := optInt(args, "order", 0)
		card, err := client(ctx).DeckMoveCard(boardID, stackID, cardID, targetStackID, order)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(card)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "deck_create_stack",
		Description: "Create a new stack in a Deck board",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"boardId": prop("boardId", "integer", "Board ID"),
				"title":   prop("title", "string", "Stack title"),
				"order":   prop("order", "integer", "Stack order (default 999)"),
			},
			Required: []string{"boardId", "title"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		boardID, err := requireInt(args, "boardId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		title, err := requireString(args, "title")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		order := optInt(args, "order", 999)
		stack, err := client(ctx).DeckCreateStack(boardID, title, order)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(stack)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "deck_assign_label_to_card",
		Description: "Assign a label to a Deck card",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"boardId": prop("boardId", "integer", "Board ID"),
				"stackId": prop("stackId", "integer", "Stack ID"),
				"cardId":  prop("cardId", "integer", "Card ID"),
				"labelId": prop("labelId", "integer", "Label ID"),
			},
			Required: []string{"boardId", "stackId", "cardId", "labelId"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		boardID, _ := requireInt(args, "boardId")
		stackID, _ := requireInt(args, "stackId")
		cardID, _ := requireInt(args, "cardId")
		labelID, err := requireInt(args, "labelId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		if err := client(ctx).DeckAssignLabel(boardID, stackID, cardID, labelID); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("label assigned")
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "deck_remove_label_from_card",
		Description: "Remove a label from a Deck card",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"boardId": prop("boardId", "integer", "Board ID"),
				"stackId": prop("stackId", "integer", "Stack ID"),
				"cardId":  prop("cardId", "integer", "Card ID"),
				"labelId": prop("labelId", "integer", "Label ID"),
			},
			Required: []string{"boardId", "stackId", "cardId", "labelId"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		boardID, _ := requireInt(args, "boardId")
		stackID, _ := requireInt(args, "stackId")
		cardID, _ := requireInt(args, "cardId")
		labelID, err := requireInt(args, "labelId")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		if err := client(ctx).DeckRemoveLabel(boardID, stackID, cardID, labelID); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("label removed")
	})
}
