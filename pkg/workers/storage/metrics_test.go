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

func (suite *MetricsTestSuite) TestPushSamplesBasics() {
	assert := suite.Require()

	tests := []struct {
		timestamp time.Time
		samples   []*MetricSample
		nMetrics  int
		nSamples  int
		earliest  time.Time
		latest    time.Time
	}{
		{
			timestamp: time.Date(2025, time.January, 1, 13, 0, 0, 0, time.UTC),
			samples: []*MetricSample{
				&MetricSample{
					Name:        "foo",
					Flag:        "c",
					Format:      "i",
					Description: "foo",
					Value:       float64(3.14),
				},
				&MetricSample{
					Name:        "bar",
					Flag:        "g",
					Format:      "i",
					Description: "bar",
					Value:       uint64(42),
				},
			},
			nMetrics: 2,
			nSamples: 2,
			earliest: time.Date(2025, time.January, 1, 13, 0, 0, 0, time.UTC),
			latest:   time.Date(2025, time.January, 1, 13, 0, 0, 0, time.UTC),
		},
		{
			timestamp: time.Date(2025, time.January, 1, 13, 0, 5, 0, time.UTC),
			samples: []*MetricSample{
				&MetricSample{
					Name:        "foo",
					Flag:        "c",
					Format:      "i",
					Description: "foo changed description",
					Value:       float64(2.71),
				},
			},
			nMetrics: 2,
			nSamples: 3,
			earliest: time.Date(2025, time.January, 1, 13, 0, 0, 0, time.UTC),
			latest:   time.Date(2025, time.January, 1, 13, 0, 5, 0, time.UTC),
		},
	}

	for _, test := range tests {
		err := suite.stg.PushMetricSamples(test.timestamp, test.samples)
		assert.NoError(err)
		assert.Len(suite.stg.cache.metricsByID, test.nMetrics)
		assert.Len(suite.stg.cache.metricsByName, test.nMetrics)
		assert.Equal(suite.stg.cache.earliest, test.earliest)
		assert.Equal(suite.stg.cache.latest, test.latest)

		for _, sample := range test.samples {
			var sampleClass string
			switch sample.Value.(type) {
			case uint64:
				sampleClass = "uint64"
			case float64:
				sampleClass = "float64"
			}

			metric, ok := suite.stg.cache.metricsByName[sample.Name]
			assert.True(ok)
			assert.Equal(sample.Name, metric.Name)
			assert.Equal(sample.Flag, metric.Flag)
			assert.Equal(sample.Format, metric.Format)
			assert.Equal(sample.Description, metric.Description)
			assert.Equal(sampleClass, metric.Class)

			var name, flag, format, description, class string
			err = suite.stg.db.QueryRow(`
				SELECT name, flag, format, description, class
				FROM metrics
				WHERE id = $1`, metric.ID).Scan(&name, &flag, &format, &description, &class)
			assert.NoError(err)
			assert.Equal(sample.Name, name)
			assert.Equal(sample.Flag, flag)
			assert.Equal(sample.Format, format)
			assert.Equal(sample.Description, description)
			assert.Equal(sampleClass, class)

			var value interface{}
			err = suite.stg.db.QueryRow(`
				SELECT value.`+sampleClass+`
				FROM metric_values
				WHERE metric_id = $1 AND timestamp = $2`, metric.ID, test.timestamp).Scan(&value)
			assert.NoError(err)
			assert.Equal(sample.Value, value)
		}

		var nRows int
		err = suite.stg.db.QueryRow("SELECT COUNT(*) FROM metrics").Scan(&nRows)
		assert.NoError(err)
		assert.Equal(test.nMetrics, nRows)

		err = suite.stg.db.QueryRow("SELECT COUNT(*) FROM metric_values").Scan(&nRows)
		assert.NoError(err)
		assert.Equal(test.nSamples, nRows)
	}

	assert.Len(suite.stg.app.Cfg().Log().Buffer().Events(), 0)
}

func TestMetricsTestSuite(t *testing.T) {
	suite.Run(t, &MetricsTestSuite{})
}
