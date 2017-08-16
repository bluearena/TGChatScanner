package requestHandler

import (
	"github.com/zwirec/TGChatScanner/dbManager"
)

type RequestHandler struct {
	dbm *dbManager.DBManager
}

func NewRequestHandler(dbinfo map[string]string) (*RequestHandler, error) {
	dbm, err := dbManager.NewDBManager(dbinfo)
	if err != nil {
		return nil, err
	} else {
		return &RequestHandler{dbm:dbm}, nil
	}
}
