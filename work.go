package goqueue

import "log"
import "strconv"

type worker interface {
	start()
	getQName() string
}

func dispatch(w worker, n int) {
	log.Println("Starting " + strconv.Itoa(n) + " workers for queue " + w.getQName())
	for i := 0; i < n; i++ {
		go func() {
			w.start()
		}()
	}
}
