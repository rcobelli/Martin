package main

import (
	"fmt"
	"time"

	"github.com/adhocore/chin"
	"github.com/gdamore/tcell/v2"
	"github.com/pkg/browser"
	"github.com/rivo/tview"
)

const TEXT_EDIT_PAGE_NAME = "textEditPage"
const ERROR_PAGE_NAME = "errorPage"

var COLUMN_HEADERS = []string{"Name", "Birthday", "How I Know Them", "Tier", "Last Contact Date", "Last Contact Note", "LinkedIn URL"}

const TABLE_CONTROLS_TEXT = "Ctrl + F = page down\nCtrl + B = page up\nCtrl + R = reload\nCtrl + T = Record birthday text\ng = top\nG = bottom"
const DETAILS_CONTROLS_TEXT = "Esc = Back"
const TIERS_CONTROLS_TEXT = "0 = No effort\n1 = Yearly effort\n2 = Quarterly effort\n3 = Active\n\n9 = Ignore"
const COLORS_TEXT = "[lightskyblue]Blue = Tier 0 w/ birthday this month\n[orange]Orange = Missing data\n[red]Red = Overdue"
const EDIT_CONTROLS_TEXT = "Enter/Tab = Save\nEsc = Cancel"

var app = tview.NewApplication()
var pages = tview.NewPages()
var info = tview.NewTextView()
var table = tview.NewTable()
var editTitleInputField = tview.NewInputField()

func main() {
	fmt.Print("Loading data...")
	spinner := chin.New()
	go spinner.Start()

	srv := setupGoogle()
	data := loadData(srv)
	spinner.Stop()
	buildUI()
	renderData(data, table)

	sortCol := 0
	table.SetSelectedFunc(func(row int, col int) {
		// Don't select the header rows, sort instead
		if row == 0 {
			// If we have already sorted by this column, reverse the sort
			if sortCol == col {
				// NOTE: This doesn't work with asc sorting the name column but I'm ok with that
				sortCol = -1 * sortCol
			} else {
				sortCol = col
			}

			sortData(sortCol, &data)
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

	// Setup "modal popup" for editing a field
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
		} else if event.Key() == tcell.KeyCtrlT { // Birthday Text on Crtl + T
			row, _ := table.GetSelection()
			// Account for the header row
			row--

			// Update the in-memory data set for contact date
			if err := updateContactStruct(&data[row], COLUMN_HEADERS[4], time.Now().Format("2006-01-02")); err != nil {
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
		}
		return event
	})

	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}
