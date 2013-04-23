package weather

import (
	"code.google.com/p/go-charset/charset"
	"github.com/jteeuwen/go-pkg-xmlx"

	"fmt"
	"go/build"
	"math"
	"os"
	"time"
)

const (
	url         = "http://forecast.weather.gov/MapClick.php?lat=30.30117&lon=-97.79243&FcstType=digitalDWML"
	arrows      = "↑↗→↘↓↙←↖"
	firstOctile = 0x2581
	maxHours    = 48
)

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

func Info() string {
	charset.CharsetDir = build.Default.GOPATH + "/src/code.google.com/p/go-charset/datafiles"
	doc := xmlx.New()
	doc.LoadUri(url, charset.NewReader)

	startTimeNodes := doc.SelectNodes("", "start-valid-time")
	endTimeNodes := doc.SelectNodes("", "end-valid-time")
	if len(endTimeNodes) == 0 || len(endTimeNodes) == 0 {
		os.Exit(1)
	}
	if len(endTimeNodes) > maxHours {
		endTimeNodes = endTimeNodes[:maxHours]
	}
	startTime, _ := time.Parse(time.RFC3339, startTimeNodes[0].Value)
	endTime, _ := time.Parse(time.RFC3339, endTimeNodes[len(endTimeNodes)-1].Value)

	tempGraph := ""
	temps := findTemps(doc)
	minTemp, maxTemp := minmax(temps)
	for _, temp := range temps {
		octile := rescale(temp, minTemp, maxTemp, 8)
		tempGraph += string(firstOctile + octile)
	}

	humidGraph := ""
	humids := findHumids(doc)
	minHumid, maxHumid := minmax(humids)
	for _, humid := range humids {
		octile := rescale(humid, minHumid, maxHumid, 8)
		humidGraph += string(firstOctile + octile)
	}

	precipGraph := ""
	precips := findPrecips(doc)
	minPrecip, maxPrecip := minmax(precips)
	for _, precip := range precips {
		octile := rescale(precip, minPrecip, maxPrecip, 8)
		precipGraph += string(firstOctile + octile)
	}

	speedGraph := ""
	speeds := findWindSpeeds(doc)
	minSpeed, maxSpeed := minmax(speeds)
	for _, speed := range speeds {
		octile := rescale(speed, minSpeed, maxSpeed, 8)
		speedGraph += string(firstOctile + octile)
	}

	dirGraph := ""
	dirs := findWindDirs(doc)
	for _, dir := range dirs {
		idx := dirIndex(dir)
		dirGraph += string([]rune(arrows)[idx])
	}

	timeFmt := "2006-01-02 15:04"
	start, end := startTime.Format(timeFmt), endTime.Format(timeFmt)
	out := fmt.Sprintf("                 %-24s%24s\n", start, end)
	tempRange := fmt.Sprintf("%d/%d", minTemp, maxTemp)
	humidRange := fmt.Sprintf("%d/%d", minHumid, maxHumid)
	precipRange := fmt.Sprintf("%d/%d", minPrecip, maxPrecip)
	speedRange := fmt.Sprintf("%d/%d", minSpeed, maxSpeed)
	out += fmt.Sprintf("Temp °F  %7s %s\n", tempRange, tempGraph)
	out += fmt.Sprintf("Humid %%  %7s %s\n", humidRange, humidGraph)
	out += fmt.Sprintf("Precip %% %7s %s\n", precipRange, precipGraph)
	out += fmt.Sprintf("Wind mph %7s %s\n", speedRange, speedGraph)
	out += fmt.Sprintf("Wind direction   %s\n", dirGraph)
	return out
}

func findTemps(doc *xmlx.Document) []int {
	temps := []int{}
	tempNodes := doc.SelectNodes("", "temperature")
	for _, node := range tempNodes {
		if node.As("", "type") == "hourly" {
			for _, kid := range node.Children {
				temps = append(temps, kid.I("", "value"))
				if len(temps) >= maxHours {
					break
				}
			}
		}
	}
	return temps
}

func findHumids(doc *xmlx.Document) []int {
	humids := []int{}
	humidNode := doc.SelectNode("", "humidity")
	for _, kid := range humidNode.Children {
		humids = append(humids, kid.I("", "value"))
		if len(humids) >= maxHours {
			break
		}
	}
	return humids
}

func findPrecips(doc *xmlx.Document) []int {
	precips := []int{}
	precipNode := doc.SelectNode("", "probability-of-precipitation")
	for _, kid := range precipNode.Children {
		precips = append(precips, kid.I("", "value"))
		if len(precips) >= maxHours {
			break
		}
	}
	return precips
}

func findWindSpeeds(doc *xmlx.Document) []int {
	speeds := []int{}
	speedNodes := doc.SelectNodes("", "wind-speed")
	for _, node := range speedNodes {
		if node.As("", "type") == "sustained" {
			for _, kid := range node.Children {
				speeds = append(speeds, kid.I("", "value"))
				if len(speeds) >= maxHours {
					break
				}
			}
		}
	}
	return speeds
}

func findWindDirs(doc *xmlx.Document) []int {
	dirs := []int{}
	dirNode := doc.SelectNode("", "direction")
	for _, kid := range dirNode.Children {
		dir := kid.I("", "value")
		dirs = append(dirs, dir)
		if len(dirs) >= maxHours {
			break
		}
	}
	return dirs
}
