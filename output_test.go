package engine

import (
	"github.com/stretchr/testify/assert"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewOutput(t *testing.T) {
	assert.IsType(t, NewOutput(), &Output{})
}

func TestOutput_ParseOutputReportToJSONFile(t *testing.T) {
	t.Run("Should not return error when report is nil", func(t *testing.T) {
		outputPath := "./tmp.json"
		assert.NoError(t, NewOutput().ParseOutputReportToJSONFile(nil, outputPath))
		content, err := ioutil.ReadFile(outputPath)
		assert.NoError(t, err)
		assert.Equal(t, string(content), "[]")
		assert.NoError(t, os.RemoveAll(outputPath))
	})
	t.Run("Should not return error when report exist content is nil", func(t *testing.T) {
		outputPath := "./tmp.json"
		ID := uuid.New().String()
		report := []Report{
			{
				ID:             ID,
				Name:           uuid.New().String(),
				Description:    uuid.New().String(),
				SourceLocation: Location{
					Filename: "/tmp",
					Line:     1,
					Column:   1,
				},
			},
		}
		assert.NoError(t, NewOutput().ParseOutputReportToJSONFile(report, outputPath))
		content, err := ioutil.ReadFile(outputPath)
		assert.NoError(t, err)
		assert.Contains(t, string(content), ID)
		assert.NoError(t, os.RemoveAll(outputPath))
	})
	t.Run("Should return error because path is invalid", func(t *testing.T) {
		outputPath := "./////"
		assert.Error(t, NewOutput().ParseOutputReportToJSONFile(nil, outputPath))
	})
}