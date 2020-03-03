package dbase

import "database/sql"

func SelectCloudSettingsByPhonePin(db *sql.DB, phone string, pin int) (*sql.Rows, error) {

	return db.Query(`
		SELECT 
			s.srv_ip,
			s.srv_user,
			s.srv_password
		FROM
			cloud_users u
		LEFT JOIN 
			cloud_settings s
			ON u.id_settings = s.id
		WHERE
			u.phone=$1
			AND u.pin=$2`,
		phone, pin)
}

func SelectCloudAuthDBByUSerPassword(db *sql.DB, user, password string) (*sql.Rows, error) {

	return db.Query(`
		SELECT 
			a.id_cloud_db,
		    a.cloud_user,
		    a.cloud_password
		FROM
			cloud_db_auth a	
		WHERE
			a.cloud_user=$1
			AND a.cloud_password=$2`,
		user, password)
}

