package main

import (
	"errors"
	"regexp"
)

type RegexReplacer struct {
	pattern  *regexp.Regexp
	template string
}

// Create RegexReplacer for LinkTitle
func NewLinkTitleRegexReplacer(l *LinkTitle) (*RegexReplacer, error) {
	if l == nil || len(l.UrlPattern) == 0 || len(l.TitleTemplate) == 0 {
		return nil, errors.New("Pattern or template was empty")
	}

	p, err := regexp.Compile(l.UrlPattern)
	if err != nil {
		return nil, err
	}
	u := "$0"
	if len(l.UrlTemplate) != 0 {
		u = l.UrlTemplate
	}

	return &RegexReplacer{
		template: "[" + l.TitleTemplate + "](" + u + ")",
		pattern:  p,
	}, nil
}

// Create RegexReplacer for LinkTitle
func NewLinkCreateRegexReplacer(l *LinkCreate) (*RegexReplacer, error) {
	if l == nil || len(l.TextPattern) == 0 || len(l.UrlTemplate) == 0 {
		return nil, errors.New("Pattern or template was empty")
	}

	p, err := regexp.Compile(l.TextPattern)
	if err != nil {
		return nil, err
	}

	return &RegexReplacer{
		template: "[$0](" + l.UrlTemplate + ")",
		pattern:  p,
	}, nil
}

func (r *RegexReplacer) Replace(message string) string {
	if r.pattern == nil {
		return message
	}

	return string(r.pattern.ReplaceAll([]byte(message), []byte(r.template)))
}
