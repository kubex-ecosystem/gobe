package types

import (
	"fmt"
	"reflect"

	logz "github.com/kubex-ecosystem/logz"
)

func AutoEncode[T any](v T, format, filePath string) ([]byte, error) {
	mapper := NewMapper[T](&v, filePath)
	if data, err := mapper.Serialize(format); err != nil {
		logz.Log("error", "AutoEncode: unknown type for serialization (", reflect.TypeOf(v).Name(), "):", err.Error())
		return nil, fmt.Errorf("error: %s", err.Error())
	} else {
		return data, nil
	}
}

func AutoDecode[T any](data []byte, target *T, format string) error {
	mapper := NewMapper[T](target, "")
	if obj, err := mapper.Deserialize(data, format); err != nil {
		logz.Log("error", "AutoDecode: unknown type for deserialization (", reflect.TypeOf(target).Name(), "):", err.Error())
		return fmt.Errorf("error: %s", err.Error())
	} else {
		if reflect.ValueOf(obj).IsValid() {
			target = obj
			logz.Log("success", "AutoDecode: deserialized object of type ", reflect.TypeOf(obj).Name())
		} else {
			logz.Log("error", "AutoDecode: deserialized object is nil")
			return fmt.Errorf("deserialized object is nil")
		}
		return nil
	}
}
