package nextcloud

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
)

type FileInfo struct {
	Href        string
	Name        string
	ContentType string
	Size        int64
	IsDir       bool
	Modified    string
}

type WebDAVMultistatus struct {
	XMLName   xml.Name         `xml:"DAV: multistatus"`
	Responses []WebDAVResponse `xml:"response"`
}

type WebDAVResponse struct {
	Href     string           `xml:"href"`
	Propstat []WebDAVPropstat `xml:"propstat"`
}

type WebDAVPropstat struct {
	Prop   WebDAVProp `xml:"prop"`
	Status string     `xml:"status"`
}

type WebDAVProp struct {
	DisplayName     string            `xml:"displayname"`
	ContentType     string            `xml:"getcontenttype"`
	ContentLength   int64             `xml:"getcontentlength"`
	LastModified    string            `xml:"getlastmodified"`
	ResourceType    WebDAVResourceType `xml:"resourcetype"`
}

type WebDAVResourceType struct {
	Collection *struct{} `xml:"collection"`
}

func (c *Client) filesPath(path string) string {
	path = strings.TrimPrefix(path, "/")
	return fmt.Sprintf("/remote.php/dav/files/%s/%s", c.Username, path)
}

func (c *Client) FilesList(path string) ([]FileInfo, error) {
	body := `<?xml version="1.0" encoding="utf-8" ?>
<d:propfind xmlns:d="DAV:">
  <d:prop>
    <d:displayname />
    <d:getcontenttype />
    <d:getcontentlength />
    <d:getlastmodified />
    <d:resourcetype />
  </d:prop>
</d:propfind>`

	var ms WebDAVMultistatus
	headers := map[string]string{
		"Content-Type": "application/xml",
		"Depth":        "1",
	}
	_, err := c.DoXML("PROPFIND", c.filesPath(path), strings.NewReader(body), headers, &ms)
	if err != nil {
		return nil, err
	}

	reqPath := c.filesPath(path)
	var files []FileInfo
	for _, r := range ms.Responses {
		// Skip the directory itself
		if r.Href == reqPath || r.Href == reqPath+"/" {
			continue
		}
		for _, ps := range r.Propstat {
			files = append(files, FileInfo{
				Href:        r.Href,
				Name:        ps.Prop.DisplayName,
				ContentType: ps.Prop.ContentType,
				Size:        ps.Prop.ContentLength,
				IsDir:       ps.Prop.ResourceType.Collection != nil,
				Modified:    ps.Prop.LastModified,
			})
		}
	}
	return files, nil
}

func (c *Client) FilesRead(path string) ([]byte, error) {
	data, status, err := c.DoRaw("GET", c.filesPath(path), nil, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", status, string(data))
	}
	return data, nil
}

func (c *Client) FilesWrite(path string, content []byte) error {
	_, status, err := c.DoRaw("PUT", c.filesPath(path), strings.NewReader(string(content)), nil)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("HTTP %d writing file", status)
	}
	return nil
}

func (c *Client) FilesMkdir(path string) error {
	_, status, err := c.DoRaw("MKCOL", c.filesPath(path), nil, nil)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("HTTP %d creating directory", status)
	}
	return nil
}

func (c *Client) FilesDelete(path string) error {
	_, status, err := c.DoRaw("DELETE", c.filesPath(path), nil, nil)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("HTTP %d deleting", status)
	}
	return nil
}

func (c *Client) FilesMove(src, dst string) error {
	headers := map[string]string{
		"Destination": c.Host + c.filesPath(dst),
		"Overwrite":   "F",
	}
	_, status, err := c.DoRaw("MOVE", c.filesPath(src), nil, headers)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("HTTP %d moving", status)
	}
	return nil
}

func (c *Client) FilesSearch(pattern string) ([]FileInfo, error) {
	// Use the Nextcloud OCS unified search API (the WebDAV SEARCH method is not supported).
	path := "/ocs/v2.php/search/providers/files/search?term=" + url.QueryEscape(pattern) + "&format=json"

	var result struct {
		OCS struct {
			Data struct {
				Entries []struct {
					Title      string `json:"title"`
					Icon       string `json:"icon"`
					Attributes struct {
						FileID string `json:"fileId"`
						Path   string `json:"path"`
					} `json:"attributes"`
				} `json:"entries"`
			} `json:"data"`
		} `json:"ocs"`
	}
	err := c.DoJSON("GET", path, nil, &result)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, e := range result.OCS.Data.Entries {
		files = append(files, FileInfo{
			Href:  e.Attributes.Path,
			Name:  e.Title,
			IsDir: e.Icon == "icon-folder",
		})
	}
	return files, nil
}
