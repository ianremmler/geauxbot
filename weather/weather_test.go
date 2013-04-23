package weather_test

import (
	"github.com/ianremmler/igopher/weather"

	"testing"
)

func TestWeather(t *testing.T) {
	t.Log("\n" + weather.Forecast())
}
