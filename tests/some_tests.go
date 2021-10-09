package tests

import (
	"fmt"
	"testing"

	"github.com/gookit/validate"
	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	assert := assert.New(t)

	assert.Nil(nil)

	expectValue := "expect value"
	assert.Equal(expectValue, "expect value")
}

func TestValidate(t *testing.T) {
	// assert := assert.New(t)

	m := map[string]interface{}{
		"name":  "enumA",
		"age":   100,
		"oldSt": 1,
		"newSt": 2,
		"email": "some@email.com",
	}

	v := validate.Map(m)
	v.StringRule("name", "required|string|enum:enumA,enumB")

	if v.Validate() { // validate ok
		fmt.Println("ok")
	} else {
		fmt.Println(v.Errors)       // all error messages
		fmt.Println(v.Errors.One()) // returns a random error message text
	}
}
