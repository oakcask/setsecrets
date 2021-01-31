package gcp

import (
	"context"
	"encoding/json"
	"fmt"

	gopipeline "github.com/mattn/go-pipeline"
	googleauth "golang.org/x/oauth2/google"
)

type gcloudConfigOutput struct {
	Core map[string]string `json:"core"`
}

func getProjectID(ctx context.Context) (string, error) {
	creds, e := googleauth.FindDefaultCredentials(ctx)
	if e != nil {
		return "", e
	}

	if creds.ProjectID != "" {
		return creds.ProjectID, nil
	}

	gcloudOutJSON, e := gopipeline.Output(
		[]string{"gcloud", "-q", "config", "list", "core/project", "--format=json"},
	)
	if e != nil {
		return "", fmt.Errorf("failed to invoke gcloud command: %v", e.Error())
	}
	var out gcloudConfigOutput
	if e = json.Unmarshal(gcloudOutJSON, &out); e != nil {
		return "", fmt.Errorf("failed to parse gcloud command output")
	}

	return out.Core["project"], nil
}
