package pipeline

import (
	"github.com/dereulenspiegel/node-informant/announced"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
)

type Closeable interface {
	Close()
}

type ReceivePipe interface {
	Process(in chan announced.Response) chan announced.Response
}

type ReceivePipeline struct {
	head chan announced.Response
	tail chan data.ParsedResponse
}

type ParsePipe interface {
	Process(in chan announced.Response) chan data.ParsedResponse
}

func (pipeline *ReceivePipeline) Enqueue(response announced.Response) {
	pipeline.head <- response
}

func (pipeline *ReceivePipeline) Dequeue(handler func(data.ParsedResponse)) {
	for response := range pipeline.tail {
		handler(response)
	}
}

func (pipeline *ReceivePipeline) Close() {
	close(pipeline.head)
	close(pipeline.tail)
}

func NewReceivePipeline(parsePipe ParsePipe, pipes ...ReceivePipe) *ReceivePipeline {
	head := make(chan announced.Response)
	var next_chan chan announced.Response
	for _, pipe := range pipes {
		if next_chan == nil {
			next_chan = pipe.Process(head)
		} else {
			next_chan = pipe.Process(next_chan)
		}
	}
	last_chan := parsePipe.Process(next_chan)
	return &ReceivePipeline{head: head, tail: last_chan}
}

type ProcessPipe interface {
	Process(in chan data.ParsedResponse) chan data.ParsedResponse
}

type ProcessPipeline struct {
	head chan data.ParsedResponse
	tail chan data.ParsedResponse
}

func (pipeline *ProcessPipeline) Close() {
	close(pipeline.head)
	close(pipeline.tail)
}

func (pipeline *ProcessPipeline) Enqueue(response data.ParsedResponse) {
	pipeline.head <- response
}

func (pipeline *ProcessPipeline) Dequeue(handler func(data.ParsedResponse)) {
	for response := range pipeline.tail {
		handler(response)
	}
}

func NewProcessPipeline(pipes ...ProcessPipe) *ProcessPipeline {
	head := make(chan data.ParsedResponse)
	var next_chan chan data.ParsedResponse
	for _, pipe := range pipes {
		if next_chan == nil {
			next_chan = pipe.Process(head)
		} else {
			next_chan = pipe.Process(next_chan)
		}
	}
	return &ProcessPipeline{head: head, tail: next_chan}
}
