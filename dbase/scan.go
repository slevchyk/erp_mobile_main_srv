package dbase

import (
	"database/sql"

	"github.com/slevchyk/erp_mobile_main_srv/models"
)

func ScanCloudDBSettings(rows *sql.Rows, cs *models.CloudDBSettings) error {
	return rows.Scan(&cs.SrvIP, &cs.SrvUser, &cs.SrvPassword)
}

func ScanCloudDBUser(rows *sql.Rows, cu *models.CloudDBUsers) error {
	return rows.Scan(&cu.ID, &cu.IDSettings, &cu.Phone, &cu.Pin)
}

func ScanCloudDBAuth(rows *sql.Rows, ca *models.CloudDBAuth) error {
	return rows.Scan(&ca.ID, &ca.IDCloudDB, &ca.CloudUser, &ca.CloudPassword)
}
