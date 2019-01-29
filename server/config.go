package main

// LinkTitle represents a pattern to match against an auto-detected link,
// and a template to form the title of the inserted inline link
type LinkTitle struct {
	UrlPattern    string
	TitleTemplate string
	UrlTemplate   string
}

// LinkCreate represents a pattern to match against plain text,
// and a template to form the URL of the inserted inline link
type LinkCreate struct {
	TextPattern string
	UrlTemplate string
}

// Configuration from config.json
type Configuration struct {
	LinkTitles  []*LinkTitle
	LinkCreates []*LinkCreate
}
