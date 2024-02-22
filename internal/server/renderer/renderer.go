package renderer

import (
	"bytes"
	"fmt"
	"go-todo/internal/models"
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
	if err != nil {
		return nil, fmt.Errorf("Could not execute template %s to buffer. %w", templateName, err)
	}
	return buffer.Bytes(), nil
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
	TodoListProps  TodoListProps
	LoginFormProps LoginFormProps
}

func NewHomePageProps(basePageProps BasePageProps, todoListProps TodoListProps, loginFormProps LoginFormProps) HomePageProps {
	return HomePageProps{
		BasePageProps:  basePageProps,
		TodoListProps:  todoListProps,
		LoginFormProps: loginFormProps,
	}
}
func (r *Renderer) HomePage(p HomePageProps) ([]byte, error) {
	bytes, err := r.render("home", p)
	if err != nil {
		return []byte{}, fmt.Errorf("Could not render Homepage. %w", err)
	}
	return bytes, nil
}

/*
SignupPage
*/
type SignupPageProps struct {
	BasePageProps
	SignupFormProps
}

func NewSignupPageProps(basePageProps BasePageProps, signupFormProps SignupFormProps) SignupPageProps {
	return SignupPageProps{
		BasePageProps:   basePageProps,
		SignupFormProps: signupFormProps,
	}
}
func (r *Renderer) Signup(p SignupPageProps) ([]byte, error) {
	bytes, err := r.render("signup", p)
	if err != nil {
		return []byte{}, fmt.Errorf("Could not render signup page. %w", err)
	}
	return bytes, nil
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
	bytes, err := r.render("upgrade", p)
	if err != nil {
		return []byte{}, fmt.Errorf("Could not render upgrade page. %w", err)
	}
	return bytes, nil
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
	bytes, err := r.render("success", p)
	if err != nil {
		return []byte{}, fmt.Errorf("Could not render success page. %w", err)
	}
	return bytes, nil
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
	bytes, err := r.render("cancel", p)
	if err != nil {
		return []byte{}, fmt.Errorf("Could not render cancel page. %w", err)
	}
	return bytes, nil
}

// partials

type TodoProps *models.Todo

func NewTodoProps(todo *models.Todo) TodoProps {
	return todo
}
func (r *Renderer) Todo(p TodoProps) ([]byte, error) {
	bytes, err := r.render("todo", p)
	if err != nil {
		return []byte{}, fmt.Errorf("Could not render todo element. %w", err)
	}
	return bytes, nil
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
	bytes, err := r.render("todo-list", p)
	if err != nil {
		return []byte{}, fmt.Errorf("Could not render todo list element. %w", err)
	}
	return bytes, nil
}

type LoginFormProps struct {
	EmailErrors    []string
	PasswordErrors []string
}

func NewLoginFormProps(emailErrors, passwordErrors []string) LoginFormProps {
	return LoginFormProps{
		EmailErrors:    emailErrors,
		PasswordErrors: passwordErrors,
	}
}
func (r *Renderer) LoginForm(p LoginFormProps) ([]byte, error) {
	bytes, err := r.render("login-form", p)
	if err != nil {
		return []byte{}, fmt.Errorf("Could not render login form element. %w", err)
	}
	return bytes, nil
}

type SignupFormProps struct {
	EmailErrors    []string
	UsernameErrors []string
	PasswordErrors []string
}

func NewSignupFormProps(usernameErrors, emailErrors, passwordErrors []string) SignupFormProps {
	return SignupFormProps{
		UsernameErrors: usernameErrors,
		EmailErrors:    emailErrors,
		PasswordErrors: passwordErrors,
	}
}
