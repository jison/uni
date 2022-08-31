package valuer

type identityValuer struct{}

func (v *identityValuer) ValueOne(input Value) Value {
	return input
}

func (v *identityValuer) String() string {
	return "identity"
}

func (v *identityValuer) Clone() OneInputValuer {
	return &identityValuer{}
}

func (v *identityValuer) Equal(other interface{}) bool {
	if v == nil && other == nil {
		return true
	}

	_, ok := other.(*identityValuer)
	return ok
}

func Identity() Valuer {
	return &oneInputValuer{&identityValuer{}}
}
