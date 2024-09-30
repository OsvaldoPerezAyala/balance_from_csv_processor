package refresh

import (
	"balance_from_csv_processor/repository"
	"balance_from_csv_processor/requesthandler"
)

func Task() {

	requesthandler.EnterDownTime()
	go repository.ReloadData()
	defer requesthandler.ExitDownTime()

}
