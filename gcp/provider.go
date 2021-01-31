package gcp

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/oakcask/setsecrets"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type provider struct {
	projectID string
}

// New returns the object implements Provider which utilizes
// Secret Manager on Google Cloud Platform.
func New(ctx context.Context) (setsecrets.Provider, error) {
	projectID, e := getProjectID(ctx)
	if e != nil {
		return nil, e
	}

	return &provider{
		projectID: projectID,
	}, nil
}

func (*provider) Concurrency() int {
	// this is not optimal
	return runtime.NumCPU() * 2
}

func (p *provider) GetSecret(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("empty path to secret specified")
	}

	if strings.HasPrefix(key, "projects/") {
		return doGetSecret(ctx, key)
	}

	return doGetSecret(ctx, fmt.Sprintf("projects/%s/%s", p.projectID, key))
}

func (*provider) IsThreadSafe() bool {
	return true
}

func doGetSecret(ctx context.Context, key string) (string, error) {
	client, e := secretmanager.NewClient(ctx)
	if e != nil {
		return "", e
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: key,
	}
	secretVersion, e := client.AccessSecretVersion(ctx, req)
	if e != nil {
		return "", e
	}

	return string(secretVersion.Payload.Data), nil
}
