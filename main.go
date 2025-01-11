package main

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/pkg/browser"
	"github.com/rivo/tview"
)

const TEXT_EDIT_PAGE_NAME = "textEditPage"
const ERROR_PAGE_NAME = "errorPage"

var COLUMN_HEADERS = []string{"Name", "Birthday", "How I Know Them", "Tier", "Last Contact Date", "Last Contact Note", "LinkedIn URL"}

const TABLE_CONTROLS_TEXT = "Ctrl + F = page down\nCtrl + B = page up\nCtrl + R = reload\ng = top\nG = bottom\nEnter = more detail or sort"
const DETAILS_CONTROLS_TEXT = "Esc = Exit\nCtrl + B = Last contact was birthday text"
const TIERS_CONTROLS_TEXT = "0 = No effort\n1 = Yearly effort\n2 = Quarterly effort\n3 = Active"

func createTextEditModal(p tview.Primitive, cell *tview.TableCell) tview.Primitive {
	x, y, _ := cell.GetLastPosition()

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, x, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, y, 1, false).
				AddItem(p, 1, 1, true), 0, 1, true), 0, 1, true)
}

func renderData(data []contact, table *tview.Table) {
	table.Clear()

	for v := range len(COLUMN_HEADERS) {
		table.SetCell(0, v,
			tview.NewTableCell(COLUMN_HEADERS[v]).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter))
	}

	// TODO: Color the row based on last contact date & tier
	for i := 0; i < len(data); i++ {
		table.SetCell(i+1, 0,
			tview.NewTableCell(data[i].name).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(40))
		table.SetCell(i+1, 1,
			tview.NewTableCell(data[i].birthday).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(10))
		table.SetCell(i+1, 2,
			tview.NewTableCell(data[i].org).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(40))
		table.SetCell(i+1, 3,
			tview.NewTableCell(fmt.Sprintf("%d", data[i].tier)).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(1))
		table.SetCell(i+1, 4,
			tview.NewTableCell(data[i].lastContactDate).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(10))
		table.SetCell(i+1, 5,
			tview.NewTableCell(data[i].lastContactNote).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(40))
		table.SetCell(i+1, 6,
			tview.NewTableCell(data[i].linkedinURL).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(40))
	}
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

var pages = tview.NewPages()

func main() {
	// UI Setup
	app := tview.NewApplication()
	grid := tview.NewGrid().
		SetRows(15).
		SetColumns(15)
	grid.SetTitleColor(tcell.ColorWhite).
		SetBorderColor(tcell.ColorBlack).
		SetTitle(" Martin: CLI-Based Personal CRM ").
		SetBorder(true)

	table := tview.NewTable()
	table.Select(1, 0).
		SetBorders(true).
		SetFixed(1, 1).
		SetSelectable(true, true).
		SetBorderColor(tcell.ColorAqua).
		SetBorder(true).
		SetTitle(" Contacts ").
		SetTitleColor(tcell.ColorWhiteSmoke)
	table.SetSelectable(true, true)

	info := tview.NewTextView()
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

	grid.AddItem(table, 0, 0, 15, 10, 0, 80, true).
		AddItem(info, 0, 10, 13, 5, 0, 80, false).
		AddItem(controls, 13, 10, 2, 2, 0, 80, false).
		AddItem(tiers, 13, 12, 2, 3, 0, 80, false)

	pages.AddPage("main", grid, true, true)

	srv := setupGoogle()
	data := loadData(srv)
	renderData(data, table)

	info.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Birthday on Crtl + B
		if event.Key() == tcell.KeyCtrlB {
			row, _ := table.GetSelection()
			// Account for the header row
			row--

			// Update the in-memory data set for contact date
			if err := updateContactStruct(&data[row], COLUMN_HEADERS[4], data[row].birthday); err != nil {
				showErrorModal(err)
			} // Update the in-memory data set for contact note
			if err := updateContactStruct(&data[row], COLUMN_HEADERS[5], "Birthday text"); err != nil {
				showErrorModal(err)
			}

			// Update Google for both
			if err := updateGoogleContactData(srv, &data[row]); err != nil {
				showErrorModal(err)
			}

			// Re-draw table
			renderData(data, table)
		} else if event.Key() == tcell.KeyEsc {
			app.SetFocus(table)
		}
		return event
	})

	// Setup "modal popup" for editing a field
	editTitleInputField := tview.NewInputField()
	editTitleInputField.SetDoneFunc(func(key tcell.Key) {
		pages.RemovePage(TEXT_EDIT_PAGE_NAME)

		// Don't save unless it's Enter or Tab
		if key != tcell.KeyCR && key != tcell.KeyTAB {
			return
		}

		row, col := table.GetSelection()
		// Account for the header row
		row--

		// Update the in-memory data set
		if err := updateContactStruct(&data[row], COLUMN_HEADERS[col], editTitleInputField.GetText()); err != nil {
			showErrorModal(err)
		}

		// Update Google
		if err := updateGoogleContactData(srv, &data[row]); err != nil {
			showErrorModal(err)
		}

		// Re-draw table
		renderData(data, table)
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Reload on Crtl + R
		if event.Key() == tcell.KeyCtrlR {
			data = loadData(srv)
			renderData(data, table)
			return nil
		} else if event.Key() == tcell.KeyEscape {
			app.Stop()
		}
		return event
	})

	table.SetSelectedFunc(func(row int, col int) {
		// Don't select the header rows, sort instead
		if row == 0 {
			sort.Slice(data, func(i, j int) bool {
				switch COLUMN_HEADERS[col] {
				case "Name":
					return data[i].name < data[j].name
				case "Birthday":
					return data[i].birthday < data[j].birthday
				case "How I Know Them":
					return data[i].org < data[j].org
				case "Tier":
					return data[i].tier < data[j].tier
				case "Last Contact Date":
					return data[i].lastContactDate < data[j].lastContactDate
				case "Last Contact Note":
					return data[i].lastContactNote < data[j].lastContactNote
				case "LinkedIn URL":
					return data[i].linkedinURL < data[j].linkedinURL
				default:
					panic("Unknown sort field")
				}
			})

			// Re-draw table
			renderData(data, table)
			return
		}

		cell := table.GetCell(row, col)
		// Don't edit the name, just show more info
		if col == 0 {
			info.SetText(generateDetailsString(data[row-1]))
			app.SetFocus(info)
			return
		}
		// Open populated LinkedIn URLS, don't edit
		if col == 6 && data[row-1].linkedinURL != "" {
			browser.OpenURL(data[row-1].linkedinURL)
			return
		}

		editTitleInputField.SetText(cell.Text)
		editTitleInputField.SetFieldWidth(cell.MaxWidth + 1)
		pages.AddPage(TEXT_EDIT_PAGE_NAME, createTextEditModal(editTitleInputField, cell), true, false)
		pages.ShowPage(TEXT_EDIT_PAGE_NAME)
	})

	// Update controls box contents
	info.SetFocusFunc(func() {
		controls.SetText(DETAILS_CONTROLS_TEXT)
	})
	table.SetFocusFunc(func() {
		controls.SetText(TABLE_CONTROLS_TEXT)
	})

	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}
