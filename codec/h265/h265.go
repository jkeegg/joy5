package h265

import (
	"fmt"
)

const (
	NALU_TRAIL_N    = 0
	NALU_TRAIL_R    = 1
	NALU_IDR_W_RADL = 19
	NALU_IDR_N_LP   = 20
	NALU_VPS        = 32
	NALU_SPS        = 33
	NALU_PPS        = 34
	NALU_AUD        = 35
	NALU_EOS_NUT    = 36
	NALU_EOB_NUT    = 37
	NALU_FD_NUT     = 38
)

func NALUType(b []byte) byte {
	if len(b) > 0 {
		return (b[0] >> 1) & 0x3f
	}
	return 0
}

func NALUTypeString(i byte) string {
	switch i {
	case NALU_TRAIL_N:
		return "TRAIL_N"
	case NALU_TRAIL_R:
		return "TRAIL_R"
	case NALU_IDR_W_RADL:
		return "IDR_W_RADL"
	case NALU_IDR_N_LP:
		return "IDR_N_LP"
	case NALU_VPS:
		return "VPS"
	case NALU_SPS:
		return "SPS"
	case NALU_PPS:
		return "PPS"
	case NALU_AUD:
		return "AUD"
	default:
		return fmt.Sprint(i)
	}
}

// H.265/HEVC 解析器初步实现，仅供 RTMP/FLV 封装和解封装使用
// 可根据需要扩展 NALU 解析、SPS/PPS/VPS 解析等
// 参考 codec/h264/h264.go

type Codec struct {
	ConfigBytes   []byte
	VPS, SPS, PPS map[int][]byte
	W, H          int
}
