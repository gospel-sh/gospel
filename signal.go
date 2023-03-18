package gospel

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitForInterrupt() {

	sigchan := make(chan os.Signal, 1)

	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// we wait for CTRL-C / Interrupt
	<-sigchan

}
