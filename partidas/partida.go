package partidas

import (
	"PS_Risk_server/mensajes"
	"PS_Risk_server/mensajesInternos"
	"errors"
	"math/rand"
	"sync"

	"github.com/gorilla/websocket"
)

const MaxMensajes = 100

type Partida struct {
	partidaJson
	IdCreador      int
	Territorios    []Territorio
	Jugadores      []int
	JugadoresVivos map[int]bool
	Conexiones     sync.Map //map[int]*websocket.Conn
	Mensajes       chan mensajesInternos.MensajePartida
}

type partidaJson struct {
	IdPartida   int    `json:"idSala"`
	TiempoTurno int    `json:"tiempoTurno"`
	TurnoActual int    `json:"turnoActual,omitempty"`
	Fase        int    `json:"fase,omitempty"`
	Nombre      string `json:"nombreSala"`
	Empezada    bool   `json:"empezada"`
}

func NuevaPartida(idPartida, idCreador, tiempoTurno int, nombreSala string,
	wsCreador *websocket.Conn) *Partida {
	pJson := partidaJson{
		IdPartida:   idPartida,
		TiempoTurno: tiempoTurno,
		Nombre:      nombreSala,
		Empezada:    false,
	}
	p := &Partida{
		partidaJson: pJson,
		IdCreador:   idCreador,
		Jugadores:   []int{idCreador},
		Conexiones:  sync.Map{},
		Mensajes:    make(chan mensajesInternos.MensajePartida, MaxMensajes),
	}
	p.Conexiones.Store(idCreador, wsCreador)
	return p
}

func (p *Partida) IniciarPartida(idUsuario int) error {
	if p.Empezada {
		return errors.New("la partida ya está empezada")
	}
	if idUsuario != p.IdCreador {
		return errors.New("solo el creador de la partida puede comenzarla")
	}
	if len(p.Jugadores) < 3 {
		return errors.New("número de jugadores insuficiente")
	}
	// Decidir orden de jugadores
	orden := rand.Perm(len(p.Jugadores))
	aux := make([]int, len(p.Jugadores))
	copiados := copy(aux, p.Jugadores)
	if copiados != len(p.Jugadores) {
		return errors.New("error al copiar los jugadores a un vector auxiliar")
	}
	for i := 0; i < len(p.Jugadores); i++ {
		p.Jugadores[orden[i]] = aux[i]
	}
	// Marcar que están vivos
	for _, idJugador := range p.Jugadores {
		p.JugadoresVivos[idJugador] = true
	}

	// Decidir asignación de territorios
	p.Territorios = make([]Territorio, numTerritorios)
	for i := 0; i < numTerritorios; i++ {
		t := Territorio{
			IdJugador: 0,
			NumTropas: 0,
		}
		p.Territorios = append(p.Territorios, t)
	}
	asignados := make([]int, len(p.Jugadores))
	for i := 0; i < len(p.Jugadores); {
		idTerritorio := rand.Intn(numTerritorios)
		if !contenido(asignados, idTerritorio) {
			asignados[i] = idTerritorio
			p.Territorios[idTerritorio].IdJugador = p.Jugadores[i]
			i++
		}
	}

	p.Empezada = true
	return nil
}

func (p *Partida) UnirsePartida(idUsuario int, ws *websocket.Conn) error {
	if len(p.Jugadores) >= 6 {
		return errors.New("ya se ha alcanzado el número máximo de jugadores permitido")
	}
	if p.Empezada {
		return errors.New("no se puede unir a una partida que ya ha empezado")
	}
	p.Jugadores = append(p.Jugadores, idUsuario)
	p.Conexiones.Store(idUsuario, ws)
	return nil
}

func (p *Partida) QuitarJugadorPartida(idJugador int) error {
	if !p.EstaEnPartida(idJugador) {
		return errors.New("el jugador no está en la partida, no se puede retirar")
	}
	if p.Empezada {
		p.JugadoresVivos[idJugador] = false
		p.Conexiones.Delete(idJugador)
	} else if p.IdCreador == idJugador {
		return errors.New("no se puede eliminar de la partida al creador")
	} else {
		p.Jugadores = borrar(p.Jugadores, idJugador)
		p.Conexiones.Delete(idJugador)
	}
	return nil
}

func (p *Partida) EstaEnPartida(idUsuario int) bool {
	return contenido(p.Jugadores, idUsuario)
}

func PartidaDesdeJson(estado mensajes.JsonData, idCreador int) (*Partida, error) {
	if i, ok := estado["idSala"]; !ok || i == nil {
		return &Partida{}, errors.New("el json no contiene datos sobre una partida")
	}
	var turnoActual, fase int
	conexiones := sync.Map{}
	territorios := []Territorio{}
	jugadores := []int{}
	jugadoresVivos := make(map[int]bool)
	j := estado["jugadores"].([]interface{})
	for _, datosJugador := range j {
		datos := datosJugador.(map[string]interface{})
		id := int(datos["id"].(float64))
		jugadores = append(jugadores, id)
		if sigueVivo, ok := datos["sigueVivo"]; ok {
			jugadoresVivos[id] = sigueVivo.(bool)
		}
		conexiones.Store(id, nil)
	}
	listaTerritorios := estado["listaTerritorios"]
	if listaTerritorios != nil {
		t := listaTerritorios.([]interface{})
		for _, datosTerritorio := range t {
			datos := datosTerritorio.(map[string]interface{})
			territorios = append(territorios, Territorio{
				IdJugador: int(datos["numJugador"].(float64)),
				NumTropas: int(datos["tropas"].(float64)),
			})
		}
	}
	aux, ok := estado["turnoActual"]
	if ok {
		turnoActual = int(aux.(float64))
	} else {
		turnoActual = 0
	}
	aux, ok = estado["fase"]
	if ok {
		fase = int(aux.(float64))
	} else {
		fase = 0
	}
	pJson := partidaJson{
		IdPartida:   int(estado["idSala"].(float64)),
		Empezada:    estado["empezada"].(bool),
		TiempoTurno: int(estado["tiempoTurno"].(float64)),
		TurnoActual: turnoActual,
		Fase:        fase,
		Nombre:      estado["nombreSala"].(string),
	}
	p := &Partida{
		partidaJson:    pJson,
		IdCreador:      idCreador,
		Territorios:    territorios,
		Jugadores:      jugadores,
		JugadoresVivos: jugadoresVivos,
		Conexiones:     sync.Map{},
		Mensajes:       make(chan mensajesInternos.MensajePartida, MaxMensajes),
	}
	return p, nil
}

func contenido(lista []int, valor int) bool {
	i := indice(lista, valor)
	return i >= 0
}

func indice(lista []int, valor int) int {
	for i, v := range lista {
		if v == valor {
			return i
		}
	}
	return -1
}

func borrar(lista []int, valor int) []int {
	i := indice(lista, valor)
	if i < 0 {
		return lista
	} else {
		return append(lista[:i], lista[i+1:]...)
	}
}
