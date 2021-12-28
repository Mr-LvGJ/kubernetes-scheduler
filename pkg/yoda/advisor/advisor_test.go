package advisor

import (
	"fmt"
	"testing"
)

func TestInfo(t *testing.T) {
	var res Result

	info, err := res.Init()
	if err != nil {
		t.Errorf("failed to get info , err: %v", err)
	}
	for i := range info.Info {
		fmt.Println(info.Info[i])
	}
}
