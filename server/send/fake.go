package send

import (
	log "github.com/sirupsen/logrus"
)

type Faker struct {
}

func (f Faker) Send(msg string) error {
	log.Info(msg)
	return nil
}
