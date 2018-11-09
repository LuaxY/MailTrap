package http

import (
	"html"
	"io"
	"net/http"
	"text/template"

	"MailTrap/src/database"
	"MailTrap/src/model"
	"github.com/labstack/echo"
)

func Start() {
	e := echo.New()

	reloadTemplates(e)

	e.GET("/mails", getList)
	e.GET("/mails/:id", getMail)
	e.GET("/mails/:id/view", getMailView)

	e.Logger.Fatal(e.Start(":1323"))
}

type Mail struct {
	ID      uint
	Date    string
	From    string
	To      string
	Subject string
	Body    string
	Remote  string
}

func getList(c echo.Context) error {
	reloadTemplates(c.Echo())

	var list []model.Email
	var mails []Mail

	database.DB.Find(&list)

	for _, mail := range list {
		mails = append(mails, Mail{
			ID:      mail.ID,
			Date:    mail.Date,
			From:    mail.From,
			To:      mail.To,
			Subject: mail.Subject,
			Remote:  mail.Remote,
		})
	}

	return c.Render(http.StatusOK, "list.html", struct {
		Mails []Mail
	}{
		Mails: mails,
	})
}

func getMail(c echo.Context) error {
	reloadTemplates(c.Echo())

	id := c.Param("id")

	var email model.Email

	database.DB.First(&email, id)

	return c.Render(http.StatusOK, "mail.html", Mail{
		ID:      email.ID,
		Date:    html.EscapeString(email.Date),
		From:    html.EscapeString(email.EmailFrom),
		To:      html.EscapeString(email.EmailTo),
		Subject: html.EscapeString(email.Subject),
		Body:    email.Body,
	})
}

func getMailView(c echo.Context) error {
	reloadTemplates(c.Echo())

	id := c.Param("id")

	var email model.Email

	database.DB.First(&email, id)

	return c.HTML(http.StatusOK, email.Body)
}

func reloadTemplates(e *echo.Echo) {
	t := &Template{
		templates: template.Must(template.ParseGlob("./web/*.html")),
	}

	e.Renderer = t
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
