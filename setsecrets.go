package setsecrets

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// Provider interface
type Provider interface {
	// Concurrency will be used to limit the number of go-routines
	// being run simultaneously for retrieving secrets.
	Concurrency() int
	// GetSecret fetches a secret value pointed by the key.
	// Providers that can return greater than 1 from Concurrency method,
	// must implement this method as thread-safe.
	GetSecret(ctx context.Context, key string) (string, error)
}

// SetSecrets sets environment variable by retrieved secrets with provider.
// secretMaps is a map from secret name to environment variable.
// The format of secret name will differ by provier implementations.
func SetSecrets(ctx context.Context, provider Provider, secretMaps map[string]string) error {
	var keys []string
	for key := range secretMaps {
		keys = append(keys, key)
	}

	secrets, e := LoadSecrets(ctx, provider, keys)
	if e != nil {
		return e
	}

	for key, secret := range secrets {
		e = os.Setenv(secretMaps[key], string(secret))
		if e != nil {
			return e
		}
	}

	return nil
}

// LoadSecrets returns a map which contains pairs of
// secret name and the secret value.
func LoadSecrets(ctx context.Context, provider Provider, keys []string) (map[string]string, error) {
	concurrency := provider.Concurrency()
	if concurrency > len(keys) {
		return parallelLoadSecrets(ctx, len(keys), provider, keys)
	}
	if concurrency > 1 {
		return parallelLoadSecrets(ctx, concurrency, provider, keys)
	}

	return parallelLoadSecrets(ctx, 1, provider, keys)
}

type getSecretResult struct {
	key    string
	secret string
	err    error
}

func parallelLoadSecrets(ctx context.Context, concurrency int, provider Provider, keys []string) (map[string]string, error) {
	if concurrency < 1 {
		concurrency = 1
	}

	// semaphore limits concurrency of calling provider.GetSecret
	semaphore := make(chan struct{}, concurrency)

	// set buffer size as same as len(secrets)
	// so we can keep inserting all the results.
	channel := make(chan getSecretResult, len(keys))

	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for _, key := range keys {
		wg.Add(1)

		go func(ctx context.Context, key string) {
			defer wg.Done()

			semaphore <- struct{}{}
			// ensure returning "ticket" to semaphore
			// in order to prevent from deadlocking
			defer func() {
				<-semaphore
			}()

			secret, err := provider.GetSecret(ctx, key)
			channel <- getSecretResult{
				key:    key,
				secret: secret,
				err:    err,
			}
			if err != nil {
				cancel()
			}
		}(ctx1, key)
	}
	go func() {
		wg.Wait()
		close(channel)
	}()

	result := map[string]string{}

	for out := range channel {
		if out.err != nil {
			return nil, out.err
		}
		result[out.key] = out.secret
	}
	for _, key := range keys {
		if _, ok := result[key]; !ok {
			return nil, fmt.Errorf("unexpectedly lost secret for key %s", key)
		}
	}

	return result, nil
}
