package ui

import (
	"gioui.org/layout"
)

type Modal interface {
	modalID() string
	OnResume()
	Layout(gtx layout.Context) layout.Dimensions
	OnDismiss()
	handle()
}
