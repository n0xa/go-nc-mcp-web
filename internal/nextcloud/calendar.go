package nextcloud

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// CalDAV XML types

type Multistatus struct {
	XMLName   xml.Name   `xml:"DAV: multistatus"`
	Responses []DAVResponse `xml:"response"`
}

type DAVResponse struct {
	Href     string      `xml:"href"`
	Propstat []Propstat  `xml:"propstat"`
}

type Propstat struct {
	Prop   Prop   `xml:"prop"`
	Status string `xml:"status"`
}

type Prop struct {
	DisplayName  string        `xml:"displayname"`
	CalendarData string        `xml:"urn:ietf:params:xml:ns:caldav calendar-data"`
	ResourceType ResourceType  `xml:"resourcetype"`
	CalColor     string        `xml:"http://apple.com/ns/ical/ calendar-color"`
}

type ResourceType struct {
	Calendar *struct{} `xml:"urn:ietf:params:xml:ns:caldav calendar"`
}

type Calendar struct {
	Href  string
	Name  string
	Color string
}

type CalEvent struct {
	Href string
	Data string
}

func (c *Client) calendarPath(subpath string) string {
	return fmt.Sprintf("/remote.php/dav/calendars/%s/%s", c.Username, subpath)
}

func (c *Client) CalendarList() ([]Calendar, error) {
	body := `<?xml version="1.0" encoding="utf-8" ?>
<d:propfind xmlns:d="DAV:" xmlns:cs="urn:ietf:params:xml:ns:caldav" xmlns:ic="http://apple.com/ns/ical/">
  <d:prop>
    <d:displayname />
    <d:resourcetype />
    <ic:calendar-color />
  </d:prop>
</d:propfind>`

	var ms Multistatus
	headers := map[string]string{
		"Content-Type": "application/xml",
		"Depth":        "1",
	}
	_, err := c.DoXML("PROPFIND", c.calendarPath(""), strings.NewReader(body), headers, &ms)
	if err != nil {
		return nil, err
	}

	var cals []Calendar
	for _, r := range ms.Responses {
		for _, ps := range r.Propstat {
			if ps.Prop.ResourceType.Calendar != nil {
				cals = append(cals, Calendar{
					Href:  r.Href,
					Name:  ps.Prop.DisplayName,
					Color: ps.Prop.CalColor,
				})
			}
		}
	}
	return cals, nil
}

func (c *Client) CalendarListEvents(calName, start, end string) ([]CalEvent, error) {
	body := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8" ?>
<c:calendar-query xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
  <d:prop>
    <d:getetag />
    <c:calendar-data />
  </d:prop>
  <c:filter>
    <c:comp-filter name="VCALENDAR">
      <c:comp-filter name="VEVENT">
        <c:time-range start="%s" end="%s"/>
      </c:comp-filter>
    </c:comp-filter>
  </c:filter>
</c:calendar-query>`, start, end)

	var ms Multistatus
	headers := map[string]string{
		"Content-Type": "application/xml",
		"Depth":        "1",
	}
	_, err := c.DoXML("REPORT", c.calendarPath(calName+"/"), strings.NewReader(body), headers, &ms)
	if err != nil {
		return nil, err
	}

	var events []CalEvent
	for _, r := range ms.Responses {
		for _, ps := range r.Propstat {
			if ps.Prop.CalendarData != "" {
				events = append(events, CalEvent{
					Href: r.Href,
					Data: ps.Prop.CalendarData,
				})
			}
		}
	}
	return events, nil
}

func (c *Client) CalendarListTodos(calName string) ([]CalEvent, error) {
	body := `<?xml version="1.0" encoding="utf-8" ?>
<c:calendar-query xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
  <d:prop>
    <d:getetag />
    <c:calendar-data />
  </d:prop>
  <c:filter>
    <c:comp-filter name="VCALENDAR">
      <c:comp-filter name="VTODO" />
    </c:comp-filter>
  </c:filter>
</c:calendar-query>`

	var ms Multistatus
	headers := map[string]string{
		"Content-Type": "application/xml",
		"Depth":        "1",
	}
	_, err := c.DoXML("REPORT", c.calendarPath(calName+"/"), strings.NewReader(body), headers, &ms)
	if err != nil {
		return nil, err
	}

	var todos []CalEvent
	for _, r := range ms.Responses {
		for _, ps := range r.Propstat {
			if ps.Prop.CalendarData != "" {
				todos = append(todos, CalEvent{
					Href: r.Href,
					Data: ps.Prop.CalendarData,
				})
			}
		}
	}
	return todos, nil
}

func (c *Client) CalendarGetEvent(calName, uid string) (*CalEvent, error) {
	path := c.calendarPath(calName + "/" + uid + ".ics")
	data, status, err := c.DoRaw("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", status, string(data))
	}
	return &CalEvent{
		Href: path,
		Data: string(data),
	}, nil
}

func (c *Client) CalendarPutEvent(calName, uid, icsData string) error {
	path := c.calendarPath(calName + "/" + uid + ".ics")
	headers := map[string]string{
		"Content-Type": "text/calendar; charset=utf-8",
	}
	_, status, err := c.DoRaw("PUT", path, strings.NewReader(icsData), headers)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("HTTP %d putting event", status)
	}
	return nil
}

func (c *Client) CalendarDeleteEvent(calName, uid string) error {
	path := c.calendarPath(calName + "/" + uid + ".ics")
	_, status, err := c.DoRaw("DELETE", path, nil, nil)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("HTTP %d deleting event", status)
	}
	return nil
}
