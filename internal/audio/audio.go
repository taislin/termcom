package audio

import (
	"fmt"
)

// PlayBeep produces a system beep
func PlayBeep() {
	fmt.Print("\a")
}
