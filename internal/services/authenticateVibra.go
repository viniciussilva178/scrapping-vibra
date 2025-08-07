package services

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func AuthenticateVibra(user, password string) (*rod.Page, *rod.Browser, error) {
	// Acessar a Pagina
	l := launcher.New().Headless(true)
	url := l.MustLaunch()

	browser := rod.New().ControlURL(url).MustConnect()
	// defer browser.MustClose() // Removido

	page := browser.MustPage("https://cn.vibraenergia.com.br/login/")
	page.MustWaitLoad()

	usernameInput := page.MustElement(`input[name="usuario"]`)
	usernameInput.MustInput(user)

	passwordInput := page.MustElement(`input[name="senha"]`)
	passwordInput.MustInput(password)

	submitButton := page.MustElement(`.col-md-12.form-group #btn-acessar`)
	submitButton.MustClick()
	page.MustWaitLoad()

	payableButton := page.MustElement(`#menuAcessoRevendedorContasPagar`)
	payableButton.MustClick()
	page.MustWaitLoad()

	time.Sleep(20 * time.Second)

	return page, browser, nil
}
