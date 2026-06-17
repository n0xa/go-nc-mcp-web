package nextcloud_test

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCalendarList(t *testing.T) {
	nc := loadTestEnv(t)

	cals, err := nc.CalendarList()
	if err != nil {
		t.Fatalf("CalendarList: %v", err)
	}
	t.Logf("found %d calendars", len(cals))

	if len(cals) == 0 {
		t.Skip("no calendars found for test user — create one in Nextcloud first")
	}

	for _, cal := range cals {
		t.Logf("  calendar: %q href=%s", cal.Name, cal.Href)
	}
}

func TestCalendarEventLifecycle(t *testing.T) {
	nc := loadTestEnv(t)

	cals, err := nc.CalendarList()
	if err != nil {
		t.Fatalf("CalendarList: %v", err)
	}
	if len(cals) == 0 {
		t.Skip("no calendars available")
	}

	// Use the first calendar — extract the calendar name from href
	calHref := cals[0].Href
	parts := strings.Split(strings.TrimRight(calHref, "/"), "/")
	calName := parts[len(parts)-1]
	t.Logf("using calendar %q (name=%s)", cals[0].Name, calName)

	// Create an event
	uid := fmt.Sprintf("ncmcp-test-%d", time.Now().UnixNano())
	now := time.Now()
	start := now.Add(24 * time.Hour)
	end := start.Add(1 * time.Hour)

	ics := fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//go-nc-mcp//integration-test//EN
BEGIN:VEVENT
UID:%s
DTSTART:%s
DTEND:%s
SUMMARY:Integration Test Event
DESCRIPTION:Created by go-nc-mcp integration tests
END:VEVENT
END:VCALENDAR`, uid, start.UTC().Format("20060102T150405Z"), end.UTC().Format("20060102T150405Z"))

	err = nc.CalendarPutEvent(calName, uid, ics)
	if err != nil {
		t.Fatalf("CalendarPutEvent: %v", err)
	}
	t.Logf("created event uid=%s", uid)

	// Get the event
	event, err := nc.CalendarGetEvent(calName, uid)
	if err != nil {
		t.Fatalf("CalendarGetEvent: %v", err)
	}
	if !strings.Contains(event.Data, "Integration Test Event") {
		t.Errorf("event data missing summary, got: %s", event.Data[:min(200, len(event.Data))])
	}

	// List events in range
	rangeStart := now.UTC().Format("20060102T150405Z")
	rangeEnd := now.Add(48 * time.Hour).UTC().Format("20060102T150405Z")
	events, err := nc.CalendarListEvents(calName, rangeStart, rangeEnd)
	if err != nil {
		t.Fatalf("CalendarListEvents: %v", err)
	}
	found := false
	for _, e := range events {
		if strings.Contains(e.Data, uid) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("CalendarListEvents did not return created event %s (got %d events)", uid, len(events))
	}

	// Update the event
	updatedIcs := fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//go-nc-mcp//integration-test//EN
BEGIN:VEVENT
UID:%s
DTSTART:%s
DTEND:%s
SUMMARY:Updated Integration Test Event
DESCRIPTION:Updated by go-nc-mcp integration tests
END:VEVENT
END:VCALENDAR`, uid, start.UTC().Format("20060102T150405Z"), end.UTC().Format("20060102T150405Z"))

	err = nc.CalendarPutEvent(calName, uid, updatedIcs)
	if err != nil {
		t.Fatalf("CalendarPutEvent (update): %v", err)
	}

	// Verify update
	event, err = nc.CalendarGetEvent(calName, uid)
	if err != nil {
		t.Fatalf("CalendarGetEvent after update: %v", err)
	}
	if !strings.Contains(event.Data, "Updated Integration Test Event") {
		t.Errorf("updated event missing new summary")
	}

	// Delete the event
	err = nc.CalendarDeleteEvent(calName, uid)
	if err != nil {
		t.Fatalf("CalendarDeleteEvent: %v", err)
	}
	t.Logf("deleted event %s", uid)

	// Verify deletion
	_, err = nc.CalendarGetEvent(calName, uid)
	if err == nil {
		t.Error("expected error fetching deleted event, got nil")
	}
}

func TestCalendarTodoLifecycle(t *testing.T) {
	nc := loadTestEnv(t)

	cals, err := nc.CalendarList()
	if err != nil {
		t.Fatalf("CalendarList: %v", err)
	}
	if len(cals) == 0 {
		t.Skip("no calendars available")
	}

	// Use first calendar
	parts := strings.Split(strings.TrimRight(cals[0].Href, "/"), "/")
	calName := parts[len(parts)-1]

	// Create a todo
	uid := fmt.Sprintf("ncmcp-todo-test-%d", time.Now().UnixNano())
	ics := fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//go-nc-mcp//integration-test//EN
BEGIN:VTODO
UID:%s
SUMMARY:Integration Test Todo
DESCRIPTION:Created by go-nc-mcp integration tests
STATUS:NEEDS-ACTION
END:VTODO
END:VCALENDAR`, uid)

	err = nc.CalendarPutEvent(calName, uid, ics)
	if err != nil {
		t.Fatalf("CalendarPutEvent (todo): %v", err)
	}
	t.Logf("created todo uid=%s", uid)

	// List todos
	todos, err := nc.CalendarListTodos(calName)
	if err != nil {
		t.Fatalf("CalendarListTodos: %v", err)
	}
	found := false
	for _, td := range todos {
		if strings.Contains(td.Data, uid) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("CalendarListTodos did not return created todo %s (got %d todos)", uid, len(todos))
	}

	// Update todo (mark complete)
	updatedIcs := fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//go-nc-mcp//integration-test//EN
BEGIN:VTODO
UID:%s
SUMMARY:Integration Test Todo
STATUS:COMPLETED
END:VTODO
END:VCALENDAR`, uid)

	err = nc.CalendarPutEvent(calName, uid, updatedIcs)
	if err != nil {
		t.Fatalf("CalendarPutEvent (update todo): %v", err)
	}

	// Delete todo
	err = nc.CalendarDeleteEvent(calName, uid)
	if err != nil {
		t.Fatalf("CalendarDeleteEvent (todo): %v", err)
	}
	t.Logf("deleted todo %s", uid)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
