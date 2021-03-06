package core

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

type PrintFormat int

const (
	HUMAN_PRINT_FORMAT  PrintFormat = 0
	TEXT_PRINT_FORMAT   PrintFormat = 1
	MAX_DATA_COL_LENGTH             = 120
)

func GetPrintFormat(format string) PrintFormat {
	return map[string]PrintFormat{
		"human": HUMAN_PRINT_FORMAT,
		"text":  TEXT_PRINT_FORMAT,
	}[format]
}

type ItemPrinter struct {
	bumpedColor  color.Attribute
	failColor    color.Attribute
	waitColor    color.Attribute
	successColor color.Attribute

	PrintFormat PrintFormat
}

func NewItemPrinter(ctx context.Context) *ItemPrinter {
	format := ctx.Value("format")
	if format == nil {
		format = "text"
	}
	printFormat := GetPrintFormat(format.(string))
	return &ItemPrinter{bumpedColor: color.FgYellow, failColor: color.FgRed, successColor: color.FgGreen, waitColor: color.FgWhite, PrintFormat: printFormat}
}

func (ip *ItemPrinter) Print(items ...*Item) {
	ip.FPrint(os.Stdout, items...)
}

func (ip *ItemPrinter) FPrint(w io.Writer, items ...*Item) {
	if len(items) == 0 {
		return
	}

	tw := &tabwriter.Writer{}
	tw.Init(w, 5, 0, 1, ' ', 0)
	defer tw.Flush()

	if len(items) == 1 {
		switch ip.PrintFormat {
		case TEXT_PRINT_FORMAT:
			ip.fPrintItemCompact(w, items[0])
		case HUMAN_PRINT_FORMAT:
			ip.fPrintItemDetail(tw, items[0])
		}
		return
	}

	currDay := items[0].Time().Day() - 1 // set to something different
	for _, item := range items {
		switch ip.PrintFormat {
		case TEXT_PRINT_FORMAT:
			ip.fPrintItemCompact(w, item)
		case HUMAN_PRINT_FORMAT:
			if currDay != item.Time().Day() {
				fmt.Fprintf(tw, "\t\t\t\t\n")
				fmt.Fprintf(tw, "- %s\t\t\t\n", item.Time().Format("Monday January 02"))
				currDay = item.Time().Day()
			}
			ip.fPrintItem(tw, item)
		}
	}
}

func (ip *ItemPrinter) fPrintItemDetail(w io.Writer, item *Item) {
	fmt.Fprintf(w, "%s -- %v\n", ip.doneStatus(item), item.Time().Format("Mon, 02 Jan 2006 15:04:05"))
	fmt.Fprintf(w, "InternalID: %s\n", item.internalID)
	if item.NextID() != "" {
		fmt.Fprintf(w, "Bumped to: %s\n", color.New(color.Bold).Sprintf("%s", item.NextID()))
	}
	if item.PreviousID() != "" {
		fmt.Fprintf(w, "Bumped from: %s\n", color.New(color.Bold).Sprintf("%s", item.PreviousID()))
	}
	fmt.Fprintf(w, "Data:\n %s\n\n", item.Data())
}

func (ip *ItemPrinter) fPrintItemCompact(w io.Writer, item *Item) {
	refID := ""
	if item.NextID() != "" {
		refID = "->" + item.NextID()
	}
	if item.PreviousID() != "" {
		refID = "<-" + item.PreviousID()
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%q\t%s\n", item.ID(), item.internalID, item.Status(), refID, item.Data(), item.Time().Format(time.RFC3339))
}

func (ip *ItemPrinter) fPrintItem(w io.Writer, item *Item) {
	fmt.Fprintf(w, "%s\t%q\t%v\t\n", ip.doneStatus(item), ip.truncatedData(item), item.Time().Format("15:04"))
}

func (ip *ItemPrinter) truncatedData(item *Item) string {
	truncLen := math.Min(float64(len(item.Data())), MAX_DATA_COL_LENGTH)
	return item.Data()[:int(truncLen)]
}

func (ip *ItemPrinter) doneStatus(item *Item) string {
	switch item.Status() {
	case "bumped":
		return color.New(ip.bumpedColor, color.Bold).Sprintf("⇒ %v", item.ID())
	case "done":
		return color.New(ip.successColor, color.Bold).Sprintf("✔ %v", item.ID())
	case "waiting":
		return color.New(ip.waitColor, color.Bold).Sprintf("⇒ %v", item.ID())
	case "skipped":
		return color.New(ip.failColor, color.Bold).Sprintf("✘ %v", item.ID())
	default:
		return color.New(ip.waitColor, color.Bold).Sprintf("? %v", item.ID())
	}
}
