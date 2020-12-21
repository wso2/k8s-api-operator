package registry

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestGetPathWithoutReg(t *testing.T) {
	datas :=
		[]struct {
			in  string
			out string
		}{
			{
				in:  "192.168.8.132:5000/apis-test",
				out: "apis-test",
			},
			{
				in:  "192.168.8.132:5000",
				out: "",
			},
		}

	for _, data := range datas {
		result := getPathWithoutReg(data.in)
		assert.Equal(t, result, data.out)
	}
}
