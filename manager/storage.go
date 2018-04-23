package manager

import (
	"fmt"
	"io"
	"time"

	minio "github.com/minio/minio-go"
)

func (m *Manager) uploadJob(name string, size int64, file io.Reader) (string, error) {
	jobName := fmt.Sprintf("%s-%s", name, time.Now().Format(time.RFC3339))
	metadata := map[string]string{
		"filename": name,
		"size":     fmt.Sprintf("%d", size),
	}
	if _, err := m.mc.PutObject(
		storageJobsBucketName,
		jobName,
		file,
		size,
		minio.PutObjectOptions{
			ContentType:  "application/octet-stream",
			UserMetadata: metadata,
		}); err != nil {
		return "", err
	}
	return jobName, nil
}
