package baseDatos

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"strings"

	_ "github.com/lib/pq"
)

const (
	violacionUnicidad      = "unique_violation"
	violacionClaveForanea  = "foreign_key_violation"
	violacionRestricciones = "check_violation"
)

/*
	CrearBD abre una conexión a la base de datos bbdd, borra el contenido actual
	y crea de nuevo las tablas.
*/
func CrearBD(bbdd string) (*sql.DB, error) {
	db, err := sql.Open("postgres", bbdd)
	if err != nil {
		return db, err
	}
	execScript(db, "scripts/destruirBBDD.sql")
	execScript(db, "scripts/crearBBDD.sql")
	execScript(db, "scripts/crearCuentasTest.sql")
	return db, err
}

/*
	CrearBD abre una conexión a la base de datos bbdd, borra el contenido actual
	y crea de nuevo las tablas, indicando la ruta a los scripts relativa a los
	ficheros de test.
*/
func CrearBDLocal(bbdd string) (*sql.DB, error) {
	db, err := sql.Open("postgres", bbdd)
	if err != nil {
		return db, err
	}
	execScript(db, "../scripts/destruirBBDD.sql")
	execScript(db, "../scripts/crearBBDD.sql")
	return db, err
}

// execScript ejecuta el script indicado en la base de datos db
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

func min(n1, n2 int) int {
	if n1 < n2 {
		return n1
	} else {
		return n2
	}
}

func max(n1, n2 int) int {
	if n1 < n2 {
		return n2
	} else {
		return n1
	}
}
