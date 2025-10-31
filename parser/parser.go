package parser

import (
	"encoding/xml"
	"os"
	"time"
)

type TCXData struct {
	Activities Activities `xml:"Activities"`
}

type Activities struct {
	Activity Activity `xml:"Activity"`
}

type Activity struct {
	Sport   string  `xml:"Sport,attr"`
	Id      string  `xml:"Id"`
	Laps    []Lap   `xml:"Lap"`
	Creator Creator `xml:"Creator"`
}

type Lap struct {
	StartTime string  `xml:"StartTime,attr"`
	Tracks    []Track `xml:"Track"`
}

type Track struct {
	Trackpoints []Trackpoint `xml:"Trackpoint"`
}

type Trackpoint struct {
	Time         time.Time    `xml:"Time"`
	HeartRateBpm HeartRateBpm `xml:"HeartRateBpm"`
}

type HeartRateBpm struct {
	Value int `xml:"Value"`
}

type HRDataPoint struct {
	Timestamp time.Time
	HeartRate int
}

type Creator struct {
	XMLName   xml.Name `xml:"Creator"`
	Name      string   `xml:"Name"`
	ProductId int      `xml:"ProductID"`
}

func ParseTCX(filepath string) (*TCXData, error) {
	var tcxData TCXData

	tcxFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer tcxFile.Close()

	if err := xml.NewDecoder(tcxFile).Decode(&tcxData); err != nil {
		return &TCXData{}, err
	}

	return &tcxData, nil
}

func (tcx *TCXData) GetDeviceName() string {
	return tcx.Activities.Activity.Creator.Name
}

func (tcx *TCXData) GetProductID() int {
	return tcx.Activities.Activity.Creator.ProductId
}

func (tcx *TCXData) GetHRDataPoints() ([]HRDataPoint, error) {
	var dataPoints []HRDataPoint

	for _, lap := range tcx.Activities.Activity.Laps {
		for _, track := range lap.Tracks {
			for _, trackpoint := range track.Trackpoints {
				heartRate := trackpoint.HeartRateBpm.Value

				if heartRate == 0 {
					continue
				}

				dataPoints = append(dataPoints,
					HRDataPoint{
						Timestamp: trackpoint.Time,
						HeartRate: trackpoint.HeartRateBpm.Value,
					},
				)
			}
		}
	}
	return dataPoints, nil
}
