package dbase

import "database/sql"

func SelectCloudSettingsByPhonePin(db *sql.DB, phone string, pin int) (*sql.Rows, error) {

	return db.Query(`
		SELECT 
			s.srv_ip,
			s.srv_user,
			s.srv_password
		FROM
			worker.public.cloud_users u
		LEFT JOIN 
			worker.public.cloud_settings s
			ON u.id_settings = s.id
		WHERE
			u.phone=$1
			AND u.pin=$2`,
			phone, pin)
}