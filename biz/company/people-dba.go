package company

import (
	"adserver/motor"
)

func GetPersonOfUser(onTransit *motor.Convey) bool {
	result := onTransit.GetMapped("PersonOfUser")
	if result == "" {
		store := onTransit.Store()
		if store == nil {
			goto BadError
		}
		if !onTransit.Query(
			"SELECT pessoa FROM usuarios WHERE role_name = $1",
			store.User) {
			goto BadError
		}
		if !onTransit.Next() {
			goto BadError
		}
		value, err := onTransit.Take("pessoa")
		if err != nil {
			onTransit.PutError(err.Error())
			goto BadError
		}
		result := onTransit.FormatString(value)
		onTransit.SetMapped("PersonOfUser", result)
	}
	if result != "" {
		onTransit.Set("PersonOfUser", result)
		return true
	}
BadError:
	onTransit.PutError("can't get the person of user")
	return false
}
