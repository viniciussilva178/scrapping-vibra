package pkg

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func AuthenticateVibra(user, password string) (*rod.Page, *rod.Browser, error) {
	l := launcher.New().Headless(false)
	url := l.MustLaunch()

	browser := rod.New().ControlURL(url).MustConnect()

	page := browser.MustPage("https://cn.vibraenergia.com.br/login/")
	page.MustWaitLoad()

	usernameInput := page.MustElement(`input[name="usuario"]`)
	usernameInput.MustInput(user)

	passwordInput := page.MustElement(`input[name="senha"]`)
	passwordInput.MustInput(password)

	submitButton := page.MustElement(`.col-md-12.form-group #btn-acessar`)
	submitButton.MustClick()
	page.MustWaitLoad()

	page.MustHandleDialog()
	page.MustHandleDialog()

	payableButton := page.MustElement(`#menuAcessoRevendedorContasPagar`)
	payableButton.MustClick()
	page.MustWaitStable().MustElement("#dtListaDocumentos2").MustVisible()

	page.Timeout(120 * time.Second)

	return page, browser, nil
}
