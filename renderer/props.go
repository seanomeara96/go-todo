package renderer

import "go-todo/models"

// partials

type TodoProps *models.Todo

func NewTodoProps(todo *models.Todo) TodoProps {
	return todo
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

// pages

type BasePageProps struct {
	User *models.User
}

func NewBasePageProps(user *models.User) BasePageProps {
	return BasePageProps{
		User: user,
	}
}

type HomePageLoggedOutProps struct {
	BasePageProps
}

func NewHomePageLoggedOutProps(basePageProps BasePageProps) HomePageLoggedOutProps {
	return HomePageLoggedOutProps{
		BasePageProps: basePageProps,
	}
}

type HomePageLoggedInProps struct {
	BasePageProps
	TodoListProps TodoListProps
}

func NewHomePageLoggedInProps(basePageProps BasePageProps, todoListProps TodoListProps) HomePageLoggedInProps {
	return HomePageLoggedInProps{
		BasePageProps: basePageProps,
		TodoListProps: todoListProps,
	}
}

type SignupPageProps struct {
	BasePageProps
}

func NewSignupPageProps(basePageProps BasePageProps) SignupPageProps {
	return SignupPageProps{
		BasePageProps: basePageProps,
	}
}

type UpgradePageProps struct {
	BasePageProps
}

func NewUpgradePageProps(basePageProps BasePageProps) UpgradePageProps {
	return UpgradePageProps{
		BasePageProps: basePageProps,
	}
}

type SuccessPageProps struct {
	BasePageProps
}

func NewSuccessPageProps(basePageProps BasePageProps) SuccessPageProps {
	return SuccessPageProps{
		BasePageProps: basePageProps,
	}
}

type CancelPageProps struct {
	BasePageProps
}

func NewCancelPageProps(basePageProps BasePageProps) CancelPageProps {
	return CancelPageProps{
		BasePageProps: basePageProps,
	}
}
