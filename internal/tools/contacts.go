package tools

import (
	"context"

	"go-nc-mcp-web/internal/mcp"
)

func registerContactsTools(srv *mcp.Server) {
	srv.RegisterTool(mcp.Tool{
		Name:        "nc_contacts_list_addressbooks",
		Description: "List all address books",
		InputSchema: mcp.InputSchema{Type: "object"},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		books, err := client(ctx).ContactsListAddressBooks()
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(books)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_contacts_list",
		Description: "List contacts in an address book (returns vCard data)",
		InputSchema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"addressbook": prop("addressbook", "string", "Address book name (URL slug, e.g. 'contacts')")},
			Required:   []string{"addressbook"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		book, _ := requireString(args, "addressbook")
		contacts, err := client(ctx).ContactsList(book)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(contacts)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_contacts_create",
		Description: "Create a contact. Provide full vCard data.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"addressbook": prop("addressbook", "string", "Address book name"),
				"uid":         prop("uid", "string", "Unique contact ID"),
				"vcfData":     prop("vcfData", "string", "Full vCard data"),
			},
			Required: []string{"addressbook", "uid", "vcfData"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		book, _ := requireString(args, "addressbook")
		uid, _ := requireString(args, "uid")
		vcf, _ := requireString(args, "vcfData")
		if err := client(ctx).ContactsPut(book, uid, vcf); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("contact created")
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_contacts_update",
		Description: "Update a contact. Provide full replacement vCard data.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"addressbook": prop("addressbook", "string", "Address book name"),
				"uid":         prop("uid", "string", "Contact UID"),
				"vcfData":     prop("vcfData", "string", "Full replacement vCard data"),
			},
			Required: []string{"addressbook", "uid", "vcfData"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		book, _ := requireString(args, "addressbook")
		uid, _ := requireString(args, "uid")
		vcf, _ := requireString(args, "vcfData")
		if err := client(ctx).ContactsPut(book, uid, vcf); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("contact updated")
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_contacts_delete",
		Description: "Delete a contact",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"addressbook": prop("addressbook", "string", "Address book name"),
				"uid":         prop("uid", "string", "Contact UID"),
			},
			Required: []string{"addressbook", "uid"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		book, _ := requireString(args, "addressbook")
		uid, _ := requireString(args, "uid")
		if err := client(ctx).ContactsDelete(book, uid); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("contact deleted")
	})
}
