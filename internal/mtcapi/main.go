package mtcapi

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/morethancertified/mtc-cli/internal/auth"
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

// ensureAuthenticated checks for a valid token and sets it on the request
func (c *MtcApiClient) ensureAuthenticated(req *resty.Request) error {
	token, err := auth.GetToken()
	if err != nil {
		return fmt.Errorf("authentication required: please run 'sprintctl login' first")
	}
	
	req.SetAuthToken(token)
	return nil
}

func (c *MtcApiClient) GetLesson(lessonToken string) (types.Lesson, error) {
	req := c.httpClient.R().
		SetResult(&types.Lesson{})
	
	if err := c.ensureAuthenticated(req); err != nil {
		return types.Lesson{}, err
	}
	
	res, err := req.Get("/lessons/" + lessonToken)
	if err != nil {
		return types.Lesson{}, err
	}

	results := *res.Result().(*types.Lesson)

	if len(results.Tasks) == 0 {
		return types.Lesson{}, fmt.Errorf("token is invalid")
	}

	return results, nil
}

func (c *MtcApiClient) SubmitLesson(lessonToken string, cliCommandResults []types.CLICommandResult) (types.Lesson, error) {
	req := c.httpClient.R().
		SetBody(types.SubmitLessonRequest{
			Type:              types.SubmitLessonRequestTypeCommandResults,
			CliCommandResults: cliCommandResults,
		}).
		SetResult(&types.Lesson{})
	
	if err := c.ensureAuthenticated(req); err != nil {
		return types.Lesson{}, err
	}
	
	res, err := req.Post("/lessons/" + lessonToken + "/submit")
	if err != nil {
		return types.Lesson{}, err
	}

	if res.IsError() {
		return types.Lesson{}, fmt.Errorf("%s", res.String())
	}

	return *res.Result().(*types.Lesson), nil
}

func (c *MtcApiClient) ResetLesson(lessonToken string) (types.Lesson, error) {
	req := c.httpClient.R().
		SetResult(&types.Lesson{})
	
	if err := c.ensureAuthenticated(req); err != nil {
		return types.Lesson{}, err
	}
	
	res, err := req.Post("/lessons/" + lessonToken + "/reset")
	if err != nil {
		return types.Lesson{}, err
	}
	return *res.Result().(*types.Lesson), nil
}

func ValidCUID(cuid string) bool {
	return len(cuid) >= 7 && strings.HasPrefix(cuid, "c")
}
