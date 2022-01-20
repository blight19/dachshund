package send

type Sender interface {
	Send(msg map[string]interface{}) error
}
