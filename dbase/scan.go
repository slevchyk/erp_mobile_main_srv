package dbase

import (
	"database/sql"
	"github.com/slevchyk/erp_mobile_main_srv/models"
)

func ScanCloudDBSettings(rows *sql.Rows, cs *models.CloudDBSettings) error {
	return rows.Scan(&cs.SrvIP, &cs.SrvUser, &cs.SrvPassword)
}

func ScanCloudDBAuth(rows *sql.Rows, ca *models.CloudDBAuth) error {
	return rows.Scan(&ca.IDCloudDB, &ca.CloudUser, &ca.CloudPassword)
}