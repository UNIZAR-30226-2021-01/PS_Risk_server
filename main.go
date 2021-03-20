package main

import (
	"PS_Risk_server/server"
	"fmt"
	"os"
)

func main() {
	serv, err := server.NuevoServidor(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Println(err)
		return
	}
	err = serv.Iniciar()
	if err != nil {
		fmt.Println(err)
	}
}
