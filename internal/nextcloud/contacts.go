package nextcloud

import (
	"fmt"
	"strings"
)

type AddressBook struct {
	Href string
	Name string
}

type Contact struct {
	Href string
	Data string
}

func (c *Client) contactsPath(subpath string) string {
	return fmt.Sprintf("/remote.php/dav/addressbooks/users/%s/%s", c.Username, subpath)
}

func (c *Client) ContactsListAddressBooks() ([]AddressBook, error) {
	body := `<?xml version="1.0" encoding="utf-8" ?>
<d:propfind xmlns:d="DAV:" xmlns:cr="urn:ietf:params:xml:ns:carddav">
  <d:prop>
    <d:displayname />
    <d:resourcetype />
  </d:prop>
</d:propfind>`

	var ms Multistatus
	headers := map[string]string{
		"Content-Type": "application/xml",
		"Depth":        "1",
	}
	_, err := c.DoXML("PROPFIND", c.contactsPath(""), strings.NewReader(body), headers, &ms)
	if err != nil {
		return nil, err
	}

	var books []AddressBook
	basePath := c.contactsPath("")
	for _, r := range ms.Responses {
		if r.Href == basePath || r.Href == basePath+"/" {
			continue
		}
		for _, ps := range r.Propstat {
			if ps.Prop.DisplayName != "" {
				books = append(books, AddressBook{
					Href: r.Href,
					Name: ps.Prop.DisplayName,
				})
			}
		}
	}
	return books, nil
}

func (c *Client) ContactsList(bookName string) ([]Contact, error) {
	body := `<?xml version="1.0" encoding="utf-8" ?>
<c:addressbook-query xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:carddav">
  <d:prop>
    <d:getetag />
    <c:address-data />
  </d:prop>
</c:addressbook-query>`

	headers := map[string]string{
		"Content-Type": "application/xml",
		"Depth":        "1",
	}

	// CardDAV uses similar XML structure
	type CardMultistatus struct {
		Responses []struct {
			Href     string `xml:"href"`
			Propstat []struct {
				Prop struct {
					AddressData string `xml:"urn:ietf:params:xml:ns:carddav address-data"`
				} `xml:"prop"`
			} `xml:"propstat"`
		} `xml:"response"`
	}

	rawData, status, err := c.DoRaw("REPORT", c.contactsPath(bookName+"/"), strings.NewReader(body), headers)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", status, string(rawData))
	}

	// Parse the response manually since namespace differs
	var ms CardMultistatus
	if err := parseXML(rawData, &ms); err != nil {
		return nil, err
	}

	var contacts []Contact
	for _, r := range ms.Responses {
		for _, ps := range r.Propstat {
			if ps.Prop.AddressData != "" {
				contacts = append(contacts, Contact{
					Href: r.Href,
					Data: ps.Prop.AddressData,
				})
			}
		}
	}
	return contacts, nil
}

func (c *Client) ContactsPut(bookName, uid, vcfData string) error {
	path := c.contactsPath(bookName + "/" + uid + ".vcf")
	headers := map[string]string{
		"Content-Type": "text/vcard; charset=utf-8",
	}
	_, status, err := c.DoRaw("PUT", path, strings.NewReader(vcfData), headers)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("HTTP %d putting contact", status)
	}
	return nil
}

func (c *Client) ContactsDelete(bookName, uid string) error {
	path := c.contactsPath(bookName + "/" + uid + ".vcf")
	_, status, err := c.DoRaw("DELETE", path, nil, nil)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("HTTP %d deleting contact", status)
	}
	return nil
}
