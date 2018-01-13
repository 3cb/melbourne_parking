package main

import (
	"github.com/3cb/melbourne_parking/melbourne"
	flatbuffers "github.com/google/flatbuffers/go"
)

func serialize(data []Spot, date string) []byte {
	builder := flatbuffers.NewBuilder(1024)

	s := []flatbuffers.UOffsetT{}
	for _, v := range data {
		bayID := builder.CreateString(v.BayID)
		longitude := builder.CreateString(v.Lon)
		latitude := builder.CreateString(v.Lat)
		stMarkerID := builder.CreateString(v.StMarkerID)
		status := builder.CreateString(v.Status)

		melbourne.SpotStart(builder)
		melbourne.SpotAddBayId(builder, bayID)
		melbourne.SpotAddLongitude(builder, longitude)
		melbourne.SpotAddLatitude(builder, latitude)
		melbourne.SpotAddStMarkerId(builder, stMarkerID)
		melbourne.SpotAddStatus(builder, status)
		s = append(s, melbourne.SpotEnd(builder))
	}
	melbourne.MessageStartSpotsVector(builder, len(s))
	for i := len(s) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(s[i])
	}
	spots := builder.EndVector(len(s))

	d := builder.CreateString(date)

	melbourne.MessageStart(builder)
	melbourne.MessageAddDate(builder, d)
	melbourne.MessageAddSpots(builder, spots)
	msg := melbourne.MessageEnd(builder)

	builder.Finish(msg)
	buf := builder.FinishedBytes()

	return buf
}
