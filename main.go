package main

import (
	"PS_Risk_server/server"
	"fmt"
	"os"
)

func main() {
	serv, err := server.NuevoServidor(os.Args[1], os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6])
	if err != nil {
		fmt.Println(err)
		return
	}
	err = serv.Iniciar()
	if err != nil {
		fmt.Println(err)
	}
}
