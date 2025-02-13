package storage

import (
	"testing"
	"time"

	"github.com/allenta/varnishmon/pkg/testutil"
	"github.com/stretchr/testify/suite"
)

type MetricsTestSuite struct {
	suite.Suite
	testName string
	stg      *Storage
}

func (suite *MetricsTestSuite) BeforeTest(suiteName, testName string) {
	suite.testName = testName

	app := new(MockApplication)
	app.
		On("Cfg").
		Return(testutil.NewConfig(
			suite.T(),
			"global.loglevel", "error",
			"scraper.enabled", false,
			"api.enabled", false,
			"db.file", ""))
	suite.stg = NewStorage(app)
}

func (suite *MetricsTestSuite) TestNormalizeFromTo() {
	assert := suite.Require()

	tests := []struct {
		from           time.Time
		to             time.Time
		step           int
		normalizedFrom time.Time
		normalizedTo   time.Time
		normalizedStep int
	}{
		{
			from:           time.Date(2025, time.January, 1, 13, 4, 59, 0, time.UTC),
			to:             time.Date(2025, time.January, 1, 13, 56, 30, 0, time.UTC),
			step:           300,
			normalizedFrom: time.Date(2025, time.January, 1, 13, 0, 0, 0, time.UTC),
			normalizedTo:   time.Date(2025, time.January, 1, 14, 0, 0, 0, time.UTC),
			normalizedStep: 300,
		},
		{
			from:           time.Date(2025, time.January, 1, 13, 4, 59, 0, time.UTC),
			to:             time.Date(2025, time.January, 1, 14, 0, 30, 0, time.UTC),
			step:           300,
			normalizedFrom: time.Date(2025, time.January, 1, 13, 0, 0, 0, time.UTC),
			normalizedTo:   time.Date(2025, time.January, 1, 14, 5, 0, 0, time.UTC),
			normalizedStep: 300,
		},
		{
			from:           time.Date(2025, time.January, 1, 13, 0, 0, 0, time.UTC),
			to:             time.Date(2025, time.January, 1, 14, 0, 0, 0, time.UTC),
			step:           300,
			normalizedFrom: time.Date(2025, time.January, 1, 13, 0, 0, 0, time.UTC),
			normalizedTo:   time.Date(2025, time.January, 1, 14, 5, 0, 0, time.UTC),
			normalizedStep: 300,
		},
		{
			from:           time.Date(2025, time.January, 1, 13, 4, 59, 0, time.UTC),
			to:             time.Date(2025, time.January, 1, 13, 56, 30, 0, time.UTC),
			step:           0,
			normalizedFrom: time.Date(2025, time.January, 1, 13, 4, 59, 0, time.UTC),
			normalizedTo:   time.Date(2025, time.January, 1, 13, 56, 31, 0, time.UTC),
			normalizedStep: 1,
		},
		{
			from:           time.Date(2025, time.January, 1, 13, 4, 59, 0, time.UTC),
			to:             time.Date(2025, time.January, 1, 13, 56, 30, 0, time.UTC),
			step:           -100,
			normalizedFrom: time.Date(2025, time.January, 1, 13, 4, 59, 0, time.UTC),
			normalizedTo:   time.Date(2025, time.January, 1, 13, 56, 31, 0, time.UTC),
			normalizedStep: 1,
		},
	}

	suite.stg.mutex.RLock()
	defer suite.stg.mutex.RUnlock()
	for _, test := range tests {
		from, to, step, err := suite.stg.unsafeNormalizeFromToAndStep(test.from, test.to, test.step)
		assert.Equal(test.normalizedFrom, from)
		assert.Equal(test.normalizedTo, to)
		assert.Equal(test.normalizedStep, step)
		assert.NoError(err)

		logBuffer := suite.stg.app.Cfg().Log().Buffer()
		events := logBuffer.Events()
		assert.Len(events, 0)
		logBuffer.Clear()
	}
}

func TestMetricsTestSuite(t *testing.T) {
	suite.Run(t, &MetricsTestSuite{})
}
