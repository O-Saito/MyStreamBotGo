package processors

var stopCh = make(chan struct{})

func StopProcessors() {
	close(stopCh)
}
