package manager

import (
	"fmt"
	"io"
	"os"
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

func (m *Manager) uploadRender(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	fi, err := file.Stat()
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := m.mc.PutObject(
		storageResultsBucketName,
		dest,
		file,
		fi.Size(),
		minio.PutObjectOptions{
			ContentType: "image/png",
		}); err != nil {
		return err
	}

	return nil
}
