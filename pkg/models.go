package pkg

type ServiceConfig struct {
	SqlDriverName 	 	string 		`json:"sqlDriverName"`
	SqlDataSourceName 	string 		`json:"sqlDataSourceName"`
	LogLevel			string		`json:"logLevel"`
	ServerMode			string		`json:"serverMode"`
}
