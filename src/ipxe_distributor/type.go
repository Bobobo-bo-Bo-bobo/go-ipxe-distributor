package main

func getInterfaceType(i interface{}) int {
	switch i.(type) {
	case nil:
		return TypeNil
	case bool:
		return TypeBool
	case int:
		return TypeInt
	case int32:
		return TypeInt
	case byte:
		return TypeByte
	case float64:
		return TypeFloat
	case float32:
		return TypeFloat
	case string:
		return TypeString
	default:
		return TypeOther
	}
}
