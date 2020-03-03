package dbase

import (
	"database/sql"
	"github.com/slevchyk/erp_mobile_main_srv/models"
)

func UpdateCloudUser(db *sql.DB, cu models.CloudDBUsers) (sql.Result, error) {

	var err error

	stmt, err := db.Prepare(`
			UPDATE
				cloud_users
			SET
				id_settings = $1,
			    phone = $2,
			    pin = $3
			WHERE
				id=$4
			`)
	if err != nil {
		return nil, err
	}
	res, err := stmt.Exec(cu.IDSettings, cu.Phone, cu.Pin, cu.ID)

	return res, err
}

