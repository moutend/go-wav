# go-wav

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]

[license]: https://github.com/moutend/go-wave/blob/master/LICENSE

Package `go-wav` reads and writes WAV file.

# Example

The following example concatinates `input1.wav` and `input2.wav` into `output.wav`.

```go
package main

import (
	"io"
	"io/ioutil"

	"github.com/moutend/go-wav"
)

func main() {
	// Read input1.wav and input2.wav
	i1, _ := ioutil.ReadFile("input1.wav")
	i2, _ := ioutil.ReadFile("input2.wav")

	// Create wav.File.
	a := &wav.File{}
	b := &wav.File{}

	// Unmarshal input1.wav and input2.wav.
	wav.Unmarshal(i1, a)
	wav.Unmarshal(i2, b)

	// Add input2.wav to input1.wav.
	c, _ := wav.New(a.SamplesPerSec(), a.BitsPerSample(), a.Channels())
	io.Copy(c, a)
	io.Copy(c, b)

	// Marshal input1.wav and save result.
	file, _ := wav.Marshal(c)
	ioutil.WriteFile("output.wav", file, 0644)
}
```

Note that the example assumes that the two input files have same sample rate, bit depth and channels.

## Contributing

1. Fork ([https://github.com/moutend/go-wca/fork](https://github.com/moutend/go-wca/fork))
1. Create a feature branch
1. Add changes
1. Run `go fmt` and `go test`
1. Commit your changes
1. Open a new Pull Request

## Author

[Yoshiyuki Koyanagi](https://github.com/moutend)
