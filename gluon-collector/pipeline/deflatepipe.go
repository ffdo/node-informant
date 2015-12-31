package pipeline

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ffdo/node-informant/announced"
	"github.com/ffdo/node-informant/utils"
)

// DeflatePipe tries to decompress the payload of all received Responses with
// deflate algorithm. All Responses which can't be deflated are discarded and
// written to the error log.
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
				response.Errored = true
			} else {
				response.Payload = decompressedData
			}
			out <- response
		}

	}()
	return out
}
