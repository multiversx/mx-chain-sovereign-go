package sovereign

// AbiSerializerMock -
type AbiSerializerMock struct {
	SerializeCalled   func(inputValues []any) (string, error)
	DeserializeCalled func(data string, outputValues []any) error
}

// Serialize -
func (m *AbiSerializerMock) Serialize(inputValues []any) (string, error) {
	if m.SerializeCalled != nil {
		return m.SerializeCalled(inputValues)
	}
	return "", nil
}

// Deserialize -
func (m *AbiSerializerMock) Deserialize(data string, outputValues []any) error {
	if m.DeserializeCalled != nil {
		return m.DeserializeCalled(data, outputValues)
	}
	return nil
}
