package nextcloud_test

import (
	"testing"
)

func TestDeckFullLifecycle(t *testing.T) {
	nc := loadTestEnv(t)

	// List boards — should succeed even if empty
	boards, err := nc.DeckGetBoards()
	if err != nil {
		t.Fatalf("DeckGetBoards: %v", err)
	}
	t.Logf("found %d boards", len(boards))

	// Create a test stack on the first available board, or skip if no boards
	if len(boards) == 0 {
		t.Skip("no boards available for test user — create one in Nextcloud first")
	}
	boardID := boards[0].ID

	// Get board detail
	board, err := nc.DeckGetBoard(boardID)
	if err != nil {
		t.Fatalf("DeckGetBoard(%d): %v", boardID, err)
	}
	t.Logf("board: %q (id=%d)", board.Title, board.ID)

	// Create a test stack
	stack, err := nc.DeckCreateStack(boardID, "Integration Test Stack", 999)
	if err != nil {
		t.Fatalf("DeckCreateStack: %v", err)
	}
	t.Logf("created stack: %q (id=%d)", stack.Title, stack.ID)
	stackID := stack.ID

	// Clean up stack at end (no API to delete stacks, but we'll delete cards)
	// Note: Deck API doesn't expose stack deletion — we'll leave it for manual cleanup

	// Get stacks
	stacks, err := nc.DeckGetStacks(boardID)
	if err != nil {
		t.Fatalf("DeckGetStacks(%d): %v", boardID, err)
	}
	found := false
	for _, s := range stacks {
		if s.ID == stackID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created stack %d not found in stacks list", stackID)
	}

	// Create a card
	card, err := nc.DeckCreateCard(boardID, stackID, "Test Card", "Test description", "")
	if err != nil {
		t.Fatalf("DeckCreateCard: %v", err)
	}
	t.Logf("created card: %q (id=%d)", card.Title, card.ID)
	cardID := card.ID

	if card.Title != "Test Card" {
		t.Errorf("card title = %q, want %q", card.Title, "Test Card")
	}
	if card.Description != "Test description" {
		t.Errorf("card description = %q, want %q", card.Description, "Test description")
	}

	// Get card
	fetched, err := nc.DeckGetCard(boardID, stackID, cardID)
	if err != nil {
		t.Fatalf("DeckGetCard(%d, %d, %d): %v", boardID, stackID, cardID, err)
	}
	if fetched.ID != cardID {
		t.Errorf("fetched card ID = %d, want %d", fetched.ID, cardID)
	}

	// Get cards in stack
	cards, err := nc.DeckGetCards(boardID, stackID)
	if err != nil {
		t.Fatalf("DeckGetCards: %v", err)
	}
	if len(cards) == 0 {
		t.Fatal("DeckGetCards returned 0 cards, expected at least 1")
	}

	// Update card
	updated, err := nc.DeckUpdateCard(boardID, stackID, cardID, map[string]interface{}{
		"title":       "Updated Test Card",
		"description": "Updated description",
	})
	if err != nil {
		t.Fatalf("DeckUpdateCard: %v", err)
	}
	if updated.Title != "Updated Test Card" {
		t.Errorf("updated card title = %q, want %q", updated.Title, "Updated Test Card")
	}
	if updated.Description != "Updated description" {
		t.Errorf("updated card description = %q, want %q", updated.Description, "Updated description")
	}

	// Create a second stack to test move
	stack2, err := nc.DeckCreateStack(boardID, "Integration Test Stack 2", 998)
	if err != nil {
		t.Fatalf("DeckCreateStack (second): %v", err)
	}
	t.Logf("created stack2: %q (id=%d)", stack2.Title, stack2.ID)

	// Move card to second stack
	moved, err := nc.DeckMoveCard(boardID, stackID, cardID, stack2.ID, 0)
	if err != nil {
		t.Fatalf("DeckMoveCard: %v", err)
	}
	if moved != nil && moved.StackID != stack2.ID {
		t.Errorf("moved card stackId = %d, want %d", moved.StackID, stack2.ID)
	}

	// Test labels if board has any
	if len(board.Labels) > 0 {
		labelID := board.Labels[0].ID
		err = nc.DeckAssignLabel(boardID, stack2.ID, cardID, labelID)
		if err != nil {
			t.Fatalf("DeckAssignLabel: %v", err)
		}
		t.Logf("assigned label %d to card %d", labelID, cardID)

		err = nc.DeckRemoveLabel(boardID, stack2.ID, cardID, labelID)
		if err != nil {
			t.Fatalf("DeckRemoveLabel: %v", err)
		}
		t.Logf("removed label %d from card %d", labelID, cardID)
	} else {
		t.Log("no labels on board, skipping label tests")
	}

	// Archive card
	archived, err := nc.DeckArchiveCard(boardID, stack2.ID, cardID)
	if err != nil {
		t.Fatalf("DeckArchiveCard: %v", err)
	}
	t.Logf("archived card %d", cardID)
	if archived != nil && archived.ID != cardID {
		t.Errorf("archived card ID = %d, want %d", archived.ID, cardID)
	}

	// Unarchive card
	unarchived, err := nc.DeckUnarchiveCard(boardID, stack2.ID, cardID)
	if err != nil {
		t.Fatalf("DeckUnarchiveCard: %v", err)
	}
	t.Logf("unarchived card %d", cardID)
	if unarchived != nil && unarchived.ID != cardID {
		t.Errorf("unarchived card ID = %d, want %d", unarchived.ID, cardID)
	}

	// Delete card
	err = nc.DeckDeleteCard(boardID, stack2.ID, cardID)
	if err != nil {
		t.Fatalf("DeckDeleteCard: %v", err)
	}
	t.Logf("deleted card %d", cardID)

	// Verify deletion
	_, err = nc.DeckGetCard(boardID, stack2.ID, cardID)
	if err == nil {
		t.Error("expected error fetching deleted card, got nil")
	}
}

// Regression: updating a card's title used to wipe the description because
// the Deck API requires a full PUT with all fields. Verify that updating one
// field preserves all others.
func TestDeckUpdatePreservesDescription(t *testing.T) {
	nc := loadTestEnv(t)

	boards, err := nc.DeckGetBoards()
	if err != nil {
		t.Fatalf("DeckGetBoards: %v", err)
	}
	if len(boards) == 0 {
		t.Skip("no boards available")
	}
	boardID := boards[0].ID

	stacks, err := nc.DeckGetStacks(boardID)
	if err != nil {
		t.Fatalf("DeckGetStacks: %v", err)
	}
	if len(stacks) == 0 {
		t.Skip("no stacks available")
	}
	stackID := stacks[0].ID

	// Create card with a description
	card, err := nc.DeckCreateCard(boardID, stackID, "Regression: Preserve Fields", "This description must survive updates", "")
	if err != nil {
		t.Fatalf("DeckCreateCard: %v", err)
	}
	defer nc.DeckDeleteCard(boardID, stackID, card.ID)

	// Update only the title
	updated, err := nc.DeckUpdateCard(boardID, stackID, card.ID, map[string]interface{}{
		"title": "Regression: Title Changed",
	})
	if err != nil {
		t.Fatalf("DeckUpdateCard (title only): %v", err)
	}
	if updated.Title != "Regression: Title Changed" {
		t.Errorf("title = %q, want %q", updated.Title, "Regression: Title Changed")
	}
	if updated.Description != "This description must survive updates" {
		t.Errorf("REGRESSION: description lost after title update! got %q", updated.Description)
	}

	// Update only the description
	updated2, err := nc.DeckUpdateCard(boardID, stackID, card.ID, map[string]interface{}{
		"description": "New description",
	})
	if err != nil {
		t.Fatalf("DeckUpdateCard (description only): %v", err)
	}
	if updated2.Title != "Regression: Title Changed" {
		t.Errorf("REGRESSION: title lost after description update! got %q", updated2.Title)
	}
	if updated2.Description != "New description" {
		t.Errorf("description = %q, want %q", updated2.Description, "New description")
	}
}

// Regression: moving a card to another stack used to lose the description
// because the move endpoint didn't preserve card data.
func TestDeckMovePreservesDescription(t *testing.T) {
	nc := loadTestEnv(t)

	boards, err := nc.DeckGetBoards()
	if err != nil {
		t.Fatalf("DeckGetBoards: %v", err)
	}
	if len(boards) == 0 {
		t.Skip("no boards available")
	}
	boardID := boards[0].ID

	stacks, err := nc.DeckGetStacks(boardID)
	if err != nil {
		t.Fatalf("DeckGetStacks: %v", err)
	}
	if len(stacks) < 2 {
		t.Skip("need at least 2 stacks to test move")
	}
	srcStack := stacks[0].ID
	dstStack := stacks[1].ID

	// Create card with description
	card, err := nc.DeckCreateCard(boardID, srcStack, "Move Regression Test", "Description must survive move", "")
	if err != nil {
		t.Fatalf("DeckCreateCard: %v", err)
	}

	// Move the card
	_, err = nc.DeckMoveCard(boardID, srcStack, card.ID, dstStack, 0)
	if err != nil {
		nc.DeckDeleteCard(boardID, srcStack, card.ID)
		t.Fatalf("DeckMoveCard: %v", err)
	}

	// Fetch card from new location and check description
	moved, err := nc.DeckGetCard(boardID, dstStack, card.ID)
	if err != nil {
		t.Fatalf("DeckGetCard after move: %v", err)
	}
	if moved.Description != "Description must survive move" {
		t.Errorf("REGRESSION: description lost after move! got %q", moved.Description)
	}
	if moved.Title != "Move Regression Test" {
		t.Errorf("REGRESSION: title lost after move! got %q", moved.Title)
	}

	// Cleanup
	nc.DeckDeleteCard(boardID, dstStack, card.ID)
}

func TestDeckCreateCardWithDueDate(t *testing.T) {
	nc := loadTestEnv(t)

	boards, err := nc.DeckGetBoards()
	if err != nil {
		t.Fatalf("DeckGetBoards: %v", err)
	}
	if len(boards) == 0 {
		t.Skip("no boards available")
	}
	boardID := boards[0].ID

	stacks, err := nc.DeckGetStacks(boardID)
	if err != nil {
		t.Fatalf("DeckGetStacks: %v", err)
	}
	if len(stacks) == 0 {
		t.Skip("no stacks available")
	}
	stackID := stacks[0].ID

	card, err := nc.DeckCreateCard(boardID, stackID, "Due Date Test", "", "2099-12-31T00:00:00Z")
	if err != nil {
		t.Fatalf("DeckCreateCard with due date: %v", err)
	}
	t.Logf("created card %d with duedate=%q", card.ID, card.DueDate)

	if card.DueDate == "" {
		t.Error("expected due date to be set, got empty string")
	}

	// Cleanup
	err = nc.DeckDeleteCard(boardID, stackID, card.ID)
	if err != nil {
		t.Errorf("cleanup DeckDeleteCard: %v", err)
	}
}

// Regression: sending "done" as a non-date value (e.g. 1, true) in a card
// update causes Nextcloud Deck to choke with "Failed to parse time string".
// The "done" field expects either null or a date string. Verify that our
// DeckUpdateCard doesn't blow up when callers pass "done" in the fields map,
// and that other fields still survive the update.
func TestDeckUpdateWithDoneField(t *testing.T) {
	nc := loadTestEnv(t)

	boards, err := nc.DeckGetBoards()
	if err != nil {
		t.Fatalf("DeckGetBoards: %v", err)
	}
	if len(boards) == 0 {
		t.Skip("no boards available")
	}
	boardID := boards[0].ID

	stacks, err := nc.DeckGetStacks(boardID)
	if err != nil {
		t.Fatalf("DeckGetStacks: %v", err)
	}
	if len(stacks) == 0 {
		t.Skip("no stacks available")
	}
	stackID := stacks[0].ID

	card, err := nc.DeckCreateCard(boardID, stackID, "Done Field Regression", "Must survive done update", "")
	if err != nil {
		t.Fatalf("DeckCreateCard: %v", err)
	}
	defer nc.DeckDeleteCard(boardID, stackID, card.ID)

	// Update with "done" set to a valid date string (the only safe way)
	updated, err := nc.DeckUpdateCard(boardID, stackID, card.ID, map[string]interface{}{
		"done": "2026-06-17T00:00:00+00:00",
	})
	if err != nil {
		t.Fatalf("DeckUpdateCard with done=date: %v", err)
	}
	if updated.Description != "Must survive done update" {
		t.Errorf("REGRESSION: description lost after done update! got %q", updated.Description)
	}
	if updated.Title != "Done Field Regression" {
		t.Errorf("REGRESSION: title lost after done update! got %q", updated.Title)
	}

	// Verify that setting done to null also works (un-done)
	updated2, err := nc.DeckUpdateCard(boardID, stackID, card.ID, map[string]interface{}{
		"done": nil,
	})
	if err != nil {
		t.Fatalf("DeckUpdateCard with done=nil: %v", err)
	}
	if updated2.Description != "Must survive done update" {
		t.Errorf("REGRESSION: description lost after done=nil update! got %q", updated2.Description)
	}
}

func TestDeckArchiveUnarchive(t *testing.T) {
	nc := loadTestEnv(t)

	boards, err := nc.DeckGetBoards()
	if err != nil {
		t.Fatalf("DeckGetBoards: %v", err)
	}
	if len(boards) == 0 {
		t.Skip("no boards available")
	}
	boardID := boards[0].ID

	stacks, err := nc.DeckGetStacks(boardID)
	if err != nil {
		t.Fatalf("DeckGetStacks: %v", err)
	}
	if len(stacks) == 0 {
		t.Skip("no stacks available")
	}
	stackID := stacks[0].ID

	card, err := nc.DeckCreateCard(boardID, stackID, "Archive Test Card", "Testing archive/unarchive", "")
	if err != nil {
		t.Fatalf("DeckCreateCard: %v", err)
	}
	defer nc.DeckDeleteCard(boardID, stackID, card.ID)

	archived, err := nc.DeckArchiveCard(boardID, stackID, card.ID)
	if err != nil {
		t.Fatalf("DeckArchiveCard: %v", err)
	}
	if archived != nil && archived.ID != card.ID {
		t.Errorf("archived card ID = %d, want %d", archived.ID, card.ID)
	}
	t.Logf("archived card %d", card.ID)

	unarchived, err := nc.DeckUnarchiveCard(boardID, stackID, card.ID)
	if err != nil {
		t.Fatalf("DeckUnarchiveCard: %v", err)
	}
	if unarchived != nil && unarchived.ID != card.ID {
		t.Errorf("unarchived card ID = %d, want %d", unarchived.ID, card.ID)
	}
	t.Logf("unarchived card %d", card.ID)
}
