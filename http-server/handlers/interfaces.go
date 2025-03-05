package handlers

type ServerJob interface {
	AddTask(date string, title string, comment string, repeat string) (string, error)
	GetTasks(NumberOfOutPutTasks int) ([]Task, error)
	GetTask(id string) (Task, error)
	UpdateTask(task Task) error
	DeleteTask(idTask string) error
	SearchTasks(code int, searchQuery string, NumberOfOutPutTasks int) ([]Task, error)
}
