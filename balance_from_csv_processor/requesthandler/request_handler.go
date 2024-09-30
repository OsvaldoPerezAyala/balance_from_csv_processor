package requesthandler

import (
	"net/http"
)

var healthCheckStatus = http.StatusOK
var healthCheckmessage = `{"alive": true}`

func EnterDownTime() {
	healthCheckStatus = http.StatusBadRequest
	healthCheckmessage = `{"alive": false}`
}

func ExitDownTime() {
	healthCheckStatus = http.StatusOK
	healthCheckmessage = `{"alive": true}`
}
