package model

// Структура для изменения задачи (task_description)
type UpdateTask struct {
	TaskNumber      string `json:"task_number"`
	TaskDescription string `json:"task_description"`
}
