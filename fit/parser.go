package fit

import (
	"os"
	"strings"
	"zone-finder/tcx"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
	"github.com/muktihari/fit/proto"
)

type FITData struct {
	messages   []proto.Message
	deviceInfo *mesgdef.DeviceInfo
}

const missingHeartRate = 255

func ParseFIT(filepath string) (*FITData, error) {
	fitFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer fitFile.Close()

	dec := decoder.New(fitFile)
	fit, err := dec.Decode()
	if err != nil {
		return nil, err
	}

	return &FITData{
		messages:   fit.Messages,
		deviceInfo: findDeviceInfo(fit.Messages),
	}, nil
}

func findDeviceInfo(messages []proto.Message) *mesgdef.DeviceInfo {
	for _, msg := range messages {
		if di := mesgdef.NewDeviceInfo(&msg); di != nil {
			if isValidDeviceInfo(di) {
				return di
			}
		}
	}

	return nil
}

func isValidDeviceInfo(deviceInfo *mesgdef.DeviceInfo) bool {
	if deviceInfo == nil {
		return false
	}

	manufacturer := deviceInfo.Manufacturer.String()
	return manufacturer != "" && !strings.Contains(manufacturer, "Invalid")
}

func (fit *FITData) GetHRDataPoints() ([]tcx.HRDataPoint, error) {
	var dataPoints []tcx.HRDataPoint

	for _, msg := range fit.messages {
		record := mesgdef.NewRecord(&msg)
		if record == nil || record.HeartRate == missingHeartRate {
			continue
		}

		dataPoints = append(dataPoints, tcx.HRDataPoint{
			Timestamp: record.Timestamp,
			HeartRate: int(record.HeartRate),
		})
	}

	return dataPoints, nil
}

func (fit *FITData) GetDeviceName() string {
	if fit.deviceInfo == nil {
		return "Unknown"
	}

	manufacturer := fit.deviceInfo.Manufacturer.String()
	if manufacturer == "garmin" {
		return typedef.GarminProduct(fit.deviceInfo.Product).String()
	}

	return manufacturer
}

func (fit *FITData) GetProductID() int {
	if fit.deviceInfo == nil {
		return 0
	}

	return int(fit.deviceInfo.Product)
}
