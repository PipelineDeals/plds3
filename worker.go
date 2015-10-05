package main

type Worker struct {
	id int
	q  chan *Upload
}

func NewWorker(id int, q chan *Upload) Worker {
	return Worker{
		id: id,
		q:  q,
	}
}

func (w Worker) start() {
	go func() {
		for {
			select {
			case upload := <-w.q:
				upload.Put()
				upload.WaitGroup.Done()
			}
		}
	}()
}
