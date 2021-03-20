package baseDatos

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"strings"
)

func execScript(db *sql.DB, script string) {
	file, err := ioutil.ReadFile(script)
	if err != nil {
		fmt.Println(err)
	}
	requests := strings.Split(string(file), ";")
	for _, request := range requests {
		_, err = db.Exec(request)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func NuevaBDConexionLocal(bbdd string, restablecer bool) *BD {
	//You should establish DB connection without SSL encryption, like that:
	db, err := sql.Open("postgres", bbdd)
	if err != nil {
		fmt.Println("No se ha podido abrir conexión con la base de "+
			"datos.\nError:", err)
	}
	b := &BD{
		bbdd: bbdd,
		bd:   db,
	}

	if restablecer {
		execScript(b.bd, "../scripts/destruirBBDD.sql")
		execScript(b.bd, "../scripts/crearBBDD.sql")
	}
	return b
}

func (b *BD) LeerMaxIdUsuario() int {
	var idLeido interface{}
	var id int
	err := b.bd.QueryRow("SELECT MAX(id_usuario) FROM usuario").Scan(&idLeido)
	if err != nil {
		if err == sql.ErrNoRows {
			id = 0
		} else {
			fmt.Println("Error leyendo el máximo id:", err)
		}
	} else {
		if idLeido == nil {
			id = 0
		} else {
			switch idLeidoTipo := idLeido.(type) {
			case int64: // No puede ser de otro tipo
				id = int(idLeidoTipo)
			}
		}
	}
	return id
}
