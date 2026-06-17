package nextcloud_test

import (
	"testing"
)

func TestNotesFullLifecycle(t *testing.T) {
	nc := loadTestEnv(t)

	// Create a note
	note, err := nc.NotesCreate("Integration Test Note", "This is test content.", "testing")
	if err != nil {
		t.Fatalf("NotesCreate: %v", err)
	}
	t.Logf("created note: %q (id=%d)", note.Title, note.ID)

	if note.Title != "Integration Test Note" {
		t.Errorf("note title = %q, want %q", note.Title, "Integration Test Note")
	}
	if note.Content != "This is test content." {
		t.Errorf("note content = %q, want %q", note.Content, "This is test content.")
	}
	if note.Category != "testing" {
		t.Errorf("note category = %q, want %q", note.Category, "testing")
	}

	noteID := note.ID

	// Get the note
	fetched, err := nc.NotesGet(noteID)
	if err != nil {
		t.Fatalf("NotesGet(%d): %v", noteID, err)
	}
	if fetched.ID != noteID {
		t.Errorf("fetched note ID = %d, want %d", fetched.ID, noteID)
	}
	if fetched.Content != "This is test content." {
		t.Errorf("fetched note content = %q, want %q", fetched.Content, "This is test content.")
	}

	// Search for the note
	results, err := nc.NotesSearch("Integration Test Note")
	if err != nil {
		t.Fatalf("NotesSearch: %v", err)
	}
	found := false
	for _, n := range results {
		if n.ID == noteID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("NotesSearch did not return created note %d", noteID)
	}

	// Update the note
	updated, err := nc.NotesUpdate(noteID, map[string]interface{}{
		"title":   "Updated Integration Test Note",
		"content": "Updated content.",
	})
	if err != nil {
		t.Fatalf("NotesUpdate: %v", err)
	}
	if updated.Title != "Updated Integration Test Note" {
		t.Errorf("updated note title = %q, want %q", updated.Title, "Updated Integration Test Note")
	}
	if updated.Content != "Updated content." {
		t.Errorf("updated note content = %q, want %q", updated.Content, "Updated content.")
	}

	// Delete the note
	err = nc.NotesDelete(noteID)
	if err != nil {
		t.Fatalf("NotesDelete: %v", err)
	}
	t.Logf("deleted note %d", noteID)

	// Verify deletion
	_, err = nc.NotesGet(noteID)
	if err == nil {
		t.Error("expected error fetching deleted note, got nil")
	}
}

func TestNotesSearchEmpty(t *testing.T) {
	nc := loadTestEnv(t)

	// Search with empty query should return all notes (or empty list)
	results, err := nc.NotesSearch("")
	if err != nil {
		t.Fatalf("NotesSearch (empty): %v", err)
	}
	t.Logf("NotesSearch('') returned %d notes", len(results))
}
