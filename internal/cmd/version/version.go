package version

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	cli "github.com/urfave/cli/v3"
)

// VersionInfo represents the application version information
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	Runtime   string `json:"runtime"`
}

// DisplayVersion outputs the version information as JSON
func DisplayVersion(ctx context.Context, c *cli.Command) error {
	versionInfo := VersionInfo{
		Version:   ctx.Value("buildVersion").(string),
		Commit:    ctx.Value("buildCommit").(string),
		BuildDate: ctx.Value("buildDate").(string),
		Runtime:   runtime.Version(),
	}

	jsonData, err := json.MarshalIndent(versionInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling version info to JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}
