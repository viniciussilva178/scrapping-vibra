package scraper_test

import (
	"scraper/internal/scraper"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuite(t *testing.T) {
	assert := assert.New(t)

	assert.Nil(t, scraper.ScrapeDocsVibra)
}
