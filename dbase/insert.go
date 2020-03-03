package dbase

import (
	"database/sql"
	"github.com/slevchyk/erp_mobile_main_srv/models"
)

func InsertCloudUser(db *sql.DB, cu models.CloudDBUsers) (sql.Result, error) {

	stmt, _ := db.Prepare(`
		INSERT INTO
			cloud_users (
				id_settings,
			    phone,
			    pin
				)
		VALUES ($1, $2, $3);`)

	return stmt.Exec(cu.IDSettings, cu.Phone, cu.Pin)
}
