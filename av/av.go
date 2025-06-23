package av

import (
	"fmt"
	"time"

	"github.com/nareix/joy5/codec/aac"
	"github.com/nareix/joy5/codec/h264"
	"github.com/nareix/joy5/codec/h265"
)

const (
	H264 = 1 + iota
	AAC
	H264DecoderConfig
	H264SPSPPSNALU
	AACDecoderConfig
	Metadata
	H265
	H265DecoderConfig
)

var PacketTypeString = map[int]string{
	H264:              "H264",
	AAC:               "AAC",
	H264DecoderConfig: "H264DecoderConfig",
	H264SPSPPSNALU:    "H264SPSPPSNALU",
	AACDecoderConfig:  "AACDecoderConfig",
	Metadata:          "Metadata",
	H265:              "H265",
	H265DecoderConfig: "H265DecoderConfig",
}

type Packet struct {
	Type       int
	IsKeyFrame bool
	CTime      time.Duration
	Time       time.Duration
	Data       []byte
	ASeqHdr    []byte
	VSeqHdr    []byte
	Metadata   []byte
	AAC        *aac.Codec
	H264       *h264.Codec
	H265       *h265.Codec
}

func (p Packet) String() string {
	ret := ""

	typeStr := PacketTypeString[p.Type]
	if typeStr == "" {
		typeStr = "UnknownPacketType"
	}
	ret += typeStr

	if p.IsKeyFrame {
		ret += " K"
	}

	ret += " " + fmt.Sprint(p.Time)

	if p.CTime != 0 {
		ret += " " + fmt.Sprint(p.CTime)
	}

	ret += " " + fmt.Sprint(len(p.Data))

	return ret
}

type PacketReader interface {
	ReadPacket() (Packet, error)
}

type PacketReadCloser interface {
	PacketReader
	Close() error
}

type PacketWriter interface {
	WritePacket(Packet) error
}

type PacketWriteCloser interface {
	PacketWriter
	Close() error
}
