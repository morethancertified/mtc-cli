package mtcapi

import (
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type MtcApiClient struct {
	BaseURL    string
	httpClient *resty.Client
}

func New(baseURL string) *MtcApiClient {
	httpClient := resty.New()
	httpClient.SetBaseURL(baseURL)

	return &MtcApiClient{
		BaseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *MtcApiClient) GetLesson(lessonToken string) (Lesson, error) {
	res, err := c.httpClient.R().
		// SetDebug(true).
		SetResult(&Lesson{}).
		Get("/lessons/" + lessonToken)
	if err != nil {
		return Lesson{}, err
	}
	return *res.Result().(*Lesson), nil
}

func ValidCUID(cuid string) bool {
	return len(cuid) >= 7 && strings.HasPrefix(cuid, "c")
}

type Lesson struct {
	ID          string    `json:"id"`
	CliCommands []string  `json:"cli_commands"`
	Tasks       []Task    `json:"tasks"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
