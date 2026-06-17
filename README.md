# go-nc-mcp-web

A Nextcloud MCP (Model Context Protocol) server over HTTP. Exposes Nextcloud
Deck, Calendar, Notes, Contacts, and Files as MCP tools for use with Claude
custom connectors, routines, or any MCP-compatible client.

Authenticates per-request using Nextcloud app tokens, bridged through an
OAuth 2.1 flow for clients that require it (such as Claude's custom connector
infrastructure).

No external dependencies. Pure Go standard library.

## Building

```
go build -o go-nc-mcp-web ./cmd/server/
```

Cross-compile for OpenBSD (or any target):

```
GOOS=openbsd GOARCH=amd64 go build -o go-nc-mcp-web ./cmd/server/
```

## Configuration

The server reads environment variables from `/etc/ncmcp.env` (if present) and
from the process environment. Required and optional variables:

| Variable | Required | Default | Description |
|---|---|---|---|
| `NEXTCLOUD_HOST` | Yes | | Nextcloud base URL (e.g. `https://nc.example.com`) |
| `MCP_LISTEN` | No | `:8080` | Listen address (e.g. `127.0.0.1:9443`) |
| `MCP_HOSTNAME` | No | `localhost` | Public hostname for OAuth issuer URL |

Example `/etc/ncmcp.env`:

```
NEXTCLOUD_HOST=https://nc.example.com
MCP_LISTEN=127.0.0.1:9443
MCP_HOSTNAME=mcp.example.com
```

## Authentication

The server supports two authentication methods on the `/mcp` endpoint:

- **Basic Auth**: Nextcloud username and app password
- **Bearer Token**: Base64-encoded `username:app-password` (issued by the built-in OAuth flow)

For Claude custom connectors, the OAuth 2.1 flow is used automatically:

1. Client discovers endpoints via `/.well-known/oauth-authorization-server`
2. User is redirected to `/authorize` and enters Nextcloud credentials
3. `/token` returns a Bearer token encoding the credentials
4. All subsequent MCP requests use the Bearer token

Each request is validated against the Nextcloud OCS API before processing.

## Deployment

The server is designed to run behind a TLS-terminating reverse proxy. It
serves plain HTTP only. Example setup with OpenBSD relayd:

```
http protocol wwwtls {
    tls keypair mcp.example.com
    match request header set "X-Forwarded-Proto" value "https"
    match request header set "X-Forwarded-For" value "$REMOTE_ADDR"
    match request header "Host" value "mcp.example.com*" forward to <mcp>
}

relay wwwtls {
    listen on $ext_addr port 443 tls
    protocol wwwtls
    forward to <mcp> port 9443
}
```

An OpenBSD rc.d script is provided in `deploy/rc.d/ncmcp`.

## Connecting to Claude

1. In Claude settings, create a Custom Connector
2. Set the URL to `https://mcp.example.com/mcp`
3. Claude will discover the OAuth endpoints and prompt for Nextcloud credentials
4. Generate an app password in Nextcloud (Settings > Security > Devices & sessions)

Note: Claude custom connectors require port 443 and OAuth 2.1. Non-standard
ports and Basic auth alone will not work.

## Running Tests

Integration tests run against a real Nextcloud instance. Create a `test.env`
in the project root:

```
NEXTCLOUD_HOST=https://nc.example.com
NEXTCLOUD_USERNAME=testuser
NEXTCLOUD_PASSWORD=xxxxx-xxxxx-xxxxx-xxxxx-xxxxx
```

Use a dedicated test user with an app password. Then run:

```
go test -v ./internal/nextcloud/
```

Tests create and clean up their own data. They skip gracefully if `test.env`
is missing.

## Tools

### Deck (12 tools)

| Tool | Description |
|---|---|
| `deck_get_boards` | List all Deck boards |
| `deck_get_board` | Get a Deck board by ID |
| `deck_get_stacks` | List stacks in a Deck board |
| `deck_get_cards` | List cards in a Deck stack |
| `deck_get_card` | Get a specific Deck card |
| `deck_create_card` | Create a new card in a Deck stack |
| `deck_update_card` | Update a Deck card (title, description, duedate, done) |
| `deck_delete_card` | Delete a Deck card |
| `deck_move_card` | Move a card to a different stack |
| `deck_create_stack` | Create a new stack in a Deck board |
| `deck_assign_label_to_card` | Assign a label to a Deck card |
| `deck_remove_label_from_card` | Remove a label from a Deck card |

### Calendar (9 tools)

| Tool | Description |
|---|---|
| `nc_calendar_list` | List all calendars |
| `nc_calendar_list_events` | List events in a calendar within a date range |
| `nc_calendar_get_event` | Get a calendar event by UID |
| `nc_calendar_create_event` | Create a calendar event (full iCalendar VEVENT data) |
| `nc_calendar_update_event` | Update a calendar event (full replacement iCalendar data) |
| `nc_calendar_delete_event` | Delete a calendar event |
| `nc_calendar_list_todos` | List todos/tasks in a calendar |
| `nc_calendar_create_todo` | Create a todo/task (full iCalendar VTODO data) |
| `nc_calendar_update_todo` | Update a todo/task (full replacement iCalendar data) |

### Notes (5 tools)

| Tool | Description |
|---|---|
| `nc_notes_search` | Search notes by query string (returns titles, not content) |
| `nc_notes_get` | Get a note by ID (includes full content) |
| `nc_notes_create` | Create a new note |
| `nc_notes_update` | Update a note (title, content, category) |
| `nc_notes_delete` | Delete a note |

### Contacts (5 tools)

| Tool | Description |
|---|---|
| `nc_contacts_list_addressbooks` | List all address books |
| `nc_contacts_list` | List contacts in an address book (returns vCard data) |
| `nc_contacts_create` | Create a contact (full vCard data) |
| `nc_contacts_update` | Update a contact (full replacement vCard data) |
| `nc_contacts_delete` | Delete a contact |

### Files (7 tools)

| Tool | Description |
|---|---|
| `nc_files_list` | List files and directories at a path |
| `nc_files_read` | Read the content of a file |
| `nc_files_write` | Write content to a file (creates or overwrites) |
| `nc_files_mkdir` | Create a directory |
| `nc_files_delete` | Delete a file or directory |
| `nc_files_move` | Move or rename a file/directory |
| `nc_files_search` | Search for files by name |

## License

MIT
