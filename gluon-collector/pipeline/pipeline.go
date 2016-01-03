package pipeline

import (
	"github.com/ffdo/node-informant/announced"
	"github.com/ffdo/node-informant/gluon-collector/data"
)

// ReceivePipe needs to implemented by all types which want to participate in the
// process of receiving a response from announced and transforming it into something
// usable. This can be deflating and parsing the date for example.
type ReceivePipe interface {

	// Process is called by the ReceivePipeline to enqueue new Responses into this pipe
	// and to retrieve the outgoing channel with further processed Responses to connect
	// to the next ReceivePipe.
	Process(in chan announced.Response) chan announced.Response
}

// ReceivePipeline is the type which connects all ReceivePipes to a ParsePipe.
type ReceivePipeline struct {
	head chan announced.Response
	tail chan data.ParsedResponse
}

// NewReceivePipeline creates a new ReceivePipeline which puts all received Responses
// through all ReceivePipes and at the end let the result be parsed by the ParsePipe.
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

// Enqueue gives a common interface to enqueue received Responses into the ReceivePipeline
func (pipeline *ReceivePipeline) Enqueue(response announced.Response) {
	pipeline.head <- response
}

// Dequeue gives a common interface to pull parsed Responses from the ReceivePipeline
func (pipeline *ReceivePipeline) Dequeue(handler func(data.ParsedResponse)) {
	for response := range pipeline.tail {
		handler(response)
	}
}

// ParsePipe needs to be implemented by a type which wants to parse to the received
// Response into usable information which can then be used by the ProcessPipeline
type ParsePipe interface {
	Process(in chan announced.Response) chan data.ParsedResponse
}

func (pipeline *ReceivePipeline) Close() error {
	close(pipeline.head)
	close(pipeline.tail)
	return nil
}

func FeedParsedResponseReaders(readers []data.ParsedResponseReader, store data.Nodeinfostore, response data.ParsedResponse) {
	for _, reader := range readers {
		reader(store, response)
	}
}
