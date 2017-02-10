package proto

func MarshalWithBytes(pb Message, data []byte) ([]byte, error) {
	p := NewBuffer(data)
	err := p.Marshal(pb)
	var state errorState
	if err != nil && !state.shouldContinue(err, nil) {
		return nil, err
	}
	return p.buf, err
}
