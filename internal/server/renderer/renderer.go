package renderer

import (
	"bytes"
	"go-todo/internal/models"
	"go-todo/internal/server/logger"
	"html/template"
)

type Renderer struct {
	tmpl   *template.Template
	logger *logger.Logger
}

func NewRenderer(tmpl *template.Template, logger *logger.Logger) *Renderer {
	return &Renderer{
		tmpl:   tmpl,
		logger: logger,
	}
}

func (r *Renderer) render(templateName string, data any) ([]byte, error) {
	var buffer bytes.Buffer
	err := r.tmpl.ExecuteTemplate(&buffer, templateName, data)
	if err != nil {
		r.logger.Debug(err.Error())
		return nil, err
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
		r.logger.Error("Could not render Homepage")
	}
	return bytes, err
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
		r.logger.Error("Could not render signup page")
	}
	return bytes, err
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
		r.logger.Error("Could not render upgrade page.")
	}
	return bytes, err
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
		r.logger.Error("Could not render success page")
	}
	return bytes, err
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
		r.logger.Error("Could not render cancel page")
	}
	return bytes, err
}

// partials

type TodoProps *models.Todo

func NewTodoProps(todo *models.Todo) TodoProps {
	return todo
}
func (r *Renderer) Todo(p TodoProps) ([]byte, error) {
	bytes, err := r.render("todo", p)
	if err != nil {
		r.logger.Error("Could not render todo element")
	}
	return bytes, err
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
		r.logger.Error("Could not render todo list element")
	}
	return bytes, err
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
		r.logger.Error("Could not render login form element")
	}
	return bytes, err
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
