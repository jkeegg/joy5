package flvio

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/nareix/joy5/utils/bits/pio"
)

func TsToTime(ts int64) time.Duration {
	return time.Millisecond * time.Duration(ts)
}

func TimeToTs(tm time.Duration) int64 {
	return int64(tm / time.Millisecond)
}

const (
	TAG_AUDIO = 8
	TAG_VIDEO = 9
	TAG_AMF0  = 18
	TAG_AMF3  = 15
)

func TagTypeString(v uint8) string {
	switch v {
	case TAG_VIDEO:
		return "VIDEO"
	case TAG_AUDIO:
		return "AUDIO"
	case TAG_AMF0:
		return "AMF0"
	case TAG_AMF3:
		return "AMF3"
	}
	return fmt.Sprint(v)
}

func FrameTypeString(v uint8) string {
	switch v {
	case FRAME_INTER:
		return "INTER"
	case FRAME_KEY:
		return "KEY"
	}
	return fmt.Sprint(v)
}

const (
	SOUND_LPCM_PLATFORM_ENDIAN  = 0
	SOUND_ADPCM                 = 1
	SOUND_MP3                   = 2
	SOUND_LPCM_LITTLE_ENDIAN    = 3
	SOUND_NELLYMOSER_16KHZ_MONO = 4
	SOUND_NELLYMOSER_8KHZ_MONO  = 5
	SOUND_NELLYMOSER            = 6
	SOUND_ALAW                  = 7
	SOUND_MULAW                 = 8
	SOUND_EXHEADER              = 9
	SOUND_AAC                   = 10
	SOUND_SPEEX                 = 11
	SOUND_MP3_8K                = 14
	SOUND_NATIVE                = 15
	SOUND_AC3                   = 20
	SOUND_EAC3                  = 21
	SOUND_OPUS                  = 22
	SOUND_FLAC                  = 23

	SOUND_5_5Khz = 0
	SOUND_11Khz  = 1
	SOUND_22Khz  = 2
	SOUND_44Khz  = 3

	SOUND_8BIT  = 0
	SOUND_16BIT = 1

	SOUND_MONO   = 0
	SOUND_STEREO = 1

	AAC_SEQHDR = 0
	AAC_RAW    = 1
)

const (
	APT_SequenceStart      = 0
	APT_CodedFrames        = 1
	APT_SequenceEnd        = 2
	APT_MultichannelConfig = 4
	APT_Multitrack         = 5
	APT_ModEx              = 7
)

const (
	AVC_SEQHDR = 0
	AVC_NALU   = 1
	AVC_EOS    = 2

	FRAME_KEY   = 1
	FRAME_INTER = 2

	VIDEO_H264 = 7
	VIDEO_H265 = 12 // 业界约定的 H.265/HEVC FLV CodecID
	VIDEO_VP8  = 13
	VIDEO_VP9  = 14
	VIDEO_AV1  = 15
)

const (
	VPT_SequenceStart        = 0
	VPT_CodedFrames          = 1
	VPT_SequenceEnd          = 2
	VPT_CodedFramesX         = 3
	VPT_Metadata             = 4
	VPT_MPEG2TSSequenceStart = 5
	VPT_Multitrack           = 6
	VPT_ModEx                = 7
)

const (
	ac3  = "ac-3"
	eac3 = "ec-3"
	opus = "Opus"
	mp3  = ".mp3"
	flac = "fLaC"
	aac  = "mp4a"
)

const (
	vp8  = "vp08"
	vp9  = "vp09"
	av1  = "av01"
	avc  = "avc1"
	hevc = "hvc1"
)

type Tag struct {
	Type uint8

	/*
		SoundFormat: UB[4]
		0 = Linear PCM, platform endian
		1 = ADPCM
		2 = MP3
		3 = Linear PCM, little endian
		4 = Nellymoser 16-kHz mono
		5 = Nellymoser 8-kHz mono
		6 = Nellymoser
		7 = G.711 A-law logarithmic PCM
		8 = G.711 mu-law logarithmic PCM
		9 = reserved
		10 = AAC
		11 = Speex
		14 = MP3 8-Khz
		15 = Device-specific sound
		Formats 7, 8, 14, and 15 are reserved for internal use
		AAC is supported in Flash Player 9,0,115,0 and higher.
		Speex is supported in Flash Player 10 and higher.
	*/
	SoundFormat uint8

	/*
		SoundRate: UB[2]
		Sampling rate
		0 = 5.5-kHz For AAC: always 3
		1 = 11-kHz
		2 = 22-kHz
		3 = 44-kHz
	*/
	SoundRate uint8

	/*
		SoundSize: UB[1]
		0 = snd8Bit
		1 = snd16Bit
		Size of each sample.
		This parameter only pertains to uncompressed formats.
		Compressed formats always decode to 16 bits internally
	*/
	SoundSize uint8

	/*
		SoundType: UB[1]
		0 = sndMono
		1 = sndStereo
		Mono or stereo sound For Nellymoser: always 0
		For AAC: always 1
	*/
	SoundType uint8

	/*
		0: AAC sequence header
		1: AAC raw
	*/
	AACPacketType uint8

	/*
		1: keyframe (for AVC, a seekable frame)
		2: inter frame (for AVC, a non- seekable frame)
		3: disposable inter frame (H.263 only)
		4: generated keyframe (reserved for server use only)
		5: video info/command frame
	*/
	FrameType uint8

	/*
		1: JPEG (currently unused)
		2: Sorenson H.263
		3: Screen video
		4: On2 VP6
		5: On2 VP6 with alpha channel
		6: Screen video version 2
		7: AVC
	*/
	VideoFormat uint8

	/*
		0: AVC sequence header
		1: AVC NALU
		2: AVC end of sequence (lower level NALU sequence ender is not required or supported)
	*/
	AVCPacketType uint8

	Time  uint32
	CTime int32

	StreamId uint32

	Header, Data []byte
}

func (t Tag) DebugFields() []interface{} {
	p := []interface{}{"Type", TagTypeString(t.Type), "Time", t.Time, "Len", len(t.Data)}

	switch t.Type {
	case TAG_VIDEO:
		p = append(p, "FrameType")
		p = append(p, FrameTypeString(t.FrameType))

		p = append(p, "VideoFormat")
		p = append(p, t.VideoFormat)

		switch t.VideoFormat {
		case VIDEO_H264, VIDEO_H265:
			p = append(p, "AVCPacketType")
			p = append(p, t.AVCPacketType)
		}

		if t.CTime != 0 {
			p = append(p, "Ctime")
			p = append(p, t.CTime)
		}

	case TAG_AMF0, TAG_AMF3:
		amf3 := t.Type == TAG_AMF3
		arr, _ := ParseAMFVals(t.Data, amf3)
		arrjs, _ := json.Marshal(arr)
		p = append(p, "Data")
		p = append(p, string(arrjs))
	}

	p = append(p, "Header")
	p = append(p, fmt.Sprintf("%x", t.Header))
	return p
}

func (t Tag) MaxHeaderLen() int {
	return 24
}

func (t *Tag) parseAudioHeader(b []byte) (n int, err error) {
	var flags uint8
	if flags, err = pio.ReadU8(b, &n); err != nil {
		return
	}
	t.SoundFormat = flags >> 4

	if t.SoundFormat != SOUND_EXHEADER {
		t.SoundRate = (flags >> 2) & 0x3
		t.SoundSize = (flags >> 1) & 0x1
		t.SoundType = flags & 0x1
		switch t.SoundFormat {
		case SOUND_AAC:
			if t.AACPacketType, err = pio.ReadU8(b, &n); err != nil {
				return
			}
		case SOUND_ALAW, SOUND_MULAW:
			return
		default:
			fmt.Println("SoundFormat, ", t.SoundFormat)
		}
	} else {
		var audioFourCC string
		audioPacketType := flags & 0xf
		for audioPacketType == APT_ModEx {
			var modExDataSize8 uint8
			if modExDataSize8, err = pio.ReadU8(b, &n); err != nil {
				return
			}
			var modExDataSize int
			modExDataSize = int(modExDataSize8) + 1
			if modExDataSize == 256 {
				var modExDataSize16 uint16
				if modExDataSize16, err = pio.ReadU16BE(b, &n); err != nil {
					return
				}
				modExDataSize = int(modExDataSize16) + 1
			}
			modExData := make([]byte, modExDataSize)
			if modExData, err = pio.ReadBytes(b, &n, modExDataSize); err != nil {
				return
			}
			fmt.Println("modExData: ", modExData)
			if flags, err = pio.ReadU8(b, &n); err != nil {
				return
			}
			audioPacketModExType := flags >> 4
			audioPacketType = flags & 0xf
			if audioPacketModExType == 0 {
				audioTimestampNanoOffset := int32(pio.U24BE(modExData[0:3]))
				fmt.Println("audioTimestampNanoOffset: ", audioTimestampNanoOffset)
			}
		}
		var isAudioMultitrack bool
		var audioMultitrackType uint8
		isAudioMultitrack = false
		if audioPacketType == APT_Multitrack {
			isAudioMultitrack = true
			if flags, err = pio.ReadU8(b, &n); err != nil {
				return
			}
			audioMultitrackType = flags >> 4
			audioPacketType = flags & 0xf
			if audioMultitrackType != 2 {
				if audioFourCC, err = pio.ReadString(b, &n, 4); err != nil {
					return
				}
				fmt.Println("audioFourCC: ", audioFourCC)
			}
		} else {
			if audioFourCC, err = pio.ReadString(b, &n, 4); err != nil {
				return
			}
			fmt.Println("audioFourCC: ", audioFourCC)
			switch audioFourCC {
			case ac3:
				t.SoundFormat = SOUND_AC3
			case eac3:
				t.SoundFormat = SOUND_EAC3
			case flac:
				t.SoundFormat = SOUND_FLAC
			case mp3:
				t.SoundFormat = SOUND_MP3
			case opus:
				t.SoundFormat = SOUND_OPUS
			case aac:
				t.SoundFormat = SOUND_AAC
			default:
				fmt.Println("audioFourCC: ", audioFourCC)
			}
		}

		// Process audio data
		if isAudioMultitrack {
			if audioMultitrackType == 2 {
				// ManyTracksManyCodecs
			}
		}

		switch audioPacketType {
		case APT_MultichannelConfig:
			if flags, err = pio.ReadU8(b, &n); err != nil {
				return
			}
			audioChannelOrder := flags
			if flags, err = pio.ReadU8(b, &n); err != nil {
				return
			}
			channelCount := flags
			if audioChannelOrder == 2 {
				// AudioChannelOrder.Custom
				audioChannelMapping := make([]uint8, channelCount)
				if audioChannelMapping, err = pio.ReadBytes(b, &n, int(channelCount)); err != nil {
					return
				}
				fmt.Println("audioChannelMapping:", audioChannelMapping)
			}
			if audioChannelOrder == 1 {
				// AudioChannelOrder.Native
				var audioChannelFlags uint32
				if audioChannelFlags, err = pio.ReadU32BE(b, &n); err != nil {
					return
				}
				fmt.Println("audioChannelFlags:", audioChannelFlags)
			}
		case APT_SequenceEnd:
		case APT_SequenceStart:
			switch audioFourCC {
			case aac:
				// AacSequenceHeader
				t.AACPacketType = 0
			case flac:
				// FlacSequenceHeader
				t.AACPacketType = 0
			case mp3:
			case opus:
				// OpusSequenceHeader
				t.AACPacketType = 0
			case ac3, eac3:
			default:

			}
		case APT_CodedFrames:
			switch audioFourCC {
			case ac3, eac3:
				// Ac3CodedData
			case flac:
				// FlacCodedData
				t.AACPacketType = 1
			case mp3:
				// Mp3CodedData
			case opus:
				// OpusCodedData
				t.AACPacketType = 1
			case aac:
				// AacCodedData
				t.AACPacketType = 1
			default:

			}
		default:
			fmt.Println("audioPacketType:", audioPacketType)
		}

		// if isAudioMultitrack && audioMultitrackType != 0 && positionDataPtrToNextAudioTrack(sizeOfAudioTrack) {
		// 	continue
		// }
	}

	return
}

func (t Tag) fillAudioHeader(b []byte) (n int) {
	var flags uint8
	flags |= t.SoundFormat << 4
	flags |= t.SoundRate << 2
	flags |= t.SoundSize << 1
	flags |= t.SoundType
	pio.WriteU8(b, &n, flags)

	switch t.SoundFormat {
	case SOUND_AAC:
		pio.WriteU8(b, &n, t.AACPacketType)
	}

	return
}

func (t *Tag) parseVideoHeader(b []byte) (n int, err error) {
	var flags uint8
	if flags, err = pio.ReadU8(b, &n); err != nil {
		return
	}

	isExVideoHeader := flags>>7 != 0
	t.FrameType = flags >> 4 & 0x7

	var videoPacketType byte
	if !isExVideoHeader {
		t.VideoFormat = flags & 0xf

		switch t.VideoFormat {
		case VIDEO_H264, VIDEO_H265:
			if t.AVCPacketType, err = pio.ReadU8(b, &n); err != nil {
				return
			}
			var v int32
			if v, err = pio.ReadI24BE(b, &n); err != nil {
				return
			}
			t.CTime = v
		}
	} else {
		var videoFourCC string
		videoPacketType = flags & 0xf

		for videoPacketType == VPT_ModEx {
			var modExDataSize8 uint8
			if modExDataSize8, err = pio.ReadU8(b, &n); err != nil {
				return
			}
			var modExDataSize int
			modExDataSize = int(modExDataSize8) + 1
			if modExDataSize == 256 {
				var modExDataSize16 uint16
				if modExDataSize16, err = pio.ReadU16BE(b, &n); err != nil {
					return
				}
				modExDataSize = int(modExDataSize16) + 1
			}
			modExData := make([]byte, modExDataSize)
			if modExData, err = pio.ReadBytes(b, &n, modExDataSize); err != nil {
				return
			}
			fmt.Println("modExData: ", modExData)
			if flags, err = pio.ReadU8(b, &n); err != nil {
				return
			}
			videoPacketModExType := flags >> 4
			videoPacketType = flags & 0xf
			if videoPacketModExType == 0 {
				// VideoPacketModExType.TimestampOffsetNano
				videoTimestampNanoOffset := int32(pio.U24BE(modExData[0:3]))
				fmt.Println("videoTimestampNanoOffset: ", videoTimestampNanoOffset)
			}
		}

		var isVideoMultitrack bool
		var videoMultitrackType uint8
		if videoPacketType != VPT_Metadata && t.FrameType == 5 { // VideoFrameType.Command
			fmt.Println("Got video Command")
		} else if videoPacketType == VPT_Multitrack {
			fmt.Println("Got multitrack")
			isVideoMultitrack = true
			if flags, err = pio.ReadU8(b, &n); err != nil {
				return
			}
			videoMultitrackType = flags >> 4
			videoPacketType = flags & 0xf
			if videoMultitrackType != 2 {
				// AvMultitrackType.ManyTracksManyCodecs
				if videoFourCC, err = pio.ReadString(b, &n, 4); err != nil {
					return
				}
				fmt.Println("audioFourCC: ", videoFourCC)
			}
		} else {
			videoFourCC, err = pio.ReadString(b, &n, 4)
			if err != nil {
				return
			}
			switch videoFourCC {
			case vp8:
				t.VideoFormat = VIDEO_VP8
			case vp9:
				t.VideoFormat = VIDEO_VP9
			case av1:
				t.VideoFormat = VIDEO_AV1
			case avc:
				t.VideoFormat = VIDEO_H264
			case hevc:
				t.VideoFormat = VIDEO_H265
			default:
				fmt.Println("videoFourCC: ", videoFourCC)
			}
		}

		if isVideoMultitrack {
			// ManyTracksManyCodecs
			if videoMultitrackType == 2 {
				if videoFourCC, err = pio.ReadString(b, &n, 4); err != nil {
					return
				}
				fmt.Println("audioFourCC: ", videoFourCC)
				if flags, err = pio.ReadU8(b, &n); err != nil {
					return
				}
				videoTrackId := flags
				fmt.Println("videoTrackId", videoTrackId)
				if videoMultitrackType != 0 {
					// AvMultitrackType.OneTrack
					var sizeOfVideoTrack uint32
					if sizeOfVideoTrack, err = pio.ReadU24BE(b, &n); err != nil {
						return
					}
					fmt.Println("sizeOfVideoTrack", sizeOfVideoTrack)
				}
			}
		}

		switch videoPacketType {
		case VPT_Metadata:
			t.AVCPacketType = 1
		case VPT_SequenceEnd:
		case VPT_SequenceStart:
			t.AVCPacketType = 0
		case VPT_MPEG2TSSequenceStart:
			if videoFourCC == av1 {
				// body contains a video descriptor to start the sequence
				t.AVCPacketType = 0
			}
		case APT_CodedFrames:
			switch videoFourCC {
			case vp8, vp9:
				// body contains series of coded full frames
				t.AVCPacketType = 1
			case av1:
				// body contains one or more OBUs representing a single temporal unit
				t.AVCPacketType = 1
			case avc, hevc:
				t.AVCPacketType = 1
				var v int32
				if v, err = pio.ReadI24BE(b, &n); err != nil {
					return
				}
				t.CTime = v
			}
		case VPT_CodedFramesX:
			switch videoFourCC {
			case avc, hevc:
				t.AVCPacketType = 1
			}
		default:
			fmt.Println("videoPacketType", videoPacketType)
		}
	}

	return
}

func (t Tag) fillVideoHeader(b []byte) (n int) {
	pio.WriteU8(b, &n, t.FrameType<<4|t.VideoFormat)

	switch t.VideoFormat {
	case VIDEO_H264, VIDEO_H265:
		pio.WriteU8(b, &n, t.AVCPacketType)
		pio.WriteI24BE(b, &n, int32(t.CTime))
	}
	return
}

func (t Tag) FillHeader(b []byte) (n int) {
	switch t.Type {
	case TAG_AUDIO:
		return t.fillAudioHeader(b)

	case TAG_VIDEO:
		return t.fillVideoHeader(b)
	}

	return
}

func (t *Tag) ParseHeader(b []byte) (n int, err error) {
	switch t.Type {
	case TAG_AUDIO:
		if n, err = t.parseAudioHeader(b); err != nil {
			return
		}

	case TAG_VIDEO:
		if n, err = t.parseVideoHeader(b); err != nil {
			return
		}
	}

	t.Header = b[:n]
	return
}

func (t *Tag) Parse(b []byte) (err error) {
	var n int
	if n, err = t.ParseHeader(b); err != nil {
		return
	}
	t.Data = b[n:]
	return
}

const (
	// TypeFlagsReserved UB[5]
	// TypeFlagsAudio    UB[1] Audio tags are present
	// TypeFlagsReserved UB[1] Must be 0
	// TypeFlagsVideo    UB[1] Video tags are present
	FILE_HAS_AUDIO = 0x4
	FILE_HAS_VIDEO = 0x1
)

func ParseTagHeader(b []byte) (tag Tag, datalen int, err error) {
	tagtype := b[0]
	tag = Tag{Type: tagtype}
	datalen = int(pio.U24BE(b[1:4]))

	var tslo uint32
	var tshi uint8
	tslo = pio.U24BE(b[4:7])
	tshi = b[7]

	tag.Time = tslo | uint32(tshi)<<24
	tag.StreamId = pio.U24BE(b[8:11])
	return
}

func ReadTag(r io.Reader, b []byte, malloc func(int) ([]byte, error)) (tag Tag, err error) {
	if _, err = io.ReadFull(r, b[:TagHeaderLength]); err != nil {
		return
	}
	var datalen int
	if tag, datalen, err = ParseTagHeader(b); err != nil {
		return
	}

	var data []byte
	if data, err = malloc(datalen); err != nil {
		return
	}
	if _, err = io.ReadFull(r, data); err != nil {
		return
	}

	if err = tag.Parse(data); err != nil {
		return
	}

	if _, err = io.ReadFull(r, b[:4]); err != nil {
		return
	}
	return
}

const TagHeaderLength = 11

func FillTagHeader(b []byte, tag Tag, datalen int) {
	b[0] = tag.Type
	pio.PutU24BE(b[1:4], uint32(datalen))
	pio.PutU24BE(b[4:7], uint32(tag.Time&0xffffff))
	b[7] = uint8(tag.Time >> 24)
	pio.PutU24BE(b[8:11], tag.StreamId)
}

const TagTrailerLength = 4

func FillTagTrailer(b []byte, datalen int) {
	pio.PutU32BE(b[0:4], uint32(datalen+TagHeaderLength))
}

func WriteTag(w io.Writer, tag Tag, b []byte) (err error) {
	data := tag.Data

	n := tag.FillHeader(b[TagHeaderLength:])
	datalen := len(data) + n

	FillTagHeader(b, tag, datalen)
	n += TagHeaderLength

	if _, err = w.Write(b[:n]); err != nil {
		return
	}

	if _, err = w.Write(data); err != nil {
		return
	}

	FillTagTrailer(b, datalen)
	if _, err = w.Write(b[:TagTrailerLength]); err != nil {
		return
	}

	return
}

const FileHeaderLength = 13

func FillFileHeader(b []byte, flags uint8) {
	// 'FLV', version 1
	pio.PutU32BE(b[0:4], 0x464c5601)
	b[4] = flags

	// DataOffset: UI32 Offset in bytes from start of file to start of body (that is, size of header)
	// The DataOffset field usually has a value of 9 for FLV version 1.
	pio.PutU32BE(b[5:9], 9)

	// PreviousTagSize0: UI32 Always 0
	pio.PutU32BE(b[9:13], 0)

	return
}

func ParseFileHeader(b []byte) (flags uint8, skip int, err error) {
	flv := pio.U24BE(b[0:3])
	if flv != 0x464c56 { // 'FLV'
		err = fmt.Errorf("CC3Invalid")
		return
	}
	flags = b[4]

	skip = int(pio.U32BE(b[5:9])) - 9
	if skip < 0 {
		err = fmt.Errorf("FileHeaderDataSizeInvalid")
		return
	}

	return
}
