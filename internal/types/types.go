package types

import (
	"time"
)

type CLICommandResult struct {
	ExitCode int    `json:"exit_code"`
	Command  string `json:"command"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

type Lesson struct {
	ID          string    `json:"id"`
	CliCommands []string  `json:"cli_commands"`
	Tasks       []Task    `json:"tasks"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Task struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Status        string    `json:"status"`
	AiExplanation string    `json:"ai_explanation"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SubmitLessonRequestType string

const (
	SubmitLessonRequestTypeCommandResults SubmitLessonRequestType = "COMMAND_RESULTS"
)

type SubmitLessonRequest struct {
	Type              SubmitLessonRequestType `json:"type"`
	CliCommandResults []CLICommandResult      `json:"cli_command_results"`
}

type ActiveLesson struct {
	LessonID       string `json:"lessonId"`
	UserLessonID   string `json:"userLessonId"`
	LessonToken    string `json:"lessonToken"`
	Title          string `json:"title"`
	CourseTitle    string `json:"courseTitle"`
	LastAccessedAt string `json:"lastAccessedAt"`
}
