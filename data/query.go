package data

const (
	CreateUser    = `INSERT INTO users (email, password, created) VALUES (?, ?, ?);`
	ResetPassword = `UPDATE users SET password = ?, updated = ? WHERE email = ?;`
)
