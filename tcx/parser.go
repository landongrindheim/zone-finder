package tcx

import (
	"encoding/xml"
	"os"
	"time"
	"zone-finder/types"
)

type TCXData struct {
	Activities activities `xml:"Activities"`
}

type activities struct {
	Activity activity `xml:"Activity"`
}

type activity struct {
	Sport   string  `xml:"Sport,attr"`
	Id      string  `xml:"Id"`
	Laps    []lap   `xml:"Lap"`
	Creator creator `xml:"Creator"`
}

type lap struct {
	StartTime string  `xml:"StartTime,attr"`
	Tracks    []track `xml:"Track"`
}

type track struct {
	Trackpoints []trackpoint `xml:"Trackpoint"`
}

type trackpoint struct {
	Time         time.Time    `xml:"Time"`
	HeartRateBpm heartRateBpm `xml:"HeartRateBpm"`
}

type heartRateBpm struct {
	Value int `xml:"Value"`
}

type creator struct {
	Name      string `xml:"Name"`
	ProductId int    `xml:"ProductID"`
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

func (tcx *TCXData) GetHRDataPoints() ([]types.HRDataPoint, error) {
	var dataPoints []types.HRDataPoint

	for _, lap := range tcx.Activities.Activity.Laps {
		for _, track := range lap.Tracks {
			for _, trackpoint := range track.Trackpoints {
				heartRate := trackpoint.HeartRateBpm.Value

				if heartRate == 0 {
					continue
				}

				dataPoints = append(dataPoints,
					types.HRDataPoint{
						Timestamp: trackpoint.Time,
						HeartRate: trackpoint.HeartRateBpm.Value,
					},
				)
			}
		}
	}
	return dataPoints, nil
}
