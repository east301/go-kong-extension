package kongext

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func Test_CreateYAMLConfigFromFile(t *testing.T) {
	//
	config, err := CreateYAMLConfigFromFile("__INVALID__")
	if config != nil || err == nil {
		t.Error()
	}

	//
	path := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(path, []byte("INVALID"), 0640)

	config, err = CreateYAMLConfigFromFile(path)
	if config != nil || err == nil {
		t.Error()
	}

	//
	os.WriteFile(path, []byte("foo: bar"), 0640)

	config, err = CreateYAMLConfigFromFile(path)
	if config == nil || err != nil {
		t.Error()
	}

	if v := config("foo"); v.IsAbsent() || v.MustGet() != "bar" {
		t.Error()
	}
	if v := config("hoge"); v.IsPresent() {
		t.Error()
	}
}

func Test_CreateMapConfig(t *testing.T) {
	//
	obj := map[string]any{
		"foo":  "bar",
		"hoge": map[string]any{"piyo": 12345},
		"abc":  []string{"a", "b", "c"},
	}

	config := CreateMapConfig(obj)

	if v, ok := config("foo").Get(); !ok || v != "bar" {
		t.Fail()
	}
	if v, ok := config("hoge.piyo").Get(); !ok || v != 12345 {
		t.Fail()
	}
	if v, ok := config("abc").Get(); !ok || slices.Compare(v.([]string), []string{"a", "b", "c"}) != 0 {
		t.Fail()
	}
	if _, ok := config("invalid").Get(); ok {
		t.Fail()
	}
}

func Test_CreatStructConfig_FieldName(t *testing.T) {
	//
	type HogeStruct struct {
		Piyo int
	}
	type RootStruct struct {
		Foo  string
		Hoge HogeStruct
		ABC  []string
	}

	obj := RootStruct{
		Foo:  "bar",
		Hoge: HogeStruct{Piyo: 12345},
		ABC:  []string{"a", "b", "c"},
	}

	config := CreateStructConfig(obj)

	if v, ok := config("Foo").Get(); !ok || v != "bar" {
		t.Errorf("Foo: %+v, %+v", v, ok)
	}
	if v, ok := config("Hoge.Piyo").Get(); !ok || v != 12345 {
		t.Error()
	}
	if v, ok := config("ABC").Get(); !ok || slices.Compare(v.([]string), []string{"a", "b", "c"}) != 0 {
		t.Error()
	}
	if _, ok := config("invalid").Get(); ok {
		t.Error()
	}
}

func Test_CreatStructConfig_JSONTag(t *testing.T) {
	//
	type HogeStruct struct {
		Piyo int `json:"piyo"`
	}
	type RootStruct struct {
		Foo  string     `json:"foo"`
		Hoge HogeStruct `json:"hoge"`
		ABC  []string   `json:"abc"`
	}

	obj := RootStruct{
		Foo:  "bar",
		Hoge: HogeStruct{Piyo: 12345},
		ABC:  []string{"a", "b", "c"},
	}

	config := CreateStructConfig(obj)

	if v, ok := config("foo").Get(); !ok || v != "bar" {
		t.Errorf("Foo: %+v, %+v", v, ok)
	}
	if v, ok := config("hoge.piyo").Get(); !ok || v != 12345 {
		t.Error()
	}
	if v, ok := config("abc").Get(); !ok || slices.Compare(v.([]string), []string{"a", "b", "c"}) != 0 {
		t.Error()
	}
	if _, ok := config("invalid").Get(); ok {
		t.Error()
	}
}

func Test_CreatStructConfig_KongTag(t *testing.T) {
	//
	type HogeStruct struct {
		Piyo int `name:"piyo"`
	}
	type RootStruct struct {
		Foo  string     `name:"foo"`
		Hoge HogeStruct `name:"hoge"`
		ABC  []string   `name:"abc"`
	}

	obj := RootStruct{
		Foo:  "bar",
		Hoge: HogeStruct{Piyo: 12345},
		ABC:  []string{"a", "b", "c"},
	}

	config := CreateStructConfig(obj)

	if v, ok := config("foo").Get(); !ok || v != "bar" {
		t.Errorf("Foo: %+v, %+v", v, ok)
	}
	if v, ok := config("hoge.piyo").Get(); !ok || v != 12345 {
		t.Error()
	}
	if v, ok := config("abc").Get(); !ok || slices.Compare(v.([]string), []string{"a", "b", "c"}) != 0 {
		t.Error()
	}
	if _, ok := config("invalid").Get(); ok {
		t.Error()
	}
}
