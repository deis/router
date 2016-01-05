package modeler

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const (
	prefix string = "sample"
	tag    string = "sample"
)

var (
	sampleData = make(map[string]string, 0)
	m          = NewModeler(prefix, tag)
)

type SampleModel struct {
	SampleString      string          `sample:"a_string"`
	SampleInt         int             `sample:"an_int"`
	SampleBool        bool            `sample:"a_bool"`
	SampleStringSlice []string        `sample:"a_string_slice"`
	SampleSubModel    *SampleSubModel `sample:"a_submodel"`
	UnmappedString    string
}

func newSampleModel() *SampleModel {
	return &SampleModel{SampleSubModel: &SampleSubModel{}}
}

type BadSampleModel struct {
	SampleString      string         `sample:"a_string"`
	SampleInt         int            `sample:"an_int"`
	SampleBool        bool           `sample:"a_bool"`
	SampleStringSlice []string       `sample:"a_string_slice"`
	SampleSubModel    SampleSubModel `sample:"a_submodel"`
	UnmappedString    string
}

func newBadSampleModel() *BadSampleModel {
	return &BadSampleModel{SampleSubModel: SampleSubModel{}}
}

type SampleSubModel struct {
	SampleString      string   `sample:"a_string"`
	SampleInt         int      `sample:"an_int"`
	SampleBool        bool     `sample:"a_bool"`
	SampleStringSlice []string `sample:"a_string_slice"`
}

func TestNilLiteral(t *testing.T) {
	err := m.MapToModel(sampleData, nil)
	checkError(t, "modeler.NilLiteralModelError", err)
}

func TestNil(t *testing.T) {
	var sampleModel *SampleModel
	err := m.MapToModel(sampleData, sampleModel)
	checkError(t, "modeler.NilModelError", err)
}

func TestNonPointer(t *testing.T) {
	sampleModel := SampleModel{}
	err := m.MapToModel(sampleData, sampleModel)
	checkError(t, "modeler.NonPointerModelError", err)
}

func TestNonStructPointer(t *testing.T) {
	sampleModel := "foo"
	err := m.MapToModel(sampleData, &sampleModel)
	checkError(t, "modeler.NonStructPointerModelError", err)
}

func TestNonPointerSubModel(t *testing.T) {
	sampleModel := newBadSampleModel()
	err := m.MapToModel(sampleData, sampleModel)
	checkError(t, "modeler.NonPointerModelError", err)
}

func TestMapping(t *testing.T) {
	sampleModel := newSampleModel()
	err := m.MapToModel(sampleData, sampleModel)
	if err != nil {
		t.Error(err)
	}
	checkStringField(t, sampleData[prefix+"/a_string"], sampleModel.SampleString)
	checkIntField(t, sampleData[prefix+"/an_int"], sampleModel.SampleInt)
	checkBoolField(t, sampleData[prefix+"/a_bool"], sampleModel.SampleBool)
	checkStringSliceField(t, sampleData[prefix+"/a_string_slice"], sampleModel.SampleStringSlice)
	checkStringField(t, sampleData[prefix+"/a_submodel.a_string"], sampleModel.SampleSubModel.SampleString)
	checkIntField(t, sampleData[prefix+"/a_submodel.an_int"], sampleModel.SampleSubModel.SampleInt)
	checkBoolField(t, sampleData[prefix+"/a_submodel.a_bool"], sampleModel.SampleSubModel.SampleBool)
	checkStringSliceField(t, sampleData[prefix+"/a_submodel.a_string_slice"], sampleModel.SampleSubModel.SampleStringSlice)
}

func checkError(t *testing.T, want string, err error) {
	if err == nil {
		t.Errorf("Expected a %s, but did not receive any error", want)
		t.FailNow()
	}
	if got := reflect.TypeOf(err).String(); want != got {
		t.Errorf("Expected a %s, but got a %s", want, got)
	}
}

func checkStringField(t *testing.T, want string, got string) {
	if want != got {
		t.Errorf("Expected %s, but got %s", want, got)
	}
}

func checkIntField(t *testing.T, want string, got int) {
	wantInt, err := strconv.Atoi(want)
	if err != nil {
		t.Error(err)
	}
	if wantInt != got {
		t.Errorf("Expected %d, but got %d", wantInt, got)
	}
}

func checkBoolField(t *testing.T, want string, got bool) {
	wantBool, err := strconv.ParseBool(want)
	if err != nil {
		t.Error(err)
	}
	if wantBool != got {
		t.Errorf("Expected %t, but got %t", wantBool, got)
	}
}

func checkStringSliceField(t *testing.T, want string, got []string) {
	wantStringSlice := strings.Split(want, ",")
	if !reflect.DeepEqual(wantStringSlice, got) {
		t.Errorf("Expected %s, but got %s", wantStringSlice, got)
	}
}

func TestMain(m *testing.M) {
	initializeSampleData()
	os.Exit(m.Run())
}

func initializeSampleData() {
	// Deliberately setting each of these to string representations of non-zero values for their
	// respective types.
	sampleData[prefix+"/a_string"] = "foobar"
	sampleData[prefix+"/an_int"] = "5"
	sampleData[prefix+"/a_bool"] = "true"
	sampleData[prefix+"/a_string_slice"] = "foo,bar,baz"
	// Key/values for a model within the model...
	sampleData[prefix+"/a_submodel.a_string"] = "foobar"
	sampleData[prefix+"/a_submodel.an_int"] = "5"
	sampleData[prefix+"/a_submodel.a_bool"] = "true"
	sampleData[prefix+"/a_submodel.a_string_slice"] = "foo,bar,baz"
}
