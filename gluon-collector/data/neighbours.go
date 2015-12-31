package data

/*
{
  "neighbours": {
    "batadv": {
      "c6:71:20:2c:c4:18": {
        "neighbours": {
          "16:cf:21:30:c9:30": {
            "lastseen": 101.55,
            "tq": 84
          },
          "ea:97:f7:06:2e:0c": {
            "lastseen": 3.94,
            "tq": 208
          }
        }
      }
    },
    "node_id": "c46e1f2cc418",
    "wifi": {
      "c6:71:20:2c:c4:18": {
        "neighbours": {
          "16:cf:21:30:c9:30": {
            "inactive": 40,
            "noise": -95,
            "signal": -85
          },
          "16:cf:21:62:f7:fe": {
            "inactive": 4760,
            "noise": -95,
            "signal": -90
          },
          "ea:97:f7:06:2e:0c": {
            "inactive": 20,
            "noise": -95,
            "signal": -73
          }
        }
      }
    }
  }
}
*/

type WifiLink struct {
	Inactive int `json:"inactive"`
	Noise    int `json:"nois"`
	Signal   int `json:"signal"`
}

type BatmanLink struct {
	Lastseen float64 `json:"lastseen"`
	Tq       int     `json:"tq"`
}

type BatadvNeighbours struct {
	Neighbours map[string]BatmanLink `json:"neighbours"`
}

type WifiNeighbours struct {
	Neighbours map[string]WifiLink `json:"neighbours"`
}

type NeighbourStruct struct {
	Batadv map[string]BatadvNeighbours `json:"batadv"`
	//WifiNeighbours map[string]WifiNeighbours   `json:"wifi"`
	NodeId string `json:"node_id"`
}
