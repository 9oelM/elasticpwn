package EPUtils

import (
	"fmt"
)

func ValidateStringFlag(flagValue interface{}, defaultFlagValue interface{}, flagName string) bool {
	hasError := false
	switch flagValue.(type) {
	case string:
		{
			defaultFlagValueStr, ok := defaultFlagValue.(string)
			if flagValue == defaultFlagValueStr && ok {
				fmt.Printf("%v option is not set. Please try again.\n", flagName)
				hasError = true
			}
		}
	}

	return hasError
}

func ValidatePositiveInt(flagValue int, flagName string) bool {
	hasError := false
	if flagValue < 1 {
		fmt.Printf("%v option can't be less than 1. Input a positive number.\n", flagName)
		hasError = true
	}

	return hasError
}
