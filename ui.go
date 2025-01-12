package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func createTextEditModal(p tview.Primitive, cell *tview.TableCell) tview.Primitive {
	x, y, _ := cell.GetLastPosition()

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, x, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, y, 1, false).
				AddItem(p, 1, 1, true), 0, 1, true), 0, 1, true)
}

func showErrorModal(err error) {
	modal := tview.NewModal().
		SetText(err.Error()).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage(ERROR_PAGE_NAME)
		})

	pages.AddPage(ERROR_PAGE_NAME, modal, true, false)
	pages.ShowPage(ERROR_PAGE_NAME)
}

func renderData(data []contact, table *tview.Table) {
	table.Clear()

	for v := range len(COLUMN_HEADERS) {
		table.SetCell(0, v,
			tview.NewTableCell(COLUMN_HEADERS[v]).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter))
	}

	for i := 0; i < len(data); i++ {
		transparent := true
		textColor := tcell.ColorWhite
		backgroundColor := tcell.ColorBlack

		if data[i].tier == 0 && data[i].birthday != "" && strings.HasPrefix(data[i].birthday, fmt.Sprintf("%02d/", time.Now().Month())) {
			// Color tier 0 blue if they have a birthday this month
			textColor = tcell.ColorLightSkyBlue
		} else if (data[i].tier == 1 || data[i].tier == 2) && data[i].lastContactDate == "" {
			// Color tier 1 & 2 orange if we are missing data
			textColor = tcell.ColorOrange
		} else if data[i].tier == 1 && isTooLongAgo(data[i].lastContactDate, 12) {
			// Color tier 1 red if contact was more than 12 months ago
			textColor = tcell.ColorRed
		} else if data[i].tier == 2 && isTooLongAgo(data[i].lastContactDate, 4) {
			// Color tier 2 red if contact was more than 4 months ago
			textColor = tcell.ColorRed
		}

		table.SetCell(i+1, 0,
			tview.NewTableCell(data[i].name).
				SetTextColor(textColor).
				SetBackgroundColor(backgroundColor).
				SetTransparency(transparent).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(40))
		table.SetCell(i+1, 1,
			tview.NewTableCell(data[i].birthday).
				SetTextColor(textColor).
				SetBackgroundColor(backgroundColor).
				SetTransparency(transparent).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(10))
		table.SetCell(i+1, 2,
			tview.NewTableCell(data[i].org).
				SetTextColor(textColor).
				SetBackgroundColor(backgroundColor).
				SetTransparency(transparent).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(40))
		table.SetCell(i+1, 3,
			tview.NewTableCell(fmt.Sprintf("%d", data[i].tier)).
				SetTextColor(textColor).
				SetBackgroundColor(backgroundColor).
				SetTransparency(transparent).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(1))
		table.SetCell(i+1, 4,
			tview.NewTableCell(data[i].lastContactDate).
				SetTextColor(textColor).
				SetBackgroundColor(backgroundColor).
				SetTransparency(transparent).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(10))
		table.SetCell(i+1, 5,
			tview.NewTableCell(data[i].lastContactNote).
				SetTextColor(textColor).
				SetBackgroundColor(backgroundColor).
				SetTransparency(transparent).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(40))
		table.SetCell(i+1, 6,
			tview.NewTableCell(data[i].linkedinURL).
				SetTextColor(textColor).
				SetBackgroundColor(backgroundColor).
				SetTransparency(transparent).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(40))
	}
}

func buildUI() {
	// UI Setup
	grid := tview.NewGrid().
		SetRows(15).
		SetColumns(16)
	grid.SetTitleColor(tcell.ColorWhite).
		SetBorderColor(tcell.ColorBlack).
		SetTitle(" Martin: CLI-Based Personal CRM ").
		SetBorder(true)

	table.Select(1, 0).
		SetBorders(true).
		SetFixed(1, 1).
		SetSelectable(true, true).
		SetBorderColor(tcell.ColorAqua).
		SetBorder(true).
		SetTitle(" Contacts ").
		SetTitleColor(tcell.ColorWhiteSmoke)
	table.SetSelectable(true, true)

	info.SetWrap(true).
		SetWordWrap(true).
		SetBorderColor(tcell.ColorAqua).
		SetTitleColor(tcell.ColorWhiteSmoke).
		SetTitle(" Details ").
		SetBorder(true)

	controls := tview.NewTextView()
	controls.SetWrap(true).
		SetWordWrap(true).
		SetBorderColor(tcell.ColorAqua).
		SetTitleColor(tcell.ColorWhiteSmoke).
		SetTitle(" Controls ").
		SetBorder(true)

	tiers := tview.NewTextView()
	tiers.SetWrap(true).
		SetWordWrap(true).
		SetBorderColor(tcell.ColorAqua).
		SetTitleColor(tcell.ColorWhiteSmoke).
		SetTitle(" Tiers ").
		SetBorder(true)
	tiers.SetText(TIERS_CONTROLS_TEXT)

	colors := tview.NewTextView()
	colors.SetWrap(true).
		SetWordWrap(true).
		SetBorderColor(tcell.ColorAqua).
		SetTitleColor(tcell.ColorWhiteSmoke).
		SetTitle(" Colors ").
		SetBorder(true)
	colors.SetText(COLORS_TEXT).
		SetDynamicColors(true)

	grid.AddItem(table, 0, 0, 15, 10, 0, 80, true).
		AddItem(info, 0, 10, 12, 6, 0, 80, false).
		AddItem(controls, 12, 10, 3, 2, 0, 80, false).
		AddItem(tiers, 12, 12, 3, 2, 0, 80, false).
		AddItem(colors, 12, 14, 3, 2, 0, 80, false)

	pages.AddPage("main", grid, true, true)

	info.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.SetFocus(table)
		}
		return event
	})

	// Update controls box contents
	info.SetFocusFunc(func() {
		controls.SetText(DETAILS_CONTROLS_TEXT)
	})
	table.SetFocusFunc(func() {
		controls.SetText(TABLE_CONTROLS_TEXT)
	})
	editTitleInputField.SetFocusFunc(func() {
		controls.SetText(EDIT_CONTROLS_TEXT)
	})
}
