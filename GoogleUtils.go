package Configuration

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"context"
	"fmt"
	"hash/crc32"
)

//func main() {
//	name := "projects/286247678501/secrets/knowme-authentication-key/versions/latest"
//	secretValue, err := AccessSecretVersion(name)
//	if err != nil {
//		log.Error().Msgf("Error :%v", err)
//	}
//	log.Info().Msgf("Secret :%v", secretValue)
//}

func AccessSecretVersion(name string) (string, error) {
	secretValue := ""
	// Create the client.
	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return secretValue, fmt.Errorf("failed to create secretmanager client: %w", err)
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return secretValue, fmt.Errorf("failed to access secret version: %w", err)
	}

	// Verify the data checksum.
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
	if checksum != *result.Payload.DataCrc32C {
		return secretValue, fmt.Errorf("Data corruption detected.")
	}

	secretValue = string(result.Payload.Data)
	return secretValue, nil
}
