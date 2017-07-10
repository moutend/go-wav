package wav

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

func TestNew(t *testing.T) {
	var a *File
	var err error

	if a, err = New(44100, 17, 2); err == nil {
		t.Fatalf("error must not be nil")
	}

	if a, err = New(44100, 16, 2); err != nil {
		t.Fatal(err)
	}
	if a.FormatTag() != WAVE_FORMAT_PCM {
		t.Fatalf("FormatTag should be %d but got %d", WAVE_FORMAT_PCM, a.FormatTag)
	}

	if a, err = New(96000, 24, 1); err != nil {
		t.Fatal(err)
	}
	if a.FormatTag() != WAVE_FORMAT_EXTENSIBLE {
		t.Fatalf("FormatTag should be %d but got %d", WAVE_FORMAT_EXTENSIBLE, a.FormatTag)
	}

	return
}

func TestUnmarshal(t *testing.T) {
	var audio *File
	var filename string
	var file []byte
	var err error

	tt := []struct {
		samples  int
		bits     int
		channels int
	}{
		{22050, 16, 1},
		{44100, 16, 2},
		{48000, 16, 2},
		{96000, 24, 2},
		{192000, 24, 2},
	}

	for _, v := range tt {
		filename = fmt.Sprintf("./testdata/%vHz-%vbit-%vch-empty.wav", v.samples, v.bits, v.channels)
		if file, err = ioutil.ReadFile(filename); err != nil {
			t.Fatal(err)
		}
		audio = &File{}
		if err = Unmarshal(file, audio); err != nil {
			t.Fatal(err)
		}
		if audio.SamplesPerSec() != v.samples {
			t.Errorf("expected: %v actual: %v (%v)\n", v.samples, audio.SamplesPerSec, filename)
		}
		if audio.BitsPerSample() != v.bits {
			t.Errorf("expected: %v actual: %v (%v)\n", v.bits, audio.BitsPerSample, filename)
		}
		if audio.Channels() != v.channels {
			t.Errorf("expected: %v actual: %v\n (%v)", v.channels, audio.Channels, filename)
		}
	}
	return
}

func TestMarshal(t *testing.T) {
	var actualBytes, expectedBytes, file []byte
	var audio *File
	var err error

	filenames := []string{
		"./testdata/sawtooth.wav",
	}
	for _, filename := range filenames {
		if file, err = ioutil.ReadFile(filename); err != nil {
			t.Fatal(err)
		}
		audio = &File{}
		if err = Unmarshal(file, audio); err != nil {
			t.Fatal(err)
		}
		if expectedBytes, err = ioutil.ReadFile(filename); err != nil {
			t.Fatal(err)
		}
		if actualBytes, err = Marshal(audio); err != nil {
			t.Fatal(err)
		}

		sizeOfExpectedBytes := len(expectedBytes)
		sizeOfActualBytes := len(actualBytes)

		if sizeOfExpectedBytes != sizeOfActualBytes {
			t.Fatalf("expected: %d actual: %d (%v)", sizeOfExpectedBytes, sizeOfActualBytes, filename)
		}
		for i, b := range expectedBytes {
			if b != actualBytes[i] {
				t.Fatalf("[%v] expected: %v actual: %v\n(%v)", i, b, actualBytes[i], filename)
			}
		}
	}
	return
}

func TestRead_(t *testing.T) {
	var audio *File
	var rawdata []byte
	var buf []byte
	var err error
	var file []byte

	audio = &File{}

	if file, err = ioutil.ReadFile("./testdata/sawtooth.wav"); err != nil {
		t.Fatal(err)
	}
	if err = Unmarshal(file, audio); err != nil {
		t.Fatal(err)
	}
	if rawdata, err = ioutil.ReadFile("./testdata/sawtooth.raw"); err != nil {
		t.Fatal(err)
	}
	if buf, err = ioutil.ReadAll(audio); err != nil {
		t.Fatal(err)
	}

	size := len(rawdata)

	for i := 0; i < size; i++ {
		if buf[i] != rawdata[i] {
			t.Fatalf("[%v] expected: %v actual: %v", i, rawdata[i], buf[i])
		}
	}
	return
}

func TestWrite_(t *testing.T) {
	var n int64
	var err error

	file, _ := ioutil.ReadFile("./testdata/sawtooth.wav")
	src := &File{}
	Unmarshal(file, src)
	dest, _ := New(src.SamplesPerSec(), src.BitsPerSample(), src.Channels())

	if n, err = io.Copy(dest, src); err != nil {
		t.Fatal(err)
	}
	if n != int64(src.Length()) {
		t.Errorf("expect: %v actual: %v", n, dest.Length())
	}
	return
}

func TestBytes(t *testing.T) {
	var audio *File
	var actualBytes, expectedBytes, file []byte
	var err error

	audio = &File{}
	if file, err = ioutil.ReadFile("./testdata/sawtooth.wav"); err != nil {
		t.Fatal(err)
	}
	if err = Unmarshal(file, audio); err != nil {
		t.Fatal(err)
	}
	if expectedBytes, err = ioutil.ReadFile("./testdata/sawtooth.raw"); err != nil {
		t.Fatal(err)
	}

	actualBytes = audio.Bytes()
	sizeOfExpectedBytes := len(expectedBytes)
	sizeOfActualBytes := len(actualBytes)

	if sizeOfExpectedBytes != sizeOfActualBytes {
		t.Fatalf("expected: %d actual: %d", sizeOfExpectedBytes, sizeOfActualBytes)
	}
	for i, b := range expectedBytes {
		if b != actualBytes[i] {
			t.Fatalf("[%v] expected: %v actual: %v\n", i, b, actualBytes[i])
		}
	}
	return
}

func TestInt32s(t *testing.T) {
	var audio *File
	var actualBytes, expectedBytes, file []byte
	var err error

	audio = &File{}
	if file, err = ioutil.ReadFile("./testdata/sawtooth.wav"); err != nil {
		t.Fatal(err)
	}
	if err = Unmarshal(file, audio); err != nil {
		t.Fatal(err)
	}
	if expectedBytes, err = ioutil.ReadFile("./testdata/sawtooth.s32"); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, audio.Int32s())
	actualBytes = buf.Bytes()

	sizeOfExpectedBytes := len(expectedBytes)
	sizeOfActualBytes := len(actualBytes)

	if sizeOfExpectedBytes != sizeOfActualBytes {
		t.Fatalf("expected: %d actual: %d", sizeOfExpectedBytes, sizeOfActualBytes)
	}
	for i, b := range expectedBytes {
		if b != actualBytes[i] {
			t.Fatalf("[%v] expected: %v actual: %v\n", i, b, actualBytes[i])
		}
	}
	return
}

func TestFloat64s(t *testing.T) {
	var audio *File
	var actualBytes, expectedBytes, file []byte
	var err error

	audio = &File{}
	if file, err = ioutil.ReadFile("./testdata/sawtooth.wav"); err != nil {
		t.Fatal(err)
	}
	if err = Unmarshal(file, audio); err != nil {
		t.Fatal(err)
	}
	if expectedBytes, err = ioutil.ReadFile("./testdata/sawtooth.f64"); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, audio.Float64s())
	actualBytes = buf.Bytes()

	sizeOfExpectedBytes := len(expectedBytes)
	sizeOfActualBytes := len(actualBytes)

	if sizeOfExpectedBytes != sizeOfActualBytes {
		t.Fatalf("expected: %d actual: %d", sizeOfExpectedBytes, sizeOfActualBytes)
	}
	for i, b := range expectedBytes {
		if b != actualBytes[i] {
			t.Fatalf("[%v] expected: %v actual: %v\n", i, b, actualBytes[i])
		}
	}
	return
}
