module get-address-api

go 1.13

replace get-address-api/step2 => ./step2

replace get-address-api/step3 => ./step3

require (
	get-address-api/step2 v0.0.0-00010101000000-000000000000
	get-address-api/step3 v0.0.0-00010101000000-000000000000
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
)
