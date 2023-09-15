package renderer

import (
	"bytes"
	"html/template"
)

type Renderer struct {
	tmpl *template.Template
}

func NewRenderer(tmpl *template.Template) *Renderer {
	return &Renderer{
		tmpl: tmpl,
	}
}

func (r *Renderer) render(templateName string, data any) ([]byte, error) {
	var buffer bytes.Buffer
	err := r.tmpl.ExecuteTemplate(&buffer, templateName, data)
	return buffer.Bytes(), err
}

func (r *Renderer) HomePage(p HomePageLoggedInProps) ([]byte, error) {
	return r.render("home", p)
}

func (r *Renderer) Signup(p SignupPageProps) ([]byte, error) {
	return r.render("signup", p)
}

func (r *Renderer) Upgrade(p UpgradePageProps) ([]byte, error) {
	return r.render("upgrade", p)
}

func (r *Renderer) Success(p SuccessPageProps) ([]byte, error) {
	return r.render("success", p)
}

func (r *Renderer) Cancel(p CancelPageProps) ([]byte, error) {
	return r.render("cancel", p)
}

// partials

func (r *Renderer) TodoList(p TodoListProps) ([]byte, error) {
	return r.render("todo-list", p)
}

func (r *Renderer) Todo(p TodoProps) ([]byte, error) {
	return r.render("todo", p)
}
