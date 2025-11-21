package data

const CreateUsersTable = `CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL,
	created INTEGER NOT NULL,
	updated INTEGER
) STRICT;`
