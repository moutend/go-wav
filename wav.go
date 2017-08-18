// Package wav reads and writes wave (.wav) file.
package wav

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

const (
	WAVE_FORMAT_PCM        = 0x1
	WAVE_FORMAT_EXTENSIBLE = 0xFFFE
)

// File represents WAV audio file.
type File struct {
	formatTag      uint16
	channels       uint16
	samplesPerSec  uint32
	avgBytesPerSec uint32
	blockAlign     uint16
	bitsPerSample  uint16
	length         uint32
	data           []byte
	offset         int
}

// Duration returns playback time in second.
func (v *File) Duration() time.Duration {
	return time.Duration(v.Length()/v.BlockAlign()) * time.Second
}

// FormatTag returns either
// 0x1 (WAVE_FORMAT_PCM) or
// 0xFFFE (WAVE_FORMAT_EXTENSIBLE).
func (v *File) FormatTag() uint16 {
	return v.formatTag
}

// Channels returns number of channels.
func (v *File) Channels() int {
	return int(v.channels)
}

// SamplesPerSec returns number of samples per second.
// For example, CD quality audio is encoded as 44100 samples per second.
func (v *File) SamplesPerSec() int {
	return int(v.samplesPerSec)
}

// Samples returns number of the samples that the audio contains.
// For example, 10 seconds of the stereo audio which is encoded 16 bit / 44.1 kHz contains 882000 samples.
func (v *File) Samples() int {
	return int(v.length) / int(v.blockAlign/v.channels)
}

// AvgBytesPerSec returns average bytes per second.
func (v *File) AvgBytesPerSec() int {
	return int(v.avgBytesPerSec)
}

// BlockAlign returns block align size in byte.
func (v *File) BlockAlign() int {
	return int(v.blockAlign)
}

// BitsPerSample returns bits per sample.
func (v *File) BitsPerSample() int {
	return int(v.bitsPerSample)
}

// Length returns size of the audio except for headers in bytes.
// The returned value is same as len(v.Bytes()).
func (v *File) Length() int {
	return int(v.length)
}

// Read reads audio samples byte by byte.
func (v *File) Read(p []byte) (int, error) {
	length := v.Length()
	size := len(p)

	for i := 0; i < size; i++ {
		if v.offset >= length {
			return i, io.EOF
		}
		p[i] = v.data[v.offset]
		v.offset++
	}

	return size, nil
}

// Write writes audio samples byte by byte.
func (v *File) Write(b []byte) (n int, err error) {
	size := len(b)

	for n = 0; n < size; n++ {
		v.data = append(v.data, b[n])
	}
	v.length += uint32(size)
	return
}

// Bytes returns audio samples as byte slice.
func (v *File) Bytes() []byte {
	return v.data
}

// String returns textual representation of audio.
func (v *File) String() string {
	return fmt.Sprintf("%v kHz / %v bit %v channel(s)", v.SamplesPerSec(), v.BitsPerSample(), v.Channels())
}

// Float64s returns audio samples as slice of float64.
func (v *File) Float64s() []float64 {
	const scale = 1 << 31
	samples := v.Samples()
	s32 := v.Int32s()
	f64 := make([]float64, samples)

	for i := 0; i < samples; i++ {
		f64[i] = float64(s32[i]) / scale
	}

	return f64
}

// Int32s returns audio samples as slice of int32.
func (v *File) Int32s() []int32 {
	switch v.BitsPerSample() {
	case 8:
		return v.fromS8ToInt32s()
	case 16:
		return v.fromS16ToInt32s()
	case 24:
		return v.fromS24ToInt32s()
	case 32:
		return v.fromS32ToInt32s()
	default:
		return []int32{}
	}
}

func (v *File) fromS8ToInt32s() []int32 {
	const scale = 1 << 24
	samples := v.Samples()
	s8 := make([]int8, samples)
	s32 := make([]int32, samples)

	binary.Read(bytes.NewBuffer(v.data), binary.LittleEndian, &s8)

	for i := 0; i < samples; i++ {
		s32[i] = int32(s8[i]) * scale
	}

	return s32
}

func (v *File) fromS16ToInt32s() []int32 {
	const scale = 1 << 16
	samples := v.Samples()
	s16 := make([]int16, samples)
	s32 := make([]int32, samples)

	binary.Read(bytes.NewBuffer(v.data), binary.LittleEndian, &s16)

	for i := 0; i < samples; i++ {
		s32[i] = int32(s16[i]) * scale
	}

	return s32
}

func (v *File) fromS24ToInt32s() []int32 {
	const scale = 1 << 8
	length := v.Length()
	samples := v.Samples()
	data := make([]byte, samples*4)
	s32 := make([]int32, samples)

	for i := 0; i < length; i += 3 {
		n := (i / 3) + 1
		data[n+i] = v.data[i]
		data[n+i+1] = v.data[i+1]
		data[n+i+2] = v.data[i+2]
	}

	binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &s32)

	for i := 0; i < samples; i++ {
		s32[i] *= scale
	}

	return s32
}

func (v *File) fromS32ToInt32s() []int32 {
	samples := v.Samples()
	s32 := make([]int32, samples)

	binary.Read(bytes.NewBuffer(v.data), binary.LittleEndian, &s32)

	return s32
}

// Unmarshal parses WAV formatted audio and store data into *File.
func Unmarshal(stream []byte, audio *File) (err error) {
	if audio == nil {
		err = fmt.Errorf("error: nil WAVE stream")
		return
	}

	reader := bytes.NewReader(stream)
	binary.Read(io.NewSectionReader(reader, 20, 2), binary.LittleEndian, &audio.formatTag)

	if !(audio.formatTag == WAVE_FORMAT_PCM || audio.formatTag == WAVE_FORMAT_EXTENSIBLE) {
		err = fmt.Errorf("error: invalid format tag '%v'", audio.formatTag)
		return
	}

	binary.Read(io.NewSectionReader(reader, 22, 2), binary.LittleEndian, &audio.channels)
	binary.Read(io.NewSectionReader(reader, 24, 4), binary.LittleEndian, &audio.samplesPerSec)
	binary.Read(io.NewSectionReader(reader, 28, 4), binary.LittleEndian, &audio.avgBytesPerSec)
	binary.Read(io.NewSectionReader(reader, 32, 2), binary.LittleEndian, &audio.blockAlign)
	binary.Read(io.NewSectionReader(reader, 34, 2), binary.LittleEndian, &audio.bitsPerSample)

	if audio.formatTag == WAVE_FORMAT_PCM {
		binary.Read(io.NewSectionReader(reader, 40, 4), binary.LittleEndian, &audio.length)
	} else if audio.formatTag == WAVE_FORMAT_EXTENSIBLE {
		binary.Read(io.NewSectionReader(reader, 76, 4), binary.LittleEndian, &audio.length)
	}

	buf := new(bytes.Buffer)
	if audio.formatTag == WAVE_FORMAT_PCM {
		io.Copy(buf, io.NewSectionReader(reader, 44, int64(audio.length)))
	} else if audio.formatTag == WAVE_FORMAT_EXTENSIBLE {
		io.Copy(buf, io.NewSectionReader(reader, 80, int64(audio.length)))
	}
	audio.data = buf.Bytes()

	return
}

// Marshal returns audio data as WAV formatted data.
func Marshal(v *File) (stream []byte, err error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, []byte("RIFF"))

	if v.formatTag == WAVE_FORMAT_PCM {
		binary.Write(buf, binary.LittleEndian, uint32(v.length+36))
	} else if v.formatTag == WAVE_FORMAT_EXTENSIBLE {
		binary.Write(buf, binary.LittleEndian, uint32(v.length+72))
	} else {
		err = fmt.Errorf("error: invalid format tag")
		return
	}

	binary.Write(buf, binary.BigEndian, []byte("WAVEfmt "))

	if v.formatTag == WAVE_FORMAT_PCM {
		binary.Write(buf, binary.LittleEndian, uint32(16))
	} else {
		binary.Write(buf, binary.LittleEndian, uint32(40))
	}

	binary.Write(buf, binary.LittleEndian, v.formatTag)
	binary.Write(buf, binary.LittleEndian, v.channels)
	binary.Write(buf, binary.LittleEndian, v.samplesPerSec)
	binary.Write(buf, binary.LittleEndian, v.avgBytesPerSec)
	binary.Write(buf, binary.LittleEndian, v.blockAlign)
	binary.Write(buf, binary.LittleEndian, v.bitsPerSample)

	if v.formatTag == WAVE_FORMAT_EXTENSIBLE {
		binary.Write(buf, binary.LittleEndian, uint16(22)) // cbSize
		// validBitsPerSample
		binary.Write(buf, binary.LittleEndian, v.bitsPerSample)
		// channelMask
		binary.Write(buf, binary.LittleEndian, uint32(getChannelMask(v.channels)))
		//binary.Write(buf, binary.LittleEndian, uint16(0))            // reserved
		guid := [16]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x80, 0x00, 0x00, 0xaa, 0x00, 0x38, 0x9b, 0x71}
		binary.Write(buf, binary.BigEndian, guid)
		binary.Write(buf, binary.BigEndian, []byte("fact"))                           // fact chunk is an optional chunk
		binary.Write(buf, binary.LittleEndian, uint32(4))                             // 4 bytes
		binary.Write(buf, binary.LittleEndian, uint32(v.length/uint32(v.blockAlign))) // zero padding
	}

	binary.Write(buf, binary.BigEndian, []byte("data"))
	binary.Write(buf, binary.LittleEndian, v.length)
	binary.Write(buf, binary.LittleEndian, v.data)
	stream = buf.Bytes()

	return
}

func getChannelMask(c uint16) (mask uint32) {
	if c == 1 {
		mask = 0x4
	} else if c == 2 {
		mask = 0x3 //
	} else if c == 4 {
		mask = 0x33
	} else if c == 6 {
		mask = 0x3f
	} else if c == 8 {
		mask = 0x63f
	}
	return
}

// New creates an empty File.
func New(samplesPerSec, bitsPerSample, channels int) (*File, error) {
	audio := &File{}

	if bitsPerSample > 16 {
		audio.formatTag = WAVE_FORMAT_EXTENSIBLE
	} else {
		audio.formatTag = WAVE_FORMAT_PCM
	}
	if bitsPerSample%8 != 0 {
		return nil, fmt.Errorf("wav: invalid bits per sample (%v bit)", bitsPerSample)
	}

	audio.samplesPerSec = uint32(samplesPerSec)
	audio.channels = uint16(channels)
	audio.bitsPerSample = uint16(bitsPerSample)
	audio.blockAlign = audio.channels * audio.bitsPerSample / 8
	audio.avgBytesPerSec = audio.samplesPerSec * uint32(audio.blockAlign)
	audio.data = []byte{}

	return audio, nil
}
