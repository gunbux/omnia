package main

import (
	"strings"

	"github.com/gunbux/omnia/calc"
	"github.com/gunbux/omnia/currency"
	"github.com/gunbux/omnia/finder"
)

// Result items for the non-app modes. Each implements list.DefaultItem.

type calcItem struct {
	value float64
}

func (c calcItem) Title() string       { return "= " + calc.Format(c.value) }
func (c calcItem) FilterValue() string { return calc.Format(c.value) }
func (c calcItem) Description() string {
	if alts := calc.Alternates(c.value); alts != nil {
		return strings.Join(alts, "  ")
	}
	return ""
}

// Text is what gets copied to the clipboard.
func (c calcItem) Text() string { return calc.Format(c.value) }

type currencyItem struct {
	amount float64
	from   string
	to     string
	result float64
	rate   float64
}

func (c currencyItem) Title() string {
	return currency.FormatAmount(c.result) + " " + c.to
}

func (c currencyItem) Description() string {
	name := currency.Names[c.to]
	if name == "" {
		name = c.to
	}
	return name + "  ·  1 " + c.from + " = " + currency.FormatAmount(c.rate) + " " + c.to
}

func (c currencyItem) FilterValue() string { return c.to }
func (c currencyItem) Text() string        { return currency.FormatAmount(c.result) }

type fileItem struct {
	finder.Result
}

func (f fileItem) Title() string       { return f.Name() }
func (f fileItem) Description() string { return f.Dir() }
func (f fileItem) FilterValue() string { return f.Path }
func (f fileItem) Text() string        { return f.Path }

type aiItem struct {
	prompt string
}

func (a aiItem) Title() string       { return "Ask Claude: " + a.prompt }
func (a aiItem) Description() string { return "Enter to ask" }
func (a aiItem) FilterValue() string { return a.prompt }

// infoItem is a non-actionable message shown in the result list.
type infoItem struct {
	title string
	desc  string
}

func (i infoItem) Title() string       { return i.title }
func (i infoItem) Description() string { return i.desc }
func (i infoItem) FilterValue() string { return i.title }

// copyable is implemented by items whose text can be copied with Ctrl+Y.
type copyable interface {
	Text() string
}
