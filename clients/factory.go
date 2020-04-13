package clients

import "github.com/Evi1/awsl/model"

// NewClients new clients
func NewClients(conf []model.Out) []Client {
	r := make([]Client, 0, len(conf))
	for _, v := range conf {
		switch v.Type {
		case "direct":
			r = append(r, DirectOut{})
		case "awsl":
			r = append(r, NewAWSL(v.Awsl.Host, v.Awsl.Port, v.Awsl.URI, v.Awsl.Auth, v.Awsl.BackUp))
		/*case "ahl":
		r = append(r, NewAHL(v.Awsl.Host, v.Awsl.Port, v.Awsl.URI, v.Awsl.Auth))*/
		case "h2c":
			r = append(r, NewH2C(v.Awsl.Host, v.Awsl.Port, v.Awsl.URI, v.Awsl.Auth, v.Awsl.BackUp))
		default:
			panic(v.Type)
		}
	}
	return r
}
