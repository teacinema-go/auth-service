package enum

type IdentifierType string

const (
	IdentifierPhone IdentifierType = "phone"
	IdentifierEmail IdentifierType = "email"
)

func (i IdentifierType) IsValid() bool {
	switch i {
	case IdentifierPhone, IdentifierEmail:
		return true
	}
	return false
}
