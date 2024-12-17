package mtcapi

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/morethancertified/mtc-cli/internal/types"
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

func (c *MtcApiClient) GetLesson(lessonToken string) (types.Lesson, error) {
	res, err := c.httpClient.R().
		// SetDebug(true).
		SetResult(&types.Lesson{}).
		Get("/lessons/" + lessonToken)
	if err != nil {
		return types.Lesson{}, err
	}
	return *res.Result().(*types.Lesson), nil
}

func (c *MtcApiClient) SubmitLesson(lessonToken string, cliCommandResults []types.CLICommandResult) (types.Lesson, error) {
	res, err := c.httpClient.R().
		// SetDebug(true).
		SetBody(types.SubmitLessonRequest{
			Type:              types.SubmitLessonRequestTypeCommandResults,
			CliCommandResults: cliCommandResults,
		}).
		SetResult(&types.Lesson{}).
		Post("/lessons/" + lessonToken + "/submit")
	if err != nil {
		return types.Lesson{}, err
	}

	if res.IsError() {
		return types.Lesson{}, fmt.Errorf("%s", res.String())
	}

	return *res.Result().(*types.Lesson), nil
}

func (c *MtcApiClient) ResetLesson(lessonToken string) (types.Lesson, error) {
	res, err := c.httpClient.R().
		SetResult(&types.Lesson{}).
		Post("/lessons/" + lessonToken + "/reset")
	if err != nil {
		return types.Lesson{}, err
	}
	return *res.Result().(*types.Lesson), nil
}

func ValidCUID(cuid string) bool {
	return len(cuid) >= 7 && strings.HasPrefix(cuid, "c")
}
