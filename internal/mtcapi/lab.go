package mtcapi

import (
	"fmt"

	"github.com/morethancertified/mtc-cli/internal/types"
)

// GetLabInfo fetches information about a lab
func (c *MtcApiClient) GetLabInfo(userLessonID string) (types.LabInfo, error) {
	res, err := c.httpClient.R().
		SetResult(&types.LabInfo{}).
		Get("/labs/" + userLessonID)
	if err != nil {
		return types.LabInfo{}, err
	}

	if res.IsError() {
		return types.LabInfo{}, fmt.Errorf("API error: %s", res.String())
	}

	return *res.Result().(*types.LabInfo), nil
}

// GetLabFiles fetches all files for a lab (public, bootstrap, and other)
func (c *MtcApiClient) GetLabFiles(userLessonID string) ([]types.LabFile, error) {
	res, err := c.httpClient.R().
		SetResult(&types.LabFiles{}).
		Get("/labs/" + userLessonID + "/files")
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, fmt.Errorf("API error: %s", res.String())
	}

	result := res.Result().(*types.LabFiles)
	var files []types.LabFile

	// Combine all file types
	files = append(files, result.Files.Public...)
	files = append(files, result.Files.Bootstrap...)
	files = append(files, result.Files.Other...)

	return files, nil
}

// GetLabPublicFiles fetches only public files for a lab
func (c *MtcApiClient) GetLabPublicFiles(userLessonID string) ([]types.LabFile, error) {
	res, err := c.httpClient.R().
		SetResult(&types.LabPublicFiles{}).
		Get("/labs/" + userLessonID + "/files/public")
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, fmt.Errorf("API error: %s", res.String())
	}

	result := res.Result().(*types.LabPublicFiles)
	return result.Files, nil
}

// GetLabFileURL fetches a pre-signed URL for a specific file
func (c *MtcApiClient) GetLabFileURL(userLessonID string, filePath string) (types.LabFileURL, error) {
	res, err := c.httpClient.R().
		SetResult(&types.LabFileURL{}).
		Get("/labs/" + userLessonID + "/files/" + filePath)
	if err != nil {
		return types.LabFileURL{}, err
	}

	if res.IsError() {
		return types.LabFileURL{}, fmt.Errorf("API error: %s", res.String())
	}

	return *res.Result().(*types.LabFileURL), nil
}
