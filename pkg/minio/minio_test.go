package minio

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	core "k8s.io/client-go/testing"
)

type stats struct {
	name string
	size int64
}

func (s stats) Name() string {
	return s.name
}
func (s stats) Size() int64 {
	return s.size
}
func (s stats) IsDir() bool {
	return false
}
func (s stats) ModTime() time.Time {
	return time.Time{}
}
func (s stats) Mode() os.FileMode {
	return 0
}
func (s stats) Sys() interface{} {
	return 0
}

func TestUploadFunction(t *testing.T) {
	minioConfig := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "minio-config",
			Namespace: "kubeless",
		},
		Data: map[string]string{
			"maxFileSize": "50Mi",
		},
	}
	// Fake successful job
	uploadFakeJob := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "kubeless",
			Name:      "upload-file",
		},
		Status: batchv1.JobStatus{
			Succeeded: 1,
		},
	}
	file := "/path/to/func.ext"
	checksum := "abcdefghijklm1234567890"

	stat := stats{
		name: "func.ext",
		size: 1000,
	}

	cli := &fake.Clientset{}
	cli.Fake.AddReactor("get", "jobs", func(action core.Action) (bool, runtime.Object, error) {
		return true, &uploadFakeJob, nil
	})
	cli.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		return true, &minioConfig, nil
	})

	// It should return a valid URL
	url, err := UploadFunction(file, checksum, stat, cli)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if url != fmt.Sprintf("http://minio.kubeless:9000/functions/%s.%s", path.Base(file), checksum) {
		t.Errorf("Unexpected url %s", url)
	}

	// It should throw an error if the file size is too big
	stat = stats{
		name: "func.ext",
		size: 52428801,
	}
	url, err = UploadFunction(file, checksum, stat, cli)
	if err == nil {
		t.Error("The function should return an error trying to upload a file too big")
	}
}
