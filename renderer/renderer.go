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

type BasePageProps struct {
	User *models.User
}

func NewBasePageProps(user *models.User) BasePageProps {
	return BasePageProps{
		User: user,
	}
}

/*
	Homepage
*/
type HomePageProps struct {
	BasePageProps
	TodoListProps *TodoListProps
	LoginFormProps *LoginFormProps
}
func NewHomePageProps(basePageProps BasePageProps, todoListProps *TodoListProps, loginFormProps *LoginFormProps) HomePageProps {
	return HomePageProps{
		BasePageProps: basePageProps,
		TodoListProps: todoListProps,
		LoginFormProps: loginFormProps,
	}
}
func (r *Renderer) HomePage(p HomePageProps) ([]byte, error) {
	return r.render("home", p)
}

/*
	SignupPage
*/
type SignupPageProps struct {
	BasePageProps
}
func NewSignupPageProps(basePageProps BasePageProps) SignupPageProps {
	return SignupPageProps{
		BasePageProps: basePageProps,
	}
}
func (r *Renderer) Signup(p SignupPageProps) ([]byte, error) {
	return r.render("signup", p)
}

/*
	UpgradePage
*/
type UpgradePageProps struct {
	BasePageProps
}
func NewUpgradePageProps(basePageProps BasePageProps) UpgradePageProps {
	return UpgradePageProps{
		BasePageProps: basePageProps,
	}
}
func (r *Renderer) Upgrade(p UpgradePageProps) ([]byte, error) {
	return r.render("upgrade", p)
}

/*
	Success Page
*/
type SuccessPageProps struct {
	BasePageProps
}
func NewSuccessPageProps(basePageProps BasePageProps) SuccessPageProps {
	return SuccessPageProps{
		BasePageProps: basePageProps,
	}
}
func (r *Renderer) Success(p SuccessPageProps) ([]byte, error) {
	return r.render("success", p)
}

/*
	Cancel Page
*/
type CancelPageProps struct {
	BasePageProps
}
func NewCancelPageProps(basePageProps BasePageProps) CancelPageProps {
	return CancelPageProps{
		BasePageProps: basePageProps,
	}
}
func (r *Renderer) Cancel(p CancelPageProps) ([]byte, error) {
	return r.render("cancel", p)
}

// partials

type TodoProps *models.Todo
func NewTodoProps(todo *models.Todo) TodoProps {
	return todo
}
func (r *Renderer) Todo(p TodoProps) ([]byte, error) {
	return r.render("todo", p)
}



type TodoListProps struct {
	Todos            []*models.Todo
	CanCreateNewTodo bool
}
func NewTodoListProps(todoList []*models.Todo, canCreateNewTodo bool) TodoListProps {
	return TodoListProps{
		Todos:            todoList,
		CanCreateNewTodo: canCreateNewTodo,
	}
}
func (r *Renderer) TodoList(p TodoListProps) ([]byte, error) {
	return r.render("todo-list", p)
}


type LoginFormProps struct {
	EmailErrors *[]string
	PasswordErrors *[]string
}
func NewLoginFormProps(emailErrors, passwordErrors *[]string){
	return LoginFormProps{
		EmailErrors: emailErrors,
		PasswordErrors: passwordErrors,
	}
}
func (r *Renderer) LoginForm(p LoginFormProps)([]byte, error) {
	return r.render("login-form", p)
}