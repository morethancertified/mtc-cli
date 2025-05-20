package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/morethancertified/mtc-cli/internal/mtcapi"
	"github.com/morethancertified/mtc-cli/internal/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:     "init <lesson-token>",
	Short:   "Initialize a lab environment",
	Args:    cobra.ExactArgs(1),
	Example: "mtc init cm4ppz694200blze51ts1234",
	Run: func(cmd *cobra.Command, args []string) {
		lessonToken := args[0]
		publicOnly, _ := cmd.Flags().GetBool("public-only")
		
		apiClient := mtcapi.New(viper.GetString("api_base_url"))
		
		// Get lab info
		fmt.Println("Fetching lab information...")
		labInfo, err := apiClient.GetLabInfo(lessonToken)
		if err != nil {
			fmt.Printf("Error getting lab information: %s\n", err)
			return
		}
		
		fmt.Printf("Initializing lab: %s\n", labInfo.Title)
		
		// Create lab directory
		labDir := sanitizeDirectoryName(labInfo.Title)
		if dirFlag, _ := cmd.Flags().GetString("dir"); dirFlag != "" {
			labDir = dirFlag
		}
		
		// Clean up the directory name
		labDir = filepath.Clean(labDir)
		
		// Create the directory if it doesn't exist
		if _, err := os.Stat(labDir); os.IsNotExist(err) {
			if err := os.MkdirAll(labDir, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %s\n", labDir, err)
				return
			}
		}
		
		// Get file listing
		fmt.Println("Fetching lab files...")
		var files []types.LabFile
		
		if publicOnly {
			files, err = apiClient.GetLabPublicFiles(lessonToken)
		} else {
			files, err = apiClient.GetLabFiles(lessonToken)
		}
		
		if err != nil {
			fmt.Printf("Error getting lab files: %s\n", err)
			return
		}
		
		// Download files
		fmt.Printf("Downloading %d files...\n", len(files))
		
		for i, file := range files {
			// Determine the target path
			var filePath string
			
			// If it's a public file, extract it to the root of the lab directory
			if strings.HasPrefix(file.Path, "public/") {
				// Remove the "public/" prefix
				targetPath := strings.TrimPrefix(file.Path, "public/")
				filePath = filepath.Join(labDir, targetPath)
				fmt.Printf("[%d/%d] Downloading %s to %s...\n", i+1, len(files), file.Path, targetPath)
			} else {
				// Keep the original path for other files
				filePath = filepath.Join(labDir, file.Path)
				fmt.Printf("[%d/%d] Downloading %s...\n", i+1, len(files), file.Path)
			}
			
			// Create subdirectories if needed
			fileDir := filepath.Dir(filePath)
			if err := os.MkdirAll(fileDir, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %s\n", fileDir, err)
				continue
			}
			
			// Download file
			if err := downloadFile(file.URL, filePath); err != nil {
				fmt.Printf("Error downloading %s: %s\n", file.Path, err)
				continue
			}
		}
		
		// No need to save lab metadata
		
		fmt.Printf("\nLab initialized successfully in %s\n", labDir)
		fmt.Println("You can now cd into the directory and start working on the lab.")
	},
}

func downloadFile(url, filePath string) error {
	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	
	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// sanitizeDirectoryName cleans up a string to be used as a directory name
func sanitizeDirectoryName(name string) string {
	// Remove quotes
	name = strings.Trim(name, "\"'`")
	
	// Replace problematic characters with underscores
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_", // Replace spaces with underscores
	)
	
	// Ensure the name doesn't have any remaining problematic characters
	sanitized := replacer.Replace(name)
	
	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)
	
	// If the name is empty after sanitization, use a default name
	if sanitized == "" {
		return "lab"
	}
	
	return sanitized
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("public-only", "p", false, "Download only public files")
	initCmd.Flags().StringP("dir", "d", "", "Directory to initialize the lab in (defaults to lab title)")
}
