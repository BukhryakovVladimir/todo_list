package model

// Структура для добавления и для отметки о выполнении задачи (task_description)
type Task struct {
	TaskDescription string `json:"task_description"`
	IsCompleted     bool   `json:"is_completed"`
}
