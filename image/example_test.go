package image_test

import (
	"fmt"

	"github.com/fun7257/xai-sdk-go/image"
)

// Preferred multi-image shape uses Samples + WithN (construct request offline).
func ExampleClient_Prepare() {
	cli := image.New(nil) // Prepare only builds the request; no RPC
	req, err := cli.Prepare("a cat", "grok-imagine-image", image.WithN(2))
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println(req.GetPrompt() != "" && req.GetN() == 2)
	// Output: true
}
