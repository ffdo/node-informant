package data

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/announced"
	"github.com/dereulenspiegel/node-informant/utils"
)

type ReceivePipe interface {
	Process(in chan announced.Response) chan announced.Response
}

type ReceivePipeline struct {
	head chan announced.Response
	tail chan ParsedResponse
}

type ParsePipe interface {
	Process(in chan announced.Response) chan ParsedResponse
}

func (pipeline *ReceivePipeline) Enqueue(response announced.Response) {
	pipeline.head <- response
}

func (pipeline *ReceivePipeline) Dequeue(handler func(ParsedResponse)) {
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
	Process(in chan ParsedResponse) chan ParsedResponse
}

type ProcessPipeline struct {
	head chan ParsedResponse
	tail chan ParsedResponse
}

func (pipeline *ProcessPipeline) Close() {
	close(pipeline.head)
	close(pipeline.tail)
}

func (pipeline *ProcessPipeline) Enqueue(response ParsedResponse) {
	pipeline.head <- response
}

func (pipeline *ProcessPipeline) Dequeue(handler func(ParsedResponse)) {
	for response := range pipeline.tail {
		handler(response)
	}
}

func NewProcessPipeline(pipes ...ProcessPipe) *ProcessPipeline {
	head := make(chan ParsedResponse)
	var next_chan chan ParsedResponse
	for _, pipe := range pipes {
		if next_chan == nil {
			next_chan = pipe.Process(head)
		} else {
			next_chan = pipe.Process(next_chan)
		}
	}
	return &ProcessPipeline{head: head, tail: next_chan}
}

type DeflatePipe struct {
}

func (d *DeflatePipe) Process(in chan announced.Response) chan announced.Response {
	out := make(chan announced.Response)
	go func() {
		for response := range in {
			decompressedData, err := utils.Deflate(response.Payload)
			if err != nil {
				log.WithFields(log.Fields{
					"error":   err,
					"client":  response.ClientAddr,
					"payload": response.Payload,
				}).Error("Error deflating response")
			} else {
				response.Payload = decompressedData
				out <- response
			}
		}
	}()
	return out
}

type JsonParsePipe struct {
}

func (j *JsonParsePipe) Process(in chan announced.Response) chan ParsedResponse {
	out := make(chan ParsedResponse)
	go func() {
		for response := range in {
			respondInfo := &RespondNodeinfo{}
			err := json.Unmarshal(response.Payload, respondInfo)
			if err != nil {
				log.WithFields(log.Fields{
					"error":  err,
					"client": response.ClientAddr,
					"json":   string(response.Payload),
				}).Error("Error parsing json")
			} else {
				if respondInfo.Nodeinfo != nil {
					out <- NodeinfoResponse{
						Nodeinfo: *respondInfo.Nodeinfo,
					}
				}
				if respondInfo.Statistics != nil {
					out <- StatisticsResponse{
						Statistics: *respondInfo.Statistics,
					}
				}
				if respondInfo.Neighbours != nil {
					out <- NeighbourReponse{
						Neighbours: *respondInfo.Neighbours,
					}
				}
			}
		}
	}()
	return out
}
