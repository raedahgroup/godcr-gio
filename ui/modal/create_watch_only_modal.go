package modal

import (
	"fmt"

	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/values"
)

const CreateWatchOnly = "create_watch_only_modal"

type CreateWatchOnlyModal struct {
	*load.Load
	randomID string
	modal    decredmaterial.Modal

	walletName     decredmaterial.Editor
	extendedPubKey decredmaterial.Editor

	isLoading      bool
	materialLoader material.LoaderStyle

	callback   func(walletName, extPubKey string, m *CreateWatchOnlyModal) bool // return true to dismiss dialog
	btnPositve decredmaterial.Button

	btnNegative decredmaterial.Button
}

func NewCreateWatchOnlyModal(l *load.Load) *CreateWatchOnlyModal {
	cm := &CreateWatchOnlyModal{
		Load:        l,
		randomID:    fmt.Sprintf("%s-%d", CreateWatchOnly, generateRandomNumber()),
		modal:       *l.Theme.ModalFloatTitle(),
		btnPositve:  l.Theme.Button(new(widget.Clickable), values.String(values.StrImport)),
		btnNegative: l.Theme.Button(new(widget.Clickable), values.String(values.StrCancel)),
	}

	cm.btnPositve.TextSize, cm.btnNegative.TextSize = values.TextSize16, values.TextSize16
	cm.btnPositve.Font.Weight, cm.btnNegative.Font.Weight = text.Bold, text.Bold

	cm.walletName = l.Theme.Editor(new(widget.Editor), "Wallet name")
	cm.walletName.Editor.SingleLine, cm.walletName.Editor.Submit = true, true

	cm.extendedPubKey = l.Theme.EditorPassword(new(widget.Editor), "Extended public key")
	cm.extendedPubKey.Editor.Submit = true

	th := material.NewTheme(gofont.Collection())
	cm.materialLoader = material.Loader(th)

	return cm
}

func (cm *CreateWatchOnlyModal) ModalID() string {
	return cm.randomID
}

func (cm *CreateWatchOnlyModal) OnResume() {
}

func (cm *CreateWatchOnlyModal) OnDismiss() {

}

func (cm *CreateWatchOnlyModal) Show() {
	cm.ShowModal(cm)
}

func (cm *CreateWatchOnlyModal) Dismiss() {
	cm.DismissModal(cm)
}

func (cm *CreateWatchOnlyModal) SetLoading(loading bool) {
	cm.isLoading = loading
}

func (cm *CreateWatchOnlyModal) SetError(err string) {
	if err == "" {
		cm.extendedPubKey.ClearError()
	} else {
		cm.extendedPubKey.SetError(err)
	}
}

func (cm *CreateWatchOnlyModal) WatchOnlyCreated(callback func(walletName, extPubKey string, m *CreateWatchOnlyModal) bool) *CreateWatchOnlyModal {
	cm.callback = callback
	return cm
}

func (cm *CreateWatchOnlyModal) Handle() {

	if editorsNotEmpty(cm.walletName.Editor, cm.extendedPubKey.Editor) ||
		handleSubmitEvent(cm.walletName.Editor, cm.extendedPubKey.Editor) {
		for cm.btnPositve.Button.Clicked() {
			cm.SetLoading(true)
			if cm.callback(cm.walletName.Editor.Text(), cm.extendedPubKey.Editor.Text(), cm) {
				cm.Dismiss()
			}
		}
	}

	if cm.btnNegative.Button.Clicked() {
		if !cm.isLoading {
			cm.Dismiss()
		}
	}
}

func (cm *CreateWatchOnlyModal) Layout(gtx layout.Context) D {
	w := []layout.Widget{
		func(gtx C) D {
			t := cm.Theme.H6(values.String(values.StrImportWatchingOnlyWallet))
			t.Font.Weight = text.Bold
			return t.Layout(gtx)
		},
		func(gtx C) D {
			return cm.walletName.Layout(gtx)
		},
		func(gtx C) D {
			return cm.extendedPubKey.Layout(gtx)
		},
		func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {

						cm.btnNegative.Background = cm.Theme.Color.Surface
						cm.btnNegative.Color = cm.Theme.Color.Primary
						return cm.btnNegative.Layout(gtx)
					}),
					layout.Rigid(func(gtx C) D {
						if cm.isLoading {
							return cm.materialLoader.Layout(gtx)
						}
						cm.btnPositve.Background, cm.btnPositve.Color = cm.Theme.Color.Surface, cm.Theme.Color.Primary
						return cm.btnPositve.Layout(gtx)
					}),
				)
			})
		},
	}

	return cm.modal.Layout(gtx, w, 850)
}
