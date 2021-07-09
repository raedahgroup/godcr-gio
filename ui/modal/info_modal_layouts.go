package modal

import (
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"

	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/renderers"
	"github.com/planetdecred/godcr/ui/values"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

const (
	VerifyMessageInfoTemplate      = "VerifyMessageInfo"
	SignMessageInfoTemplate        = "SignMessageInfo"
	PrivacyInfoTemplate            = "PrivacyInfo"
	SetupMixerInfoTemplate         = "ConfirmSetupMixer"
	SecurityToolsInfoTemplate      = "SecurityToolsInfo"
	TransactionDetailsInfoTemplate = "TransactionDetailsInfoInfo"
)

func verifyMessageInfo(th *decredmaterial.Theme) []layout.Widget {
	return []layout.Widget{
		func(gtx C) D {
			text := `<span style="text-color: gray">
						After you or your counterparty has genrated a signature, you can use this form to verify the
				 		validity of the signature.
				 		<br /> Once you have entered the address, the message and the corresponding signature, you will see <font color="success">VALID</font> 
						if the signature appropriately matches the address and message, otherwise <font color="danger">INVALID</font>.
					</span>`

			return renderers.RenderHTML(text, th).Layout(gtx)
		},
	}
}

func signMessageInfo(th *decredmaterial.Theme) []layout.Widget {
	text := `<span style="text-color: gray">
				Signing a message with an address' private key allows you to prove that 
				you are the owner of a given address  to a possible counterparty.
			</span>`

	return []layout.Widget{
		renderers.RenderHTML(text, th).Layout,
	}
}

func privacyInfo(th *decredmaterial.Theme) []layout.Widget {
	return []layout.Widget{
		func(gtx C) D {
			return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					ic := mustIcon(widget.NewIcon(icons.ImageLens))
					ic.Color = th.Color.Gray
					return ic.Layout(gtx, values.MarginPadding8)
				}),
				layout.Rigid(func(gtx C) D {
					text := th.Body1("When you turn on the mixer, your unmixed DCRs in this wallet (unmixed balance) will be gradually mixed.")
					text.Color = th.Color.Gray
					return layout.Inset{Left: values.MarginPadding10}.Layout(gtx, text.Layout)
				}),
			)
		},
		func(gtx C) D {
			txt := th.Body1("Important: keep this app opened while mixer is running.")
			txt.Font.Weight = text.Bold
			return txt.Layout(gtx)
		},
		func(gtx C) D {
			return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					ic := mustIcon(widget.NewIcon(icons.ImageLens))
					ic.Color = th.Color.Gray
					return ic.Layout(gtx, values.MarginPadding8)
				}),
				layout.Rigid(func(gtx C) D {
					text := th.Body1("Mixer will automatically stop when unmixed balance are fully mixed.")
					text.Color = th.Color.Gray
					return layout.Inset{Left: values.MarginPadding10}.Layout(gtx, text.Layout)
				}),
			)
		},
	}
}

func setupMixerInfo(th *decredmaterial.Theme) []layout.Widget {
	text := `<span style="text-color: gray">
				Two dedicated accounts (“mixed” & “unmixed”) will be created in order to use the mixer.
				<b>This action cannot be undone.</b>
			</span>`

	return []layout.Widget{
		renderers.RenderHTML(text, th).Layout,
	}
}

func transactionDetailsInfo(th *decredmaterial.Theme) []layout.Widget {
	text := `<span style="text-color: gray">Tap on <span style="text-color: primary">blue text</span> to copy the item</span>`

	return []layout.Widget{
		renderers.RenderHTML(text, th).Layout,
	}
}
