package nextcloud_test

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestContactsListAddressBooks(t *testing.T) {
	nc := loadTestEnv(t)

	books, err := nc.ContactsListAddressBooks()
	if err != nil {
		t.Fatalf("ContactsListAddressBooks: %v", err)
	}
	t.Logf("found %d address books", len(books))

	if len(books) == 0 {
		t.Skip("no address books found — Nextcloud should have a default 'contacts' book")
	}

	for _, b := range books {
		t.Logf("  addressbook: %q href=%s", b.Name, b.Href)
	}
}

func TestContactsFullLifecycle(t *testing.T) {
	nc := loadTestEnv(t)

	books, err := nc.ContactsListAddressBooks()
	if err != nil {
		t.Fatalf("ContactsListAddressBooks: %v", err)
	}
	if len(books) == 0 {
		t.Skip("no address books available")
	}

	// Find a writable address book (skip system-generated ones)
	var bookName string
	for _, b := range books {
		parts := strings.Split(strings.TrimRight(b.Href, "/"), "/")
		name := parts[len(parts)-1]
		if strings.HasPrefix(name, "z-") {
			t.Logf("skipping system book %q (%s)", b.Name, name)
			continue
		}
		bookName = name
		t.Logf("using address book %q (name=%s)", b.Name, bookName)
		break
	}
	if bookName == "" {
		t.Skip("no writable address book found — create a 'contacts' address book for the test user")
	}

	// Create a contact
	uid := fmt.Sprintf("ncmcp-test-%d", time.Now().UnixNano())
	vcf := fmt.Sprintf(`BEGIN:VCARD
VERSION:3.0
UID:%s
FN:Integration Test Contact
N:Contact;Integration Test;;;
EMAIL:test@example.com
TEL:+1-555-0199
NOTE:Created by go-nc-mcp integration tests
END:VCARD`, uid)

	err = nc.ContactsPut(bookName, uid, vcf)
	if err != nil {
		t.Fatalf("ContactsPut: %v", err)
	}
	t.Logf("created contact uid=%s", uid)

	// List contacts
	contacts, err := nc.ContactsList(bookName)
	if err != nil {
		t.Fatalf("ContactsList: %v", err)
	}
	found := false
	for _, c := range contacts {
		if strings.Contains(c.Data, uid) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ContactsList did not return created contact %s (got %d contacts)", uid, len(contacts))
	}

	// Update the contact
	updatedVcf := fmt.Sprintf(`BEGIN:VCARD
VERSION:3.0
UID:%s
FN:Updated Test Contact
N:Contact;Updated Test;;;
EMAIL:updated@example.com
TEL:+1-555-0200
NOTE:Updated by go-nc-mcp integration tests
END:VCARD`, uid)

	err = nc.ContactsPut(bookName, uid, updatedVcf)
	if err != nil {
		t.Fatalf("ContactsPut (update): %v", err)
	}

	// Verify update
	contacts, err = nc.ContactsList(bookName)
	if err != nil {
		t.Fatalf("ContactsList after update: %v", err)
	}
	foundUpdated := false
	for _, c := range contacts {
		if strings.Contains(c.Data, uid) && strings.Contains(c.Data, "Updated Test Contact") {
			foundUpdated = true
			break
		}
	}
	if !foundUpdated {
		t.Error("contact not found with updated name after ContactsPut update")
	}

	// Delete the contact
	err = nc.ContactsDelete(bookName, uid)
	if err != nil {
		t.Fatalf("ContactsDelete: %v", err)
	}
	t.Logf("deleted contact %s", uid)

	// Verify deletion
	contacts, err = nc.ContactsList(bookName)
	if err != nil {
		t.Fatalf("ContactsList after delete: %v", err)
	}
	for _, c := range contacts {
		if strings.Contains(c.Data, uid) {
			t.Errorf("contact %s still present after deletion", uid)
		}
	}
}
