package modal

import (
	"fmt"

	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
)

const TextInput = "text_input_modal"

type TextInputModal struct {
	*InfoModal

	IsLoading bool

	textInput decredmaterial.Editor
	callback  func(string, *TextInputModal) bool
}

func NewTextInputModal(l *load.Load) *TextInputModal {
	tm := &TextInputModal{
		InfoModal: NewInfoModal(l),
	}

	tm.randomID = fmt.Sprintf("%s-%d", TextInput, generateRandomNumber())

	tm.textInput = l.Theme.Editor(new(widget.Editor), "Hint")
	tm.textInput.Editor.SingleLine, tm.textInput.Editor.Submit = true, true

	return tm
}

func (tm *TextInputModal) Show() {
	tm.ShowModal(tm)
}

func (tm *TextInputModal) Dismiss() {
	tm.DismissModal(tm)
}

func (tm *TextInputModal) Hint(hint string) *TextInputModal {
	tm.textInput.Hint = hint
	return tm
}

func (tm *TextInputModal) PositiveButton(text string, callback func(string, *TextInputModal) bool) *TextInputModal {
	tm.positiveButtonText = text
	tm.callback = callback
	return tm
}

func (tm *TextInputModal) SetError(err string) {
	if err == "" {
		tm.textInput.ClearError()
	} else {
		tm.textInput.SetError(err)
	}
}

func (tm *TextInputModal) Handle() {

	for tm.btnPositve.Button.Clicked() {
		if tm.IsLoading {
			continue
		}

		tm.IsLoading = true
		tm.SetError("")
		if tm.callback(tm.textInput.Editor.Text(), tm) {
			tm.Dismiss()
		}
	}

	for tm.btnNegative.Button.Clicked() {
		if !tm.IsLoading {
			tm.Dismiss()
			tm.negativeButtonClicked()
		}
	}
}

func (tm *TextInputModal) Layout(gtx layout.Context) D {

	var w []layout.Widget

	if tm.dialogTitle != "" {
		w = append(w, tm.titleLayout())
	}

	w = append(w, tm.textInput.Layout)
	if tm.negativeButtonText != "" || tm.positiveButtonText != "" {
		w = append(w, tm.actionButtonsLayout())
	}

	return tm.modal.Layout(gtx, w, 850)
}
