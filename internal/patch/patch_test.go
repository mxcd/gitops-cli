package patch

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanRelativeFilePath(t *testing.T) {
	testCases := []struct {
		name        string
		filePath    string
		expected    string
		expectError bool
	}{
		{name: "plain relative path", filePath: "applications/dev/service-foo/values.yaml", expected: "applications/dev/service-foo/values.yaml"},
		{name: "path is cleaned", filePath: "applications/./dev/../dev/values.yaml", expected: "applications/dev/values.yaml"},
		{name: "traversal inside the repository is allowed", filePath: "a/../b.yaml", expected: "b.yaml"},
		{name: "empty path", filePath: "", expectError: true},
		{name: "blank path", filePath: "   ", expectError: true},
		{name: "absolute path", filePath: "/etc/passwd", expectError: true},
		{name: "parent traversal", filePath: "../values.yaml", expectError: true},
		{name: "nested parent traversal", filePath: "a/../../values.yaml", expectError: true},
		{name: "leading dash", filePath: "-flag.yaml", expectError: true},
		{name: "leading dash hidden behind current directory", filePath: "./-flag.yaml", expectError: true},
		{name: "leading dash hidden behind parent traversal", filePath: "a/../-flag.yaml", expectError: true},
		{name: "leading colon pathspec magic", filePath: ":!values.yaml", expectError: true},
		{name: "leading colon hidden behind current directory", filePath: "./:(top)values.yaml", expectError: true},
		{name: "current directory", filePath: ".", expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cleanedFilePath, err := cleanRelativeFilePath(testCase.filePath)
			if testCase.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, cleanedFilePath)
		})
	}
}

func TestValidatePatchBatch(t *testing.T) {
	validPatches := []Patch{{Selector: ".service.image.tag", Value: "v1.0.0"}}

	testCases := []struct {
		name        string
		batch       PatchBatch
		expectError bool
	}{
		{
			name: "valid batch with two files",
			batch: PatchBatch{
				Actor: "ci-bot",
				Files: []FilePatch{
					{FilePath: "applications/dev/service-a/values.yaml", Patches: validPatches},
					{FilePath: "applications/dev/service-b/values.yaml", Patches: validPatches},
				},
			},
		},
		{
			name: "valid batch without actor",
			batch: PatchBatch{
				Files: []FilePatch{{FilePath: "applications/dev/service-a/values.yaml", Patches: validPatches}},
			},
		},
		{
			name:        "no files",
			batch:       PatchBatch{Files: []FilePatch{}},
			expectError: true,
		},
		{
			name:        "nil files",
			batch:       PatchBatch{},
			expectError: true,
		},
		{
			name: "empty file path",
			batch: PatchBatch{
				Files: []FilePatch{{FilePath: "", Patches: validPatches}},
			},
			expectError: true,
		},
		{
			name: "absolute file path",
			batch: PatchBatch{
				Files: []FilePatch{{FilePath: "/etc/passwd", Patches: validPatches}},
			},
			expectError: true,
		},
		{
			name: "traversing file path",
			batch: PatchBatch{
				Files: []FilePatch{{FilePath: "../values.yaml", Patches: validPatches}},
			},
			expectError: true,
		},
		{
			name: "nested traversing file path",
			batch: PatchBatch{
				Files: []FilePatch{{FilePath: "a/../../values.yaml", Patches: validPatches}},
			},
			expectError: true,
		},
		{
			name: "file path starting with a dash",
			batch: PatchBatch{
				Files: []FilePatch{{FilePath: "-flag.yaml", Patches: validPatches}},
			},
			expectError: true,
		},
		{
			name: "duplicate file paths",
			batch: PatchBatch{
				Files: []FilePatch{
					{FilePath: "a/./b.yaml", Patches: validPatches},
					{FilePath: "a/b.yaml", Patches: validPatches},
				},
			},
			expectError: true,
		},
		{
			name: "file without patches",
			batch: PatchBatch{
				Files: []FilePatch{{FilePath: "applications/dev/service-a/values.yaml", Patches: []Patch{}}},
			},
			expectError: true,
		},
		{
			name: "patch without selector",
			batch: PatchBatch{
				Files: []FilePatch{{FilePath: "applications/dev/service-a/values.yaml", Patches: []Patch{{Selector: "", Value: "v1.0.0"}}}},
			},
			expectError: true,
		},
		{
			name: "second file is invalid",
			batch: PatchBatch{
				Files: []FilePatch{
					{FilePath: "applications/dev/service-a/values.yaml", Patches: validPatches},
					{FilePath: "../escape.yaml", Patches: validPatches},
				},
			},
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := ValidatePatchBatch(testCase.batch)
			if testCase.expectError {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidPatchBatch), "error should wrap ErrInvalidPatchBatch")
				return
			}
			assert.NoError(t, err)
		})
	}
}
