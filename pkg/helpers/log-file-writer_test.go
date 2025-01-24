package helpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LogFileWriterTestSuite struct {
	suite.Suite
	tmpDir   string
	fileName string
}

func (suite *LogFileWriterTestSuite) BeforeTest(suiteName, testName string) {
	// Build temporal log file name.
	suite.fileName = filepath.Join(suite.tmpDir, testName+".log")

	// Clean up any previous trace of the log file.
	os.Remove(suite.fileName)
}

func (suite *LogFileWriterTestSuite) TearDownTest() {
	// Clean up: remove log file.
	os.Remove(suite.fileName)
}

func (suite *LogFileWriterTestSuite) TestInitializationAndClose() {
	assert := suite.Require()

	// Initialize log file writer.
	log, err := NewLogFileWriter(suite.fileName, 0744, false)
	assert.Nil(err)
	assert.Equal(suite.fileName, log.name)
	assert.Equal(os.FileMode(0x1e4), log.mode)
	assert.Equal(false, log.redirect)
	assert.NotNil(log.file)

	// Close log file.
	err = log.Close()
	assert.Nil(err)

	// Make sure file is closed.
	err = log.file.Close()
	assert.IsType((*os.PathError)(nil), err)
}

func (suite *LogFileWriterTestSuite) TestAppendToFile() {
	assert := suite.Require()

	// Create log file with some contents.
	file, err := os.Create(suite.fileName)
	assert.Nil(err)
	file.Write([]byte("This was already in the file"))
	err = file.Close()
	assert.Nil(err)

	// Initialize log file writer.
	log, err := NewLogFileWriter(suite.fileName, 0744, false)
	assert.Nil(err)

	// Write to log.
	log.Write([]byte("This was written to LogFileWriter"))

	// Check log contents.
	file, err = os.Open(suite.fileName)
	assert.Nil(err)
	buffer := make([]byte, 500)
	nBytes, err := file.Read(buffer)
	assert.Nil(err)
	expectedResult := "This was already in the file" +
		"This was written to LogFileWriter"
	assert.Equal(expectedResult, string(buffer[:nBytes]))
	assert.Equal(len(expectedResult), nBytes)

	// Close log file.
	err = log.Close()
	assert.Nil(err)
}

func (suite *LogFileWriterTestSuite) TestRedirect() {
	assert := suite.Require()

	// Initialize log file writer.
	log, err := NewLogFileWriter(suite.fileName, 0744, true)
	assert.Nil(err)

	// Write to stdout, stderr and log.
	os.Stdout.Write([]byte("This was written to stdout"))
	os.Stderr.Write([]byte("This was written to stderr"))
	log.Write([]byte("This was written to LogFileWriter"))

	// Check log contents.
	file, err := os.Open(log.name)
	assert.Nil(err)
	buffer := make([]byte, 500)
	nBytes, err := file.Read(buffer)
	assert.Nil(err)
	expectedResult := "This was written to stdout" +
		"This was written to stderr" + "This was written to LogFileWriter"
	assert.Equal(expectedResult, string(buffer[:nBytes]))
	assert.Equal(len(expectedResult), nBytes)

	// Close log file.
	err = log.Close()
	assert.Nil(err)
}

func (suite *LogFileWriterTestSuite) TestReopen() {
	assert := suite.Require()

	// Initialize log file writer.
	log, err := NewLogFileWriter(suite.fileName, 0744, false)
	assert.Nil(err)

	// Reopen log and check file changes.
	old_file := log.file
	log.Reopen()
	assert.NotEqual(old_file, log.file)

	// Check old file is already closed.
	err = old_file.Close()
	assert.IsType((*os.PathError)(nil), err)

	// Close log file.
	err = log.Close()
	assert.Nil(err)
}

func TestLogFileWriterTestSuite(t *testing.T) {
	suite.Run(t, &LogFileWriterTestSuite{
		tmpDir: t.TempDir(),
	})
}
