package tests

import (
	"fmt"
	"testing"

	at "github.com/rafa-mori/gobe/internal/types"
	atc "github.com/rafa-mori/gobe/internal/types"
	gl "github.com/rafa-mori/gobe/logger"
)

func TestEncodeJSONInputProducesJSONOutput(t *testing.T) {
	input := map[string]string{"key": "value"}
	expected := `{"key":"value"}`

	output, err := atc.AutoEncode[map[string]string](input, "json", "./TestEncodeJSONInputProducesJSONOutput.json")
	if err != nil {
		gl.Log("error", fmt.Sprintf("unexpected error: %v", err))
		t.Fatalf("unexpected error: %v", err)
	}

	if string(output) != expected {
		gl.Log("error", fmt.Sprintf("expected '%s', got '%s'", expected, string(output)))
		t.Errorf("expected %s, got %s", expected, string(output))
	}
}

func TestDecodeJSONInputProducesCorrectStruct(t *testing.T) {
	input := []byte(`{"key":"value"}`)
	var obj map[string]string

	mapper := at.NewMapperPtr[map[string]string](&obj, "")
	out, err := mapper.Deserialize(input, "json")
	if err != nil {
		gl.Log("error", fmt.Sprintf("unexpected error: %v", err))
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		gl.Log("error", "expected non-nil output, got nil")
		t.Fatalf("expected non-nil output, got nil")
	}

	outA := *out
	output := *outA

	if output != nil && output["key"] != "value" {
		gl.Log("error", fmt.Sprintf("expected value to be 'value', got '%s'", (output)["key"]))
		t.Errorf("expected value to be 'value', got '%s'", (output)["key"])
	}
	gl.Log("info", fmt.Sprintf("expected value to be 'value', got '%s'", (output)["key"]))
}

func TestEncodeUnsupportedTypeReturnsError(t *testing.T) {
	input := make(chan int)

	_, err := atc.AutoEncode[chan int](input, "json", "./TestEncodeUnsupportedTypeReturnsError.json")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestDecodeUnsupportedTypeReturnsError(t *testing.T) {
	input := []byte(`{"key":"value"}`)
	var output chan int

	err := atc.AutoDecode[chan int](input, &output, "json")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestEncodeTOMLInputProducesTOMLOutput(t *testing.T) {
	input := map[string]string{"key": "value"}
	expected := "key = 'value'\n"

	output, err := atc.AutoEncode[map[string]string](input, "toml", "./TestEncodeTOMLInputProducesTOMLOutput.toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(output) != expected {
		gl.Log("error", fmt.Sprintf("expected %s, got %s", expected, string(output)))
		t.Errorf("expected %s, got '%s'", expected, string(output))
	}
}

func TestEncodeTOMLInputProducesTOMLOutputB(t *testing.T) {
	input := map[string]string{"key": "value"}
	expected := "key = \"value\"\n"

	output, err := atc.AutoEncode[map[string]string](input, "env", "./TestEncodeTOMLInputProducesTOMLOutput.toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !at.IsEqual(string(output), expected) {
		gl.Log("error", fmt.Sprintf("expected %s, got %s", expected, string(output)))
		t.Errorf("expected %s, got '%s'", expected, string(output))
	} else {
		gl.Log("info", fmt.Sprintf("expected %s, got %s", expected, string(output)))
	}
}

func TestDecodeTOMLInputProducesCorrectStruct(t *testing.T) {
	input := []byte(`key = "value"`)
	var output map[string]string

	err := atc.AutoDecode[map[string]string](input, &output, "toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !at.IsEqual(output["key"], "value") {
		gl.Log("error", fmt.Sprintf("expected value to be 'value', got '%s'", output["key"]))
		t.Errorf("expected value to be 'value', got '%v'", output["key"])
	}
}

func TestEncodeYAMLInputProducesYAMLOutput(t *testing.T) {
	input := map[string]string{"key": "value"}
	expected := "key: value\n"

	output, err := atc.AutoEncode[map[string]string](input, "yaml", "./TestEncodeYAMLInputProducesYAMLOutput.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(output) != expected {
		t.Errorf("expected %s, got %s", expected, string(output))
	}
}

func TestDecodeYAMLInputProducesCorrectStruct(t *testing.T) {
	input := []byte("key: value\n\n")
	var output = make(map[string]string)

	err := atc.AutoDecode[map[string]string](input, &output, "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output["key"] != "value" {
		t.Errorf("expected value to be 'value', got %s", output["key"])
	}
}

func TestEncodeXMLInputProducesXMLOutput(t *testing.T) {
	input := struct {
		XMLName struct{} `xml:"root"`
		Key     string   `xml:"key"`
	}{Key: "value"}
	expected := `<root><key>value</key></root>`

	output, err := atc.AutoEncode[struct {
		XMLName struct{} `xml:"root"`
		Key     string   `xml:"key"`
	}](input, "xml", "./TestEncodeXMLInputProducesXMLOutput.xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(output) != expected {
		t.Errorf("expected %s, got %s", expected, string(output))
	}
}

func TestEncodeXMLInputProducesXMLOutputB(t *testing.T) {
	input := struct {
		XMLName struct{} `xml:"root"`
		Key     string   `xml:"key"`
	}{Key: "value"}
	expected := `<root><key>value</key></root>`

	output, err := atc.AutoEncode[struct {
		XMLName struct{} `xml:"root"`
		Key     string   `xml:"key"`
	}](input, "xml", "./TestEncodeXMLInputProducesXMLOutputB.xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(output) != expected {
		t.Errorf("expected %s, got %s", expected, string(output))
	}
}

func TestDecodeXMLInputProducesCorrectStruct(t *testing.T) {
	input := []byte(`<root><key>value</key></root>`)
	var output struct {
		Key string `xml:"key"`
	}

	err := atc.AutoDecode[struct {
		Key string `xml:"key"`
	}](input, &output, "xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Key != "value" {
		t.Errorf("expected value to be 'value', got %s", output.Key)
	}
}
