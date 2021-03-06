// automatically generated by the FlatBuffers compiler, do not modify

package melbourne

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Spot struct {
	_tab flatbuffers.Table
}

func GetRootAsSpot(buf []byte, offset flatbuffers.UOffsetT) *Spot {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Spot{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Spot) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Spot) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Spot) BayId() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Spot) Longitude() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Spot) Latitude() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Spot) StMarkerId() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Spot) Status() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func SpotStart(builder *flatbuffers.Builder) {
	builder.StartObject(5)
}
func SpotAddBayId(builder *flatbuffers.Builder, bayId flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(bayId), 0)
}
func SpotAddLongitude(builder *flatbuffers.Builder, longitude flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(longitude), 0)
}
func SpotAddLatitude(builder *flatbuffers.Builder, latitude flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(latitude), 0)
}
func SpotAddStMarkerId(builder *flatbuffers.Builder, stMarkerId flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(stMarkerId), 0)
}
func SpotAddStatus(builder *flatbuffers.Builder, status flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(4, flatbuffers.UOffsetT(status), 0)
}
func SpotEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
