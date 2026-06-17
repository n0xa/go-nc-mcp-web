package tools

import (
	"context"

	"go-nc-mcp-web/internal/mcp"
)

func registerCalendarTools(srv *mcp.Server) {
	srv.RegisterTool(mcp.Tool{
		Name:        "nc_calendar_list",
		Description: "List all calendars",
		InputSchema: mcp.InputSchema{Type: "object"},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		cals, err := client(ctx).CalendarList()
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(cals)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_calendar_list_events",
		Description: "List events in a calendar within a date range",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"calendar": prop("calendar", "string", "Calendar name (URL slug, e.g. 'personal')"),
				"start":    prop("start", "string", "Start date/time in iCal format (e.g. 20250101T000000Z)"),
				"end":      prop("end", "string", "End date/time in iCal format (e.g. 20250201T000000Z)"),
			},
			Required: []string{"calendar", "start", "end"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		cal, _ := requireString(args, "calendar")
		start, _ := requireString(args, "start")
		end, _ := requireString(args, "end")
		events, err := client(ctx).CalendarListEvents(cal, start, end)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(events)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_calendar_get_event",
		Description: "Get a calendar event by UID",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"calendar": prop("calendar", "string", "Calendar name"),
				"uid":      prop("uid", "string", "Event UID (without .ics)"),
			},
			Required: []string{"calendar", "uid"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		cal, _ := requireString(args, "calendar")
		uid, _ := requireString(args, "uid")
		event, err := client(ctx).CalendarGetEvent(cal, uid)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(event)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_calendar_create_event",
		Description: "Create a calendar event. Provide full iCalendar VEVENT data.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"calendar": prop("calendar", "string", "Calendar name"),
				"uid":      prop("uid", "string", "Unique event ID"),
				"icsData":  prop("icsData", "string", "Full iCalendar data (VCALENDAR with VEVENT)"),
			},
			Required: []string{"calendar", "uid", "icsData"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		cal, _ := requireString(args, "calendar")
		uid, _ := requireString(args, "uid")
		ics, _ := requireString(args, "icsData")
		if err := client(ctx).CalendarPutEvent(cal, uid, ics); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("event created")
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_calendar_update_event",
		Description: "Update a calendar event. Provide full replacement iCalendar data.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"calendar": prop("calendar", "string", "Calendar name"),
				"uid":      prop("uid", "string", "Event UID"),
				"icsData":  prop("icsData", "string", "Full replacement iCalendar data"),
			},
			Required: []string{"calendar", "uid", "icsData"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		cal, _ := requireString(args, "calendar")
		uid, _ := requireString(args, "uid")
		ics, _ := requireString(args, "icsData")
		if err := client(ctx).CalendarPutEvent(cal, uid, ics); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("event updated")
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_calendar_delete_event",
		Description: "Delete a calendar event",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"calendar": prop("calendar", "string", "Calendar name"),
				"uid":      prop("uid", "string", "Event UID"),
			},
			Required: []string{"calendar", "uid"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		cal, _ := requireString(args, "calendar")
		uid, _ := requireString(args, "uid")
		if err := client(ctx).CalendarDeleteEvent(cal, uid); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("event deleted")
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_calendar_list_todos",
		Description: "List todos/tasks in a calendar",
		InputSchema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"calendar": prop("calendar", "string", "Calendar name")},
			Required:   []string{"calendar"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		cal, _ := requireString(args, "calendar")
		todos, err := client(ctx).CalendarListTodos(cal)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(todos)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_calendar_create_todo",
		Description: "Create a todo/task. Provide full iCalendar VTODO data.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"calendar": prop("calendar", "string", "Calendar name"),
				"uid":      prop("uid", "string", "Unique todo ID"),
				"icsData":  prop("icsData", "string", "Full iCalendar data (VCALENDAR with VTODO)"),
			},
			Required: []string{"calendar", "uid", "icsData"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		cal, _ := requireString(args, "calendar")
		uid, _ := requireString(args, "uid")
		ics, _ := requireString(args, "icsData")
		if err := client(ctx).CalendarPutEvent(cal, uid, ics); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("todo created")
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_calendar_update_todo",
		Description: "Update a todo/task. Provide full replacement iCalendar data.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"calendar": prop("calendar", "string", "Calendar name"),
				"uid":      prop("uid", "string", "Todo UID"),
				"icsData":  prop("icsData", "string", "Full replacement iCalendar data"),
			},
			Required: []string{"calendar", "uid", "icsData"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		cal, _ := requireString(args, "calendar")
		uid, _ := requireString(args, "uid")
		ics, _ := requireString(args, "icsData")
		if err := client(ctx).CalendarPutEvent(cal, uid, ics); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("todo updated")
	})
}
