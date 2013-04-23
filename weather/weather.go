package weather

import (
	"code.google.com/p/go-charset/charset"
	"github.com/jteeuwen/go-pkg-xmlx"

	"fmt"
	"go/build"
	"math"
	"time"
)

const (
	url         = "http://forecast.weather.gov/MapClick.php?lat=30.30117&lon=-97.79243&FcstType=digitalDWML"
	arrows      = "↑↗→↘↓↙←↖"
	firstOctile = 0x2581
	maxHours    = 48
)

func Forecast() string {
	charset.CharsetDir = build.Default.GOPATH + "/src/code.google.com/p/go-charset/datafiles"
	doc := xmlx.New()
	err := doc.LoadUri(url, charset.NewReader)
	if err != nil {
		return ""
	}

	startTimeNodes := doc.SelectNodes("", "start-valid-time")
	endTimeNodes := doc.SelectNodes("", "end-valid-time")
	if len(startTimeNodes) == 0 || len(endTimeNodes) == 0 {
		return ""
	}
	if len(endTimeNodes) > maxHours {
		endTimeNodes = endTimeNodes[:maxHours]
	}
	startTime, _ := time.Parse(time.RFC3339, startTimeNodes[0].Value)
	endTime, _ := time.Parse(time.RFC3339, endTimeNodes[len(endTimeNodes)-1].Value)

	temps := findVals("temperature", "hourly", doc)
	humids := findVals("humidity", "", doc)
	precips := findVals("probability-of-precipitation", "", doc)
	speeds := findVals("wind-speed", "sustained", doc)
	dirs := findVals("direction", "", doc)

	minTemp, maxTemp, tempGraph := makeGraph(temps)
	minHumid, maxHumid, humidGraph := makeGraph(humids)
	minPrecip, maxPrecip, precipGraph := makeGraph(precips)
	minSpeed, maxSpeed, speedGraph := makeGraph(speeds)

	dirGraph := ""
	for _, dir := range dirs {
		idx := dirIndex(dir)
		dirGraph += string([]rune(arrows)[idx])
	}

	timeFmt := "2006-01-02 15:04"
	start, end := startTime.Format(timeFmt), endTime.Format(timeFmt)

	tempRange := fmt.Sprintf("%3d %3d", minTemp, maxTemp)
	humidRange := fmt.Sprintf("%3d %3d", minHumid, maxHumid)
	precipRange := fmt.Sprintf("%3d %3d", minPrecip, maxPrecip)
	speedRange := fmt.Sprintf("%3d %3d", minSpeed, maxSpeed)

	out := fmt.Sprintf("         min max %-24s%24s\n", start, end)
	out += fmt.Sprintf("Temp °F  %7s %s\n", tempRange, tempGraph)
	out += fmt.Sprintf("Humid %%  %7s %s\n", humidRange, humidGraph)
	out += fmt.Sprintf("Precip %% %7s %s\n", precipRange, precipGraph)
	out += fmt.Sprintf("Wind mph %7s %s\n", speedRange, speedGraph)
	out += fmt.Sprintf("Wind dir         %s\n", dirGraph)

	return out
}

func minmax(vals []int) (int, int) {
	min, max := math.MaxInt32, math.MinInt32
	for _, v := range vals {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

func rescale(val, min, max, bins int) int {
	if min >= max {
		return 0
	}
	span := max - min
	v := (val - min) * bins / span
	if v < 0 {
		v = 0
	} else if v > bins-1 {
		v = bins - 1
	}
	return v
}

func dirIndex(dir int) int {
	return ((dir + 360/16) * 8 / 360) % 8
}

func findVals(name, typ string, doc *xmlx.Document) []int {
	vals := []int{}
	nodes := doc.SelectNodes("", name)
	for _, node := range nodes {
		if typ == "" || node.As("", "type") == typ {
			for _, kid := range node.Children {
				vals = append(vals, kid.I("", "value"))
				if len(vals) >= maxHours {
					break
				}
			}
			break // just use the first set
		}
	}
	return vals
}

func makeGraph(vals []int) (int, int, string) {
	if len(vals) == 0 {
		return 0, 0, ""
	}
	graph := ""
	min, max := minmax(vals)
	for _, val := range vals {
		octile := rescale(val, min, max, 8)
		graph += string(firstOctile + octile)
	}
	return min, max, graph
}
