package utils

import (
	"github.com/chef/omnitruck-service/logger"
	"github.com/sirupsen/logrus"
)

func AddLogFields(caller string, requestId string, logger logger.Logger) *logrus.Entry {
	fields := map[string]interface{}{
		"APIS":       caller,
		"Request_id": requestId,
	}

	return logger.WithFields(fields)
}
