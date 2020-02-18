package dbase

import (
	"database/sql"
	"github.com/slevchyk/worker_srv/models"
)

func ScanCloudDBSettings(rows *sql.Rows, cs *models.CloudDBSettings) error {
	return rows.Scan(&cs.SrvIP, &cs.SrvUser, &cs.SrvPassword)
}
