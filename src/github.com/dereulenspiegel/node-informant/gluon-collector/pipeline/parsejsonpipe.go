package pipeline

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/dereulenspiegel/node-informant/announced"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
)

type JsonParsePipe struct {
}

func (j *JsonParsePipe) Process(in chan announced.Response) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			respondInfo := &data.RespondNodeinfo{}
			err := json.Unmarshal(response.Payload, respondInfo)
			if err != nil {
				log.WithFields(log.Fields{
					"error":  err,
					"client": response.ClientAddr,
					"json":   string(response.Payload),
				}).Error("Error parsing json")
			} else {
				if respondInfo.Nodeinfo != nil {
					out <- data.NodeinfoResponse{
						Nodeinfo: *respondInfo.Nodeinfo,
					}
				}
				if respondInfo.Statistics != nil {
					out <- data.StatisticsResponse{
						Statistics: respondInfo.Statistics,
					}
				}
				if respondInfo.Neighbours != nil {
					out <- data.NeighbourReponse{
						Neighbours: respondInfo.Neighbours,
					}
				}
			}
		}
	}()
	return out
}
