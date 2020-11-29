package swagger

import (
	"github.com/getkin/kin-openapi/openapi3"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestOrderPaths(t *testing.T) {
	paths := []string{"/products/*", "/tv/", "/products/tv", "/", "/products", "/*"}
	want := []string{"/products/tv", "/products", "/tv/", "/", "/products/*", "/*"}
	orderPaths(paths)

	isErr := false
	for i := range paths {
		if paths[i] != want[i] {
			isErr = true
		}
	}
	if isErr {
		t.Errorf("Ordered paths: %v, want: %v", paths, want)
	}
}

func TestPrettyStringOrderedByPath(t *testing.T) {
	swg, err := readJSONResourceFile("test_resources/sample-swagger.json")
	if err != nil {
		t.Fatal("Error reading sample swagger definition")
	}
	prettySwg := PrettyStringOrderedByPath(swg)

	want, err := readResource("test_resources/prettified-path-ordered-swagger.json")
	if err != nil {
		t.Fatal("Error reading sample resource")
	}
	s := string(want)
	if prettySwg != s {
		t.Error("Prettified json is not ordered with swagger paths")
	}
}

func readJSONResourceFile(path string) (*openapi3.Swagger, error) {
	bytes, err := readResource(path)
	if err != nil {
		return nil, err
	}
	s := string(bytes)

	return GetSwaggerV3(&s)
}

func readResource(path string) ([]byte, error) {
	return ioutil.ReadFile(filepath.FromSlash(path))
}
