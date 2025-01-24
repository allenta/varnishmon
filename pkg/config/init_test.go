package config

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type InitTestSuite struct {
	suite.Suite
	cfg *Config
}

func (suite *InitTestSuite) BeforeTest(suiteName, testName string) {
	vpr := viper.New()
	vpr.Set("scraper.varnishstat", "/dev/null")
	suite.cfg = NewConfig(NewTestLogger(suite.T(), zerolog.ErrorLevel), vpr)
}

func (suite *InitTestSuite) TestCheckInt() {
	assert := suite.Require()

	for _, value := range []int{0, 42, 100} {
		suite.cfg.vpr.Set("foo", value)
		assert.NotPanics(func() {
			suite.cfg.checkInt("foo", 0, 100)
		})
	}

	for _, value := range []int{-1, 101} {
		suite.cfg.vpr.Set("foo", value)
		assert.Panics(func() {
			suite.cfg.checkInt("foo", 0, 100)
		})
	}
}

func (suite *InitTestSuite) TestCheckDuration() {
	assert := suite.Require()

	for _, value := range []string{"", "0s", "1ms", "42s", "1m", "100s"} {
		suite.cfg.vpr.Set("foo", value)
		assert.NotPanics(func() {
			suite.cfg.checkDuration("foo", 0*time.Second, 100*time.Second)
		})
	}

	for _, value := range []string{"-1s", "101s", "2m"} {
		suite.cfg.vpr.Set("foo", value)
		assert.Panics(func() {
			suite.cfg.checkDuration("foo", 0*time.Second, 100*time.Second)
		})
	}
}

func (suite *InitTestSuite) TestCheckLoglevel() {
	assert := suite.Require()

	for _, value := range []string{"", "debug", "info", "warn", "error", "fatal", "panic"} {
		suite.cfg.vpr.Set("foo", value)
		assert.NotPanics(func() {
			suite.cfg.checkLoglevel("foo")
		})
		_, ok := suite.cfg.vpr.Get("foo").(zerolog.Level)
		assert.True(ok)
	}

	suite.cfg.vpr.Set("foo", "whatever")
	assert.Panics(func() {
		suite.cfg.checkLoglevel("foo")
	})
}

func (suite *InitTestSuite) TestCheckIP() {
	assert := suite.Require()

	for _, value := range []string{"1.2.3.4", "2001:db8:3333:4444:5555:6666:1.2.3.4", "::11.22.33.44"} {
		suite.cfg.vpr.Set("foo", value)
		assert.NotPanics(func() {
			suite.cfg.checkIP("foo")
		})
	}

	for _, value := range []string{"", "1.2.3.256"} {
		suite.cfg.vpr.Set("foo", value)
		assert.Panics(func() {
			suite.cfg.checkIP("foo")
		})
	}
}

func (suite *InitTestSuite) TestCheckFile() {
	assert := suite.Require()

	suite.cfg.vpr.Set("foo", "/dev/null")
	assert.NotPanics(func() {
		suite.cfg.checkFile("foo")
	})

	suite.cfg.vpr.Set("foo", "/this/probably/does/not/exist")
	assert.Panics(func() {
		suite.cfg.checkFile("foo")
	})
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, &InitTestSuite{})
}
