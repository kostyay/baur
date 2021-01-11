package baur

import (
	"fmt"
	"path/filepath"

	"github.com/simplesurance/baur/v1/internal/digest"
)

type OutputType int

const (
	DockerOutput OutputType = iota
	FileOutput
)

func (o OutputType) String() string {
	switch o {
	case DockerOutput:
		return "docker"
	case FileOutput:
		return "file"

	default:
		return "invalid OutputType"
	}
}

type Output interface {
	Name() string
	String() string
	Exists() (bool, error)
	Size() (uint64, error)
	Digest() (*digest.Digest, error)
	Type() OutputType
}

func dockerOutputs(dockerClient DockerInfoClient, task *Task) ([]Output, error) {
	result := make([]Output, 0, len(task.Outputs.DockerImage))

	for _, dockerOutput := range task.Outputs.DockerImage {
		uploadInfos := make([]*UploadInfoDocker, 0, len(dockerOutput.RegistryUpload))

		for _, ru := range dockerOutput.RegistryUpload {
			uploadInfos = append(uploadInfos, &UploadInfoDocker{
				Registry:   ru.Registry,
				Repository: ru.Repository,
				Tag:        ru.Tag,
			})
		}

		d, err := NewOutputDockerImageFromIIDFile(
			dockerClient,
			dockerOutput.IDFile,
			filepath.Join(task.Directory, dockerOutput.IDFile),
			uploadInfos,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, d)
	}

	return result, nil
}

func fileOutputs(task *Task) ([]Output, error) {
	result := make([]Output, 0, len(task.Outputs.File))

	for _, fileOutput := range task.Outputs.File {
		var s3Uploads []*UploadInfoS3
		var fileCopyUploads []*UploadInfoFileCopy

		if len(fileOutput.S3Upload) == 0 && len(fileOutput.FileCopy) == 0 {
			return nil, fmt.Errorf("no upload method for output %q is specified", fileOutput.Path)
		}

		for _, s3 := range fileOutput.S3Upload {
			s3Uploads = append(s3Uploads, &UploadInfoS3{
				Bucket: s3.Bucket,
				Key:    s3.Key,
			})
		}

		for _, fc := range fileOutput.FileCopy {
			fileCopyUploads = append(fileCopyUploads, &UploadInfoFileCopy{DestinationPath: fc.Path})
		}

		result = append(result, NewOutputFile(
			fileOutput.Path,
			filepath.Join(task.Directory, fileOutput.Path),
			s3Uploads,
			fileCopyUploads,
		))
	}

	return result, nil
}

// OutputsFromTask returns the Outputs that running the task produces.
// If the outputs do not exist, the function might fail.
func OutputsFromTask(dockerClient DockerInfoClient, task *Task) ([]Output, error) {
	dockerImages, err := dockerOutputs(dockerClient, task)
	if err != nil {
		return nil, err
	}

	files, err := fileOutputs(task)
	if err != nil {
		return nil, err
	}

	result := make([]Output, 0, len(dockerImages)+len(files))
	result = append(result, dockerImages...)
	result = append(result, files...)

	return result, nil
}
