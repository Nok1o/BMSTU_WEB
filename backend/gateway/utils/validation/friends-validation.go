package validation

const (
	TypeOutComing = "outcoming"
	TypeInComing  = "incoming"
	TypeFriend    = "friend"
	IncomingNew   = "incoming_new"
)

var acceptedReqTypesSet = map[string]struct{}{
	TypeOutComing: {},
	TypeInComing:  {},
	TypeFriend:    {},
	IncomingNew:   {},
}

func ValidateFriendReqType(reqType string) bool {
	if _, ok := acceptedReqTypesSet[reqType]; !ok {
		return false
	}

	return true
}
