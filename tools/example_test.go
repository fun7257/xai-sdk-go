package tools_test

import (
	"fmt"

	"github.com/fun7257/xai-sdk-go/tools"
)

// Primary constructors validate option combinations and return errors.
func ExampleWebSearch() {
	web, err := tools.WebSearch(tools.WithAllowedDomains("example.com"))
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println(web.GetWebSearch() != nil)

	_, err = tools.WebSearch(
		tools.WithAllowedDomains("a.com"),
		tools.WithExcludedDomains("b.com"),
	)
	fmt.Println(err != nil)
	// Output:
	// true
	// true
}
