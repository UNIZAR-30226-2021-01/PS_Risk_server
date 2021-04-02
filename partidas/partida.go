package partidas

import (
	"PS_Risk_server/mensajesInternos"
	"errors"
	"math/rand"

	"github.com/gorilla/websocket"
)

const MaxMensajes = 100

type Partida struct {
	IdPartida, IdCreador           int
	TiempoTurno, TurnoActual, Fase int
	Nombre                         string
	Empezada                       bool
	Territorios                    []Territorio
	Jugadores                      []int
	JugadoresVivos                 map[int]bool
	Conexiones                     map[int]*websocket.Conn
	Mensajes                       chan mensajesInternos.MensajePartida
}

func NuevaPartida(idPartida, idCreador int, nombreSala string,
	wsCreador *websocket.Conn) Partida {
	conexiones := make(map[int]*websocket.Conn)
	conexiones[idCreador] = wsCreador
	return Partida{
		IdPartida:  idPartida,
		IdCreador:  idCreador,
		Nombre:     nombreSala,
		Empezada:   false,
		Jugadores:  []int{idCreador},
		Conexiones: conexiones,
		Mensajes:   make(chan mensajesInternos.MensajePartida, MaxMensajes),
	}
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
	p.Conexiones[idUsuario] = ws
	return nil
}

func (p *Partida) QuitarJugadorPartida(idJugador int) error {
	if !p.EstaEnPartida(idJugador) {
		return errors.New("el jugador no está en la partida, no se puede retirar")
	}
	if p.Empezada {
		p.JugadoresVivos[idJugador] = false
		delete(p.Conexiones, idJugador)
	} else if p.IdCreador == idJugador {
		return errors.New("no se puede eliminar de la partida al creador")
	} else {
		p.Jugadores = borrar(p.Jugadores, idJugador)
		delete(p.Conexiones, idJugador)
	}
	return nil
}

func (p *Partida) EstaEnPartida(idUsuario int) bool {
	return contenido(p.Jugadores, idUsuario)
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
