package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/visionik/mogcli/internal/graph"
)

// OneNoteCmd handles OneNote operations.
type OneNoteCmd struct {
	Notebooks OneNoteNotebooksCmd `cmd:"" help:"List notebooks"`
	Sections  OneNoteSectionsCmd  `cmd:"" help:"List sections in a notebook"`
	Pages     OneNotePagesCmd     `cmd:"" help:"List pages in a section"`
	Get       OneNoteGetCmd       `cmd:"" help:"Get page content"`
	Search    OneNoteSearchCmd    `cmd:"" help:"Search OneNote"`
}

// OneNoteNotebooksCmd lists notebooks.
type OneNoteNotebooksCmd struct{}

// Run executes onenote notebooks.
func (c *OneNoteNotebooksCmd) Run(root *Root) error {
	client, err := graph.NewClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	data, err := client.Get(ctx, "/me/onenote/notebooks", nil)
	if err != nil {
		return err
	}

	var resp struct {
		Value []Notebook `json:"value"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	if root.JSON {
		return outputJSON(resp.Value)
	}

	for _, nb := range resp.Value {
		fmt.Printf("%-40s %s\n", nb.DisplayName, graph.FormatID(nb.ID))
	}
	return nil
}

// OneNoteSectionsCmd lists sections.
type OneNoteSectionsCmd struct {
	NotebookID string `arg:"" help:"Notebook ID"`
}

// Run executes onenote sections.
func (c *OneNoteSectionsCmd) Run(root *Root) error {
	client, err := graph.NewClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	path := fmt.Sprintf("/me/onenote/notebooks/%s/sections", graph.ResolveID(c.NotebookID))

	data, err := client.Get(ctx, path, nil)
	if err != nil {
		return err
	}

	var resp struct {
		Value []Section `json:"value"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	if root.JSON {
		return outputJSON(resp.Value)
	}

	for _, s := range resp.Value {
		fmt.Printf("%-40s %s\n", s.DisplayName, graph.FormatID(s.ID))
	}
	return nil
}

// OneNotePagesCmd lists pages.
type OneNotePagesCmd struct {
	SectionID string `arg:"" help:"Section ID"`
}

// Run executes onenote pages.
func (c *OneNotePagesCmd) Run(root *Root) error {
	client, err := graph.NewClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	path := fmt.Sprintf("/me/onenote/sections/%s/pages", graph.ResolveID(c.SectionID))

	data, err := client.Get(ctx, path, nil)
	if err != nil {
		return err
	}

	var resp struct {
		Value []Page `json:"value"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	if root.JSON {
		return outputJSON(resp.Value)
	}

	for _, p := range resp.Value {
		fmt.Printf("%-40s %s\n", p.Title, graph.FormatID(p.ID))
	}
	return nil
}

// OneNoteGetCmd gets page content.
type OneNoteGetCmd struct {
	PageID string `arg:"" help:"Page ID"`
	HTML   bool   `help:"Output raw HTML"`
}

// Run executes onenote get.
func (c *OneNoteGetCmd) Run(root *Root) error {
	client, err := graph.NewClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	path := fmt.Sprintf("/me/onenote/pages/%s/content", graph.ResolveID(c.PageID))

	data, err := client.Get(ctx, path, nil)
	if err != nil {
		return err
	}

	if c.HTML || root.JSON {
		fmt.Println(string(data))
		return nil
	}

	// Strip HTML for text output
	fmt.Println(stripHTML(string(data)))
	return nil
}

// OneNoteSearchCmd searches OneNote.
type OneNoteSearchCmd struct {
	Query string `arg:"" help:"Search query"`
}

// Run executes onenote search.
func (c *OneNoteSearchCmd) Run(root *Root) error {
	client, err := graph.NewClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Search pages
	data, err := client.Get(ctx, "/me/onenote/pages", nil)
	if err != nil {
		return err
	}

	var resp struct {
		Value []Page `json:"value"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	if root.JSON {
		return outputJSON(resp.Value)
	}

	fmt.Println("Note: Full-text search requires Graph beta API")
	fmt.Println("Listing all pages instead:")
	for _, p := range resp.Value {
		fmt.Printf("%-40s %s\n", p.Title, graph.FormatID(p.ID))
	}
	return nil
}

// Notebook represents a OneNote notebook.
type Notebook struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// Section represents a OneNote section.
type Section struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// Page represents a OneNote page.
type Page struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
